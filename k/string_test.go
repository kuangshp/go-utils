package k

import (
	"reflect"
	"testing"
)

func TestSplitToSliceString(t *testing.T) {
	got := SplitToSlice[string](" a, b ,, c ", ",")
	want := []string{"a", "b", "c"}

	if !reflect.DeepEqual(got, want) {
		t.Fatalf("SplitToSlice[string]() = %#v, want %#v", got, want)
	}
}

func TestSplitToSliceIntTypes(t *testing.T) {
	gotInt := SplitToSlice[int]("1, 2,3", ",")
	wantInt := []int{1, 2, 3}
	if !reflect.DeepEqual(gotInt, wantInt) {
		t.Fatalf("SplitToSlice[int]() = %#v, want %#v", gotInt, wantInt)
	}

	gotInt64 := SplitToSlice[int64]("1|2|3", "|")
	wantInt64 := []int64{1, 2, 3}
	if !reflect.DeepEqual(gotInt64, wantInt64) {
		t.Fatalf("SplitToSlice[int64]() = %#v, want %#v", gotInt64, wantInt64)
	}

	gotUint32 := SplitToSlice[uint32]("4,5,6", ",")
	wantUint32 := []uint32{4, 5, 6}
	if !reflect.DeepEqual(gotUint32, wantUint32) {
		t.Fatalf("SplitToSlice[uint32]() = %#v, want %#v", gotUint32, wantUint32)
	}
}

func TestSplitToSliceEmpty(t *testing.T) {
	got := SplitToSlice[int]("", ",")

	if len(got) != 0 {
		t.Fatalf("SplitToSlice() = %#v, want empty slice", got)
	}
}
