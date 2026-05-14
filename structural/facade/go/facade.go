package main

import (
	"fmt"
	"strings"
)

// 场景：电商下单流程
//
// 痛点对比：
//   不用 Facade → 调用方必须了解 4 个子系统，自己按顺序编排：
//                 检查库存 → 扣款 → 锁库存 → 创建物流 → 发通知
//                 任何新入口（APP/小程序/后台补单）都要重写这段编排，一旦流程变更处处改
//
//   用 Facade  → 调用方只调 OrderFacade.PlaceOrder()
//                子系统的协调逻辑集中在 Facade 内，入口再多也只改一处

// ── 子系统 1：库存服务 ────────────────────────────────────────────────────────

type InventoryService struct {
	stock map[string]int
}

func NewInventoryService() *InventoryService {
	return &InventoryService{stock: map[string]int{
		"book":  50,
		"phone": 3,
		"pen":   0,
	}}
}

func (s *InventoryService) Check(item string, qty int) error {
	if s.stock[item] < qty {
		return fmt.Errorf("库存不足：%q 现有 %d，需要 %d", item, s.stock[item], qty)
	}
	return nil
}

func (s *InventoryService) Reserve(item string, qty int) {
	s.stock[item] -= qty
	fmt.Printf("  [库存] 锁定 %q × %d，剩余 %d\n", item, qty, s.stock[item])
}

// ── 子系统 2：支付服务 ────────────────────────────────────────────────────────

type PaymentService struct{}

func (s *PaymentService) Charge(account string, amount float64) error {
	if account == "" {
		return fmt.Errorf("支付账户不能为空")
	}
	fmt.Printf("  [支付] 账户 %q 扣款 ¥%.2f 成功\n", account, amount)
	return nil
}

// ── 子系统 3：物流服务 ────────────────────────────────────────────────────────

type ShippingService struct {
	seq int
}

func (s *ShippingService) CreateShipment(item string, qty int, addr string) string {
	s.seq++
	trackingNo := fmt.Sprintf("SF%06d", s.seq)
	fmt.Printf("  [物流] 创建运单 %s → %q × %d → %s\n", trackingNo, item, qty, addr)
	return trackingNo
}

// ── 子系统 4：通知服务 ────────────────────────────────────────────────────────

type NotificationService struct{}

func (s *NotificationService) Send(account, msg string) {
	fmt.Printf("  [通知] → %s：%s\n", account, msg)
}

// ── Facade：下单门面 ──────────────────────────────────────────────────────────
// 调用方只依赖这一个结构体，完全不感知子系统的存在

type OrderFacade struct {
	inventory    *InventoryService
	payment      *PaymentService
	shipping     *ShippingService
	notification *NotificationService
}

func NewOrderFacade() *OrderFacade {
	return &OrderFacade{
		inventory:    NewInventoryService(),
		payment:      &PaymentService{},
		shipping:     &ShippingService{},
		notification: &NotificationService{},
	}
}

type OrderRequest struct {
	Item    string
	Qty     int
	Price   float64
	Account string
	Address string
}

func (f *OrderFacade) PlaceOrder(req OrderRequest) error {
	fmt.Printf("  下单：%q × %d\n", req.Item, req.Qty)

	// 子系统编排逻辑集中在此，调用方对这些步骤一无所知
	if err := f.inventory.Check(req.Item, req.Qty); err != nil {
		return err
	}
	if err := f.payment.Charge(req.Account, req.Price*float64(req.Qty)); err != nil {
		return err
	}
	f.inventory.Reserve(req.Item, req.Qty)
	trackingNo := f.shipping.CreateShipment(req.Item, req.Qty, req.Address)
	f.notification.Send(req.Account, fmt.Sprintf("下单成功，运单号 %s", trackingNo))
	return nil
}

// ── main ─────────────────────────────────────────────────────────────────────

func main() {
	order := NewOrderFacade()

	cases := []struct {
		label string
		req   OrderRequest
	}{
		{
			label: "正常下单",
			req:   OrderRequest{Item: "book", Qty: 2, Price: 39.9, Account: "alice@example.com", Address: "北京市朝阳区"},
		},
		{
			label: "库存不足",
			req:   OrderRequest{Item: "pen", Qty: 1, Price: 5.0, Account: "bob@example.com", Address: "上海市浦东新区"},
		},
		{
			label: "账户为空",
			req:   OrderRequest{Item: "phone", Qty: 1, Price: 3999.0, Account: "", Address: "广州市天河区"},
		},
	}

	for _, c := range cases {
		fmt.Printf("━━ %s ━━\n", c.label)
		if err := order.PlaceOrder(c.req); err != nil {
			fmt.Printf("  下单失败：%v\n", err)
		}
		fmt.Println(strings.Repeat("─", 50))
	}
}
