package k

import (
	"net/url"
	"strings"
)

func EncodeURIComponent(str string) string {
	r := url.QueryEscape(str)
	r = strings.Replace(r, "+", "%20", -1)
	return r
}

// URLToMap 将URL查询字符串转换为map
// queryStr: URL查询字符串，如 "name=hello&age=20"
// 返回: map[string]string
func URLToMap(queryStr string) (map[string]string, error) {
	result := make(map[string]string)

	// 如果字符串为空，直接返回空map
	if queryStr == "" {
		return result, nil
	}
	// 使用url.ParseQuery解析查询字符串
	values, err := url.ParseQuery(queryStr)
	if err != nil {
		return nil, err
	}
	// 将Values转换为map[string]string
	for key, val := range values {
		// 如果有多个值，用逗号连接（根据需求调整）
		if len(val) > 0 {
			result[key] = val[0] // 只取第一个值
			// 如果需要所有值，可以用: result[key] = strings.Join(val, ",")
		}
	}
	return result, nil
}
