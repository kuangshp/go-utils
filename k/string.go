package k

import "strings"

// Substring 字符串截取
func Substring(s string, offset int, length uint) string {
	rs := []rune(s)
	size := len(rs)
	if offset < 0 {
		offset = size + offset
		if offset < 0 {
			offset = 0
		}
	}
	if offset > size {
		return ""
	}
	if length > uint(size)-uint(offset) {
		length = uint(size - offset)
	}
	str := string(rs[offset : offset+int(length)])
	return strings.Replace(str, "\x00", "", -1)
}

// HideString 字符串中隐藏部分字符
func HideString(origin string, start, end int, replaceChar string) string {
	size := len(origin)
	if start > size-1 || start < 0 || end < 0 || start > end {
		return origin
	}
	if end > size {
		end = size
	}
	if replaceChar == "" {
		return origin
	}
	startStr := origin[0:start]
	endStr := origin[end:size]
	replaceSize := end - start
	replaceStr := strings.Repeat(replaceChar, replaceSize)
	return startStr + replaceStr + endStr
}

// MaskEmail 隐藏邮箱中间4位
func MaskEmail(email string) string {
	if email == "" {
		return ""
	}
	parts := strings.Split(email, "@")
	if len(parts) != 2 {
		return email
	}
	localPart := parts[0]
	domain := parts[1]
	// 如果本地部分长度小于等于4，只显示第一个字符
	if len(localPart) <= 4 {
		return string(localPart[0]) + "****@" + domain
	}
	// 隐藏中间4位
	visibleStart := (len(localPart) - 4) / 2
	visibleEnd := visibleStart + 4
	return localPart[:visibleStart] + "****" + localPart[visibleEnd:] + "@" + domain
}

// MaskMobile 隐藏手机号中间4位
func MaskMobile(mobile string) string {
	if mobile == "" {
		return ""
	}
	// 如果手机号长度小于等于4，只显示前后各1位
	if len(mobile) <= 4 {
		return mobile
	}
	// 隐藏中间4位
	visibleStart := (len(mobile) - 4) / 2
	visibleEnd := visibleStart + 4
	return mobile[:visibleStart] + "****" + mobile[visibleEnd:]
}
