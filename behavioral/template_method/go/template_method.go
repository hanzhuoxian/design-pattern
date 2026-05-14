package main

import (
	"fmt"
	"strings"
)

// 场景：多格式报表导出（CSV / JSON / Markdown）
//
// 痛点对比：
//   不用模板方法 → 三种导出各自实现完整流程：连接数据源、遍历记录、写文件头/尾
//     新增 Markdown 格式需要复制粘贴整个导出骨架，只改几行格式化代码
//     修改通用步骤（进度打印、分页大小、错误重试）必须改每一个实现类
//
//   用模板方法 → 基类固定骨架：取数 → 写表头 → 逐行写入 → 写表尾 → 汇报结果
//     子类只实现"写表头/尾"和"格式化一行"，骨架代码零重复
//     新增格式只需一个新子类，ReportExporter.Export() 一行不动

// ── 数据层（模拟数据库订单表）───────────────────────────────────────────────────

type Order struct {
	ID       int
	Customer string
	Product  string
	Amount   float64
	Status   string
}

var orders = []Order{
	{1, "Alice", "MacBook Pro", 12999.00, "已完成"},
	{2, "Bob", "iPhone 15", 5999.00, "已发货"},
	{3, "Carol", "AirPods Pro", 1799.00, "待支付"},
	{4, "Dave", "iPad Air", 4799.00, "已完成"},
	{5, "Eve", "Apple Watch", 3299.00, "已取消"},
}

// ── Formatter：子类必须实现的三个可变步骤 ────────────────────────────────────────

type Formatter interface {
	Name() string
	WriteHeader(columns []string) string
	WriteRow(row []string) string
	WriteFooter(rowCount int) string
}

// ── 模板方法：Export 是骨架，固定步骤顺序，不可被子类覆盖 ───────────────────────

type ReportExporter struct {
	formatter Formatter
}

// Export 是模板方法——骨架永远是：进度提示 → 表头 → 逐行 → 表尾 → 汇报
func (e *ReportExporter) Export(orders []Order) string {
	columns := []string{"ID", "客户", "商品", "金额(元)", "状态"}
	var buf strings.Builder

	fmt.Printf("[%s] 开始导出，共 %d 条记录\n", e.formatter.Name(), len(orders))

	buf.WriteString(e.formatter.WriteHeader(columns))

	for _, o := range orders {
		row := []string{
			fmt.Sprintf("%d", o.ID),
			o.Customer,
			o.Product,
			fmt.Sprintf("%.2f", o.Amount),
			o.Status,
		}
		buf.WriteString(e.formatter.WriteRow(row))
	}

	buf.WriteString(e.formatter.WriteFooter(len(orders)))

	result := buf.String()
	fmt.Printf("[%s] 完成，%d 字节\n\n", e.formatter.Name(), len(result))
	return result
}

// ── CSV 格式 ───────────────────────────────────────────────────────────────────

type CSVFormatter struct{}

func (c *CSVFormatter) Name() string { return "CSV" }

func (c *CSVFormatter) WriteHeader(cols []string) string {
	return strings.Join(cols, ",") + "\n"
}

func (c *CSVFormatter) WriteRow(row []string) string {
	for i, v := range row {
		if strings.ContainsAny(v, ",\"") {
			row[i] = `"` + strings.ReplaceAll(v, `"`, `""`) + `"`
		}
	}
	return strings.Join(row, ",") + "\n"
}

func (c *CSVFormatter) WriteFooter(n int) string {
	return fmt.Sprintf("# total: %d rows\n", n)
}

// ── JSON 格式 ──────────────────────────────────────────────────────────────────

type JSONFormatter struct {
	rows []string // 先攒行，Footer 里组装完整数组
}

func (j *JSONFormatter) Name() string { return "JSON" }

func (j *JSONFormatter) WriteHeader(_ []string) string {
	j.rows = nil
	return ""
}

func (j *JSONFormatter) WriteRow(row []string) string {
	obj := fmt.Sprintf(
		`{"id":%s,"customer":%q,"product":%q,"amount":%s,"status":%q}`,
		row[0], row[1], row[2], row[3], row[4],
	)
	j.rows = append(j.rows, obj)
	return "" // 先不写，等 WriteFooter 一次输出
}

func (j *JSONFormatter) WriteFooter(_ int) string {
	return "[\n  " + strings.Join(j.rows, ",\n  ") + "\n]\n"
}

// ── Markdown 格式 ──────────────────────────────────────────────────────────────

type MarkdownFormatter struct{}

func (m *MarkdownFormatter) Name() string { return "Markdown" }

func (m *MarkdownFormatter) WriteHeader(cols []string) string {
	header := "| " + strings.Join(cols, " | ") + " |\n"
	sep := "|" + strings.Repeat(" :--- |", len(cols)) + "\n"
	return header + sep
}

func (m *MarkdownFormatter) WriteRow(row []string) string {
	return "| " + strings.Join(row, " | ") + " |\n"
}

func (m *MarkdownFormatter) WriteFooter(n int) string {
	return fmt.Sprintf("\n> 共 %d 条记录\n", n)
}

// ── main ───────────────────────────────────────────────────────────────────────

func main() {
	fmt.Println("━━ CSV 导出 ━━")
	fmt.Println((&ReportExporter{&CSVFormatter{}}).Export(orders))

	fmt.Println("━━ JSON 导出 ━━")
	fmt.Println((&ReportExporter{&JSONFormatter{}}).Export(orders))

	fmt.Println("━━ Markdown 导出 ━━")
	fmt.Println((&ReportExporter{&MarkdownFormatter{}}).Export(orders))

	// 体现模板方法优势：换数据（过滤"已完成"）时骨架不动，只换入参
	fmt.Println("━━ 仅导出已完成订单（骨架不变，过滤后数据传入）━━")
	completed := filterOrders(orders, "已完成")
	fmt.Println((&ReportExporter{&CSVFormatter{}}).Export(completed))
}

func filterOrders(src []Order, status string) []Order {
	var result []Order
	for _, o := range src {
		if o.Status == status {
			result = append(result, o)
		}
	}
	return result
}
