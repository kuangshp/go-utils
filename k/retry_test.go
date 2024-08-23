package k

import (
	"context"
	"errors"
	"fmt"
	"math/rand"
	"testing"
	"time"
)

// 定义要执行的函数
func test1() (any, error) {
	rand.Seed(time.Now().UnixNano()) // 使用当前时间的纳秒数作为种子，确保每次运行产生不同的随机数
	randomNumber := rand.Float64()
	fmt.Println(randomNumber, "当前值")
	if randomNumber < 0.5 {
		return nil, errors.New("执行失败,进入再次重试")
	}
	return randomNumber, nil
}

func TestRetry(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute*10)
	defer cancel()
	maxRetries := 5
	retryDelay := 2
	if err := Retry(ctx, maxRetries, retryDelay, test1, func(data any) {
		fmt.Println("重试成功:", data)
	}); err != nil {
		fmt.Printf("重试次数太多了: %s\n", err)
	}
}
