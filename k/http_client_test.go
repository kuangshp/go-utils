package k

import (
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

// 测试基础客户端构建
func TestClient_Build(t *testing.T) {
	_, err := NewClient("https://example.com").
		Timeout(5 * time.Second).
		Build()
	if err != nil {
		t.Fatalf("build client failed: %v", err)
	}
}

// 测试单个代理配置
func TestClient_Proxy(t *testing.T) {
	client, err := NewClient("https://example.com").
		Proxy("http://127.0.0.1:7890").
		Build()
	if err != nil {
		t.Fatalf("build client with proxy failed: %v", err)
	}

	tr, ok := client.raw.Transport.(*http.Transport)
	if !ok {
		t.Fatal("transport is not *http.Transport")
	}
	if tr.Proxy == nil {
		t.Error("proxy function not set")
	}
}

// 测试代理池构建与基本功能
func TestProxyPool_Basic(t *testing.T) {
	pool := NewProxyPool(
		[]string{
			"http://127.0.0.1:7890",
			"http://127.0.0.1:7891",
		},
		ProxyStrategyRoundRobin,
	)

	// 第一次获取
	u1, err := pool.Get()
	if err != nil {
		t.Fatal(err)
	}
	if u1 == nil {
		t.Fatal("expected proxy url")
	}

	// 第二次获取（轮询）
	u2, err := pool.Get()
	if err != nil {
		t.Fatal(err)
	}
	if u2.String() == u1.String() {
		t.Error("round robin not working, got same proxy")
	}
}

// 测试代理池随机策略
func TestProxyPool_Random(t *testing.T) {
	pool := NewProxyPool(
		[]string{
			"http://127.0.0.1:7890",
			"http://127.0.0.1:7891",
			"http://127.0.0.1:7892",
		},
		ProxyStrategyRandom,
	)

	seen := make(map[string]bool)
	for i := 0; i < 10; i++ {
		u, err := pool.Get()
		if err != nil {
			t.Fatal(err)
		}
		seen[u.String()] = true
	}
	if len(seen) < 2 {
		t.Error("random strategy seems not effective")
	}
}

// 测试代理上下线
func TestProxyPool_MarkUpDown(t *testing.T) {
	pool := NewProxyPool(
		[]string{"http://127.0.0.1:7890"},
		ProxyStrategyRoundRobin,
	)

	// 下线
	pool.MarkDown("http://127.0.0.1:7890")
	_, err := pool.Get()
	if err == nil {
		t.Error("expected no alive proxy error")
	}

	// 上线
	pool.MarkUp("http://127.0.0.1:7890")
	u, err := pool.Get()
	if err != nil {
		t.Fatal(err)
	}
	if u == nil {
		t.Error("proxy should be back online")
	}
}

// 测试客户端使用代理池
func TestClient_ProxyPool(t *testing.T) {
	pool := NewProxyPool(
		[]string{
			"http://127.0.0.1:7890",
			"http://127.0.0.1:7891",
		},
		ProxyStrategyRoundRobin,
	)

	client, err := NewClient("https://example.com").
		ProxyPool(pool).
		Build()
	if err != nil {
		t.Fatalf("build with proxy pool failed: %v", err)
	}

	tr, ok := client.raw.Transport.(*http.Transport)
	if !ok {
		t.Fatal("invalid transport")
	}
	if tr.Proxy == nil {
		t.Error("proxy func not set for proxy pool")
	}

	// 调用一次代理函数，确保不崩溃
	req, _ := http.NewRequest("GET", "/", nil)
	_, err = tr.Proxy(req)
	if err != nil {
		t.Fatalf("proxy func failed: %v", err)
	}
}

// 测试 GET 请求
func TestClient_Get(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
		_, _ = w.Write([]byte(`{"ok":true}`))
	}))
	defer server.Close()

	client, _ := NewClient(server.URL).Build()

	resp, err := client.Get("/", R().ExpectStatus(200))
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		t.Errorf("status code=%d", resp.StatusCode)
	}
	readAll, _ := io.ReadAll(resp.Body)
	fmt.Println("测试返回数据", string(readAll))
}

// 测试 PostJSON
func TestClient_PostJSON(t *testing.T) {
	type Req struct {
		Name string `json:"name"`
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Content-Type") != "application/json" {
			t.Error("content type not json")
		}
		w.WriteHeader(200)
	}))
	defer server.Close()

	client, _ := NewClient(server.URL).Build()
	_, err := client.PostJSON("/", Req{Name: "test"})
	if err != nil {
		t.Fatal(err)
	}
}

// 测试熔断器
func TestCircuitBreaker(t *testing.T) {
	cb := NewCircuitBreaker()
	cb.MaxFailures = 2

	// 连续失败
	cb.RecordFailure()
	cb.RecordFailure()
	if !cb.Allow() {
		// 刚失败两次还没到 OpenTimeout，应该是 open
		t.Log("circuit is open as expected")
	} else {
		t.Error("circuit should be open")
	}
}

// 测试响应缓存
func TestResponseCache(t *testing.T) {
	cache := NewResponseCache(1 * time.Minute)
	_, _ = NewClient("").
		ResponseCache(cache).
		Build()

	// 模拟缓存命中逻辑（内部逻辑验证）
	cache.set("https://example.com", &cacheEntry{
		respBody: []byte("cached"),
		status:   200,
	})

	entry, ok := cache.get("https://example.com")
	if !ok {
		t.Fatal("cache not found")
	}
	if string(entry.respBody) != "cached" {
		t.Error("cache content mismatch")
	}
}
