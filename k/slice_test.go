package k

import (
	"fmt"
	"reflect"
	"sort"
	"testing"
)

// TestSortListMap 测试按自定义规则对切片排序。
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

// TestSortBy 测试按多个字段组合排序，并在比较结果相等时保留原顺序。
func TestSortBy(t *testing.T) {
	type user struct {
		ID    int
		Name  string
		Score int
	}
	users := []user{
		{ID: 1, Name: "alice", Score: 90},
		{ID: 2, Name: "bob", Score: 80},
		{ID: 3, Name: "cindy", Score: 90},
		{ID: 4, Name: "david", Score: 90},
	}

	SortBy(
		users,
		DescBy(func(item user) int {
			return item.Score
		}),
		AscBy(func(item user) string {
			return item.Name
		}),
	)
	want := []user{
		{ID: 1, Name: "alice", Score: 90},
		{ID: 3, Name: "cindy", Score: 90},
		{ID: 4, Name: "david", Score: 90},
		{ID: 2, Name: "bob", Score: 80},
	}

	if !reflect.DeepEqual(users, want) {
		t.Fatalf("SortBy() = %#v, want %#v", users, want)
	}
}

// TestSortByStable 测试所有比较器相等时保持输入顺序。
func TestSortByStable(t *testing.T) {
	type user struct {
		ID    int
		Score int
	}
	users := []user{
		{ID: 1, Score: 90},
		{ID: 2, Score: 90},
		{ID: 3, Score: 80},
	}

	SortBy(users, DescBy(func(item user) int {
		return item.Score
	}))
	want := []user{
		{ID: 1, Score: 90},
		{ID: 2, Score: 90},
		{ID: 3, Score: 80},
	}

	if !reflect.DeepEqual(users, want) {
		t.Fatalf("SortBy() = %#v, want %#v", users, want)
	}
}

// TestIsContains 测试判断切片是否包含指定元素。
func TestIsContains(t *testing.T) {
	fmt.Println(IsContains([]int64{1, 2, 3}, 2))
}

// TestForEach 测试遍历切片并传入元素和索引。
func TestForEach(t *testing.T) {
	ForEach([]int64{1, 2, 3}, func(item int64, index int) {
		fmt.Println(item, index)
	})
}

// TestMap 测试将切片元素映射为新的切片。
func TestMap(t *testing.T) {
	list1 := Map([]int64{1, 2, 3}, func(item int64, index int) int64 {
		return item * 2
	})
	fmt.Println(list1)
}

// TestEvery 测试所有元素都满足条件时的判断逻辑。
func TestEvery(t *testing.T) {
	every := Every([]int64{1, 2, 3}, func(item int64, index int) bool {
		return item%2 == 0
	})
	fmt.Println(every)
}

// TestSome 测试任一元素满足条件时的判断逻辑。
func TestSome(t *testing.T) {
	some := Some([]int64{1, 2, 3}, func(item int64, index int) bool {
		return item%2 == 0
	})
	fmt.Println(some)
}

// TestFilter 测试按条件过滤切片元素。
func TestFilter(t *testing.T) {
	filter := Filter([]int64{1, 2, 3}, func(item int64, index int) bool {
		return item%2 == 0
	})
	fmt.Println(filter)
}

// TestReduce 测试从左到右聚合切片元素。
func TestReduce(t *testing.T) {
	var list1 = []int64{1, 2, 3}
	reduce := Reduce(list1, func(pre int64, cur int64, index int) int64 {
		return pre + cur
	}, 0)
	fmt.Println(reduce)
}

// TestGroupBy 测试按指定 key 对切片元素分组。
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

// TestToMap 测试将切片转换为 map。
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

// TestFind 测试查找第一个满足条件的元素。
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

// TestReduceBy 测试通过自定义 reducer 聚合切片元素。
func TestReduceBy(t *testing.T) {
	got := ReduceBy([]int{1, 2, 3}, 10, func(item int, agg int, index int) int {
		return agg + item + index
	})

	if got != 19 {
		t.Fatalf("ReduceBy() = %d, want 19", got)
	}
}

// TestReduceRight 测试从右到左聚合切片元素。
func TestReduceRight(t *testing.T) {
	got := ReduceRight([]string{"a", "b", "c"}, func(agg string, item string, index int) string {
		return agg + item
	}, "")

	if got != "cba" {
		t.Fatalf("ReduceRight() = %q, want %q", got, "cba")
	}
}

// TestGroupByWithMapper 测试分组时同时映射分组值。
func TestGroupByWithMapper(t *testing.T) {
	type user struct {
		Name string
		Age  int
	}
	users := []user{
		{Name: "alice", Age: 20},
		{Name: "bob", Age: 20},
		{Name: "cindy", Age: 30},
	}

	got := GroupByWithMapper(users, func(item user) int {
		return item.Age
	}, func(item user) string {
		return item.Name
	})
	want := map[int][]string{
		20: {"alice", "bob"},
		30: {"cindy"},
	}

	if !reflect.DeepEqual(got, want) {
		t.Fatalf("GroupByWithMapper() = %#v, want %#v", got, want)
	}
}

// TestDifference 测试返回第一个切片中独有的元素。
func TestDifference(t *testing.T) {
	got := Difference([]int{1, 2, 3, 4, 5, 6}, []int{1, 2, 3})
	want := []int{4, 5, 6}

	if !reflect.DeepEqual(got, want) {
		t.Fatalf("Difference() = %#v, want %#v", got, want)
	}
}

// TestIntersect 测试返回两个切片的交集元素。
func TestIntersect(t *testing.T) {
	got := Intersect([]int{1, 2, 3}, []int{3, 2, 4, 2})
	want := []int{3, 2, 2}

	if !reflect.DeepEqual(got, want) {
		t.Fatalf("Intersect() = %#v, want %#v", got, want)
	}
}

// TestUnion 测试合并多个切片并去重。
func TestUnion(t *testing.T) {
	got := Union([]int{1, 2, 2}, []int{2, 3}, []int{4})
	sort.Ints(got)
	want := []int{1, 2, 3, 4}

	if !reflect.DeepEqual(got, want) {
		t.Fatalf("Union() = %#v, want %#v", got, want)
	}
}

// TestDistinct 测试保留切片中的唯一元素。
func TestDistinct(t *testing.T) {
	got := Distinct([]int{1, 1, 2, 3, 2, 4})
	want := []int{1, 2, 3, 4}

	if !reflect.DeepEqual(got, want) {
		t.Fatalf("Distinct() = %#v, want %#v", got, want)
	}
}

// TestDistinctBy 测试按 key 函数对切片去重，并保留首次出现的元素顺序。
func TestDistinctBy(t *testing.T) {
	type user struct {
		ID   int
		Name string
	}
	users := []user{
		{ID: 1, Name: "alice"},
		{ID: 2, Name: "bob"},
		{ID: 1, Name: "alice again"},
		{ID: 3, Name: "cindy"},
		{ID: 2, Name: "bob again"},
	}

	got := DistinctBy(users, func(item user) int {
		return item.ID
	})
	want := []user{
		{ID: 1, Name: "alice"},
		{ID: 2, Name: "bob"},
		{ID: 3, Name: "cindy"},
	}

	if !reflect.DeepEqual(got, want) {
		t.Fatalf("DistinctBy() = %#v, want %#v", got, want)
	}
}

// TestDiffIds 测试比较数据库 ID 和请求 ID 后返回新增、删除集合。
func TestDiffIds(t *testing.T) {
	toInsert, toUpdate, toDelete := DiffIds([]int{1, 2, 3}, []int{2, 3, 4})
	sort.Ints(toInsert)
	sort.Ints(toUpdate)
	sort.Ints(toDelete)

	if !reflect.DeepEqual(toInsert, []int{4}) {
		t.Fatalf("DiffIds() toInsert = %#v, want %#v", toInsert, []int{4})
	}
	if !reflect.DeepEqual(toUpdate, []int{2, 3}) {
		t.Fatalf("DiffIds() toUpdate = %#v, want %#v", toUpdate, []int{2, 3})
	}
	if !reflect.DeepEqual(toDelete, []int{1}) {
		t.Fatalf("DiffIds() toDelete = %#v, want %#v", toDelete, []int{1})
	}
}

// TestDiffByKey 测试按 key 比较数据库对象和请求对象后返回新增、更新、删除集合。
func TestDiffByKey(t *testing.T) {
	type user struct {
		ID   int
		Name string
	}
	dbUsers := []user{
		{ID: 1, Name: "alice old"},
		{ID: 2, Name: "bob old"},
		{ID: 3, Name: "cindy old"},
	}
	reqUsers := []user{
		{ID: 0, Name: "new without id"},
		{ID: 2, Name: "bob new"},
		{ID: 4, Name: "david new"},
	}

	toCreate, toUpdate, toDelete := DiffByKey(dbUsers, reqUsers, func(item user) int {
		return item.ID
	})
	wantCreate := []user{
		{ID: 0, Name: "new without id"},
		{ID: 4, Name: "david new"},
	}
	wantUpdate := []user{
		{ID: 2, Name: "bob new"},
	}
	wantDelete := []user{
		{ID: 1, Name: "alice old"},
		{ID: 3, Name: "cindy old"},
	}

	sort.Slice(toDelete, func(i, j int) bool {
		return toDelete[i].ID < toDelete[j].ID
	})

	if !reflect.DeepEqual(toCreate, wantCreate) {
		t.Fatalf("DiffByKey() toCreate = %#v, want %#v", toCreate, wantCreate)
	}
	if !reflect.DeepEqual(toUpdate, wantUpdate) {
		t.Fatalf("DiffByKey() toUpdate = %#v, want %#v", toUpdate, wantUpdate)
	}
	if !reflect.DeepEqual(toDelete, wantDelete) {
		t.Fatalf("DiffByKey() toDelete = %#v, want %#v", toDelete, wantDelete)
	}
}

// TestIndex 测试查找元素索引以及未找到时返回 -1。
func TestIndex(t *testing.T) {
	if got := Index([]string{"a", "b", "c"}, "b"); got != 1 {
		t.Fatalf("Index() existing = %d, want 1", got)
	}
	if got := Index([]string{"a", "b", "c"}, "x"); got != -1 {
		t.Fatalf("Index() missing = %d, want -1", got)
	}
}

// TestFindNotFound 测试 Find 未命中时返回零值和 false。
func TestFindNotFound(t *testing.T) {
	got, ok := Find([]int{1, 2, 3}, func(item int) bool {
		return item > 10
	})

	if ok {
		t.Fatalf("Find() ok = true, want false")
	}
	if got != 0 {
		t.Fatalf("Find() zero value = %d, want 0", got)
	}
}

// TestDistinctByField 测试按结构体字段对切片去重。
func TestDistinctByField(t *testing.T) {
	type user struct {
		ID   int
		Name string
	}
	users := []user{
		{ID: 1, Name: "alice"},
		{ID: 2, Name: "bob"},
		{ID: 1, Name: "alice again"},
	}

	got := DistinctByField(users, "ID")
	want := []user{
		{ID: 1, Name: "alice"},
		{ID: 2, Name: "bob"},
	}

	if !reflect.DeepEqual(got, want) {
		t.Fatalf("DistinctByField() = %#v, want %#v", got, want)
	}
}

// TestChunkSlice 测试按指定大小拆分切片。
func TestChunkSlice(t *testing.T) {
	got := ChunkSlice([]int{1, 2, 3, 4, 5}, 2)
	want := [][]int{{1, 2}, {3, 4}, {5}}

	if !reflect.DeepEqual(got, want) {
		t.Fatalf("ChunkSlice() = %#v, want %#v", got, want)
	}
	if got := ChunkSlice([]int{1, 2, 3}, 0); got != nil {
		t.Fatalf("ChunkSlice() with zero size = %#v, want nil", got)
	}
}

// TestFlatten 测试将二维切片扁平化一层。
func TestFlatten(t *testing.T) {
	builder := From([][]int{{1, 2}, {3}, {4, 5}})
	got := Flatten(builder)
	want := []int{1, 2, 3, 4, 5}

	if !reflect.DeepEqual(got, want) {
		t.Fatalf("Flatten() = %#v, want %#v", got, want)
	}
}

// ExampleFlatten 演示 Flatten 的基础用法。
func ExampleFlatten() {
	builder := From([][]int{{1, 2}, {3}})
	fmt.Println(Flatten(builder))
	// Output: [1 2 3]
}
