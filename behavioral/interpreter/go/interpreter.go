package main

// 场景：电商促销规则引擎
//
// 痛点对比：
//   不用解释器 → if/else 硬编码促销规则，运营每改一条规则都要修改代码重新上线
//                 AND / OR 组合规则指数级膨胀分支，规则之间互相干扰难以维护
//
//   用解释器  → 规则以字符串存入数据库，解析为表达式树后对用户数据直接求值
//                 新增/修改规则只改数据，代码不动；AND / OR 节点自由组合，运行时生效

import (
	"fmt"
	"strconv"
	"strings"
)

// Context 保存待求值的用户数据
type Context struct {
	vars map[string]interface{}
}

func NewContext(vars map[string]interface{}) *Context {
	return &Context{vars: vars}
}

func (c *Context) Get(key string) interface{} {
	return c.vars[key]
}

// Expression 解释器统一接口
type Expression interface {
	Interpret(ctx *Context) bool
}

// ── 终结符表达式 ──────────────────────────────────────────────────────────────

// BoolExpression 处理 field == true/false
type BoolExpression struct {
	field    string
	expected bool
}

func (e *BoolExpression) Interpret(ctx *Context) bool {
	v, ok := ctx.Get(e.field).(bool)
	return ok && v == e.expected
}

// NumericExpression 处理 field >/</>=/<= number
type NumericExpression struct {
	field string
	op    string
	value float64
}

func (e *NumericExpression) Interpret(ctx *Context) bool {
	var v float64
	switch r := ctx.Get(e.field).(type) {
	case int:
		v = float64(r)
	case float64:
		v = r
	default:
		return false
	}
	switch e.op {
	case ">":
		return v > e.value
	case "<":
		return v < e.value
	case ">=":
		return v >= e.value
	case "<=":
		return v <= e.value
	}
	return false
}

// ── 非终结符表达式 ────────────────────────────────────────────────────────────

// AndExpression 逻辑与
type AndExpression struct{ left, right Expression }

func (e *AndExpression) Interpret(ctx *Context) bool {
	return e.left.Interpret(ctx) && e.right.Interpret(ctx)
}

// OrExpression 逻辑或
type OrExpression struct{ left, right Expression }

func (e *OrExpression) Interpret(ctx *Context) bool {
	return e.left.Interpret(ctx) || e.right.Interpret(ctx)
}

// ── 解析器：将规则字符串构建成表达式树 ───────────────────────────────────────

func parse(rule string) Expression {
	rule = strings.TrimSpace(rule)

	// OR 优先级最低，先分割
	if idx := strings.Index(strings.ToUpper(rule), " OR "); idx != -1 {
		return &OrExpression{
			left:  parse(rule[:idx]),
			right: parse(rule[idx+4:]),
		}
	}
	if idx := strings.Index(strings.ToUpper(rule), " AND "); idx != -1 {
		return &AndExpression{
			left:  parse(rule[:idx]),
			right: parse(rule[idx+5:]),
		}
	}

	// 数值比较（注意 >= 和 <= 要先于 > 和 < 匹配）
	for _, op := range []string{">=", "<=", ">", "<"} {
		if idx := strings.Index(rule, op); idx != -1 {
			field := strings.TrimSpace(rule[:idx])
			val, _ := strconv.ParseFloat(strings.TrimSpace(rule[idx+len(op):]), 64)
			return &NumericExpression{field: field, op: op, value: val}
		}
	}

	// 布尔比较
	if idx := strings.Index(rule, "=="); idx != -1 {
		field := strings.TrimSpace(rule[:idx])
		val := strings.TrimSpace(rule[idx+2:])
		return &BoolExpression{field: field, expected: val == "true"}
	}

	return &BoolExpression{field: rule, expected: true}
}

// ── 演示 ──────────────────────────────────────────────────────────────────────

func main() {
	// 促销规则来自数据库，运营可随时修改，无需改代码
	rules := []struct {
		name string
		expr string
	}{
		{"双十一VIP专属折扣", "vip == true AND order_count > 5"},
		{"新客免运费", "order_count <= 1"},
		{"高消费或高等级用户奖励", "level > 3 OR spend > 1000"},
	}

	users := []map[string]interface{}{
		{"name": "Alice", "vip": true, "order_count": 8, "level": 2, "spend": 500.0},
		{"name": "Bob", "vip": false, "order_count": 1, "level": 1, "spend": 50.0},
		{"name": "Carol", "vip": true, "order_count": 3, "level": 5, "spend": 300.0},
	}

	for _, rule := range rules {
		expr := parse(rule.expr) // 解析一次，可重复使用
		fmt.Printf("\n规则「%s」: %s\n", rule.name, rule.expr)
		for _, u := range users {
			ctx := NewContext(u)
			fmt.Printf("  %-6s → %v\n", u["name"], expr.Interpret(ctx))
		}
	}
}
