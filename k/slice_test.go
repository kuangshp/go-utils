package k

import (
	"fmt"
	"testing"
)

func TestSortListMap(t *testing.T) {
	list1 := []map[string]int64{
		{"name": 10},
		{"name": 2},
		{"name": 3},
	}
	SortListMap(list1, func(p1, p2 map[string]int64) bool {
		return p1["name"] < p2["name"]
	})
	fmt.Println(list1)
}

func TestIsContains(t *testing.T) {
	fmt.Println(IsContains([]int64{1, 2, 3}, 2))
}

func TestForEach(t *testing.T) {
	ForEach([]int64{1, 2, 3}, func(item int64, index int) {
		fmt.Println(item, index)
	})
}

func TestMap(t *testing.T) {
	list1 := Map([]int64{1, 2, 3}, func(item int64, index int) int64 {
		return item * 2
	})
	fmt.Println(list1)
}

func TestEvery(t *testing.T) {
	every := Every([]int64{1, 2, 3}, func(item int64, index int) bool {
		return item%2 == 0
	})
	fmt.Println(every)
}

func TestSome(t *testing.T) {
	some := Some([]int64{1, 2, 3}, func(item int64, index int) bool {
		return item%2 == 0
	})
	fmt.Println(some)
}

func TestFilter(t *testing.T) {
	filter := Filter([]int64{1, 2, 3}, func(item int64, index int) bool {
		return item%2 == 0
	})
	fmt.Println(filter)
}

func TestReduce(t *testing.T) {
	var list1 = []int64{1, 2, 3}
	reduce := Reduce(list1, func(pre int64, cur int64, index int) int64 {
		return pre + cur
	}, 0)
	fmt.Println(reduce)
}

func TestGroupBy(t *testing.T) {
	list1 := []map[string]int64{
		{"name": 10, "age": 20},
		{"name": 2, "age": 20},
		{"name": 3, "age": 30},
	}
	groupBy := GroupBy(list1, func(item map[string]int64) int64 {
		return item["age"]
	})
	// {"20":[{"age":20,"name":10},{"age":20,"name":2}],"30":[{"age":30,"name":3}]}
	fmt.Println(MapToString(groupBy))
}

func TestToMap(t *testing.T) {
	list1 := []map[string]int64{
		{"name": 10, "age": 20},
		{"name": 2, "age": 20},
		{"name": 3, "age": 30},
	}
	toMap := ToMap(list1, func(item map[string]int64) (int64, int64) {
		return item["age"], item["name"]
	})
	fmt.Println(MapToString(toMap))
}

func TestFind(t *testing.T) {
	list1 := []map[string]int64{
		{"name": 10, "age": 20},
		{"name": 2, "age": 20},
		{"name": 3, "age": 30},
	}
	find, isOk := Find(list1, func(m map[string]int64) bool {
		return m["age"] > 20
	})
	fmt.Println(find, isOk)
}
