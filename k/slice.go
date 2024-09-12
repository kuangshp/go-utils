package k

import "sort"

// SortListMap 给list的结构体排序
func SortListMap[T any, dataList ~[]T](collection dataList, predicate func(p1, p2 T) bool) {
	sort.Slice(collection, func(i, j int) bool {
		return predicate(collection[i], collection[j])
	})
}

// IsContains 判断一个元素是否包括在list里面
func IsContains[T comparable](slice []T, target T) bool {
	for _, a := range slice {
		if a == target {
			return true
		}
	}
	return false
}

// ForEach 循环数据
func ForEach[T any](collection []T, iteratee func(item T, index int)) {
	for i := range collection {
		iteratee(collection[i], i)
	}
}

// Map 循环一个list,生成一个新的list
func Map[T any, R any](collection []T, iteratee func(item T, index int) R) []R {
	result := make([]R, len(collection))
	for i := range collection {
		result[i] = iteratee(collection[i], i)
	}
	return result
}

// Every 判断list中每个元素都要满足条件才返回true,只要一个不满足就返回false
func Every[T any](slice []T, predicate func(item T, index int) bool) bool {
	for i, v := range slice {
		if !predicate(v, i) {
			return false
		}
	}
	return true
}

// Some list中只要一个满足条件就返回true
func Some[T any](slice []T, predicate func(item T, index int) bool) bool {
	for i, v := range slice {
		if predicate(v, i) {
			return true
		}
	}

	return false
}

// Filter 过滤数据
func Filter[T any, Slice ~[]T](collection Slice, predicate func(item T, index int) bool) Slice {
	result := make(Slice, 0, len(collection))
	for i := range collection {
		if predicate(collection[i], i) {
			result = append(result, collection[i])
		}
	}
	return result
}

// Reduce 循环list,生成一个新的值
func Reduce[T any, R any](collection []T, accumulator func(agg R, item T, index int) R, initial R) R {
	for i := range collection {
		initial = accumulator(initial, collection[i], i)
	}

	return initial
}

func ReduceBy[T any, U any](slice []T, initial U, reducer func(item T, agg U, index int) U) U {
	accumulator := initial
	for i, v := range slice {
		accumulator = reducer(v, accumulator, i)
	}
	return accumulator
}

func ReduceRight[T any, R any](collection []T, accumulator func(agg R, item T, index int) R, initial R) R {
	for i := len(collection) - 1; i >= 0; i-- {
		initial = accumulator(initial, collection[i], i)
	}

	return initial
}

// GroupBy 分组,返回map[key][]list
func GroupBy[T any, U comparable, Slice ~[]T](collection Slice, iteratee func(item T) U) map[U]Slice {
	result := map[U]Slice{}
	for i := range collection {
		key := iteratee(collection[i])

		result[key] = append(result[key], collection[i])
	}
	return result
}

// Difference 取差集,slice=[1,2,3,4,5,6],comparedSlice = [1,2,3] 最后返回[4,5,6]
func Difference[T comparable](slice, comparedSlice []T) []T {
	result := []T{}
	for _, v := range slice {
		if !IsContains(comparedSlice, v) {
			result = append(result, v)
		}
	}
	return result
}

// Intersect 交集
func Intersect[T comparable](slice1, slice2 []T) []T {
	slice1Map := make(map[T]struct{})
	for _, s1 := range slice1 {
		slice1Map[s1] = struct{}{}
	}

	var ret []T
	for _, s2 := range slice2 {
		_, ok := slice1Map[s2]
		if ok {
			ret = append(ret, s2)
		}
	}

	return ret
}

// Union 求并集
func Union[T comparable](slices ...[]T) []T {
	elementMap := make(map[T]struct{})
	for _, sc := range slices {
		for _, element := range sc {
			elementMap[element] = struct{}{}
		}
	}
	retSlice := make([]T, 0, len(elementMap))
	for element := range elementMap {
		retSlice = append(retSlice, element)
	}
	return retSlice
}

// Distinct 直接剩下唯一的 [1,1,2,3,4,4,5] => [1,2,3,4,5]
func Distinct[T comparable](slice []T) []T {
	result := []T{}
	exists := map[T]bool{}
	for _, t := range slice {
		if exists[t] {
			continue
		}
		exists[t] = true
		result = append(result, t)
	}
	return result
}

// SliceToMap 将slice其中抽取几个字段转换为map
func SliceToMap[T any, K comparable, V any](collection []T, transform func(item T) (K, V)) map[K]V {
	result := make(map[K]V, len(collection))
	for i := range collection {
		k, v := transform(collection[i])
		result[k] = v
	}

	return result
}
