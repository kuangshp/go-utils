package k

import (
	"strconv"
	"strings"

	"golang.org/x/crypto/bcrypt"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

// Substring 字符串截取
// s 字符串
// offset 起始位置
// length 长度
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
// mobile 手机号字符串
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

// MakePassword 明文转换为密文
func MakePassword(password string) (string, error) {
	newPasswordByte, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(newPasswordByte), nil
}

// CheckPassword 校验密文
// encryptedPassword密文
//
//	password 明文
func CheckPassword(encryptedPassword, password string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(encryptedPassword), []byte(password))
	if err != nil {
		return false
	}
	return true
}

// JoinStr 字符串拼接
// sep 拼接符号
// parts 要拼接的字符串数组
func JoinStr(sep string, parts ...string) string {
	var result []string
	for _, part := range parts {
		if part != "" {
			result = append(result, part)
		}
	}
	return strings.Join(result, sep)
}

// Case2Camel 下划线转驼峰(大驼峰)
func Case2Camel(name string) string {
	name = strings.Replace(name, "_", " ", -1)
	name = cases.Title(language.Und).String(name)
	return strings.Replace(name, " ", "", -1)
}

// LowerCamelCase 转换为小驼峰
func LowerCamelCase(name string) string {
	name = Case2Camel(name)
	return strings.ToLower(name[:1]) + name[1:]
}

// SplitToSlice 分割字符串，自动转为 []string 或整数切片。
// T 只能传 string 或 Go 内置整数类型。
func SplitToSlice[T string | int | int8 | int16 | int32 | int64 | uint | uint8 | uint16 | uint32 | uint64](raw string, sep string) []T {
	var result []T
	if raw == "" {
		return result
	}

	parts := strings.Split(raw, sep)
	for _, s := range parts {
		s = strings.TrimSpace(s)
		if s == "" {
			continue
		}

		var val T
		switch any(val).(type) {
		case string:
			val = any(s).(T)
		case int:
			n, _ := strconv.Atoi(s)
			val = any(n).(T)
		case int8:
			n, _ := strconv.ParseInt(s, 10, 8)
			val = any(int8(n)).(T)
		case int16:
			n, _ := strconv.ParseInt(s, 10, 16)
			val = any(int16(n)).(T)
		case int32:
			n, _ := strconv.ParseInt(s, 10, 32)
			val = any(int32(n)).(T)
		case int64:
			n, _ := strconv.ParseInt(s, 10, 64)
			val = any(n).(T)
		case uint:
			n, _ := strconv.ParseUint(s, 10, 0)
			val = any(uint(n)).(T)
		case uint8:
			n, _ := strconv.ParseUint(s, 10, 8)
			val = any(uint8(n)).(T)
		case uint16:
			n, _ := strconv.ParseUint(s, 10, 16)
			val = any(uint16(n)).(T)
		case uint32:
			n, _ := strconv.ParseUint(s, 10, 32)
			val = any(uint32(n)).(T)
		case uint64:
			n, _ := strconv.ParseUint(s, 10, 64)
			val = any(n).(T)
		}
		result = append(result, val)
	}
	return result
}
