package main

import (
	"fmt"
	"strings"
	"time"
)

// 场景：构建 API HTTP 请求
//
// 痛点对比：
//   不用 Builder → NewRequest(method, url, body, timeout, retries, token, proxy, ...)
//   参数一多，调用时根本看不清哪个参数是什么，可选项也无法跳过
//
//   用 Builder  → 链式调用，按需叠加，一目了然

type Request struct {
	Method    string
	URL       string
	Headers   map[string]string
	Body      string
	Timeout   time.Duration
	Retries   int
	AuthToken string
}

func (r Request) String() string {
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("%s %s\n", r.Method, r.URL))
	for k, v := range r.Headers {
		sb.WriteString(fmt.Sprintf("  %-20s %s\n", k+":", v))
	}
	if r.Body != "" {
		sb.WriteString(fmt.Sprintf("  %-20s %s\n", "Body:", r.Body))
	}
	if r.Timeout > 0 {
		sb.WriteString(fmt.Sprintf("  %-20s %v\n", "Timeout:", r.Timeout))
	}
	if r.Retries > 0 {
		sb.WriteString(fmt.Sprintf("  %-20s %d\n", "Retries:", r.Retries))
	}
	return sb.String()
}

// ── Builder ──────────────────────────────────────────────────────────────────

type RequestBuilder struct {
	request Request
}

func NewRequest(method, url string) *RequestBuilder {
	return &RequestBuilder{
		request: Request{
			Method:  method,
			URL:     url,
			Headers: make(map[string]string),
		},
	}
}

func (b *RequestBuilder) Header(key, value string) *RequestBuilder {
	b.request.Headers[key] = value
	return b
}

func (b *RequestBuilder) Body(body string) *RequestBuilder {
	b.request.Body = body
	return b
}

func (b *RequestBuilder) Timeout(d time.Duration) *RequestBuilder {
	b.request.Timeout = d
	return b
}

func (b *RequestBuilder) Retries(n int) *RequestBuilder {
	b.request.Retries = n
	return b
}

func (b *RequestBuilder) Auth(token string) *RequestBuilder {
	b.request.AuthToken = token
	b.request.Headers["Authorization"] = "Bearer " + token
	return b
}

func (b *RequestBuilder) Build() Request {
	return b.request
}

// ── Director：封装业务层固定的构建流程 ────────────────────────────────────────
// Director 的价值：调用方只关心"我要什么"，不用关心"怎么组装"
// 不同 Director 可以面向不同场景（内网服务 vs 外部 API vs 异步任务）

type APIDirector struct {
	baseURL string
	token   string
}

func NewAPIDirector(baseURL, token string) *APIDirector {
	return &APIDirector{baseURL: baseURL, token: token}
}

// 标准读接口：统一加鉴权 + JSON Accept + 30s 超时
func (d *APIDirector) GET(path string) Request {
	return NewRequest("GET", d.baseURL+path).
		Header("Accept", "application/json").
		Auth(d.token).
		Timeout(30 * time.Second).
		Build()
}

// 标准写接口：额外加重试，超时放宽到 60s
func (d *APIDirector) POST(path, body string) Request {
	return NewRequest("POST", d.baseURL+path).
		Header("Content-Type", "application/json").
		Header("Accept", "application/json").
		Auth(d.token).
		Body(body).
		Timeout(60 * time.Second).
		Retries(3).
		Build()
}

// ── main ─────────────────────────────────────────────────────────────────────

func main() {
	// 1. 直接用 Builder 自由组合——简单请求不需要任何多余参数
	fmt.Println("━━ 简单 GET（无鉴权）━━")
	req1 := NewRequest("GET", "https://api.example.com/public/items").
		Header("Accept", "application/json").
		Timeout(10 * time.Second).
		Build()
	fmt.Println(req1)

	// 2. 复杂请求按需叠加，每一步意图清晰
	fmt.Println("━━ 带鉴权 + 重试的 POST ━━")
	req2 := NewRequest("POST", "https://api.example.com/orders").
		Header("Content-Type", "application/json").
		Auth("user-secret-token").
		Body(`{"item":"book","qty":2}`).
		Timeout(10 * time.Second).
		Retries(3).
		Build()
	fmt.Println(req2)

	// 3. Director 封装固定规范，调用方只需传业务参数
	//    同一套鉴权/超时/重试策略不用到处重复
	fmt.Println("━━ 通过 Director 构建（规范统一）━━")
	api := NewAPIDirector("https://api.example.com", "prod-token-xyz")

	fmt.Println(api.GET("/users/42"))
	fmt.Println(api.POST("/users", `{"name":"Alice","role":"admin"}`))
}
