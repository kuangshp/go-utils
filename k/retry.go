package k

import (
	"context"
	"errors"
	"fmt"
	"time"
)

// Retry 重试执行函数,maxRetries:最大重试次数,retryDelayBase:每次重试间隔时间基数,operationFn:操作函数,successFn:执行成功函数
func retry(ctx context.Context, maxRetries, retryDelayBase, retryDelayInterval int, operationFn func(data ...any) (any, error), successFn func(data any), args ...any) error {
	retryDelay := time.Second * time.Duration(retryDelayBase)
	for attempt := 1; attempt <= maxRetries; attempt++ {
		fmt.Println(fmt.Sprintf("重试次数:%d", attempt), time.Now())
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}
		if data, err := operationFn(args...); err == nil {
			successFn(data)
			return nil
		}

		if attempt < maxRetries {
			time.Sleep(retryDelay)
			retryDelay *= time.Duration(retryDelayInterval) // 每次重试将间隔时间翻倍
		}
	}
	return errors.New("执行失败")
}

func Retry(operationFn func(data ...any) (any, error), successFn func(data any), args ...any) error {
	var maxRetries = 5         // 最大重试次数
	var retryDelayBase = 1     // 重试间隔时间基数
	var retryDelayInterval = 1 // 重试间隔
	var maxTime = 10           // 最大时间
	var newArgs []any
	switch len(args) {
	case 0:
		newArgs = nil
	case 1:
		if v, isOk := args[0].(int); isOk {
			maxRetries = v
		}
	case 2:
		if v, isOk := args[0].(int); isOk {
			maxRetries = v
		}
		if v, isOk := args[1].(int); isOk {
			retryDelayBase = v
		}
	case 3:
		if v, isOk := args[0].(int); isOk {
			maxRetries = v
		}
		if v, isOk := args[1].(int); isOk {
			retryDelayBase = v
		}
		if v, isOk := args[2].(int); isOk {
			retryDelayInterval = v
		}
	case 4:
		if v, isOk := args[0].(int); isOk {
			maxRetries = v
		}
		if v, isOk := args[1].(int); isOk {
			retryDelayBase = v
		}
		if v, isOk := args[2].(int); isOk {
			retryDelayInterval = v
		}
		if v, isOk := args[3].(int); isOk {
			maxTime = v
		}
	default:
		if v, isOk := args[0].(int); isOk {
			maxRetries = v
		}
		if v, isOk := args[1].(int); isOk {
			retryDelayBase = v
		}
		if v, isOk := args[2].(int); isOk {
			retryDelayInterval = v
		}
		if v, isOk := args[3].(int); isOk {
			maxTime = v
		}
		newArgs = args[4:]
	}
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute*time.Duration(maxTime))
	defer cancel()
	return retry(ctx, maxRetries, retryDelayBase, retryDelayInterval, operationFn, successFn, newArgs...)
}
