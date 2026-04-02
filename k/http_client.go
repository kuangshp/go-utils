package k

// http_client.go —— 生产级 HTTP 客户端
//
// 设计目标：
//   - Builder 模式构建客户端，所有配置链式调用，类型安全，IDE 全量补全
//   - 请求级参数通过 R() 构建，与客户端配置严格分离
//   - 可靠性能力（重试/熔断/限速/缓存）内置且按固定顺序执行，无顺序依赖问题
//   - 重试直接复用同包 retry.go，不重复定义
//
// 执行顺序（每次请求）：
//  缓存命中 → 熔断检查 → 限速等待 → 发送（含重试）→ 日志 → 指标 → 熔断记录 → 写缓存
//
// 快速开始：
//
//	// 1. 构建客户端（程序启动时一次性创建，全局复用）
//	client, err := NewClient("https://api.example.com").
//	    Timeout(10 * time.Second).
//	    BearerToken(func() string { return tokenStore.Get() }).
//	    Retry(WithMaxRetries(3), WithRetryDelay(500*time.Millisecond), WithDelayMultiplier(2.0)).
//	    CircuitBreaker(NewCircuitBreaker()).
//	    Logger(log.Printf).
//	    Build()
//
//	// 2. 发送请求（每次调用时按需附加请求级参数）
//	resp, err := client.Get("/users",
//	    R().QueryParams(map[string]string{"page": "1"}).ExpectStatus(200),
//	)
//
//	// 3. 读取响应
//	var users []User
//	err = client.ReadJSON(resp, &users)

import (
	"bytes"
	"compress/gzip"
	"context"
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math/rand"
	"mime/multipart"
	"net"
	"net/http"
	"net/url"
	"os"
	"sort"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

// ═══════════════════════════════════════════════════════
// 错误定义
// ═══════════════════════════════════════════════════════

// HTTPError 表示服务端返回了非预期的 HTTP 状态码。
// 当请求设置了 ExpectStatus 且实际状态码不匹配时返回此错误。
//
// 用法：
//
//	resp, err := client.Get("/resource", R().ExpectStatus(200))
//	var httpErr *HTTPError
//	if errors.As(err, &httpErr) {
//	    fmt.Println(httpErr.StatusCode, string(httpErr.Body))
//	}
type HTTPError struct {
	StatusCode int    // 实际返回的 HTTP 状态码
	Status     string // 状态描述，例如 "404 Not Found"
	Body       []byte // 响应体原始内容，用于错误详情提取
}

func (e *HTTPError) Error() string {
	return fmt.Sprintf("http error: %s (code=%d)", e.Status, e.StatusCode)
}

// ErrCircuitOpen 熔断器处于开启状态时返回此错误，表示请求被拒绝
var ErrCircuitOpen = errors.New("circuit breaker is open")

// ═══════════════════════════════════════════════════════
// Proxy Pool 代理池（新增，支持轮询/随机/健康检查）
// ═══════════════════════════════════════════════════════

type ProxyStrategy string

const (
	ProxyStrategyRandom     ProxyStrategy = "random"      // 随机选取代理
	ProxyStrategyRoundRobin ProxyStrategy = "round_robin" // 轮询选取代理
	ProxyStrategyWeight     ProxyStrategy = "weight"      // 加权选取代理
)

// Proxy 代理实例结构体
type Proxy struct {
	URL   string // 代理地址
	alive bool   // 代理是否存活
}

// ProxyPool 代理池（并发安全）
type ProxyPool struct {
	mu       sync.RWMutex
	proxies  []*Proxy
	strategy ProxyStrategy
	index    int // 轮询使用的游标
}

// NewProxyPool 创建代理池
// urls: 代理地址列表
// strategy: 代理选取策略
func NewProxyPool(urls []string, strategy ProxyStrategy) *ProxyPool {
	p := &ProxyPool{
		strategy: strategy,
		proxies:  make([]*Proxy, 0, len(urls)),
	}
	for _, u := range urls {
		p.proxies = append(p.proxies, &Proxy{URL: u, alive: true})
	}
	rand.New(rand.NewSource(time.Now().UnixNano()))
	return p
}

// Get 按照配置的策略获取一个可用代理
func (p *ProxyPool) Get() (*url.URL, error) {
	p.mu.RLock()
	defer p.mu.RUnlock()

	if len(p.proxies) == 0 {
		return nil, nil
	}

	// 过滤出可用代理
	alive := make([]*Proxy, 0)
	for _, px := range p.proxies {
		if px.alive {
			alive = append(alive, px)
		}
	}
	if len(alive) == 0 {
		return nil, errors.New("no alive proxy in pool")
	}

	var target *Proxy
	switch p.strategy {
	case ProxyStrategyRandom:
		target = alive[rand.Intn(len(alive))]
	case ProxyStrategyRoundRobin:
		target = alive[p.index%len(alive)]
		p.index++
	default:
		target = alive[0]
	}

	return url.Parse(strings.TrimSpace(target.URL))
}

// MarkDown 标记某个代理下线
func (p *ProxyPool) MarkDown(proxyURL string) {
	p.mu.Lock()
	defer p.mu.Unlock()
	for _, px := range p.proxies {
		if px.URL == proxyURL {
			px.alive = false
		}
	}
}

// MarkUp 标记某个代理上线
func (p *ProxyPool) MarkUp(proxyURL string) {
	p.mu.Lock()
	defer p.mu.Unlock()
	for _, px := range p.proxies {
		if px.URL == proxyURL {
			px.alive = true
		}
	}
}

// ═══════════════════════════════════════════════════════
// ClientBuilder —— 客户端构建器
// ═══════════════════════════════════════════════════════

// ClientBuilder 以链式调用方式配置并构建 HTTPClient。
// 所有方法均返回 *ClientBuilder 自身，支持连续调用。
// 调用 Build() 完成构建，Build() 会校验配置合法性。
//
// 默认值：
//   - Timeout:          50s
//   - HandshakeTimeout: 10s
//   - ResponseTimeout:  10s
//   - KeepAlive:        30s
//   - MaxIdleConns:     100
//   - MaxConnsPerHost:  10
//   - IdleConnTimeout:  90s
type ClientBuilder struct {
	baseURL          string
	timeout          time.Duration
	handshakeTimeout time.Duration
	responseTimeout  time.Duration
	keepAlive        time.Duration
	maxIdleConns     int
	maxConnsPerHost  int
	idleConnTimeout  time.Duration
	tlsConfig        *tls.Config
	proxyURL         string     // 单个代理地址
	proxyPool        *ProxyPool // 代理池（新增）
	compressed       bool
	checkRedirect    func(*http.Request, []*http.Request) error
	bearerTokenFn    func() string
	basicUsername    string
	basicPassword    string
	signFn           func(*http.Request) error
	logger           func(format string, args ...any)
	metrics          *Metrics
	circuitBreaker   *CircuitBreaker
	rateLimiter      *RateLimiter
	responseCache    *ResponseCache
	retryOpts        []Option
}

// NewClient 创建 ClientBuilder。
//
// 参数：
//   - baseURL: 所有请求路径的前缀，例如 "https://api.example.com/v1"。
//     末尾的 "/" 会自动去除。发起请求时传入的 path 会拼接在此后面。
//     若传空字符串，则每次请求须传完整 URL。
//
// 示例：
//
//	client, err := NewClient("https://api.example.com").Timeout(10*time.Second).Build()
func NewClient(baseURL string) *ClientBuilder {
	return &ClientBuilder{
		baseURL:          strings.TrimRight(baseURL, "/"),
		timeout:          50 * time.Second,
		handshakeTimeout: 10 * time.Second,
		responseTimeout:  10 * time.Second,
		keepAlive:        30 * time.Second,
		maxIdleConns:     100,
		maxConnsPerHost:  10,
		idleConnTimeout:  90 * time.Second,
	}
}

// ─── 超时配置 ──────────────────────────────────────────

// Timeout 设置单次请求的整体超时（含连接、发送、读取响应全过程）。
//
// 参数：
//   - d: 超时时长，例如 10*time.Second。设为 0 表示不限制（不推荐）。
//
// 默认值：50s
func (b *ClientBuilder) Timeout(d time.Duration) *ClientBuilder {
	b.timeout = d
	return b
}

// HandshakeTimeout 设置 TLS 握手的超时时间。
// 仅对 HTTPS 请求生效，超时后返回握手失败错误。
//
// 参数：
//   - d: 超时时长，例如 5*time.Second。
//
// 默认值：10s
func (b *ClientBuilder) HandshakeTimeout(d time.Duration) *ClientBuilder {
	b.handshakeTimeout = d
	return b
}

// ResponseTimeout 设置等待服务端返回响应头的超时时间。
// 不包含读取响应体的时间；如需限制整体耗时请用 Timeout。
//
// 参数：
//   - d: 超时时长，例如 5*time.Second。
//
// 默认值：10s
func (b *ClientBuilder) ResponseTimeout(d time.Duration) *ClientBuilder {
	b.responseTimeout = d
	return b
}

// ─── 连接池配置 ────────────────────────────────────────

// KeepAlive 设置 TCP KeepAlive 探测报文的发送间隔。
// 用于检测长连接是否存活，避免连接被中间网络设备静默断开。
//
// 参数：
//   - d: 探测间隔，例如 30*time.Second。设为负值可禁用 KeepAlive。
//
// 默认值：30s
func (b *ClientBuilder) KeepAlive(d time.Duration) *ClientBuilder {
	b.keepAlive = d
	return b
}

// MaxIdleConns 设置连接池中最大空闲连接总数（跨所有 host）。
// 数值越大，高并发时复用连接的概率越高，但占用内存也越多。
//
// 参数：
//   - n: 最大空闲连接数，例如 200。
//
// 默认值：100
func (b *ClientBuilder) MaxIdleConns(n int) *ClientBuilder {
	b.maxIdleConns = n
	return b
}

// MaxConnsPerHost 设置对单个 host 的最大并发连接数。
// 超出限制时新请求会等待已有连接释放。
//
// 参数：
//   - n: 每 host 最大连接数，例如 20。
//
// 默认值：10
func (b *ClientBuilder) MaxConnsPerHost(n int) *ClientBuilder {
	b.maxConnsPerHost = n
	return b
}

// IdleConnTimeout 设置连接在连接池中空闲多久后被关闭回收。
//
// 参数：
//   - d: 空闲超时时长，例如 60*time.Second。
//
// 默认值：90s
func (b *ClientBuilder) IdleConnTimeout(d time.Duration) *ClientBuilder {
	b.idleConnTimeout = d
	return b
}

// ─── TLS / 代理 / 压缩 ────────────────────────────────

// TLS 配置 HTTPS 的 TLS 参数。
// 常用场景：跳过证书校验（测试环境）、指定客户端证书（mTLS）。
//
// 参数：
//   - cfg: *tls.Config，为 nil 时使用系统默认配置。
//
// 示例（跳过证书校验，仅限测试）：
//
//	.TLS(&tls.Config{InsecureSkipVerify: true})
func (b *ClientBuilder) TLS(cfg *tls.Config) *ClientBuilder {
	b.tlsConfig = cfg
	return b
}

// Proxy 设置单个 HTTP 代理服务器地址。
func (b *ClientBuilder) Proxy(proxyURL string) *ClientBuilder {
	b.proxyURL = proxyURL
	b.proxyPool = nil // 与代理池互斥
	return b
}

// ProxyPool 设置代理池，支持多代理自动轮换（新增）
func (b *ClientBuilder) ProxyPool(pool *ProxyPool) *ClientBuilder {
	b.proxyPool = pool
	b.proxyURL = "" // 与单个代理互斥
	return b
}

// Compression 启用请求/响应的 gzip/deflate 压缩。
// 启用后请求头自动添加 Accept-Encoding: deflate, gzip，
// ReadBody 会自动解压 gzip 响应体。
func (b *ClientBuilder) Compression() *ClientBuilder {
	b.compressed = true
	return b
}

// CheckRedirect 自定义 HTTP 重定向行为。
//
// 参数：
//   - fn: 重定向回调函数，签名与标准库一致。
//     返回 nil 表示跟随重定向；
//     返回 http.ErrUseLastResponse 表示停止重定向并使用当前响应；
//     返回其他错误表示终止请求。
//     传入 nil 恢复默认行为（最多跟随 10 次重定向）。
func (b *ClientBuilder) CheckRedirect(fn func(*http.Request, []*http.Request) error) *ClientBuilder {
	b.checkRedirect = fn
	return b
}

// ─── 鉴权 ──────────────────────────────────────────────

// BearerToken 设置动态 Bearer Token 鉴权。
// 每次请求发送前调用 fn 获取最新 token，适合 JWT 定期刷新的场景。
// 与 BasicAuth 互斥，后调用的生效。
//
// 参数：
//   - fn: 返回 token 字符串的函数，返回空字符串时不注入鉴权头。
//
// 示例：
//
//	.BearerToken(func() string { return tokenStore.Get() })
func (b *ClientBuilder) BearerToken(fn func() string) *ClientBuilder {
	b.bearerTokenFn = fn
	b.basicUsername = ""
	b.basicPassword = ""
	return b
}

// BasicAuth 设置 HTTP Basic Auth 鉴权（用户名 + 密码）。
// 与 BearerToken 互斥，后调用的生效。
//
// 参数：
//   - username: 用户名
//   - password: 密码（会以 Base64 编码写入 Authorization 头）
func (b *ClientBuilder) BasicAuth(username, password string) *ClientBuilder {
	b.basicUsername = username
	b.basicPassword = password
	b.bearerTokenFn = nil
	return b
}

// Sign 设置请求签名钩子，在鉴权头写入之后、请求发送之前执行。
// 适合需要对完整请求内容（含 header）计算签名的场景，如 AWS Signature、HMAC-SHA256。
//
// 参数：
//   - fn: 接收 *http.Request 并向其添加签名相关 header，返回 error 时请求终止。
//
// 示例：
//
//	.Sign(func(req *http.Request) error {
//	    sig := hmac.Sign(req)
//	    req.Header.Set("X-Signature", sig)
//	    return nil
//	})
func (b *ClientBuilder) Sign(fn func(*http.Request) error) *ClientBuilder {
	b.signFn = fn
	return b
}

// ─── 可观测 ────────────────────────────────────────────

// Logger 设置请求日志函数。
// 每次请求完成后（无论成功或失败）自动打印方法、URL、状态码和耗时。
//
// 参数：
//   - fn: 日志函数，签名与 fmt.Printf / log.Printf 兼容。
//     传入 nil 禁用日志。
//
// 示例：
//
//	.Logger(log.Printf)
//	.Logger(func(format string, args ...any) { slog.Info(fmt.Sprintf(format, args...)) })
func (b *ClientBuilder) Logger(fn func(format string, args ...any)) *ClientBuilder {
	b.logger = fn
	return b
}

// Metrics 注入指标收集器，自动统计请求总数、错误数和平均耗时。
//
// 参数：
//   - m: *Metrics 实例，调用方持有引用，随时可读取统计数据。
//
// 示例：
//
//	m := &Metrics{}
//	client, _ := NewClient("...").Metrics(m).Build()
//	// 任意时刻读取
//	fmt.Println(m.Summary())
func (b *ClientBuilder) Metrics(m *Metrics) *ClientBuilder {
	b.metrics = m
	return b
}

// ─── 可靠性 ────────────────────────────────────────────

// CircuitBreaker 注入熔断器。
// 连续失败超过阈值后自动开启熔断，拒绝后续请求（返回 ErrCircuitOpen），
// 经过冷却期后进入半开状态探测，连续成功后自动恢复。
//
// 参数：
//   - cb: *CircuitBreaker 实例，通过 NewCircuitBreaker() 创建。
//     创建后可修改 MaxFailures / OpenTimeout / HalfOpenSuccesses 字段自定义阈值。
func (b *ClientBuilder) CircuitBreaker(cb *CircuitBreaker) *ClientBuilder {
	b.circuitBreaker = cb
	return b
}

// RateLimiter 注入令牌桶限速器。
// 请求在获取令牌之前会阻塞等待，context 取消时立即返回错误。
//
// 参数：
//   - rl: *RateLimiter 实例，通过 NewRateLimiter(rps, burst) 创建。
func (b *ClientBuilder) RateLimiter(rl *RateLimiter) *ClientBuilder {
	b.rateLimiter = rl
	return b
}

// ResponseCache 注入内存响应缓存。
// 仅缓存 GET 请求且状态码为 200 的响应，其他方法和状态码直接透传。
// 缓存命中时直接返回缓存内容，不经过熔断、限速、重试等环节。
//
// 参数：
//   - rc: *ResponseCache 实例，通过 NewResponseCache(ttl) 创建。
func (b *ClientBuilder) ResponseCache(rc *ResponseCache) *ClientBuilder {
	b.responseCache = rc
	return b
}

// Retry 设置自动重试策略，直接复用同包 retry.go 的 Option。
// 未调用此方法时请求只执行一次，不重试。
//
// 参数：
//   - opts: 一个或多个 retry.go 中定义的 Option：
//     WithMaxRetries(n)         —— 最大重试次数（默认 5）
//     WithRetryDelay(d)         —— 首次重试前等待时间（默认 1s）
//     WithDelayMultiplier(f)    —— 每次等待时间的增长倍数（默认 1.0，不增长）
//     WithMaxTime(d)            —— 重试的最大总时长（默认 10min）
//
// 重试判断规则：
//   - 网络错误（err != nil）→ 触发重试
//   - HTTP 5xx（非 501）    → 触发重试
//   - 其余状态码（4xx/2xx）→ 立即终止，不重试
//
// 示例：
//
//	.Retry(WithMaxRetries(3), WithRetryDelay(500*time.Millisecond), WithDelayMultiplier(2.0))
func (b *ClientBuilder) Retry(opts ...Option) *ClientBuilder {
	b.retryOpts = opts
	return b
}

// Build 校验配置并构建 HTTPClient。
// 校验失败（如 baseURL 格式非法、代理地址无法解析）时返回 error。
// 构建成功的 HTTPClient 可安全地被多个 goroutine 并发使用。
func (b *ClientBuilder) Build() (*HTTPClient, error) {
	if b.baseURL != "" {
		if _, err := url.ParseRequestURI(b.baseURL); err != nil {
			return nil, fmt.Errorf("invalid baseURL %q: %w", b.baseURL, err)
		}
	}

	transport := &http.Transport{
		DialContext: (&net.Dialer{
			Timeout:   30 * time.Second,
			KeepAlive: b.keepAlive,
		}).DialContext,
		TLSHandshakeTimeout:   b.handshakeTimeout,
		ResponseHeaderTimeout: b.responseTimeout,
		DisableCompression:    !b.compressed,
		MaxIdleConns:          b.maxIdleConns,
		MaxConnsPerHost:       b.maxConnsPerHost,
		IdleConnTimeout:       b.idleConnTimeout,
		ForceAttemptHTTP2:     true,
	}

	if b.tlsConfig != nil {
		transport.TLSClientConfig = b.tlsConfig
	}

	// 代理池优先于单个代理
	if b.proxyPool != nil {
		transport.Proxy = func(_ *http.Request) (*url.URL, error) {
			return b.proxyPool.Get()
		}
	} else if b.proxyURL != "" {
		raw := b.proxyURL
		if !strings.HasPrefix(raw, "http") {
			raw = "http://" + raw
		}
		proxy, err := url.Parse(strings.TrimSpace(raw))
		if err != nil {
			return nil, fmt.Errorf("invalid proxy URL: %w", err)
		}
		transport.Proxy = http.ProxyURL(proxy)
	}

	return &HTTPClient{
		builder: b,
		raw: &http.Client{
			Timeout:       b.timeout,
			Transport:     transport,
			CheckRedirect: b.checkRedirect,
		},
	}, nil
}

// ═══════════════════════════════════════════════════════
// HTTPClient
// ═══════════════════════════════════════════════════════

// HTTPClient 生产级 HTTP 客户端，并发安全，应全局复用。
// 通过 NewClient(...).Build() 创建，不要直接实例化此结构体。
type HTTPClient struct {
	builder *ClientBuilder
	raw     *http.Client
}

// ═══════════════════════════════════════════════════════
// RequestBuilder —— 请求级配置
// ═══════════════════════════════════════════════════════

// requestConfig 存储单次请求的可选参数
type requestConfig struct {
	ctx          context.Context
	headers      map[string]string
	queryParams  map[string]string
	formData     url.Values
	file         *File
	expectStatus []int
}

// RequestBuilder 构建单次请求的可选参数，通过 R() 创建后链式调用。
// 可以预先构建基础配置后复用：
//
//	base := R().Headers(map[string]string{"X-Tenant": "abc"})
//	client.Get("/a", base.QueryParams(...))
//	client.Get("/b", base.QueryParams(...))
type RequestBuilder struct {
	cfg requestConfig
}

// R 创建空的 RequestBuilder，作为请求级配置的起点。
// 不需要额外参数时，HTTP 方法可不传此参数。
func R() *RequestBuilder {
	return &RequestBuilder{}
}

// Context 为本次请求注入 context，用于控制超时或取消。
// 注入的 context 优先于客户端的 Timeout 设置。
//
// 参数：
//   - ctx: 请求上下文，通常由 context.WithTimeout 或 context.WithCancel 创建。
//
// 示例：
//
//	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
//	defer cancel()
//	client.Get("/slow", R().Context(ctx))
func (r *RequestBuilder) Context(ctx context.Context) *RequestBuilder {
	r.cfg.ctx = ctx
	return r
}

// Headers 设置本次请求的额外 HTTP header。
// 会覆盖客户端的默认 header（如 Accept），但不影响鉴权 header（后注入）。
//
// 参数：
//   - h: key-value 形式的 header 映射，例如 map[string]string{"X-Request-ID": "abc"}。
func (r *RequestBuilder) Headers(h map[string]string) *RequestBuilder {
	r.cfg.headers = h
	return r
}

// QueryParams 追加本次请求的 URL query 参数。
// 会追加到 URL 末尾，与 path 中已有的 query string 合并。
//
// 参数：
//   - p: key-value 形式的参数映射，例如 map[string]string{"page": "1", "size": "20"}。
//     参数值会自动进行 URL 编码。
func (r *RequestBuilder) QueryParams(p map[string]string) *RequestBuilder {
	r.cfg.queryParams = p
	return r
}

// FormData 设置 application/x-www-form-urlencoded 表单数据。
// 设置后会覆盖请求体，并自动设置 Content-Type。
// 若同时设置了 File，则改用 multipart/form-data 格式。
//
// 参数：
//   - v: url.Values 类型的表单字段，支持同一字段多个值。
func (r *RequestBuilder) FormData(v url.Values) *RequestBuilder {
	r.cfg.formData = v
	return r
}

// File 设置文件上传，必须配合 FormData 使用（即使表单字段为空也需传 url.Values{}）。
// 设置后请求自动切换为 multipart/form-data 格式。
//
// 参数：
//   - f: *File 描述上传文件的元信息和内容来源（Content 字节或 Path 路径）。
//
// 示例：
//
//	client.UploadFile("/upload", &File{
//	    FieldName: "avatar",
//	    FileName:  "photo.png",
//	    Path:      "/tmp/photo.png",
//	}, url.Values{"desc": {"profile photo"}})
func (r *RequestBuilder) File(f *File) *RequestBuilder {
	r.cfg.file = f
	return r
}

// ExpectStatus 声明本次请求期望的 HTTP 状态码白名单。
// 若实际状态码不在列表中，请求自动返回 *HTTPError（含状态码和响应体）。
// 未调用此方法时任何状态码都不会触发错误（调用方自行判断）。
//
// 参数：
//   - codes: 一个或多个期望的状态码，例如 200、201。
//
// 示例：
//
//	R().ExpectStatus(200)           // 只接受 200
//	R().ExpectStatus(200, 201)      // 接受 200 或 201
func (r *RequestBuilder) ExpectStatus(codes ...int) *RequestBuilder {
	r.cfg.expectStatus = codes
	return r
}

// ═══════════════════════════════════════════════════════
// 核心执行（内部方法）
// ═══════════════════════════════════════════════════════

// buildURL 将相对路径拼接到 baseURL。
// 空 path 返回 baseURL；空 baseURL 返回原始 path；否则确保中间只有一个 "/"。
func (c *HTTPClient) buildURL(path string) string {
	if path == "" {
		return c.builder.baseURL
	}
	if c.builder.baseURL == "" {
		return path
	}
	return c.builder.baseURL + "/" + strings.TrimLeft(path, "/")
}

// do 构建 *http.Request 并委托给 execute 执行。
// 负责：URL 拼接、默认 header、query params、form data / 文件上传、鉴权、签名、状态码校验。
func (c *HTTPClient) do(method, path string, body io.Reader, contentType string, rb *RequestBuilder) (*http.Response, error) {
	var cfg requestConfig
	if rb != nil {
		cfg = rb.cfg
	}

	ctx := cfg.ctx
	if ctx == nil {
		ctx = context.Background()
	}

	req, err := http.NewRequestWithContext(ctx, method, c.buildURL(path), body)
	if err != nil {
		return nil, err
	}

	// 默认 header
	req.Header.Set("Accept", "*/*")
	if c.builder.compressed {
		req.Header.Set("Accept-Encoding", "deflate, gzip")
	}
	if contentType != "" {
		req.Header.Set("Content-Type", contentType)
	}
	// 调用方 header 可覆盖上面的默认值
	for k, v := range cfg.headers {
		req.Header.Set(k, v)
	}

	// query 参数追加到 URL
	if len(cfg.queryParams) > 0 {
		q := req.URL.Query()
		for k, v := range cfg.queryParams {
			q.Add(k, v)
		}
		req.URL.RawQuery = q.Encode()
	}

	// form data / 文件上传
	if cfg.formData != nil {
		if cfg.file != nil {
			// 有文件时使用 multipart/form-data
			if err = attachFile(req, cfg.formData, cfg.file); err != nil {
				return nil, err
			}
		} else {
			// 纯表单使用 application/x-www-form-urlencoded
			encoded := []byte(cfg.formData.Encode())
			req.Body = io.NopCloser(bytes.NewReader(encoded))
			req.ContentLength = int64(len(encoded))
			req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		}
	}

	// 鉴权（BearerToken 优先于 BasicAuth）
	if c.builder.bearerTokenFn != nil {
		if token := c.builder.bearerTokenFn(); token != "" {
			req.Header.Set("Authorization", "Bearer "+token)
		}
	} else if c.builder.basicUsername != "" {
		req.SetBasicAuth(c.builder.basicUsername, c.builder.basicPassword)
	}

	// 签名（在鉴权之后执行，可对完整 header 签名）
	if c.builder.signFn != nil {
		if err = c.builder.signFn(req); err != nil {
			return nil, fmt.Errorf("sign request: %w", err)
		}
	}

	resp, err := c.execute(req)
	if err != nil {
		return nil, err
	}

	// 状态码白名单校验
	if len(cfg.expectStatus) > 0 {
		match := false
		for _, code := range cfg.expectStatus {
			if resp.StatusCode == code {
				match = true
				break
			}
		}
		if !match {
			respBody, _ := io.ReadAll(resp.Body)
			resp.Body.Close()
			return nil, &HTTPError{StatusCode: resp.StatusCode, Status: resp.Status, Body: respBody}
		}
	}

	return resp, nil
}

// execute 按固定顺序执行所有可靠性能力，顺序不可更改：
//  1. 缓存命中（仅 GET）→ 直接返回，跳过后续所有步骤
//  2. 熔断检查          → 开启时返回 ErrCircuitOpen
//  3. 限速等待          → 令牌不足时阻塞，context 取消时返回错误
//  4. body 缓存         → 读取并缓存请求体字节，供重试时重放
//  5. 发送请求（含重试）→ 复用 retry.go 的 RetryWithContext
//  6. 记录日志
//  7. 更新指标
//  8. 更新熔断状态
//  9. 写入响应缓存（仅 GET 200）
func (c *HTTPClient) execute(req *http.Request) (*http.Response, error) {
	b := c.builder

	// ① 缓存命中（仅 GET）
	if b.responseCache != nil && req.Method == http.MethodGet {
		if cached, ok := b.responseCache.get(req.URL.String()); ok {
			return &http.Response{
				StatusCode: cached.status,
				Status:     http.StatusText(cached.status),
				Header:     cached.headers.Clone(),
				Body:       io.NopCloser(bytes.NewReader(cached.respBody)),
				Request:    req,
			}, nil
		}
	}

	// ② 熔断检查
	if b.circuitBreaker != nil && !b.circuitBreaker.Allow() {
		return nil, ErrCircuitOpen
	}

	// ③ 限速
	if b.rateLimiter != nil {
		if err := b.rateLimiter.Wait(req.Context()); err != nil {
			return nil, fmt.Errorf("rate limiter: %w", err)
		}
	}

	// ④ 缓存请求 body，供重试时重放（http.Request.Body 只能读一次）
	var bodyBytes []byte
	if req.Body != nil && req.Body != http.NoBody {
		var err error
		bodyBytes, err = io.ReadAll(req.Body)
		req.Body.Close()
		if err != nil {
			return nil, err
		}
	}

	// ⑤ 发送请求（含重试）
	//    operationFn 内部：
	//      - 网络错误 / 5xx（非 501）→ 返回普通 error，RetryWithContext 会重试
	//      - 4xx / 2xx / 501        → 返回 NonRetryable，RetryWithContext 立即终止
	var finalResp *http.Response
	start := time.Now()

	operationFn := func(args ...any) (any, error) {
		if bodyBytes != nil {
			req.Body = io.NopCloser(bytes.NewReader(bodyBytes))
		}
		resp, err := c.raw.Do(req)
		if err != nil {
			return nil, err // 网络错误，触发重试
		}
		if resp.StatusCode >= 500 && resp.StatusCode != http.StatusNotImplemented {
			io.Copy(io.Discard, resp.Body)
			resp.Body.Close()
			return nil, fmt.Errorf("server error: %s", resp.Status) // 5xx，触发重试
		}
		return resp, NonRetryable(nil) // 其余状态码，立即终止
	}

	successFn := func(data any) {
		finalResp = data.(*http.Response)
	}

	var execErr error
	if len(b.retryOpts) > 0 {
		execErr = RetryWithContext(req.Context(), operationFn, successFn, b.retryOpts...)
	} else {
		// 未配置重试：直接执行一次
		data, err := operationFn()
		if err != nil && !IsNonRetryable(err) {
			execErr = err
		} else if data != nil {
			finalResp = data.(*http.Response)
		}
	}

	elapsed := time.Since(start)

	// ⑥ 日志
	if b.logger != nil {
		ms := float64(elapsed.Milliseconds())
		if execErr != nil {
			b.logger("[HTTP] %s %s — error: %v (%.2fms)", req.Method, req.URL, execErr, ms)
		} else {
			b.logger("[HTTP] %s %s — %d (%.2fms)", req.Method, req.URL, finalResp.StatusCode, ms)
		}
	}

	// ⑦ 指标
	if b.metrics != nil {
		b.metrics.TotalRequests.Add(1)
		b.metrics.TotalDurationMs.Add(elapsed.Milliseconds())
		if execErr != nil || (finalResp != nil && finalResp.StatusCode >= 500) {
			b.metrics.ErrorRequests.Add(1)
		}
	}

	// ⑧ 熔断状态更新
	if b.circuitBreaker != nil {
		if execErr != nil || (finalResp != nil && finalResp.StatusCode >= 500) {
			b.circuitBreaker.RecordFailure()
		} else {
			b.circuitBreaker.RecordSuccess()
		}
	}

	if execErr != nil {
		return nil, execErr
	}

	// ⑨ 写入响应缓存（仅 GET 200，读取 body 后重新装填以供调用方读取）
	if b.responseCache != nil && req.Method == http.MethodGet && finalResp.StatusCode == http.StatusOK {
		body, readErr := io.ReadAll(finalResp.Body)
		finalResp.Body.Close()
		if readErr == nil {
			b.responseCache.set(req.URL.String(), &cacheEntry{
				respBody: body,
				headers:  finalResp.Header.Clone(),
				status:   finalResp.StatusCode,
			})
			finalResp.Body = io.NopCloser(bytes.NewReader(body))
		}
	}

	return finalResp, nil
}

// ═══════════════════════════════════════════════════════
// HTTP 方法
// ═══════════════════════════════════════════════════════

// Get 发送 HTTP GET 请求。
//
// 参数：
//   - path: 请求路径，拼接在 baseURL 之后。传完整 URL 时 baseURL 须为空。
//   - rb:   可选请求配置，通过 R() 构建，不需要时可省略。
func (c *HTTPClient) Get(path string, rb ...*RequestBuilder) (*http.Response, error) {
	return c.do(http.MethodGet, path, nil, "", first(rb))
}

// Post 发送 HTTP POST 请求。
//
// 参数：
//   - path:        请求路径
//   - body:        请求体，通常为 strings.NewReader / bytes.NewReader，为 nil 时不发送 body
//   - contentType: Content-Type，例如 "application/json"、"text/plain"
//   - rb:          可选请求配置
func (c *HTTPClient) Post(path string, body io.Reader, contentType string, rb ...*RequestBuilder) (*http.Response, error) {
	return c.do(http.MethodPost, path, body, contentType, first(rb))
}

// Put 发送 HTTP PUT 请求，参数同 Post。
func (c *HTTPClient) Put(path string, body io.Reader, contentType string, rb ...*RequestBuilder) (*http.Response, error) {
	return c.do(http.MethodPut, path, body, contentType, first(rb))
}

// Patch 发送 HTTP PATCH 请求，参数同 Post。
func (c *HTTPClient) Patch(path string, body io.Reader, contentType string, rb ...*RequestBuilder) (*http.Response, error) {
	return c.do(http.MethodPatch, path, body, contentType, first(rb))
}

// Delete 发送 HTTP DELETE 请求。
//
// 参数：
//   - path: 请求路径
//   - rb:   可选请求配置
func (c *HTTPClient) Delete(path string, rb ...*RequestBuilder) (*http.Response, error) {
	return c.do(http.MethodDelete, path, nil, "", first(rb))
}

// Head 发送 HTTP HEAD 请求，响应体为空，常用于检查资源是否存在。
//
// 参数：
//   - path: 请求路径
//   - rb:   可选请求配置
func (c *HTTPClient) Head(path string, rb ...*RequestBuilder) (*http.Response, error) {
	return c.do(http.MethodHead, path, nil, "", first(rb))
}

// PostJSON 将 data 序列化为 JSON 后以 POST 方式发送。
// 自动设置 Content-Type: application/json。
//
// 参数：
//   - path: 请求路径
//   - data: 任意可被 json.Marshal 序列化的值（struct、map 等）
//   - rb:   可选请求配置
func (c *HTTPClient) PostJSON(path string, data any, rb ...*RequestBuilder) (*http.Response, error) {
	b, err := json.Marshal(data)
	if err != nil {
		return nil, err
	}
	return c.Post(path, bytes.NewReader(b), "application/json", rb...)
}

// PutJSON 将 data 序列化为 JSON 后以 PUT 方式发送，参数同 PostJSON。
func (c *HTTPClient) PutJSON(path string, data any, rb ...*RequestBuilder) (*http.Response, error) {
	b, err := json.Marshal(data)
	if err != nil {
		return nil, err
	}
	return c.Put(path, bytes.NewReader(b), "application/json", rb...)
}

// PatchJSON 将 data 序列化为 JSON 后以 PATCH 方式发送，参数同 PostJSON。
func (c *HTTPClient) PatchJSON(path string, data any, rb ...*RequestBuilder) (*http.Response, error) {
	b, err := json.Marshal(data)
	if err != nil {
		return nil, err
	}
	return c.Patch(path, bytes.NewReader(b), "application/json", rb...)
}

// UploadFile 以 multipart/form-data 格式上传文件。
//
// 参数：
//   - path:        上传接口路径
//   - file:        *File 描述文件元信息和内容来源
//   - extraFields: 附带的表单字段，无额外字段时传 nil 或 url.Values{}
//   - rb:          可选请求配置
func (c *HTTPClient) UploadFile(path string, file *File, extraFields url.Values, rb ...*RequestBuilder) (*http.Response, error) {
	r := first(rb)
	if r == nil {
		r = R()
	}
	r.FormData(extraFields).File(file)
	return c.do(http.MethodPost, path, nil, "", r)
}

// first 取切片第一个元素，供可变参数的 HTTP 方法内部使用
func first(rbs []*RequestBuilder) *RequestBuilder {
	if len(rbs) > 0 {
		return rbs[0]
	}
	return nil
}

// ═══════════════════════════════════════════════════════
// 响应读取
// ═══════════════════════════════════════════════════════

// ReadBody 读取并关闭响应体，自动处理 gzip 解压。
// 调用后 resp.Body 已关闭，不可再次读取。
//
// 参数：
//   - resp: *http.Response，不可为 nil。
//
// 返回：
//   - []byte: 原始响应体内容（gzip 时为解压后内容）
func (c *HTTPClient) ReadBody(resp *http.Response) ([]byte, error) {
	defer resp.Body.Close()
	var reader io.Reader = resp.Body
	if strings.Contains(resp.Header.Get("Content-Encoding"), "gzip") {
		gr, err := gzip.NewReader(resp.Body)
		if err != nil {
			return nil, err
		}
		defer gr.Close()
		reader = gr
	}
	return io.ReadAll(reader)
}

// ReadBodyString 读取响应体并以字符串形式返回，内部调用 ReadBody。
//
// 参数：
//   - resp: *http.Response，不可为 nil。
func (c *HTTPClient) ReadBodyString(resp *http.Response) (string, error) {
	b, err := c.ReadBody(resp)
	if err != nil {
		return "", err
	}
	return string(b), nil
}

// ReadJSON 将响应体反序列化到 target 指向的结构体，内部调用 ReadBody。
// 调用后 resp.Body 已关闭。
//
// 参数：
//   - resp:   *http.Response，为 nil 时返回错误。
//   - target: 指向目标结构体的指针，例如 &User{} 或 &[]User{}。
//
// 示例：
//
//	var user User
//	err = client.ReadJSON(resp, &user)
func (c *HTTPClient) ReadJSON(resp *http.Response, target any) error {
	if resp == nil {
		return errors.New("response is nil")
	}
	b, err := c.ReadBody(resp)
	if err != nil {
		return err
	}
	return json.Unmarshal(b, target)
}

// ═══════════════════════════════════════════════════════
// 可靠性组件
// ═══════════════════════════════════════════════════════

// ─── 熔断器 ────────────────────────────────────────────

type circuitState int

const (
	stateClosed   circuitState = iota // 正常，所有请求放行
	stateOpen                         // 熔断，所有请求拒绝
	stateHalfOpen                     // 半开，放行探测请求
)

// CircuitBreaker 三态熔断器，状态流转：Closed → Open → HalfOpen → Closed。
//
// 创建后可直接修改公开字段调整阈值：
//
//	cb := NewCircuitBreaker()
//	cb.MaxFailures = 10          // 允许更多失败次数
//	cb.OpenTimeout = time.Minute // 熔断持续更久
type CircuitBreaker struct {
	mu          sync.Mutex
	state       circuitState
	failures    int
	successes   int
	lastFailure time.Time

	// MaxFailures 连续失败多少次后开启熔断，默认 5
	MaxFailures int
	// OpenTimeout 熔断持续时间，超过后进入半开状态，默认 30s
	OpenTimeout time.Duration
	// HalfOpenSuccesses 半开状态需要多少次连续成功才完全恢复，默认 2
	HalfOpenSuccesses int
}

// NewCircuitBreaker 创建使用默认阈值的熔断器。
// 默认值：MaxFailures=5，OpenTimeout=30s，HalfOpenSuccesses=2。
func NewCircuitBreaker() *CircuitBreaker {
	return &CircuitBreaker{
		MaxFailures:       5,
		OpenTimeout:       30 * time.Second,
		HalfOpenSuccesses: 2,
	}
}

// Allow 判断当前是否允许发送请求。
// 熔断开启时返回 false；半开状态或正常状态返回 true。
// 此方法由 HTTPClient 内部调用，通常不需要外部直接调用。
func (cb *CircuitBreaker) Allow() bool {
	cb.mu.Lock()
	defer cb.mu.Unlock()
	if cb.state == stateOpen {
		if time.Since(cb.lastFailure) > cb.OpenTimeout {
			cb.state = stateHalfOpen
			cb.successes = 0
			return true
		}
		return false
	}
	return true
}

// RecordSuccess 记录一次成功，由 HTTPClient 内部在请求成功后调用。
// 半开状态下连续成功达到阈值时切换回 Closed。
func (cb *CircuitBreaker) RecordSuccess() {
	cb.mu.Lock()
	defer cb.mu.Unlock()
	cb.failures = 0
	if cb.state == stateHalfOpen {
		cb.successes++
		if cb.successes >= cb.HalfOpenSuccesses {
			cb.state = stateClosed
		}
	}
}

// RecordFailure 记录一次失败，由 HTTPClient 内部在请求失败后调用。
// 连续失败达到阈值时切换到 Open 状态。
func (cb *CircuitBreaker) RecordFailure() {
	cb.mu.Lock()
	defer cb.mu.Unlock()
	cb.failures++
	cb.lastFailure = time.Now()
	if cb.failures >= cb.MaxFailures {
		cb.state = stateOpen
	}
}

// State 返回熔断器当前状态的字符串描述："closed"、"open" 或 "half-open"。
// 可用于监控上报或调试日志。
func (cb *CircuitBreaker) State() string {
	cb.mu.Lock()
	defer cb.mu.Unlock()
	switch cb.state {
	case stateClosed:
		return "closed"
	case stateOpen:
		return "open"
	case stateHalfOpen:
		return "half-open"
	}
	return "unknown"
}

// ─── 限速器 ────────────────────────────────────────────

// RateLimiter 基于令牌桶算法的限速器，并发安全。
// 令牌按照 rps 速率持续补充，桶的容量由 burst 决定。
type RateLimiter struct {
	mu         sync.Mutex
	tokens     float64   // 当前令牌数
	maxTokens  float64   // 桶的最大容量（= burst）
	refillRate float64   // 每秒补充的令牌数（= rps）
	lastRefill time.Time // 上次补充时间，用于计算增量
}

// NewRateLimiter 创建令牌桶限速器。
//
// 参数：
//   - rps:   每秒允许的稳定请求数，例如 100.0 表示每秒 100 个请求
//   - burst: 瞬时峰值（桶的容量），例如 20 表示瞬间最多 20 个并发请求
//
// 示例：
//
//	rl := NewRateLimiter(100, 20) // 稳定 100 rps，瞬时最多 20
func NewRateLimiter(rps float64, burst int) *RateLimiter {
	return &RateLimiter{
		tokens:     float64(burst),
		maxTokens:  float64(burst),
		refillRate: rps,
		lastRefill: time.Now(),
	}
}

// Wait 阻塞等待直到获取到令牌，context 取消时立即返回错误。
// 此方法由 HTTPClient 内部调用，通常不需要外部直接调用。
func (r *RateLimiter) Wait(ctx context.Context) error {
	for {
		r.mu.Lock()
		now := time.Now()
		r.tokens = minFloat(r.maxTokens, r.tokens+now.Sub(r.lastRefill).Seconds()*r.refillRate)
		r.lastRefill = now
		if r.tokens >= 1 {
			r.tokens--
			r.mu.Unlock()
			return nil
		}
		wait := time.Duration((1-r.tokens)/r.refillRate*1000) * time.Millisecond
		r.mu.Unlock()
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(wait):
		}
	}
}

func minFloat(a, b float64) float64 {
	if a < b {
		return a
	}
	return b
}

// ─── 响应缓存 ──────────────────────────────────────────

// cacheEntry 单条缓存记录
type cacheEntry struct {
	respBody  []byte
	headers   http.Header
	status    int
	expiresAt time.Time
}

// ResponseCache 基于内存的 TTL 缓存，仅缓存 GET 请求且状态码为 200 的响应。
// 后台 goroutine 按 TTL 间隔定期清理过期条目，读写均并发安全。
type ResponseCache struct {
	mu      sync.RWMutex
	entries map[string]*cacheEntry
	ttl     time.Duration
}

// NewResponseCache 创建响应缓存。
//
// 参数：
//   - ttl: 缓存有效期，例如 5*time.Minute。过期后下次请求会重新发送并刷新缓存。
func NewResponseCache(ttl time.Duration) *ResponseCache {
	c := &ResponseCache{entries: make(map[string]*cacheEntry), ttl: ttl}
	go c.evictLoop()
	return c
}

// evictLoop 后台定期清理过期条目，随 ResponseCache 生命周期运行
func (c *ResponseCache) evictLoop() {
	t := time.NewTicker(c.ttl)
	defer t.Stop()
	for range t.C {
		now := time.Now()
		c.mu.Lock()
		for k, e := range c.entries {
			if now.After(e.expiresAt) {
				delete(c.entries, k)
			}
		}
		c.mu.Unlock()
	}
}

func (c *ResponseCache) get(key string) (*cacheEntry, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	e, ok := c.entries[key]
	if !ok || time.Now().After(e.expiresAt) {
		return nil, false
	}
	return e, true
}

func (c *ResponseCache) set(key string, e *cacheEntry) {
	c.mu.Lock()
	defer c.mu.Unlock()
	e.expiresAt = time.Now().Add(c.ttl)
	c.entries[key] = e
}

// ─── 指标 ──────────────────────────────────────────────

// Metrics 并发安全的请求指标收集器，通过原子操作更新，无锁竞争。
//
// 使用方式：
//
//	m := &Metrics{}
//	client, _ := NewClient("...").Metrics(m).Build()
//	// 随时读取
//	fmt.Println(m.Summary())
type Metrics struct {
	TotalRequests   atomic.Int64 // 请求总数
	ErrorRequests   atomic.Int64 // 失败请求数（网络错误或 5xx）
	TotalDurationMs atomic.Int64 // 所有请求的累计耗时（毫秒）
}

// Summary 返回格式化的指标摘要字符串，格式：total=N errors=N avg_ms=N
func (m *Metrics) Summary() string {
	total := m.TotalRequests.Load()
	avg := int64(0)
	if total > 0 {
		avg = m.TotalDurationMs.Load() / total
	}
	return fmt.Sprintf("total=%d errors=%d avg_ms=%d", total, m.ErrorRequests.Load(), avg)
}

// ═══════════════════════════════════════════════════════
// 文件上传辅助
// ═══════════════════════════════════════════════════════

// File 描述一个待上传的文件，由 Content 或 Path 二选一提供内容来源。
// Content 非 nil 时优先使用，Path 用于从磁盘流式读取大文件。
type File struct {
	FieldName string // multipart 表单字段名，例如 "avatar"
	FileName  string // 文件名，例如 "photo.png"（显示在服务端）
	Content   []byte // 直接提供文件字节内容（优先于 Path）
	Path      string // 从磁盘读取的文件路径，例如 "/tmp/upload.csv"
}

// attachFile 将文件和表单字段写入 multipart 请求体，并设置对应的 Content-Type
func attachFile(req *http.Request, values url.Values, f *File) error {
	buf := &bytes.Buffer{}
	w := multipart.NewWriter(buf)
	for key, vals := range values {
		for _, val := range vals {
			if err := w.WriteField(key, val); err != nil {
				return err
			}
		}
	}
	part, err := w.CreateFormFile(f.FieldName, f.FileName)
	if err != nil {
		return err
	}
	if f.Content != nil {
		if _, err = part.Write(f.Content); err != nil {
			return err
		}
	} else if f.Path != "" {
		file, err := os.Open(f.Path)
		if err != nil {
			return err
		}
		defer file.Close()
		if _, err = io.Copy(part, file); err != nil {
			return err
		}
	}
	if err = w.Close(); err != nil {
		return err
	}
	req.Body = io.NopCloser(buf)
	req.ContentLength = int64(buf.Len())
	req.Header.Set("Content-Type", w.FormDataContentType())
	return nil
}

// ═══════════════════════════════════════════════════════
// 便捷独立函数（共享单例客户端）
// ═══════════════════════════════════════════════════════
//
// 这些函数使用一个进程级共享的 HTTPClient 单例，无需手动创建客户端。
// 适合简单脚本或单次请求场景，生产服务中建议使用 NewClient 创建专属客户端
// 以便独立配置超时、重试、鉴权等参数。

var (
	sharedOnce   sync.Once
	sharedClient *HTTPClient
)

// getShared 返回共享单例客户端，首次调用时懒初始化
func getShared() *HTTPClient {
	sharedOnce.Do(func() {
		sharedClient, _ = NewClient("").Build()
	})
	return sharedClient
}

// HttpGet 使用共享客户端发送 GET 请求。
//
// 参数：
//   - rawURL: 完整请求 URL，例如 "https://api.example.com/users"
//   - rb:     可选请求配置
func HttpGet(rawURL string, rb ...*RequestBuilder) (*http.Response, error) {
	return getShared().Get(rawURL, rb...)
}

// HttpPost 使用共享客户端发送 POST 请求。
//
// 参数：
//   - rawURL:      完整请求 URL
//   - body:        请求体
//   - contentType: Content-Type
//   - rb:          可选请求配置
func HttpPost(rawURL string, body io.Reader, contentType string, rb ...*RequestBuilder) (*http.Response, error) {
	return getShared().Post(rawURL, body, contentType, rb...)
}

// HttpPostJSON 使用共享客户端将 data 序列化为 JSON 后发送 POST 请求。
//
// 参数：
//   - rawURL: 完整请求 URL
//   - data:   可被 json.Marshal 序列化的值
//   - rb:     可选请求配置
func HttpPostJSON(rawURL string, data any, rb ...*RequestBuilder) (*http.Response, error) {
	return getShared().PostJSON(rawURL, data, rb...)
}

// HttpPut 使用共享客户端发送 PUT 请求，参数同 HttpPost。
func HttpPut(rawURL string, body io.Reader, contentType string, rb ...*RequestBuilder) (*http.Response, error) {
	return getShared().Put(rawURL, body, contentType, rb...)
}

// HttpPatch 使用共享客户端发送 PATCH 请求，参数同 HttpPost。
func HttpPatch(rawURL string, body io.Reader, contentType string, rb ...*RequestBuilder) (*http.Response, error) {
	return getShared().Patch(rawURL, body, contentType, rb...)
}

// HttpDelete 使用共享客户端发送 DELETE 请求。
//
// 参数：
//   - rawURL: 完整请求 URL
//   - rb:     可选请求配置
func HttpDelete(rawURL string, rb ...*RequestBuilder) (*http.Response, error) {
	return getShared().Delete(rawURL, rb...)
}

// ParseHttpResponse 将响应体反序列化到 obj（兼容旧接口，新代码推荐使用 client.ReadJSON）。
//
// 参数：
//   - resp: *http.Response，为 nil 时返回错误
//   - obj:  指向目标结构体的指针
func ParseHttpResponse(resp *http.Response, obj any) error {
	if resp == nil {
		return errors.New("response is nil")
	}
	defer resp.Body.Close()
	return json.NewDecoder(resp.Body).Decode(obj)
}

// ═══════════════════════════════════════════════════════
// 工具函数
// ═══════════════════════════════════════════════════════

// ConvertMapToQueryString 将 map 转换为按 key 字典序排序、经过 URL 编码的 query string。
// 结果可直接追加到 URL "?" 之后，特殊字符已编码。
//
// 参数：
//   - param: 待转换的参数 map，value 支持任意类型（以 fmt.Sprintf("%v") 格式化）
//
// 示例：
//
//	ConvertMapToQueryString(map[string]any{"b": 2, "a": "hello world"})
//	// 返回 "a=hello+world&b=2"
func ConvertMapToQueryString(param map[string]any) string {
	if len(param) == 0 {
		return ""
	}
	keys := make([]string, 0, len(param))
	for k := range param {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	var sb strings.Builder
	for i, k := range keys {
		sb.WriteString(url.QueryEscape(k))
		sb.WriteByte('=')
		sb.WriteString(url.QueryEscape(fmt.Sprintf("%v", param[k])))
		if i < len(keys)-1 {
			sb.WriteByte('&')
		}
	}
	return sb.String()
}

// StructToURLValues 将结构体转换为 url.Values，只处理有 json tag 的导出字段。
// 内部通过 JSON 序列化再反序列化实现，嵌套结构体的子字段会被展平为 JSON 字符串。
//
// 参数：
//   - v: 任意结构体值（或指针），未导出字段和无 json tag 的字段会被忽略。
//
// 返回：
//   - url.Values: 每个字段对应一个 key，value 为 fmt.Sprintf("%v") 格式化的字符串。
//
// 示例：
//
//	type Query struct {
//	    Page int    `json:"page"`
//	    Size int    `json:"size"`
//	}
//	values, err := StructToURLValues(Query{Page: 1, Size: 20})
//	// values = url.Values{"page": ["1"], "size": ["20"]}
func StructToURLValues(v any) (url.Values, error) {
	b, err := json.Marshal(v)
	if err != nil {
		return nil, err
	}
	var m map[string]any
	if err = json.Unmarshal(b, &m); err != nil {
		return nil, err
	}
	result := make(url.Values, len(m))
	for k, val := range m {
		result.Set(k, fmt.Sprintf("%v", val))
	}
	return result, nil
}
