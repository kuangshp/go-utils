package k

import "testing"

func BenchmarkParseDateRange(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_, _, _ = ParseDateRange("2026-01-01", "2026-01-31")
	}
}

func BenchmarkDateTimeStrToTime(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_, _ = DateTimeStrToTime("2026-01-01 12:00:00")
	}
}

func BenchmarkCheckDateTime(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_ = CheckDateTime("2026-01-01 12:00:00")
	}
}
