package k

import "testing"

func TestIf(t *testing.T) {
	if got := If(true, 1, 2); got != 1 {
		t.Fatalf("If true = %d", got)
	}
	if got := If(false, "a", "b"); got != "b" {
		t.Fatalf("If false = %s", got)
	}
}

func TestIfLazy(t *testing.T) {
	called := false
	got := IfLazy(true, func() int {
		called = true
		return 9
	}, 0)
	if !called || got != 9 {
		t.Fatalf("IfLazy true: called=%v got=%d", called, got)
	}

	called = false
	got = IfLazy(false, func() int {
		called = true
		return 9
	}, 3)
	if called || got != 3 {
		t.Fatalf("IfLazy false: called=%v got=%d", called, got)
	}
}
