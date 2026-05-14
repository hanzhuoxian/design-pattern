package main

import (
	"fmt"
	"strings"
	"time"
)

// 场景：报表查询服务
//
// 痛点对比：
//   不用 Proxy → 调用方直接 new ReportService，启动即建立 DB 连接（慢）
//               每次 Query 都打 DB，相同参数重复查询白白浪费
//               鉴权逻辑散落在每个调用方，遗漏一处就是漏洞
//
//   用 Proxy  → 调用方只面对 Reporter 接口
//               Proxy 统一处理：懒加载连接、鉴权拦截、结果缓存
//               真实服务零修改，调用方对这三层逻辑完全无感知

// ── Subject 接口 ──────────────────────────────────────────────────────────────

type Reporter interface {
	Query(dept, month string) (string, error)
}

// ── Real Subject：真实报表服务（模拟昂贵的 DB 操作）─────────────────────────

type ReportService struct {
	connected bool
}

func (s *ReportService) connect() {
	fmt.Println("  [DB] 建立数据库连接（耗时操作，模拟 200ms）...")
	time.Sleep(200 * time.Millisecond)
	s.connected = true
	fmt.Println("  [DB] 连接成功")
}

func (s *ReportService) Query(dept, month string) (string, error) {
	if !s.connected {
		s.connect()
	}
	fmt.Printf("  [DB] 执行查询 dept=%q month=%q\n", dept, month)
	return fmt.Sprintf("报表[%s-%s]: 营收 ¥%.0fW", dept, month, float64(len(dept)+len(month))*12.5), nil
}

// ── Proxy：懒加载 + 鉴权 + 缓存，三层关注点对真实服务透明 ────────────────────

type ReportProxy struct {
	real   *ReportService // nil 直到首次真正使用，懒加载
	cache  map[string]string
	caller string
	admins map[string]bool
}

func NewReportProxy(caller string) *ReportProxy {
	return &ReportProxy{
		cache:  make(map[string]string),
		caller: caller,
		admins: map[string]bool{"alice": true, "bob": true},
	}
}

func (p *ReportProxy) Query(dept, month string) (string, error) {
	// 1. 鉴权：非授权账户直接拦截，真实服务根本不会被触碰
	if !p.admins[p.caller] {
		return "", fmt.Errorf("鉴权失败：%q 无报表查询权限", p.caller)
	}

	// 2. 缓存：命中则直接返回，跳过 DB 查询
	key := dept + ":" + month
	if v, ok := p.cache[key]; ok {
		fmt.Printf("  [Cache] 命中 %q，跳过 DB\n", key)
		return v, nil
	}

	// 3. 懒加载：首次真正需要时才创建真实服务并建立 DB 连接
	if p.real == nil {
		fmt.Println("  [Proxy] 首次使用，懒加载 ReportService...")
		p.real = &ReportService{}
	}

	result, err := p.real.Query(dept, month)
	if err == nil {
		p.cache[key] = result
	}
	return result, err
}

// ── main ─────────────────────────────────────────────────────────────────────

func main() {
	sep := strings.Repeat("─", 50)

	// 1. 懒加载：构造 Proxy 本身无任何 DB 开销，首次 Query 才触发连接
	fmt.Println("━━ 懒加载（Proxy 构造无 DB 开销，首次 Query 才建立连接）━━")
	proxy := NewReportProxy("alice")
	fmt.Println("  Proxy 已创建，DB 连接尚未建立（real == nil）")
	result, _ := proxy.Query("研发部", "2024-03")
	fmt.Println("  结果:", result)
	fmt.Println(sep)

	// 2. 缓存：相同参数第二次调用不打 DB；不同参数才真正查询
	fmt.Println("━━ 缓存（相同查询命中缓存，DB 沉默）━━")
	result, _ = proxy.Query("研发部", "2024-03") // 命中缓存
	fmt.Println("  结果:", result)
	result, _ = proxy.Query("销售部", "2024-03") // 新参数，打 DB
	fmt.Println("  结果:", result)
	fmt.Println(sep)

	// 3. 鉴权：未授权账户被 Proxy 拦截，真实服务从未被初始化
	fmt.Println("━━ 鉴权（非授权账户被拦截，真实服务不受影响）━━")
	guestProxy := NewReportProxy("guest")
	_, err := guestProxy.Query("研发部", "2024-03")
	fmt.Println("  错误:", err)
	fmt.Printf("  guestProxy.real == nil: %v（DB 连接从未建立）\n", guestProxy.real == nil)
	fmt.Println(sep)

	// 4. 接口统一：调用方只依赖 Reporter，可随时用真实服务替换 Proxy
	fmt.Println("━━ 接口统一（调用方面对 Reporter 接口，不感知 Proxy 的存在）━━")
	var r Reporter = NewReportProxy("bob")
	result, _ = r.Query("产品部", "2024-04")
	fmt.Println("  结果:", result)
}
