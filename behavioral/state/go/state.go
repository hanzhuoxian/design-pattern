package main

import "fmt"

// 场景：电商订单状态机
//
// 痛点对比：
//   不用状态模式 → Order 里每个方法都要 if/switch 判断 state 字段：
//     Pay()  里判断 "只有 pending 才能支付"
//     Ship() 里判断 "只有 paid 才能发货"
//     Cancel() 里分别处理 pending/paid/shipped 三种情况…
//     每新增一个状态（如"退款中"）要改遍所有方法，非法操作验证极易遗漏。
//
//   用状态模式  → 每个状态对象只负责自己允许的操作和转换目标。
//     Order 把请求委托给当前状态对象，自身代码永不改动。
//     新增"退款中"状态只需加一个 struct，已有状态零改动。

// ── 状态接口 ──────────────────────────────────────────────────────────────────

type OrderState interface {
	Pay(o *Order) error
	Ship(o *Order) error
	Deliver(o *Order) error
	Cancel(o *Order) error
	Refund(o *Order) error
	Name() string
}

// ── Order：只持有当前状态，所有方法委托给状态对象 ────────────────────────────

type Order struct {
	ID    string
	state OrderState
}

func NewOrder(id string) *Order {
	o := &Order{ID: id, state: &PendingState{}}
	fmt.Printf("[订单 %s] 创建  当前状态：%s\n", o.ID, o.state.Name())
	return o
}

func (o *Order) transition(next OrderState) {
	fmt.Printf("[订单 %s] %s → %s\n", o.ID, o.state.Name(), next.Name())
	o.state = next
}

func (o *Order) Pay() error     { return o.state.Pay(o) }
func (o *Order) Ship() error    { return o.state.Ship(o) }
func (o *Order) Deliver() error { return o.state.Deliver(o) }
func (o *Order) Cancel() error  { return o.state.Cancel(o) }
func (o *Order) Refund() error  { return o.state.Refund(o) }

// ── 辅助：统一拒绝非法操作 ───────────────────────────────────────────────────

func deny(op, state string) error {
	return fmt.Errorf("非法操作：[%s] 在 %q 状态下不可执行", op, state)
}

// ── 状态 1：待支付 ────────────────────────────────────────────────────────────

type PendingState struct{}

func (s *PendingState) Name() string { return "待支付" }

func (s *PendingState) Pay(o *Order) error {
	o.transition(&PaidState{})
	return nil
}

func (s *PendingState) Cancel(o *Order) error {
	o.transition(&CancelledState{})
	return nil
}

func (s *PendingState) Ship(o *Order) error    { return deny("Ship", s.Name()) }
func (s *PendingState) Deliver(o *Order) error { return deny("Deliver", s.Name()) }
func (s *PendingState) Refund(o *Order) error  { return deny("Refund", s.Name()) }

// ── 状态 2：已支付（待发货）──────────────────────────────────────────────────

type PaidState struct{}

func (s *PaidState) Name() string { return "已支付" }

func (s *PaidState) Ship(o *Order) error {
	o.transition(&ShippedState{})
	return nil
}

func (s *PaidState) Cancel(o *Order) error {
	// 已付款取消 → 自动发起退款，而非直接关单
	fmt.Printf("[订单 %s] 已付款取消，自动发起退款\n", o.ID)
	o.transition(&RefundingState{})
	return nil
}

func (s *PaidState) Refund(o *Order) error {
	o.transition(&RefundingState{})
	return nil
}

func (s *PaidState) Pay(o *Order) error     { return deny("Pay", s.Name()) }
func (s *PaidState) Deliver(o *Order) error { return deny("Deliver", s.Name()) }

// ── 状态 3：已发货 ────────────────────────────────────────────────────────────

type ShippedState struct{}

func (s *ShippedState) Name() string { return "已发货" }

func (s *ShippedState) Deliver(o *Order) error {
	o.transition(&DeliveredState{})
	return nil
}

func (s *ShippedState) Refund(o *Order) error {
	fmt.Printf("[订单 %s] 发货后退款，需先拦截物流\n", o.ID)
	o.transition(&RefundingState{})
	return nil
}

func (s *ShippedState) Pay(o *Order) error    { return deny("Pay", s.Name()) }
func (s *ShippedState) Ship(o *Order) error   { return deny("Ship", s.Name()) }
func (s *ShippedState) Cancel(o *Order) error { return deny("Cancel", s.Name()) }

// ── 状态 4：已收货 ────────────────────────────────────────────────────────────

type DeliveredState struct{}

func (s *DeliveredState) Name() string           { return "已收货" }
func (s *DeliveredState) Pay(o *Order) error     { return deny("Pay", s.Name()) }
func (s *DeliveredState) Ship(o *Order) error    { return deny("Ship", s.Name()) }
func (s *DeliveredState) Deliver(o *Order) error { return deny("Deliver", s.Name()) }
func (s *DeliveredState) Cancel(o *Order) error  { return deny("Cancel", s.Name()) }
func (s *DeliveredState) Refund(o *Order) error  { return deny("Refund", s.Name()) }

// ── 状态 5：退款中 ────────────────────────────────────────────────────────────

type RefundingState struct{}

func (s *RefundingState) Name() string { return "退款中" }

// 退款到账后由平台调用 Cancel 关单
func (s *RefundingState) Cancel(o *Order) error {
	o.transition(&CancelledState{})
	return nil
}

func (s *RefundingState) Pay(o *Order) error     { return deny("Pay", s.Name()) }
func (s *RefundingState) Ship(o *Order) error    { return deny("Ship", s.Name()) }
func (s *RefundingState) Deliver(o *Order) error { return deny("Deliver", s.Name()) }
func (s *RefundingState) Refund(o *Order) error  { return deny("Refund", s.Name()) }

// ── 状态 6：已取消（终态）────────────────────────────────────────────────────

type CancelledState struct{}

func (s *CancelledState) Name() string           { return "已取消" }
func (s *CancelledState) Pay(o *Order) error     { return deny("Pay", s.Name()) }
func (s *CancelledState) Ship(o *Order) error    { return deny("Ship", s.Name()) }
func (s *CancelledState) Deliver(o *Order) error { return deny("Deliver", s.Name()) }
func (s *CancelledState) Cancel(o *Order) error  { return deny("Cancel", s.Name()) }
func (s *CancelledState) Refund(o *Order) error  { return deny("Refund", s.Name()) }

// ── 辅助：打印操作结果 ────────────────────────────────────────────────────────

func try(label string, err error) {
	if err != nil {
		fmt.Printf("  [ERR] %s：%v\n", label, err)
	} else {
		fmt.Printf("  [OK]  %s\n", label)
	}
}

// ── main ─────────────────────────────────────────────────────────────────────

func main() {
	fmt.Println("━━ 正常流程：下单 → 支付 → 发货 → 确认收货 ━━")
	o1 := NewOrder("ORD-001")
	try("支付", o1.Pay())
	try("发货", o1.Ship())
	try("确认收货", o1.Deliver())
	try("收货后再支付（非法）", o1.Pay())

	fmt.Println("\n━━ 未付款直接取消 ━━")
	o2 := NewOrder("ORD-002")
	try("取消", o2.Cancel())
	try("已取消再取消（非法）", o2.Cancel())

	fmt.Println("\n━━ 已付款后取消（自动退款 → 关单）━━")
	o3 := NewOrder("ORD-003")
	try("支付", o3.Pay())
	try("取消（自动退款）", o3.Cancel())
	try("退款到账，平台关单", o3.Cancel()) // RefundingState.Cancel → CancelledState

	fmt.Println("\n━━ 发货后申请退款 ━━")
	o4 := NewOrder("ORD-004")
	try("支付", o4.Pay())
	try("发货", o4.Ship())
	try("申请退款（拦截物流）", o4.Refund())
	try("退款中再次发货（非法）", o4.Ship())

	fmt.Println("\n━━ 非法操作：跳过支付直接发货 ━━")
	o5 := NewOrder("ORD-005")
	try("直接发货（非法）", o5.Ship())
	try("支付", o5.Pay())
	try("发货", o5.Ship())
}
