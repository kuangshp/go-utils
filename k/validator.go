package k

import (
	"encoding/json"
	"fmt"
	"net"
	"net/mail"
	"net/url"
	"reflect"
	"regexp"
	"strconv"
	"strings"
	"time"
	"unicode"
)

var (
	alphaMatcher           *regexp.Regexp = regexp.MustCompile(`^[a-zA-Z]+$`)
	letterRegexMatcher     *regexp.Regexp = regexp.MustCompile(`[a-zA-Z]`)
	numberRegexMatcher     *regexp.Regexp = regexp.MustCompile(`\d`)
	intStrMatcher          *regexp.Regexp = regexp.MustCompile(`^[\+-]?\d+$`)
	urlMatcher             *regexp.Regexp = regexp.MustCompile(`^((ftp|http|https?):\/\/)?(\S+(:\S*)?@)?((([1-9]\d?|1\d\d|2[01]\d|22[0-3])(\.(1?\d{1,2}|2[0-4]\d|25[0-5])){2}(?:\.([0-9]\d?|1\d\d|2[0-4]\d|25[0-4]))|(([a-zA-Z0-9]+([-\.][a-zA-Z0-9]+)*)|((www\.)?))?(([a-z\x{00a1}-\x{ffff}0-9]+-?-?)*[a-z\x{00a1}-\x{ffff}0-9]+)(?:\.([a-z\x{00a1}-\x{ffff}]{2,}))?))(:(\d{1,5}))?((\/|\?|#)[^\s]*)?$`)
	dnsMatcher             *regexp.Regexp = regexp.MustCompile(`^(?:[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?\.)+[a-zA-Z]{2,}$`)
	emailMatcher           *regexp.Regexp = regexp.MustCompile(`^[a-z0-9._%+\-]+@[a-z0-9.\-]+\.[a-z]{2,4}$`)
	chineseMobileMatcher   *regexp.Regexp = regexp.MustCompile(`^1(?:3\d|4[4-9]|5[0-35-9]|6[67]|7[013-8]|8\d|9\d)\d{8}$`)                                                                                                                                                                                 // 中国手机号码
	chineseIdMatcher       *regexp.Regexp = regexp.MustCompile(`^(\d{17})([0-9]|X|x)$`)                                                                                                                                                                                                                   // 身份证号码
	chineseMatcher         *regexp.Regexp = regexp.MustCompile("[\u4e00-\u9fa5]")                                                                                                                                                                                                                         // 中文
	chinesePhoneMatcher    *regexp.Regexp = regexp.MustCompile(`\d{3}-\d{8}|\d{4}-\d{7}|\d{4}-\d{8}`)                                                                                                                                                                                                     // 电话电话
	creditCardMatcher      *regexp.Regexp = regexp.MustCompile(`^(?:4[0-9]{12}(?:[0-9]{3})?|5[1-5][0-9]{14}|(222[1-9]|22[3-9][0-9]|2[3-6][0-9]{2}|27[01][0-9]|2720)[0-9]{12}|6(?:011|5[0-9][0-9])[0-9]{12}|3[47][0-9]{13}|3(?:0[0-5]|[68][0-9])[0-9]{11}|(?:2131|1800|35\\d{3})\\d{11}|6[27][0-9]{14})$`) // 信用卡号
	base64Matcher          *regexp.Regexp = regexp.MustCompile(`^(?:[A-Za-z0-9+\\/]{4})*(?:[A-Za-z0-9+\\/]{2}==|[A-Za-z0-9+\\/]{3}=|[A-Za-z0-9+\\/]{4})$`)
	base64URLMatcher       *regexp.Regexp = regexp.MustCompile(`^([A-Za-z0-9_-]{4})*([A-Za-z0-9_-]{2}(==)?|[A-Za-z0-9_-]{3}=?)?$`)
	binMatcher             *regexp.Regexp = regexp.MustCompile(`^(0b)?[01]+$`)
	hexMatcher             *regexp.Regexp = regexp.MustCompile(`^(#|0x|0X)?[0-9a-fA-F]+$`)
	visaMatcher            *regexp.Regexp = regexp.MustCompile(`^4[0-9]{12}(?:[0-9]{3})?$`)
	masterCardMatcher      *regexp.Regexp = regexp.MustCompile(`^5[1-5][0-9]{14}$`)
	americanExpressMatcher *regexp.Regexp = regexp.MustCompile(`^3[47][0-9]{13}$`)
	unionPay               *regexp.Regexp = regexp.MustCompile("^62[0-5]\\d{13,16}$")
	chinaUnionPay          *regexp.Regexp = regexp.MustCompile(`^62[0-9]{14,17}$`)
	dateRegex              *regexp.Regexp = regexp.MustCompile(`^\d{4}-\d{2}-\d{2}$`)                   // 匹配 YYYY-MM-DD
	dateTimeRegex          *regexp.Regexp = regexp.MustCompile(`^\d{4}-\d{2}-\d{2} \d{2}:\d{2}:\d{2}$`) // 匹配 YYYY-MM-DD HH:MM:SS
)

var (
	// Identity card formula
	factor = [17]int{7, 9, 10, 5, 8, 4, 2, 1, 6, 3, 7, 9, 10, 5, 8, 4, 2}
	// ID verification bit
	verifyStr = [11]string{"1", "0", "X", "9", "8", "7", "6", "5", "4", "3", "2"}
	// Starting year of ID card
	birthStartYear = 1900
	// Province code
	provinceKv = map[string]struct{}{
		"11": {},
		"12": {},
		"13": {},
		"14": {},
		"15": {},
		"21": {},
		"22": {},
		"23": {},
		"31": {},
		"32": {},
		"33": {},
		"34": {},
		"35": {},
		"36": {},
		"37": {},
		"41": {},
		"42": {},
		"43": {},
		"44": {},
		"45": {},
		"46": {},
		"50": {},
		"51": {},
		"52": {},
		"53": {},
		"54": {},
		"61": {},
		"62": {},
		"63": {},
		"64": {},
		"65": {},
		//"71": {},
		//"81": {},
		//"82": {},
	}
)

// IsAlpha 判断是否为字母,包括大写，小写
func IsAlpha(str string) bool {
	return alphaMatcher.MatchString(str)
}

// IsAllUpper 判断是否全部为大写字母
func IsAllUpper(str string) bool {
	for _, r := range str {
		if !unicode.IsUpper(r) {
			return false
		}
	}
	return str != ""
}

// IsAllLower 判断是否全部为小写字母
func IsAllLower(str string) bool {
	for _, r := range str {
		if !unicode.IsLower(r) {
			return false
		}
	}
	return str != ""
}

// IsASCII 判断字符串是否全部为ASCII
func IsASCII(str string) bool {
	for i := 0; i < len(str); i++ {
		if str[i] > unicode.MaxASCII {
			return false
		}
	}
	return true
}

// ContainUpper 判断字符串里面是否包含大写字母
func ContainUpper(str string) bool {
	for _, r := range str {
		if unicode.IsUpper(r) && unicode.IsLetter(r) {
			return true
		}
	}
	return false
}

// ContainLower 判断字符串里面是否包括小写字符
func ContainLower(str string) bool {
	for _, r := range str {
		if unicode.IsLower(r) && unicode.IsLetter(r) {
			return true
		}
	}
	return false
}

// ContainLetter 检查字符串是否至少包含一个字母。
func ContainLetter(str string) bool {
	return letterRegexMatcher.MatchString(str)
}

// ContainNumber 判断字符中至少包含一个数字
func ContainNumber(input string) bool {
	return numberRegexMatcher.MatchString(input)
}

// IsJSON 判断是否为json
func IsJSON(str string) bool {
	var js json.RawMessage
	return json.Unmarshal([]byte(str), &js) == nil
}

// IsNumberStr 判断是否为字符串数字
func IsNumberStr(s string) bool {
	return IsIntStr(s) || IsFloatStr(s)
}

// IsFloatStr 判断是否为浮点型字符串
func IsFloatStr(str string) bool {
	_, e := strconv.ParseFloat(str, 64)
	return e == nil
}

// IsIntStr 判断是否为整形字符串
func IsIntStr(str string) bool {
	return intStrMatcher.MatchString(str)
}

// IsIp 判断是否为ip地址
func IsIp(ipStr string) bool {
	ip := net.ParseIP(ipStr)
	return ip != nil
}

// IsIpV4 判断是否为ipv4
func IsIpV4(ipStr string) bool {
	ip := net.ParseIP(ipStr)
	if ip == nil {
		return false
	}
	return ip.To4() != nil
}

// IsIpV6 判断是否为ipv6
func IsIpV6(ipStr string) bool {
	ip := net.ParseIP(ipStr)
	if ip == nil {
		return false
	}
	return ip.To4() == nil && len(ip) == net.IPv6len
}

// IsPort 判断是为端口号
func IsPort(str string) bool {
	if i, err := strconv.ParseInt(str, 10, 64); err == nil && i > 0 && i < 65536 {
		return true
	}
	return false
}

// IsUrl 是否为url地址
func IsUrl(str string) bool {
	if str == "" || len(str) >= 2083 || len(str) <= 3 || strings.HasPrefix(str, ".") {
		return false
	}
	u, err := url.Parse(str)
	if err != nil {
		return false
	}
	if strings.HasPrefix(u.Host, ".") {
		return false
	}
	if u.Host == "" && (u.Path != "" && !strings.Contains(u.Path, ".")) {
		return false
	}

	return urlMatcher.MatchString(str)
}

// IsDns check if the string is dns.
// Play: https://go.dev/play/p/jlYApVLLGTZ
func IsDns(dns string) bool {
	return dnsMatcher.MatchString(dns)
}

// IsEmail 是否为email地址
func IsEmail(email string) bool {
	_, err := mail.ParseAddress(email)
	return err == nil
}

// IsDate 判断是否为时间格式
func IsDate(date string) bool {
	return dateRegex.MatchString(date)
}

// IsDateTime 判断是否datetime时间格式
func IsDateTime(dateTime string) bool {
	return dateTimeRegex.MatchString(dateTime)
}

// IsChineseMobile 是否为中国手机号码
func IsChineseMobile(mobileNum string) bool {
	return chineseMobileMatcher.MatchString(mobileNum)
}

// IsChineseIdNum 是否为中国大陆身份证号码
func IsChineseIdNum(id string) bool {
	// All characters should be numbers, and the last digit can be either x or X
	if !chineseIdMatcher.MatchString(id) {
		return false
	}

	// Verify province codes and complete all province codes according to GB/T2260
	_, ok := provinceKv[id[0:2]]
	if !ok {
		return false
	}

	// Verify birthday, must be greater than birthStartYear and less than the current year
	birthStr := fmt.Sprintf("%s-%s-%s", id[6:10], id[10:12], id[12:14])
	birthday, err := time.Parse("2006-01-02", birthStr)
	if err != nil || birthday.After(time.Now()) || birthday.Year() < birthStartYear {
		return false
	}

	// Verification code
	sum := 0
	for i, c := range id[:17] {
		v, _ := strconv.Atoi(string(c))
		sum += v * factor[i]
	}

	return verifyStr[sum%11] == strings.ToUpper(id[17:18])
}

// IsContainChinese 是否包括中文
func IsContainChinese(s string) bool {
	return chineseMatcher.MatchString(s)
}

// IsChinesePhone 是否为中国电话号码
func IsChinesePhone(phone string) bool {
	return chinesePhoneMatcher.MatchString(phone)
}

// IsCreditCard 是否为信用卡号
func IsCreditCard(creditCart string) bool {
	return creditCardMatcher.MatchString(creditCart)
}

// IsBase64 是否为base64字符
func IsBase64(base64 string) bool {
	return base64Matcher.MatchString(base64)
}

// IsEmptyString 是否为空字符串
func IsEmptyString(str string) bool {
	return len(str) == 0
}

// IsRegexMatch 是否为正则
func IsRegexMatch(str, regex string) bool {
	reg := regexp.MustCompile(regex)
	return reg.MatchString(str)
}

// IsZeroValue 是否为空值
func IsZeroValue(value any) bool {
	if value == nil {
		return true
	}

	rv := reflect.ValueOf(value)
	if rv.Kind() == reflect.Ptr {
		rv = rv.Elem()
	}

	if !rv.IsValid() {
		return true
	}

	switch rv.Kind() {
	case reflect.String:
		return rv.Len() == 0
	case reflect.Bool:
		return !rv.Bool()
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return rv.Int() == 0
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		return rv.Uint() == 0
	case reflect.Float32, reflect.Float64:
		return rv.Float() == 0
	case reflect.Ptr, reflect.Chan, reflect.Func, reflect.Interface, reflect.Slice, reflect.Map:
		return rv.IsNil()
	}
	return reflect.DeepEqual(rv.Interface(), reflect.Zero(rv.Type()).Interface())
}
