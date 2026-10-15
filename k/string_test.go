package k

import (
	"reflect"
	"strings"
	"testing"
)

func TestMakePasswordAndCheck(t *testing.T) {
	hash, err := MakePassword("Secret#123")
	if err != nil {
		t.Fatalf("MakePassword: %v", err)
	}
	if hash == "" || hash == "Secret#123" {
		t.Fatalf("hash 异常: %q", hash)
	}
	if !CheckPassword(hash, "Secret#123") {
		t.Fatal("正确密码应通过")
	}
	if CheckPassword(hash, "wrong") {
		t.Fatal("错误密码应失败")
	}
}

func TestMaskMobile(t *testing.T) {
	got := MaskMobile("13800138000")
	if !strings.Contains(got, "****") || len(got) != len("13800138000") {
		t.Fatalf("MaskMobile = %q", got)
	}
	if MaskMobile("") != "" {
		t.Fatal("空手机号应返回空串")
	}
}

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
