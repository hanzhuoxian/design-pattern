package main

import (
	"fmt"
	"log"
)

// 场景：订单系统向用户发送通知，支持邮件、短信、Slack 三种渠道。
//
// 没有工厂方法时，OrderService 里会出现：
//
//	switch channel {
//	case "email": notifier = &EmailNotifier{...}
//	case "sms":   notifier = &SMSNotifier{...}
//	}
//
// 每次新增渠道都要改业务代码，违反开闭原则。
//
// 工厂方法的解法：把"如何创建通知器"委托给具体工厂，
// OrderService 只依赖 NotifierFactory 接口，新增渠道只需加新工厂。

// Notifier 产品接口
type Notifier interface {
	Send(to, message string) error
}

// NotifierFactory 工厂接口，子类决定创建哪种 Notifier
type NotifierFactory interface {
	Create(config map[string]string) (Notifier, error)
}

// --- Email ---

type EmailNotifier struct {
	smtpHost string
	smtpPort string
}

func (n *EmailNotifier) Send(to, message string) error {
	fmt.Printf("[Email] to=%s smtp=%s:%s msg=%q\n", to, n.smtpHost, n.smtpPort, message)
	return nil
}

type EmailFactory struct{}

func (EmailFactory) Create(config map[string]string) (Notifier, error) {
	host, ok := config["smtp_host"]
	if !ok {
		return nil, fmt.Errorf("email factory: missing smtp_host")
	}
	return &EmailNotifier{smtpHost: host, smtpPort: config["smtp_port"]}, nil
}

// --- SMS ---

type SMSNotifier struct {
	apiKey string
}

func (n *SMSNotifier) Send(to, message string) error {
	fmt.Printf("[SMS]   to=%s key=%s msg=%q\n", to, n.apiKey, message)
	return nil
}

type SMSFactory struct{}

func (SMSFactory) Create(config map[string]string) (Notifier, error) {
	key, ok := config["api_key"]
	if !ok {
		return nil, fmt.Errorf("sms factory: missing api_key")
	}
	return &SMSNotifier{apiKey: key}, nil
}

// --- Slack ---

type SlackNotifier struct {
	webhookURL string
}

func (n *SlackNotifier) Send(to, message string) error {
	fmt.Printf("[Slack] to=%s webhook=%s msg=%q\n", to, n.webhookURL, message)
	return nil
}

type SlackFactory struct{}

func (SlackFactory) Create(config map[string]string) (Notifier, error) {
	url, ok := config["webhook_url"]
	if !ok {
		return nil, fmt.Errorf("slack factory: missing webhook_url")
	}
	return &SlackNotifier{webhookURL: url}, nil
}

// OrderService 是业务代码，注入工厂而非具体类型。
// 新增通知渠道时，这里不需要任何改动。
type OrderService struct {
	factory NotifierFactory
	config  map[string]string
}

func (s *OrderService) PlaceOrder(orderID, contact string) {
	notifier, err := s.factory.Create(s.config)
	if err != nil {
		log.Printf("create notifier: %v", err)
		return
	}
	msg := fmt.Sprintf("order %s placed successfully", orderID)
	if err := notifier.Send(contact, msg); err != nil {
		log.Printf("send notification: %v", err)
	}
}

func main() {
	cases := []struct {
		label   string
		factory NotifierFactory
		config  map[string]string
		contact string
	}{
		{
			label:   "email channel",
			factory: EmailFactory{},
			config:  map[string]string{"smtp_host": "smtp.example.com", "smtp_port": "465"},
			contact: "user@example.com",
		},
		{
			label:   "sms channel",
			factory: SMSFactory{},
			config:  map[string]string{"api_key": "sk-abc123"},
			contact: "+8613800138000",
		},
		{
			label:   "slack channel",
			factory: SlackFactory{},
			config:  map[string]string{"webhook_url": "https://hooks.slack.com/xxx"},
			contact: "#orders",
		},
	}

	for _, c := range cases {
		fmt.Printf("\n=== %s ===\n", c.label)
		svc := &OrderService{factory: c.factory, config: c.config}
		svc.PlaceOrder("ORD-2026-001", c.contact)
	}
}
