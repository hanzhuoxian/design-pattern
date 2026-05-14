package main

import "fmt"

// 场景：电商结账集成多家第三方支付 SDK
//
// 痛点对比：
//   不用适配器 → 业务代码直接调用各 SDK 原生方法（AliPay / WeChatPay），接口各不相同
//                 切换支付渠道要改 OrderService；单位换算（元→分）等细节散落在业务层
//
//   用适配器  → 每个 SDK 包一层 Adapter，统一实现 PaymentProcessor 接口
//                 OrderService 只依赖接口，切换渠道只换注入对象，业务代码零改动

// ========== 第三方 SDK：支付宝（不可修改）==========

type AliPaySDK struct{}

func (a *AliPaySDK) AliPay(orderID string, amount float64) {
	fmt.Printf("[支付宝] 订单 %s 支付 %.2f 元\n", orderID, amount)
}

// ========== 第三方 SDK：微信支付（不可修改）==========

type WeChatPaySDK struct{}

func (w *WeChatPaySDK) WeChatPay(transactionID string, amountCents int) {
	fmt.Printf("[微信支付] 交易 %s 支付 %d 分\n", transactionID, amountCents)
}

// ========== 新系统期望的统一支付接口 ==========

type PaymentProcessor interface {
	Pay(orderID string, amountYuan float64)
}

// ========== 适配器：支付宝 → 统一接口 ==========

type AliPayAdapter struct {
	sdk *AliPaySDK
}

func (a *AliPayAdapter) Pay(orderID string, amountYuan float64) {
	a.sdk.AliPay(orderID, amountYuan)
}

func NewAliPayAdapter() PaymentProcessor {
	return &AliPayAdapter{sdk: &AliPaySDK{}}
}

// ========== 适配器：微信支付 → 统一接口（含单位转换）==========

type WeChatPayAdapter struct {
	sdk *WeChatPaySDK
}

func (w *WeChatPayAdapter) Pay(orderID string, amountYuan float64) {
	w.sdk.WeChatPay(orderID, int(amountYuan*100)) // 元 → 分
}

func NewWeChatPayAdapter() PaymentProcessor {
	return &WeChatPayAdapter{sdk: &WeChatPaySDK{}}
}

// ========== 客户端：只依赖统一接口，无感知底层差异 ==========

type OrderService struct {
	payment PaymentProcessor
}

func (o *OrderService) Checkout(orderID string, amount float64) {
	fmt.Printf("开始结账，订单：%s\n", orderID)
	o.payment.Pay(orderID, amount)
	fmt.Println("结账完成")
}

func main() {
	fmt.Println("=== 使用支付宝 ===")
	svc := &OrderService{payment: NewAliPayAdapter()}
	svc.Checkout("ORDER-001", 99.50)

	fmt.Println()

	fmt.Println("=== 切换微信支付（OrderService 代码零改动）===")
	svc = &OrderService{payment: NewWeChatPayAdapter()}
	svc.Checkout("ORDER-002", 128.00)
}
