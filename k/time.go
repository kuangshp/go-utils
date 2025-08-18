package k

import (
	"errors"
	"regexp"
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
func DateStrToTime(dateTimeStr string) time.Time {
	local, _ := time.LoadLocation("Asia/Shanghai")
	dateTime, _ := time.ParseInLocation("2006-01-02", dateTimeStr, local)
	return dateTime
}

// DateTimeStrToTime 时间字符串转换时间格式
func DateTimeStrToTime(dateTimeStr string) (time.Time, error) {
	layouts := []string{
		"2006-01-02 15:04:05", // 完整日期时间格式
		"2006-01-02",          // 仅日期格式
	}
	// 尝试用所有支持的格式解析
	var parsedTime time.Time
	var err error

	for _, layout := range layouts {
		parsedTime, err = time.Parse(layout, dateTimeStr)
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
