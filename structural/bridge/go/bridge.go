package main

// MessageImplementor 实现层接口：定义底层发送能力
type MessageImplementor interface {
	Send(text, to string)
}

// --- Implementor 维度：发送渠道 ---
type MessageSMS struct{}

func (m *MessageSMS) Send(text, to string) {
	println("SMS → " + to + ": " + text)
}

type MessageEmail struct{}

func (m *MessageEmail) Send(text, to string) {
	println("Email → " + to + ": " + text)
}

// AbstractionMessage 抽象层接口：定义上层业务行为
type AbstractionMessage interface {
	SendMessage(text, to string)
}

// --- Abstraction 维度：消息类型 ---
type Abstraction struct {
	implementor MessageImplementor
}

func NewAbstraction(impl MessageImplementor) AbstractionMessage {
	return &Abstraction{implementor: impl}
}

func (a *Abstraction) SendMessage(text, to string) {
	a.implementor.Send(text, to)
}

// UrgentAbstraction 是 RefinedAbstraction，在抽象层扩展行为
type UrgentAbstraction struct {
	implementor MessageImplementor
}

func NewUrgentAbstraction(impl MessageImplementor) AbstractionMessage {
	return &UrgentAbstraction{implementor: impl}
}

func (u *UrgentAbstraction) SendMessage(text, to string) {
	u.implementor.Send("[URGENT] "+text, to)
}

func main() {
	sms := &MessageSMS{}
	email := &MessageEmail{}

	// 普通消息 × 两种渠道
	NewAbstraction(sms).SendMessage("Hello", "1234567890")
	NewAbstraction(email).SendMessage("Hello", "user@example.com")

	// 加急消息 × 两种渠道
	NewUrgentAbstraction(sms).SendMessage("Server is down", "1234567890")
	NewUrgentAbstraction(email).SendMessage("Server is down", "oncall@example.com")
}
