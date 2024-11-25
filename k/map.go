package k

import (
	"encoding/json"
	"fmt"
	"sort"
)

// MapToString 将map转换为字符串输出
func MapToString[T any](result T) string {
	jsonBytes, _ := json.Marshal(result)
	return string(jsonBytes)
}

// MapKeySort 将map的key进行ASCII排序
func MapKeySort[T any](m map[string]T, options ...string) string {
	// 将map中全部的key到切片中
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	// 对切片进行排序
	sort.Strings(keys)
	// 循环拼接返回
	result := ""
	for _, item := range keys {
		if len(options) > 0 {
			if result != "" {
				result += fmt.Sprintf("%s%v", options[0], fmt.Sprintf("%s=%v", item, m[item]))
			} else {
				result += fmt.Sprintf("%s=%v", item, m[item])
			}
		} else {
			if result != "" {
				result += fmt.Sprintf("&%v", fmt.Sprintf("%s=%v", item, m[item]))
			} else {
				result += fmt.Sprintf("%s=%v", item, m[item])
			}
		}
	}
	return result
}

// Keys 获取map的全部key
func Keys[K comparable, V any](m map[K]V) []K {
	keys := make([]K, len(m))
	var i int
	for k := range m {
		keys[i] = k
		i++
	}
	return keys
}

// Values 获取map全部的value
func Values[K comparable, V any](m map[K]V) []V {
	values := make([]V, len(m))
	var i int
	for _, v := range m {
		values[i] = v
		i++
	}
	return values
}

// KeysBy 根据条件来获取map的key
func KeysBy[K comparable, V any, T any](m map[K]V, mapper func(item K) T) []T {
	keys := make([]T, 0, len(m))
	for k := range m {
		keys = append(keys, mapper(k))
	}
	return keys
}

// ValuesBy 根据条件来获取map的值
func ValuesBy[K comparable, V any, T any](m map[K]V, mapper func(item V) T) []T {
	keys := make([]T, 0, len(m))
	for _, v := range m {
		keys = append(keys, mapper(v))
	}
	return keys
}

// Merge 合并多个map
func Merge[K comparable, V any](maps ...map[K]V) map[K]V {
	result := make(map[K]V, 0)
	for _, m := range maps {
		for k, v := range m {
			result[k] = v
		}
	}
	return result
}

// ForEachMap 循环map
func ForEachMap[K comparable, V any](m map[K]V, iteratee func(key K, value V)) {
	for k, v := range m {
		iteratee(k, v)
	}
}

// FilterMap 过滤map对象
func FilterMap[K comparable, V any](m map[K]V, predicate func(key K, value V) bool) map[K]V {
	result := make(map[K]V)
	for k, v := range m {
		if predicate(k, v) {
			result[k] = v
		}
	}
	return result
}

// FilterByKeys 根据key过滤数据
func FilterByKeys[K comparable, V any](m map[K]V, keys []K) map[K]V {
	result := make(map[K]V)
	for k, v := range m {
		if IsContains(keys, k) {
			result[k] = v
		}
	}
	return result
}

// FilterByValues 根据value值来过滤数据
func FilterByValues[K comparable, V comparable](m map[K]V, values []V) map[K]V {
	result := make(map[K]V)
	for k, v := range m {
		if IsContains(values, v) {
			result[k] = v
		}
	}
	return result
}

// OmitBy  从map中删除元素
func OmitBy[K comparable, V any](m map[K]V, predicate func(key K, value V) bool) map[K]V {
	result := make(map[K]V)
	for k, v := range m {
		if !predicate(k, v) {
			result[k] = v
		}
	}
	return result
}

// OmitByKeys 根据key来删除元素
func OmitByKeys[K comparable, V any](m map[K]V, keys []K) map[K]V {
	result := make(map[K]V)
	for k, v := range m {
		if !IsContains(keys, k) {
			result[k] = v
		}
	}
	return result
}

// OmitByValues 根据value字段来删除
func OmitByValues[K comparable, V comparable](m map[K]V, values []V) map[K]V {
	result := make(map[K]V)
	for k, v := range m {
		if !IsContains(values, v) {
			result[k] = v
		}
	}
	return result
}

// MapKeys 抽取map中全部的key
func MapKeys[K comparable, V any, T comparable](m map[K]V, iteratee func(key K, value V) T) map[T]V {
	result := make(map[T]V, len(m))
	for k, v := range m {
		result[iteratee(k, v)] = v
	}
	return result
}

// MapValues 抽取map中全部的value
func MapValues[K comparable, V any, T any](m map[K]V, iteratee func(key K, value V) T) map[K]T {
	result := make(map[K]T, len(m))
	for k, v := range m {
		result[k] = iteratee(k, v)
	}
	return result
}

// HasKey map中是否包括这个key
func HasKey[K comparable, V any](m map[K]V, key K) bool {
	_, hasKey := m[key]
	return hasKey
}
