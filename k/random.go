package k

import (
	"math/rand"
)

// RandInt 随机生成一个整数[1,2)
func RandInt(min, max int) int {
	if min == max {
		return min
	}
	if max < min {
		min, max = max, min
	}
	return rand.Intn(max-min) + min
}
