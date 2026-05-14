package main

import (
	"fmt"
	"strings"
)

// 场景：电商购物车 —— 对多种商品类型执行多种操作（计税、开票、折扣、导出）
//
// 痛点对比：
//   不用访问者 → 每新增一种操作（计税、开票、折扣、CSV 导出…），
//     就要给 Item 接口加方法，Product / Service / Subscription 全部跟着改
//     N 种操作 × M 种类型 = N×M 次散乱修改，类型定义越来越臃肿
//
//   用访问者  → Item 接口只有一个 Accept(Visitor) 方法，类型定义永久稳定
//     新增操作只需实现一个新的 Visitor，完全不碰任何 Item 代码
//     M 种类型固定，N 种操作线性扩展：只需新增 N 个 Visitor

// ── 元素接口：稳定不变 ────────────────────────────────────────────────────────

type Item interface {
	Accept(v Visitor)
	Name() string
}

// ── 访问者接口：每种操作实现一次 ──────────────────────────────────────────────

type Visitor interface {
	VisitProduct(p *Product)
	VisitService(s *Service)
	VisitSubscription(sub *Subscription)
}

// ── 三种商品类型（Accept 之后类型定义不再修改）────────────────────────────────

type Product struct {
	name  string
	price float64
	qty   int
}

func (p *Product) Name() string     { return p.name }
func (p *Product) Accept(v Visitor) { v.VisitProduct(p) }

type Service struct {
	name      string
	unitPrice float64
	hours     float64
}

func (s *Service) Name() string     { return s.name }
func (s *Service) Accept(v Visitor) { v.VisitService(s) }

type Subscription struct {
	name    string
	monthly float64
	months  int
}

func (sub *Subscription) Name() string     { return sub.name }
func (sub *Subscription) Accept(v Visitor) { v.VisitSubscription(sub) }

// ── 访问者 1：计税（商品13%，服务6%，订阅免税）────────────────────────────────

type TaxCalculator struct {
	total float64
}

func (t *TaxCalculator) VisitProduct(p *Product) {
	sub := p.price * float64(p.qty)
	tax := sub * 0.13
	fmt.Printf("[计税] %-20s 小计 %7.2f  税(13%%) %.2f\n", p.name, sub, tax)
	t.total += sub + tax
}

func (t *TaxCalculator) VisitService(s *Service) {
	sub := s.unitPrice * s.hours
	tax := sub * 0.06
	fmt.Printf("[计税] %-20s 小计 %7.2f  税(6%%)  %.2f\n", s.name, sub, tax)
	t.total += sub + tax
}

func (t *TaxCalculator) VisitSubscription(sub *Subscription) {
	amount := sub.monthly * float64(sub.months)
	fmt.Printf("[计税] %-20s 小计 %7.2f  免税\n", sub.name, amount)
	t.total += amount
}

// ── 访问者 2：生成发票明细 ────────────────────────────────────────────────────

type InvoiceRenderer struct {
	lines []string
}

func (ir *InvoiceRenderer) VisitProduct(p *Product) {
	ir.lines = append(ir.lines, fmt.Sprintf(
		"  商品  %-18s x%d   单价 %.2f", p.name, p.qty, p.price))
}

func (ir *InvoiceRenderer) VisitService(s *Service) {
	ir.lines = append(ir.lines, fmt.Sprintf(
		"  服务  %-18s %.1f小时 单价 %.2f/h", s.name, s.hours, s.unitPrice))
}

func (ir *InvoiceRenderer) VisitSubscription(sub *Subscription) {
	ir.lines = append(ir.lines, fmt.Sprintf(
		"  会员  %-18s %d个月  月费 %.2f", sub.name, sub.months, sub.monthly))
}

func (ir *InvoiceRenderer) Print() {
	fmt.Println("┌──────────────── 电子发票 ────────────────┐")
	for _, l := range ir.lines {
		fmt.Println(l)
	}
	fmt.Println("└──────────────────────────────────────────┘")
}

// ── 访问者 3：VIP 折扣（商品9折，服务95折，订阅不打折）────────────────────────

type DiscountApplier struct {
	savings float64
}

func (d *DiscountApplier) VisitProduct(p *Product) {
	saved := p.price * float64(p.qty) * 0.10
	fmt.Printf("[折扣] %-20s 节省 %.2f（9折）\n", p.name, saved)
	d.savings += saved
}

func (d *DiscountApplier) VisitService(s *Service) {
	saved := s.unitPrice * s.hours * 0.05
	fmt.Printf("[折扣] %-20s 节省 %.2f（95折）\n", s.name, saved)
	d.savings += saved
}

func (d *DiscountApplier) VisitSubscription(sub *Subscription) {
	fmt.Printf("[折扣] %-20s 无折扣\n", sub.name)
}

// ── 访问者 4：导出 CSV（演示扩展性，Item 代码一行未改）────────────────────────

type CSVExporter struct {
	rows []string
}

func (c *CSVExporter) VisitProduct(p *Product) {
	c.rows = append(c.rows, fmt.Sprintf("商品,%s,%d,%.2f", p.name, p.qty, p.price))
}

func (c *CSVExporter) VisitService(s *Service) {
	c.rows = append(c.rows, fmt.Sprintf("服务,%s,%.1fh,%.2f", s.name, s.hours, s.unitPrice))
}

func (c *CSVExporter) VisitSubscription(sub *Subscription) {
	c.rows = append(c.rows, fmt.Sprintf("会员,%s,%d月,%.2f", sub.name, sub.months, sub.monthly))
}

func (c *CSVExporter) Result() string {
	return "类型,名称,数量,单价\n" + strings.Join(c.rows, "\n")
}

// ── 购物车：元素集合，负责驱动遍历 ───────────────────────────────────────────

type Cart struct {
	items []Item
}

func (c *Cart) Add(items ...Item) {
	c.items = append(c.items, items...)
}

// Accept 实现双分派：Cart 遍历元素，每个元素再回调 Visitor 对应的方法
// 运行时同时由 item 的具体类型 + visitor 的具体类型共同决定执行哪段逻辑
func (c *Cart) Accept(v Visitor) {
	for _, item := range c.items {
		item.Accept(v)
	}
}

// ── main ─────────────────────────────────────────────────────────────────────

func main() {
	cart := &Cart{}
	cart.Add(
		&Product{name: "机械键盘", price: 599.00, qty: 2},
		&Product{name: "显示器", price: 1899.00, qty: 1},
		&Service{name: "系统安装服务", unitPrice: 200.00, hours: 3},
		&Subscription{name: "云存储会员", monthly: 15.00, months: 12},
	)

	fmt.Println("━━ 含税金额汇总 ━━")
	tax := &TaxCalculator{}
	cart.Accept(tax)
	fmt.Printf("    含税总计：%.2f 元\n", tax.total)

	fmt.Println("\n━━ 发票明细 ━━")
	invoice := &InvoiceRenderer{}
	cart.Accept(invoice)
	invoice.Print()

	fmt.Println("\n━━ VIP 折扣预览 ━━")
	discount := &DiscountApplier{}
	cart.Accept(discount)
	fmt.Printf("    合计可省：%.2f 元\n", discount.savings)

	// 体现访问者优势：新增"导出 CSV"只需加一个 Visitor，Item 类型代码零改动
	fmt.Println("\n━━ 新操作：导出 CSV（新增 Visitor，不改任何 Item）━━")
	csv := &CSVExporter{}
	cart.Accept(csv)
	fmt.Println(csv.Result())
}
