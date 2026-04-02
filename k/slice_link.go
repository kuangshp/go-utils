package k

import (
	"math/rand"
	"reflect"
	"sort"
	"time"
)

// Builder 流式链式构造器
type Builder[T any] struct {
	data []T
}

// From 创建链式实例
// Example: builder := k.From([]int{1,2,3})
func From[T any](data []T) *Builder[T] {
	return &Builder[T]{data: data}
}

// Build 结束链式，返回结果
// Example: builder.Filter(...).Sort(...).Build()
func (b *Builder[T]) Build() []T {
	return b.data
}

// -----------------------------------------------------------------------------
// 👇 以下全部是 纯链式方法（均返回 *Builder[T]）
// -----------------------------------------------------------------------------

// Sort 自定义排序
// Example: builder.Sort(func(a, b int) bool { return a < b })
func (b *Builder[T]) Sort(less func(a, b T) bool) *Builder[T] {
	sort.Slice(b.data, func(i, j int) bool {
		return less(b.data[i], b.data[j])
	})
	return b
}

// Filter 过滤
// Example: builder.Filter(func(item int, idx int) bool { return item > 0 })
func (b *Builder[T]) Filter(pred func(item T, idx int) bool) *Builder[T] {
	res := make([]T, 0, len(b.data))
	for i, item := range b.data {
		if pred(item, i) {
			res = append(res, item)
		}
	}
	b.data = res
	return b
}

// Distinct 去重
// Example: builder.Distinct()
func (b *Builder[T]) Distinct() *Builder[T] {
	seen := make(map[any]struct{})
	res := make([]T, 0, len(b.data))
	for _, item := range b.data {
		if _, ok := seen[item]; !ok {
			seen[item] = struct{}{}
			res = append(res, item)
		}
	}
	b.data = res
	return b
}

// DistinctByField 按结构体字段去重
// Example: builder.DistinctByField("Name")
func (b *Builder[T]) DistinctByField(field string) *Builder[T] {
	seen := make(map[any]struct{})
	res := make([]T, 0, len(b.data))
	for _, item := range b.data {
		val := reflect.ValueOf(item).FieldByName(field).Interface()
		if _, ok := seen[val]; !ok {
			seen[val] = struct{}{}
			res = append(res, item)
		}
	}
	b.data = res
	return b
}

// ForEach 遍历
// Example: builder.ForEach(func(item int, idx int) { fmt.Println(item) })
func (b *Builder[T]) ForEach(fn func(item T, idx int)) *Builder[T] {
	for i, item := range b.data {
		fn(item, i)
	}
	return b
}

// Reverse 反转切片
// Example: builder.Reverse()
func (b *Builder[T]) Reverse() *Builder[T] {
	for i, j := 0, len(b.data)-1; i < j; i, j = i+1, j-1 {
		b.data[i], b.data[j] = b.data[j], b.data[i]
	}
	return b
}

// Take 取前 N 个
// Example: builder.Take(3)
func (b *Builder[T]) Take(n int) *Builder[T] {
	if n <= 0 {
		b.data = []T{}
		return b
	}
	if n > len(b.data) {
		n = len(b.data)
	}
	b.data = b.data[:n]
	return b
}

// TakeRight 取后 N 个
// Example: builder.TakeRight(2)
func (b *Builder[T]) TakeRight(n int) *Builder[T] {
	if n <= 0 {
		b.data = []T{}
		return b
	}
	start := len(b.data) - n
	if start < 0 {
		start = 0
	}
	b.data = b.data[start:]
	return b
}

// Skip 跳过前 N 个
// Example: builder.Skip(2)
func (b *Builder[T]) Skip(n int) *Builder[T] {
	if n >= len(b.data) {
		b.data = []T{}
		return b
	}
	b.data = b.data[n:]
	return b
}

// SkipRight 跳过后 N 个
// Example: builder.SkipRight(1)
func (b *Builder[T]) SkipRight(n int) *Builder[T] {
	if n >= len(b.data) {
		b.data = []T{}
		return b
	}
	end := len(b.data) - n
	b.data = b.data[:end]
	return b
}

// Shuffle 随机打乱
// Example: builder.Shuffle()
func (b *Builder[T]) Shuffle() *Builder[T] {
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	r.Shuffle(len(b.data), func(i, j int) {
		b.data[i], b.data[j] = b.data[j], b.data[i]
	})
	return b
}

// Pull 移除指定值
// Example: builder.Pull(1, 3, 5)
func (b *Builder[T]) Pull(values ...T) *Builder[T] {
	set := make(map[any]struct{})
	for _, v := range values {
		set[v] = struct{}{}
	}
	res := make([]T, 0, len(b.data))
	for _, item := range b.data {
		if _, ok := set[item]; !ok {
			res = append(res, item)
		}
	}
	b.data = res
	return b
}

// PullAt 移除指定索引
// Example: builder.PullAt(0, 2)
func (b *Builder[T]) PullAt(indexes ...int) *Builder[T] {
	idxSet := make(map[int]struct{})
	for _, i := range indexes {
		idxSet[i] = struct{}{}
	}
	res := make([]T, 0, len(b.data))
	for i, item := range b.data {
		if _, ok := idxSet[i]; !ok {
			res = append(res, item)
		}
	}
	b.data = res
	return b
}

// OrderBy 多字段排序（支持升降）
// Example: builder.OrderBy([]string{"Age","ID"}, []bool{true,false})
func (b *Builder[T]) OrderBy(fields []string, asc []bool) *Builder[T] {
	sort.Slice(b.data, func(i, j int) bool {
		for idx, field := range fields {
			v1 := reflect.ValueOf(b.data[i]).FieldByName(field).Interface()
			v2 := reflect.ValueOf(b.data[j]).FieldByName(field).Interface()
			less := compare(v1, v2)
			if len(asc) > idx && !asc[idx] {
				less = !less
			}
			if less {
				return true
			}
			if compare(v2, v1) {
				return false
			}
		}
		return false
	})
	return b
}

// MapToAny 类型映射，返回新类型的 Builder
func (b *Builder[T]) MapToAny(fn func(item T, index int) any) *Builder[any] {
	result := make([]any, len(b.data))
	for i, item := range b.data {
		result[i] = fn(item, i)
	}
	return From(result)
}

// -----------------------------------------------------------------------------
// 内部比较函数
// -----------------------------------------------------------------------------
func compare(a, b any) bool {
	switch v := a.(type) {
	case int:
		return v < b.(int)
	case int64:
		return v < b.(int64)
	case float64:
		return v < b.(float64)
	case string:
		return v < b.(string)
	default:
		return false
	}
}
