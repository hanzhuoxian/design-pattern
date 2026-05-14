package main

import (
	"fmt"
	"strings"
	"time"
)

// 场景：HTTP 中间件管道
//
// 痛点对比：
//   不用责任链 → 每个 Handler 自己处理日志、鉴权、限流，逻辑全部耦合在一起
//     换一条路由想跳过鉴权，就要改 Handler 内部代码
//     新增一个中间件（如链路追踪），需要改每一个 Handler
//
//   用责任链  → 每个中间件只做一件事，决定"处理后继续"还是"直接拦截"
//     路由级别自由组合中间件，Handler 只写业务逻辑
//     新增/删除/调换中间件顺序，Handler 代码零改动

// ── 核心类型 ──────────────────────────────────────────────────────────────────

type Request struct {
	Method  string
	Path    string
	Headers map[string]string
	Body    string
}

type Response struct {
	Status int
	Body   string
}

func (r Response) String() string {
	return fmt.Sprintf("%d  %s", r.Status, r.Body)
}

// Handler 是链中每个节点的接口
type Handler interface {
	SetNext(Handler) Handler
	Handle(req *Request) Response
}

// ── 基类：封装链传递逻辑 ──────────────────────────────────────────────────────

type baseHandler struct {
	next Handler
}

func (b *baseHandler) SetNext(h Handler) Handler {
	b.next = h
	return h
}

func (b *baseHandler) forward(req *Request) Response {
	if b.next != nil {
		return b.next.Handle(req)
	}
	return Response{Status: 404, Body: "no handler"}
}

// ── 中间件：日志 ──────────────────────────────────────────────────────────────

type LoggerMiddleware struct{ baseHandler }

func (l *LoggerMiddleware) Handle(req *Request) Response {
	start := time.Now()
	resp := l.forward(req)
	fmt.Printf("[Logger]  %s %s → %d  (%v)\n",
		req.Method, req.Path, resp.Status, time.Since(start).Round(time.Microsecond))
	return resp
}

// ── 中间件：鉴权（检查 Authorization 头）────────────────────────────────────

type AuthMiddleware struct {
	baseHandler
	validTokens map[string]string // token → username
}

func (a *AuthMiddleware) Handle(req *Request) Response {
	token := req.Headers["Authorization"]
	if token == "" {
		fmt.Printf("[Auth]    拦截：缺少 Authorization 头\n")
		return Response{Status: 401, Body: "unauthorized: missing token"}
	}
	token = strings.TrimPrefix(token, "Bearer ")
	if _, ok := a.validTokens[token]; !ok {
		fmt.Printf("[Auth]    拦截：无效 token\n")
		return Response{Status: 401, Body: "unauthorized: invalid token"}
	}
	fmt.Printf("[Auth]    通过：%s\n", a.validTokens[token])
	return a.forward(req)
}

// ── 中间件：限流（简单计数，每个 IP 最多 N 次）───────────────────────────────

type RateLimitMiddleware struct {
	baseHandler
	counts map[string]int
	limit  int
}

func (r *RateLimitMiddleware) Handle(req *Request) Response {
	ip := req.Headers["X-Real-IP"]
	r.counts[ip]++
	if r.counts[ip] > r.limit {
		fmt.Printf("[RateLimit] 拦截：%s 已超过 %d 次/窗口\n", ip, r.limit)
		return Response{Status: 429, Body: "too many requests"}
	}
	fmt.Printf("[RateLimit] 放行：%s（第 %d 次）\n", ip, r.counts[ip])
	return r.forward(req)
}

// ── 业务 Handler：只写业务逻辑，感知不到任何中间件 ──────────────────────────

type UserHandler struct{ baseHandler }

func (u *UserHandler) Handle(req *Request) Response {
	switch req.Method + " " + req.Path {
	case "GET /users":
		return Response{Status: 200, Body: `[{"id":1,"name":"Alice"},{"id":2,"name":"Bob"}]`}
	case "POST /users":
		return Response{Status: 201, Body: fmt.Sprintf(`{"created":true,"body":"%s"}`, req.Body)}
	default:
		return Response{Status: 404, Body: "not found"}
	}
}

// ── Chain：将中间件串成链，返回链头 ──────────────────────────────────────────

func Chain(handlers ...Handler) Handler {
	for i := 0; i < len(handlers)-1; i++ {
		handlers[i].SetNext(handlers[i+1])
	}
	return handlers[0]
}

// ── main ─────────────────────────────────────────────────────────────────────

func main() {
	tokens := map[string]string{"secret-token-alice": "Alice", "secret-token-bob": "Bob"}
	handler := &UserHandler{}

	// 受保护路由：Logger → Auth → RateLimit → UserHandler
	protected := Chain(
		&LoggerMiddleware{},
		&AuthMiddleware{validTokens: tokens},
		&RateLimitMiddleware{counts: map[string]int{}, limit: 2},
		handler,
	)

	fmt.Println("━━ 正常请求（携带有效 token）━━")
	protected.Handle(&Request{
		Method:  "GET",
		Path:    "/users",
		Headers: map[string]string{"Authorization": "Bearer secret-token-alice", "X-Real-IP": "1.2.3.4"},
	})

	fmt.Println("\n━━ 缺少 Authorization 头 ━━")
	protected.Handle(&Request{
		Method:  "GET",
		Path:    "/users",
		Headers: map[string]string{"X-Real-IP": "1.2.3.5"},
	})

	fmt.Println("\n━━ 无效 token ━━")
	protected.Handle(&Request{
		Method:  "GET",
		Path:    "/users",
		Headers: map[string]string{"Authorization": "Bearer wrong-token", "X-Real-IP": "1.2.3.6"},
	})

	fmt.Println("\n━━ 限流触发（同一 IP 连续 3 次）━━")
	for i := 1; i <= 3; i++ {
		protected.Handle(&Request{
			Method:  "POST",
			Path:    "/users",
			Headers: map[string]string{"Authorization": "Bearer secret-token-bob", "X-Real-IP": "9.9.9.9"},
			Body:    fmt.Sprintf(`{"name":"user%d"}`, i),
		})
	}

	// 公开路由：只挂 Logger，跳过 Auth 和 RateLimit
	// 体现责任链优势：Handler 代码不变，仅重新组装链
	fmt.Println("\n━━ 公开路由（仅 Logger，无需 token）━━")
	public := Chain(&LoggerMiddleware{}, handler)
	public.Handle(&Request{
		Method:  "GET",
		Path:    "/users",
		Headers: map[string]string{"X-Real-IP": "5.5.5.5"},
	})
}
