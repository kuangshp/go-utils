package k

import (
	"reflect"
	"testing"
)

func TestKeysAndValues(t *testing.T) {
	m := map[string]int{"b": 2, "a": 1}
	keys := Keys(m)
	if len(keys) != 2 {
		t.Fatalf("Keys len=%d", len(keys))
	}
	values := Values(m)
	if len(values) != 2 {
		t.Fatalf("Values len=%d", len(values))
	}
	sorted := MapKeySort(m)
	if sorted != "a=1&b=2" {
		t.Fatalf("MapKeySort = %q", sorted)
	}
	_ = reflect.TypeOf(keys)
}
