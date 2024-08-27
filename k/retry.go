package k

import (
	"context"
	"errors"
	"fmt"
	"time"
)

// Retry 重试执行函数,maxRetries:最大重试次数,retryDelayBase:每次重试间隔时间基数,operationFn:操作函数,successFn:执行成功函数
func retry(ctx context.Context, maxRetries int, retryDelayBase int, operationFn func(args ...any) (any, error), successFn func(data ...any), args ...any) error {
	retryDelay := time.Second * time.Duration(retryDelayBase)
	for attempt := 1; attempt <= maxRetries; attempt++ {
		fmt.Println(fmt.Sprintf("重试次数:%d", attempt), time.Now())
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		if data, err := operationFn(args); err == nil {
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

func Retry(operationFn func(args ...any) (any, error), successFn func(args ...any), args ...any) error {
	var maxRetries = 5
	var retryDelayBase = 1
	var maxTime = 10
	var newArgs []any
	switch len(args) {
	case 0:
		newArgs = nil
	case 1:
		maxRetries = args[0].(int)
		newArgs = nil
	case 2:
		maxRetries = args[0].(int)
		retryDelayBase = args[1].(int)
		newArgs = nil
	case 3:
		maxRetries = args[0].(int)
		retryDelayBase = args[1].(int)
		maxTime = args[2].(int)
		newArgs = nil
	default:
		maxRetries = args[0].(int)
		retryDelayBase = args[1].(int)
		maxTime = args[2].(int)
		newArgs = args[3:]
	}
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute*time.Duration(maxTime))
	defer cancel()
	return retry(ctx, maxRetries, retryDelayBase, operationFn, successFn, newArgs)
}
