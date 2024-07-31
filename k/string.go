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
