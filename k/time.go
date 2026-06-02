package k

import (
	"errors"
	"regexp"
	"strconv"
	"strings"
	"time"
)

// DateToTimesInt10 时间转时间戳
func DateToTimesInt10(date time.Time) int64 {
	return date.Unix()
}

// DateToTimesInt13 时间转时间戳
func DateToTimesInt13(date time.Time) int64 {
	return date.UnixNano() / 1e6
}

// CheckDateTime 判断时间格式
func CheckDateTime(timeStr string) bool {
	// 匹配 YYYY-MM-DD
	dateRegex := regexp.MustCompile(`^\d{4}-\d{2}-\d{2}$`)
	// 匹配 YYYY-MM-DD HH:MM:SS
	dateTimeRegex := regexp.MustCompile(`^\d{4}-\d{2}-\d{2} \d{2}:\d{2}:\d{2}$`)
	return dateRegex.MatchString(timeStr) || dateTimeRegex.MatchString(timeStr)
}

// DateStrToTime 简单的时间判断
func DateStrToTime(dateTimeStr string) (time.Time, error) {
	// 空字符串直接返回零值
	if strings.TrimSpace(dateTimeStr) == "" {
		return time.Time{}, nil
	}
	return time.ParseInLocation("2006-01-02", dateTimeStr, time.Local)
}

const (
	DateTimeModeDefault = 0
	DateTimeModeStart   = 1 // 补 00:00:00
	DateTimeModeEnd     = 2 // 补 23:59:59
)

// DateTimeStrToTime 时间字符串转换 time.Time
//
// 支持：
//  1. 2026-06-02
//  2. 2026-06-02 15:04:05
//  3. 2026-06-02 15:04
//  4. 2026-06-02T15:04:05
//  5. 2026-06-02T15:04:05+08:00
//  6. 10位秒级时间戳
//  7. 13位毫秒级时间戳
func DateTimeStrToTime(dateTimeStr string, flag ...int) (time.Time, error) {
	dateTimeStr = strings.TrimSpace(dateTimeStr)
	if dateTimeStr == "" {
		return time.Time{}, nil
	}

	mode := DateTimeModeDefault
	if len(flag) > 0 {
		mode = flag[0]
	}

	// 纯数字：只允许 10 位秒时间戳 / 13 位毫秒时间戳
	if isNumber(dateTimeStr) {
		ts, err := strconv.ParseInt(dateTimeStr, 10, 64)
		if err != nil {
			return time.Time{}, errors.New("时间戳格式错误")
		}

		switch len(dateTimeStr) {
		case 10:
			return time.Unix(ts, 0).In(time.Local), nil
		case 13:
			return time.UnixMilli(ts).In(time.Local), nil
		default:
			return time.Time{}, errors.New("时间戳仅支持10位秒级或13位毫秒级")
		}
	}

	// 只有日期时，根据模式补齐开始/结束时间
	if isDateOnly(dateTimeStr) {
		switch mode {
		case DateTimeModeStart:
			dateTimeStr += " 00:00:00"
		case DateTimeModeEnd:
			dateTimeStr += " 23:59:59"
		}
	}

	layouts := []string{
		time.RFC3339,
		time.RFC3339Nano,

		"2006-01-02 15:04:05",
		"2006-01-02 15:04",

		"2006-01-02T15:04:05",
		"2006-01-02T15:04",

		"2006-01-02",
	}

	for _, layout := range layouts {
		if t, err := time.ParseInLocation(layout, dateTimeStr, time.Local); err == nil {
			return t, nil
		}
	}

	return time.Time{}, errors.New("仅支持时间格式：YYYY-MM-DD、YYYY-MM-DD HH:mm:ss、RFC3339、10位秒时间戳、13位毫秒时间戳")
}

// ParseDateRange 解析时间范围
func ParseDateRange(startDate, endDate string) (*time.Time, *time.Time, error) {
	startAt, err := DateTimeStrToTime(startDate, DateTimeModeStart)
	if err != nil {
		return nil, nil, errors.New("开始时间格式错误")
	}

	endAt, err := DateTimeStrToTime(endDate, DateTimeModeEnd)
	if err != nil {
		return nil, nil, errors.New("结束时间格式错误")
	}

	if startAt.IsZero() && endAt.IsZero() {
		return nil, nil, nil
	}

	if !startAt.IsZero() && !endAt.IsZero() && startAt.After(endAt) {
		return nil, nil, errors.New("开始时间不能大于结束时间")
	}

	var startPtr *time.Time
	var endPtr *time.Time

	if !startAt.IsZero() {
		startPtr = &startAt
	}

	if !endAt.IsZero() {
		endPtr = &endAt
	}

	return startPtr, endPtr, nil
}

func isNumber(s string) bool {
	if s == "" {
		return false
	}

	for _, r := range s {
		if r < '0' || r > '9' {
			return false
		}
	}

	return true
}

func isDateOnly(s string) bool {
	if len(s) != len("2006-01-02") {
		return false
	}

	_, err := time.ParseInLocation("2006-01-02", s, time.Local)
	return err == nil
}

// DiffBetweenTwoDays 根据开始时间和结束时间获取2个时间段之间的天数
func DiffBetweenTwoDays(startDate, endDate time.Time) int64 {
	if startDate.After(endDate) {
		startDate, endDate = endDate, startDate
	}
	// 计算时间差（秒），再转换为天数
	diffSeconds := endDate.Unix() - startDate.Unix()
	days := diffSeconds / 86400
	return days
}

var (
	layoutDateTime = "2006-01-02 15:04:05"
	layoutDate     = "2006-01-02"
)

// LocalTime 自定义一个时间类型
type LocalTime struct {
	time.Time
}

func (t *LocalTime) UnmarshalJSON(data []byte) error {
	str := strings.Trim(string(data), "\"")

	if str == "" || str == "null" {
		return nil
	}

	if ts, err := parseTimestamp(str); err == nil {
		t.Time = ts
		return nil
	}

	if tt, err := time.ParseInLocation(layoutDateTime, str, time.Local); err == nil {
		t.Time = tt
		return nil
	}

	if tt, err := time.ParseInLocation(layoutDate, str, time.Local); err == nil {
		t.Time = tt
		return nil
	}
	if tt, err := time.Parse(time.RFC3339, str); err == nil {
		t.Time = tt
		return nil
	}
	return errors.New("时间格式错误: " + str)
}

func (t LocalTime) MarshalJSON() ([]byte, error) {
	if t.Time.IsZero() {
		return []byte(`""`), nil
	}
	return []byte(`"` + t.Format(layoutDateTime) + `"`), nil
}

// IsZero 是否为空
func (t LocalTime) IsZero() bool {
	return t.Time.IsZero()
}

// StartOfDay 获取开始时间（当天 00:00:00）
func (t LocalTime) StartOfDay() time.Time {
	if t.IsZero() {
		return time.Time{}
	}
	y, m, d := t.Date()
	return time.Date(y, m, d, 0, 0, 0, 0, t.Location())
}

// EndOfDay 获取结束时间（当天 23:59:59）
func (t LocalTime) EndOfDay() time.Time {
	if t.IsZero() {
		return time.Time{}
	}
	y, m, d := t.Date()
	return time.Date(y, m, d, 23, 59, 59, 0, t.Location())
}

func parseTimestamp(str string) (time.Time, error) {
	i, err := strconv.ParseInt(str, 10, 64)
	if err != nil {
		return time.Time{}, err
	}

	// 毫秒
	if len(str) == 13 {
		return time.UnixMilli(i), nil
	}

	// 秒
	if len(str) == 10 {
		return time.Unix(i, 0), nil
	}
	return time.Time{}, errors.New("invalid timestamp")
}
