package k

import (
	"testing"
	"time"
)

func TestCheckDateTime(t *testing.T) {
	if !CheckDateTime("2024-12-20") {
		t.Fatal("期望日期格式通过")
	}
	if !CheckDateTime("2024-12-20 12:12:12") {
		t.Fatal("期望日期时间格式通过")
	}
	if CheckDateTime("2024/12/20") {
		t.Fatal("非法分隔符不应通过")
	}
}

func TestDateStrToTime(t *testing.T) {
	got, err := DateStrToTime("2024-12-20")
	if err != nil {
		t.Fatalf("DateStrToTime: %v", err)
	}
	if got.Year() != 2024 || got.Month() != 12 || got.Day() != 20 {
		t.Fatalf("got %v", got)
	}
	empty, err := DateStrToTime("  ")
	if err != nil || !empty.IsZero() {
		t.Fatalf("空串应返回零值, got %v err %v", empty, err)
	}
}

func TestDateTimeStrToTime(t *testing.T) {
	cases := []struct {
		name string
		in   string
		mode int
		want func(time.Time) bool
	}{
		{"empty", "", DateTimeModeDefault, func(tt time.Time) bool { return tt.IsZero() }},
		{"date only default", "2026-06-02", DateTimeModeDefault, func(tt time.Time) bool {
			return tt.Year() == 2026 && tt.Month() == 6 && tt.Day() == 2
		}},
		{"date start", "2026-06-02", DateTimeModeStart, func(tt time.Time) bool {
			return tt.Hour() == 0 && tt.Minute() == 0 && tt.Second() == 0
		}},
		{"date end", "2026-06-02", DateTimeModeEnd, func(tt time.Time) bool {
			return tt.Hour() == 23 && tt.Minute() == 59 && tt.Second() == 59
		}},
		{"datetime", "2026-06-02 15:04:05", DateTimeModeDefault, func(tt time.Time) bool {
			return tt.Hour() == 15 && tt.Minute() == 4 && tt.Second() == 5
		}},
		{"rfc3339", "2026-06-02T15:04:05+08:00", DateTimeModeDefault, func(tt time.Time) bool {
			return !tt.IsZero()
		}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			var (
				got time.Time
				err error
			)
			if tc.mode == DateTimeModeDefault && tc.name != "date start" && tc.name != "date end" {
				got, err = DateTimeStrToTime(tc.in)
			} else {
				got, err = DateTimeStrToTime(tc.in, tc.mode)
			}
			if err != nil {
				t.Fatalf("err=%v", err)
			}
			if !tc.want(got) {
				t.Fatalf("unexpected %v", got)
			}
		})
	}

	sec := time.Now().Unix()
	got, err := DateTimeStrToTime("1700000000")
	if err != nil || got.Unix() != 1700000000 {
		t.Fatalf("10位时间戳: got %v err %v", got, err)
	}
	_ = sec

	milli := int64(1700000000123)
	got, err = DateTimeStrToTime("1700000000123")
	if err != nil || got.UnixMilli() != milli {
		t.Fatalf("13位时间戳: got %v err %v", got, err)
	}

	if _, err := DateTimeStrToTime("not-a-time"); err == nil {
		t.Fatal("非法格式应报错")
	}
}

func TestParseDateRange(t *testing.T) {
	start, end, err := ParseDateRange("2026-01-01", "2026-01-31")
	if err != nil {
		t.Fatalf("ParseDateRange: %v", err)
	}
	if start.Hour() != 0 || end.Hour() != 23 {
		t.Fatalf("起止应补齐, start=%v end=%v", start, end)
	}

	_, _, err = ParseDateRange("2026-02-01", "2026-01-01")
	if err == nil {
		t.Fatal("开始大于结束应报错")
	}

	s, e, err := ParseDateRange("", "")
	if err != nil || !s.IsZero() || !e.IsZero() {
		t.Fatalf("空范围应成功且为零值: %v %v %v", s, e, err)
	}
}
