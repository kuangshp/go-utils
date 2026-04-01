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

// DateTimeStrToTime 时间字符串转换时间格式
func DateTimeStrToTime(dateTimeStr string) (time.Time, error) {
	// 空字符串直接返回零值
	if strings.TrimSpace(dateTimeStr) == "" {
		return time.Time{}, nil
	}
	// 尝试时间戳
	if ts, err := strconv.ParseInt(dateTimeStr, 10, 64); err == nil {
		switch len(dateTimeStr) {
		case 13: // 毫秒
			return time.UnixMilli(ts), nil
		case 10: // 秒
			return time.Unix(ts, 0), nil
		}
	}
	layouts := []string{
		"2006-01-02 15:04:05", // 完整日期时间格式
		"2006-01-02",          // 仅日期格式
		time.RFC3339,          // ISO标准格式
	}
	// 尝试用所有支持的格式解析
	var parsedTime time.Time
	var err error

	for _, layout := range layouts {
		parsedTime, err = time.ParseInLocation(layout, dateTimeStr, time.Local)
		if err == nil {
			return parsedTime, nil
		}
	}
	// 如果所有格式都解析失败，返回错误
	return time.Time{}, errors.New("仅支持时间格式: 'YYYY-MM-DD' or 'YYYY-MM-DD HH:MM:SS'")
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
