// 场景：电商平台多数据源订单汇总
//
// 痛点对比：
//   不用迭代器 → 汇总函数要针对每种数据结构写不同的遍历分支
//     新增数据源（Redis、Kafka）必须修改汇总函数本身
//     遍历细节（分页、解析 CSV）与业务统计逻辑耦合在一起
//
//   用迭代器  → 汇总/统计函数只调用 HasNext/Next，完全屏蔽存储细节
//     新增数据源只需实现 Iterator 接口，业务函数零改动
//     MultiIterator 将多个数据源串联成一个，对上层透明

package main

import (
	"encoding/csv"
	"fmt"
	"strconv"
	"strings"
)

// ── 核心类型 ──────────────────────────────────────────────────────────────────

type Order struct {
	ID     string
	Amount float64
	Status string // "paid" | "refunded" | "pending"
}

type Iterator interface {
	HasNext() bool
	Next() Order
}

// ── 数据源 1：内存缓存（slice）───────────────────────────────────────────────

type SliceIterator struct {
	data  []Order
	index int
}

func (s *SliceIterator) HasNext() bool { return s.index < len(s.data) }
func (s *SliceIterator) Next() Order {
	o := s.data[s.index]
	s.index++
	return o
}

// ── 数据源 2：数据库分页游标（模拟 LIMIT/OFFSET，每页按需加载）────────────────

type DBPage struct{ Orders []Order }

type DBCursor struct {
	pages   []DBPage
	pageIdx int
	rowIdx  int
}

func (d *DBCursor) HasNext() bool {
	for d.pageIdx < len(d.pages) {
		if d.rowIdx < len(d.pages[d.pageIdx].Orders) {
			return true
		}
		d.pageIdx++
		d.rowIdx = 0
	}
	return false
}

func (d *DBCursor) Next() Order {
	o := d.pages[d.pageIdx].Orders[d.rowIdx]
	d.rowIdx++
	return o
}

// ── 数据源 3：CSV 文件（按行解析）───────────────────────────────────────────

type CSVIterator struct {
	records [][]string
	index   int
}

func NewCSVIterator(data string) *CSVIterator {
	r := csv.NewReader(strings.NewReader(data))
	records, _ := r.ReadAll()
	if len(records) > 0 {
		records = records[1:] // skip header
	}
	return &CSVIterator{records: records}
}

func (c *CSVIterator) HasNext() bool { return c.index < len(c.records) }
func (c *CSVIterator) Next() Order {
	rec := c.records[c.index]
	c.index++
	amount, _ := strconv.ParseFloat(rec[1], 64)
	return Order{ID: rec[0], Amount: amount, Status: rec[2]}
}

// ── 多源串联迭代器 ─────────────────────────────────────────────────────────────

type MultiIterator struct {
	iters []Iterator
	cur   int
}

func (m *MultiIterator) HasNext() bool {
	for m.cur < len(m.iters) {
		if m.iters[m.cur].HasNext() {
			return true
		}
		m.cur++
	}
	return false
}

func (m *MultiIterator) Next() Order { return m.iters[m.cur].Next() }

// ── 业务逻辑：只依赖 Iterator 接口，对底层存储无感知 ─────────────────────────

func SumPaid(iter Iterator) float64 {
	var total float64
	for iter.HasNext() {
		if o := iter.Next(); o.Status == "paid" {
			total += o.Amount
		}
	}
	return total
}

func CountByStatus(iter Iterator) map[string]int {
	counts := map[string]int{}
	for iter.HasNext() {
		counts[iter.Next().Status]++
	}
	return counts
}

// ── main ─────────────────────────────────────────────────────────────────────

func main() {
	cacheOrders := []Order{
		{ID: "C001", Amount: 199.9, Status: "paid"},
		{ID: "C002", Amount: 88.0, Status: "refunded"},
		{ID: "C003", Amount: 350.0, Status: "paid"},
	}
	dbPages := []DBPage{
		{Orders: []Order{
			{ID: "D001", Amount: 120.0, Status: "paid"},
			{ID: "D002", Amount: 45.0, Status: "pending"},
		}},
		{Orders: []Order{
			{ID: "D003", Amount: 670.0, Status: "paid"},
			{ID: "D004", Amount: 30.0, Status: "refunded"},
		}},
	}
	csvData := "id,amount,status\nF001,250.0,paid\nF002,99.9,pending\nF003,500.0,paid\n"

	fmt.Println("━━ 各数据源独立统计（同一 SumPaid 函数，换迭代器即可）━━")
	fmt.Printf("  内存缓存 已支付合计：%.2f\n", SumPaid(&SliceIterator{data: cacheOrders}))
	fmt.Printf("  数据库   已支付合计：%.2f\n", SumPaid(&DBCursor{pages: dbPages}))
	fmt.Printf("  CSV文件  已支付合计：%.2f\n", SumPaid(NewCSVIterator(csvData)))

	fmt.Println("\n━━ 多源聚合（三个数据源串联，SumPaid 代码零改动）━━")
	all := &MultiIterator{iters: []Iterator{
		&SliceIterator{data: cacheOrders},
		&DBCursor{pages: dbPages},
		NewCSVIterator(csvData),
	}}
	fmt.Printf("  全部数据源 已支付合计：%.2f\n", SumPaid(all))

	fmt.Println("\n━━ 状态分布统计（复用 CountByStatus，与数据源无关）━━")
	all2 := &MultiIterator{iters: []Iterator{
		&SliceIterator{data: cacheOrders},
		&DBCursor{pages: dbPages},
		NewCSVIterator(csvData),
	}}
	for status, count := range CountByStatus(all2) {
		fmt.Printf("  %-10s %d 笔\n", status, count)
	}
}
