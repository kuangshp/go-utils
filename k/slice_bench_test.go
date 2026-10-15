package k

import "testing"

func BenchmarkMap_1k(b *testing.B) {
	in := make([]int, 1000)
	for i := range in {
		in[i] = i
	}
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = Map(in, func(item int, index int) int { return item * 2 })
	}
}

func BenchmarkFilter_1k(b *testing.B) {
	in := make([]int, 1000)
	for i := range in {
		in[i] = i
	}
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = Filter(in, func(item int, index int) bool { return item%2 == 0 })
	}
}

func BenchmarkDistinct_1k(b *testing.B) {
	in := make([]int, 1000)
	for i := range in {
		in[i] = i % 100
	}
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = Distinct(in)
	}
}

func BenchmarkDifference_1k(b *testing.B) {
	left := make([]int, 1000)
	right := make([]int, 500)
	for i := range left {
		left[i] = i
	}
	for i := range right {
		right[i] = i * 2
	}
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = Difference(left, right)
	}
}
