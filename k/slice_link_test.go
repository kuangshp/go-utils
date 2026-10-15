package k

import (
	"reflect"
	"testing"
)

func TestBuilder(t *testing.T) {
	list := []int64{1, 2, 3, 4, 5}
	result := From(list).Filter(func(item int64, index int) bool {
		return item%2 == 1
	}).Reverse().Build()
	want := []int64{5, 3, 1}
	if !reflect.DeepEqual(result, want) {
		t.Fatalf("Builder = %#v, want %#v", result, want)
	}
}

func TestBuilderMapToAny(t *testing.T) {
	list := []int64{1, 2, 3}
	result := From(list).MapToAny(func(item int64, index int) any {
		return item * 2
	}).Build()
	want := []any{int64(2), int64(4), int64(6)}
	if !reflect.DeepEqual(result, want) {
		t.Fatalf("MapToAny = %#v, want %#v", result, want)
	}
}
