package main

import "fmt"

// 场景：电商订单的促销价格计算
//
// 痛点对比：
//   不用策略模式 → Order.Checkout() 内部写一个巨大的 if-else/switch
//     促销类型新增/修改要改 Order 内部代码，违反开闭原则
//     不同促销逻辑混在一起，单测必须构造特定的 Order 状态才能覆盖各分支
//     活动切换（如双十一结束）要重启服务或改代码
//
//   用策略模式 → Order 只依赖 PricingStrategy 接口
//     新增促销只需新写一个 struct，Order 代码零改动
//     每种策略可独立单测，运行时可热替换当前促销活动

// ── 策略接口 ──────────────────────────────────────────────────────────────────

type PricingStrategy interface {
	Name() string
	CalcFinal(original float64) float64
}

// ── 具体策略：原价 ────────────────────────────────────────────────────────────

type NoDiscount struct{}

func (NoDiscount) Name() string                      { return "原价" }
func (NoDiscount) CalcFinal(original float64) float64 { return original }

// ── 具体策略：折扣（如 8 折）─────────────────────────────────────────────────

type PercentOff struct {
	Percent float64 // 0.8 = 8折
}

func (p PercentOff) Name() string { return fmt.Sprintf("%.0f折优惠", p.Percent*10) }
func (p PercentOff) CalcFinal(original float64) float64 {
	return original * p.Percent
}

// ── 具体策略：满减（满 X 减 Y）───────────────────────────────────────────────

type FullReduction struct {
	Threshold float64
	Reduction float64
}

func (f FullReduction) Name() string {
	return fmt.Sprintf("满%.0f减%.0f", f.Threshold, f.Reduction)
}
func (f FullReduction) CalcFinal(original float64) float64 {
	if original >= f.Threshold {
		return original - f.Reduction
	}
	return original
}

// ── 具体策略：阶梯满减（每满 X 减 Y，可叠加）────────────────────────────────

type TieredReduction struct {
	Threshold float64
	Reduction float64
}

func (t TieredReduction) Name() string {
	return fmt.Sprintf("每满%.0f减%.0f(叠加)", t.Threshold, t.Reduction)
}
func (t TieredReduction) CalcFinal(original float64) float64 {
	times := float64(int(original / t.Threshold))
	return original - times*t.Reduction
}

// ── 具体策略：限时秒杀（直接覆盖为秒杀价）──────────────────────────────────

type FlashSale struct {
	SalePrice float64
}

func (f FlashSale) Name() string { return "限时秒杀" }
func (f FlashSale) CalcFinal(_ float64) float64 {
	return f.SalePrice
}

// ── Order：只依赖接口，感知不到任何具体促销逻辑 ──────────────────────────────

type Order struct {
	Product  string
	Original float64
	strategy PricingStrategy
}

func NewOrder(product string, price float64, s PricingStrategy) *Order {
	return &Order{Product: product, Original: price, strategy: s}
}

// SetStrategy 运行时热替换促销策略（双十一开始/结束无需改 Order 代码）
func (o *Order) SetStrategy(s PricingStrategy) {
	o.strategy = s
}

func (o *Order) Checkout() {
	final := o.strategy.CalcFinal(o.Original)
	saved := o.Original - final
	if saved > 0 {
		fmt.Printf("  商品：%-14s  原价：%7.2f  促销：%-16s  实付：%7.2f  省：%.2f\n",
			o.Product, o.Original, o.strategy.Name(), final, saved)
	} else {
		fmt.Printf("  商品：%-14s  原价：%7.2f  促销：%-16s  实付：%7.2f\n",
			o.Product, o.Original, o.strategy.Name(), final)
	}
}

// ── main ─────────────────────────────────────────────────────────────────────

func main() {
	fmt.Println("━━ 平时：各商品使用不同定价策略 ━━")
	orders := []*Order{
		NewOrder("普通键盘", 299.00, NoDiscount{}),
		NewOrder("机械键盘", 599.00, PercentOff{Percent: 0.8}),
		NewOrder("显示器", 1299.00, FullReduction{Threshold: 1000, Reduction: 100}),
		NewOrder("桌椅套装", 2680.00, TieredReduction{Threshold: 500, Reduction: 50}),
	}
	for _, o := range orders {
		o.Checkout()
	}

	fmt.Println("\n━━ 双十一活动开始：运行时统一切换为秒杀策略，Order 代码零改动 ━━")
	flashPrices := map[string]float64{
		"普通键盘": 199.00,
		"机械键盘": 399.00,
		"显示器":  888.00,
		"桌椅套装": 1888.00,
	}
	for _, o := range orders {
		o.SetStrategy(FlashSale{SalePrice: flashPrices[o.Product]})
		o.Checkout()
	}

	fmt.Println("\n━━ 活动结束：切回满减，只改策略对象，不碰 Order ━━")
	for _, o := range orders {
		o.SetStrategy(FullReduction{Threshold: 500, Reduction: 60})
		o.Checkout()
	}

	fmt.Println("\n━━ 痛点演示：阶梯满减——若无策略模式，此逻辑只能硬塞进 Order ━━")
	big := NewOrder("服务器套件", 3200.00, TieredReduction{Threshold: 500, Reduction: 50})
	big.Checkout() // 每满500减50，3200→3200-6*50=2900
}
