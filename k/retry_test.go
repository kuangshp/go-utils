package k

import (
	"errors"
	"fmt"
	"math/rand"
	"testing"
	"time"
)

// 定义要执行的函数
func test1(args ...any) (any, error) {
	fmt.Println(args, "??")
	rand.New(rand.NewSource(time.Now().UnixNano())) // 使用当前时间的纳秒数作为种子，确保每次运行产生不同的随机数
	randomNumber := rand.Float64()
	fmt.Println(randomNumber, "当前值")
	if randomNumber < 0.9 {
		return nil, errors.New("执行失败,进入再次重试")
	}
	return randomNumber, nil
}

func TestRetry(t *testing.T) {
	maxRetries := 5
	retryDelay := 2
	if err := Retry(test1, func(data any) {
		fmt.Println("成功数据:", data)
	}, maxRetries, retryDelay, 2, 10, 3, 4); err != nil {
		fmt.Printf("重试次数太多了: %s\n", err)
	}
}
