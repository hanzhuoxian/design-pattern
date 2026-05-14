package main

import "fmt"

// 场景：电商订单状态变更通知
//
// 痛点对比：
//   不用观察者 → Order.ChangeStatus() 直接调用库存扣减、发邮件、通知物流、触发风控
//     新增短信通知或积分发放，必须修改 Order 代码
//     某些订单不需要物流，只能在 Order 里加 if 分支
//     单元测试 Order 时必须 mock 所有下游模块
//
//   用观察者  → Order 只广播"状态已变更"，谁关心、怎么处理由观察者自己决定
//     新增/移除观察者：Order 代码零改动
//     不同订单类型挂不同观察者集合，自由组合互不干扰
//     观察者可在运行时动态挂载/摘除（如物流服务临时维护）

// ── 核心类型 ──────────────────────────────────────────────────────────────────

type OrderStatus string

const (
	StatusPending   OrderStatus = "待付款"
	StatusPaid      OrderStatus = "已付款"
	StatusShipped   OrderStatus = "已发货"
	StatusDelivered OrderStatus = "已签收"
	StatusCancelled OrderStatus = "已取消"
)

type OrderEvent struct {
	OrderID   string
	OldStatus OrderStatus
	NewStatus OrderStatus
	Amount    float64
}

// Observer 是所有订阅者的接口
type Observer interface {
	OnOrderChanged(event OrderEvent)
}

// ── 被观察者：订单 ────────────────────────────────────────────────────────────

type Order struct {
	ID        string
	Status    OrderStatus
	Amount    float64
	observers []Observer
}

func NewOrder(id string, amount float64) *Order {
	return &Order{ID: id, Amount: amount, Status: StatusPending}
}

func (o *Order) Subscribe(ob Observer) {
	o.observers = append(o.observers, ob)
}

func (o *Order) Unsubscribe(ob Observer) {
	for i, existing := range o.observers {
		if existing == ob {
			o.observers = append(o.observers[:i], o.observers[i+1:]...)
			return
		}
	}
}

func (o *Order) ChangeStatus(newStatus OrderStatus) {
	event := OrderEvent{
		OrderID:   o.ID,
		OldStatus: o.Status,
		NewStatus: newStatus,
		Amount:    o.Amount,
	}
	o.Status = newStatus
	for _, ob := range o.observers {
		ob.OnOrderChanged(event)
	}
}

// ── 观察者：库存服务 ──────────────────────────────────────────────────────────

type InventoryService struct{}

func (s *InventoryService) OnOrderChanged(e OrderEvent) {
	switch e.NewStatus {
	case StatusPaid:
		fmt.Printf("[库存]    订单 %s 已付款 → 锁定库存\n", e.OrderID)
	case StatusDelivered:
		fmt.Printf("[库存]    订单 %s 已签收 → 扣减库存\n", e.OrderID)
	case StatusCancelled:
		fmt.Printf("[库存]    订单 %s 已取消 → 释放库存\n", e.OrderID)
	}
}

// ── 观察者：邮件通知 ──────────────────────────────────────────────────────────

type EmailService struct{}

func (s *EmailService) OnOrderChanged(e OrderEvent) {
	messages := map[OrderStatus]string{
		StatusPaid:      "您的订单已付款，正在备货",
		StatusShipped:   "您的包裹已发出，请注意查收",
		StatusDelivered: "订单已签收，感谢您的购买",
		StatusCancelled: "订单已取消，退款将在 3 个工作日内到账",
	}
	if msg, ok := messages[e.NewStatus]; ok {
		fmt.Printf("[邮件]    订单 %s → %s\n", e.OrderID, msg)
	}
}

// ── 观察者：物流系统 ──────────────────────────────────────────────────────────

type LogisticsService struct{}

func (s *LogisticsService) OnOrderChanged(e OrderEvent) {
	switch e.NewStatus {
	case StatusPaid:
		fmt.Printf("[物流]    订单 %s 已付款 → 创建物流单\n", e.OrderID)
	case StatusShipped:
		fmt.Printf("[物流]    订单 %s 已发货 → 更新运单状态\n", e.OrderID)
	case StatusCancelled:
		fmt.Printf("[物流]    订单 %s 已取消 → 撤销物流单\n", e.OrderID)
	}
}

// ── 观察者：风控系统（只关注大额订单）────────────────────────────────────────

type RiskControlService struct {
	threshold float64
}

func (s *RiskControlService) OnOrderChanged(e OrderEvent) {
	if e.Amount < s.threshold {
		return
	}
	switch e.NewStatus {
	case StatusPaid:
		fmt.Printf("[风控]    订单 %s 金额 %.0f 超阈值 → 触发人工复核\n", e.OrderID, e.Amount)
	case StatusCancelled:
		fmt.Printf("[风控]    大额订单 %s 取消 → 记录异常日志\n", e.OrderID)
	}
}

// ── 观察者：积分服务（仅签收后发放）─────────────────────────────────────────

type PointsService struct{}

func (s *PointsService) OnOrderChanged(e OrderEvent) {
	if e.NewStatus != StatusDelivered {
		return
	}
	points := int(e.Amount / 10)
	fmt.Printf("[积分]    订单 %s 已签收 → 赠送 %d 积分\n", e.OrderID, points)
}

// ── main ──────────────────────────────────────────────────────────────────────

func main() {
	inventory := &InventoryService{}
	email := &EmailService{}
	logistics := &LogisticsService{}
	risk := &RiskControlService{threshold: 1000}
	points := &PointsService{}

	// 普通订单：库存 + 邮件 + 物流 + 积分，不挂风控
	fmt.Println("━━ 普通订单完整流转（100 元）━━")
	order1 := NewOrder("ORD-001", 100)
	order1.Subscribe(inventory)
	order1.Subscribe(email)
	order1.Subscribe(logistics)
	order1.Subscribe(points)

	order1.ChangeStatus(StatusPaid)
	fmt.Println()
	order1.ChangeStatus(StatusShipped)
	fmt.Println()
	order1.ChangeStatus(StatusDelivered)

	// 大额订单：在普通观察者基础上额外挂风控
	// 体现优势：Order 代码不变，只是订阅列表不同
	fmt.Println("\n━━ 大额订单（2000 元，额外挂风控）━━")
	order2 := NewOrder("ORD-002", 2000)
	order2.Subscribe(inventory)
	order2.Subscribe(email)
	order2.Subscribe(logistics)
	order2.Subscribe(risk) // 仅大额订单挂载
	order2.Subscribe(points)

	order2.ChangeStatus(StatusPaid)

	// 订单中途取消：积分服务只关注 Delivered，自动忽略 Cancelled，无需 if 分支
	fmt.Println("\n━━ 订单取消（各观察者自行决定是否响应）━━")
	order3 := NewOrder("ORD-003", 500)
	order3.Subscribe(inventory)
	order3.Subscribe(email)
	order3.Subscribe(logistics)
	order3.Subscribe(points)

	order3.ChangeStatus(StatusPaid)
	fmt.Println()
	order3.ChangeStatus(StatusCancelled)

	// 动态摘除观察者：物流服务临时维护，运行时 Unsubscribe
	// 体现优势：无需修改 Order，也不影响其他观察者
	fmt.Println("\n━━ 动态取消订阅（物流维护，运行时摘除）━━")
	order4 := NewOrder("ORD-004", 300)
	order4.Subscribe(inventory)
	order4.Subscribe(email)
	order4.Subscribe(logistics)

	order4.ChangeStatus(StatusPaid)
	fmt.Println("  [系统]  物流服务维护中，临时摘除物流观察者...")
	order4.Unsubscribe(logistics)
	fmt.Println()
	order4.ChangeStatus(StatusShipped) // 物流不再收到通知，其余不受影响
}
