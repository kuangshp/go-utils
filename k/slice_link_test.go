package k

import (
	"fmt"
	"testing"
)

func TestBuilder(t *testing.T) {
	list := []int64{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}
	result := From(list).MapToAny(func(item int64, index int) any {
		return item * 2
	}).Reverse().Build()
	fmt.Println("结果:", MapToString(result))
}
