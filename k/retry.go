package k

import (
	"context"
	"errors"
	"fmt"
	"time"
)

// Retry 重试执行函数,maxRetries:最大重试次数,retryDelayBase:每次重试间隔时间基数,operationFn:操作函数,successFn:执行成功函数
func Retry(ctx context.Context, maxRetries int, retryDelayBase int, operationFn func() (any, error), successFn func(data any)) error {
	retryDelay := time.Second * time.Duration(retryDelayBase)
	for attempt := 1; attempt <= maxRetries; attempt++ {
		fmt.Println(fmt.Sprintf("重试次数:%d", attempt), time.Now())
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		if data, err := operationFn(); err == nil {
			successFn(data)
			return nil
		}

		if attempt < maxRetries {
			time.Sleep(retryDelay)
			retryDelay *= 2 // 每次重试将间隔时间翻倍
		}
	}
	return errors.New("执行失败")
}
