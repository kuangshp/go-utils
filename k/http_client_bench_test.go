package k

import "testing"

func BenchmarkProxyPoolGet_RoundRobin(b *testing.B) {
	pool := NewProxyPool([]string{
		"http://127.0.0.1:8001",
		"http://127.0.0.1:8002",
		"http://127.0.0.1:8003",
	}, ProxyStrategyRoundRobin)
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if _, err := pool.Get(); err != nil {
			b.Fatal(err)
		}
	}
}
