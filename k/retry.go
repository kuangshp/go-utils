package k

import (
	"context"
	"errors"
	"fmt"
	"time"
)

// RetryOptions 重试配置
type RetryOptions struct {
	MaxRetries         int
	RetryDelayBase     time.Duration
	RetryDelayInterval float64
	MaxTime            time.Duration
	Logger             func(format string, args ...any)
}

type Option func(*RetryOptions)

func WithMaxRetries(n int) Option {
	return func(o *RetryOptions) { o.MaxRetries = n }
}

func WithRetryDelay(base time.Duration) Option {
	return func(o *RetryOptions) { o.RetryDelayBase = base }
}

func WithDelayMultiplier(multiplier float64) Option {
	return func(o *RetryOptions) { o.RetryDelayInterval = multiplier }
}

func WithMaxTime(d time.Duration) Option {
	return func(o *RetryOptions) { o.MaxTime = d }
}

// WithLogger 注入可选日志回调；默认不输出到 stdout。
func WithLogger(logger func(format string, args ...any)) Option {
	return func(o *RetryOptions) { o.Logger = logger }
}

func defaultOptions() *RetryOptions {
	return &RetryOptions{
		MaxRetries:         5,
		RetryDelayBase:     time.Second,
		RetryDelayInterval: 1.0,
		MaxTime:            10 * time.Minute,
	}
}

// NonRetryableError 包装不需要重试的错误
type NonRetryableError struct {
	Err error
}

func (e *NonRetryableError) Error() string {
	return fmt.Sprintf("不可重试错误: %v", e.Err)
}

func (e *NonRetryableError) Unwrap() error {
	return e.Err
}

// NonRetryable 将错误包装为不可重试错误
func NonRetryable(err error) error {
	return &NonRetryableError{Err: err}
}

// IsNonRetryable 判断是否为不可重试错误
func IsNonRetryable(err error) bool {
	var e *NonRetryableError
	return errors.As(err, &e)
}

// Retry 重试执行，支持 Option 配置和 context 超时控制
//
// 示例：
//
//	Retry(
//	    func(args ...any) (any, error) { ... },
//	    func(data any) { ... },
//	    WithMaxRetries(3),
//	    WithRetryDelay(2*time.Second),
//	    WithDelayMultiplier(2.0),
//	    WithLogger(log.Printf),
//	)
//
// operationFn 重试的方法,successFn 成功的方法
func Retry(
	operationFn func(args ...any) (any, error),
	successFn func(data any),
	opts ...Option,
) error {
	options := defaultOptions()
	for _, opt := range opts {
		opt(options)
	}

	ctx, cancel := context.WithTimeout(context.Background(), options.MaxTime)
	defer cancel()

	return retry(ctx, options, operationFn, successFn)
}

func RetryWithContext(
	ctx context.Context,
	operationFn func(args ...any) (any, error),
	successFn func(data any),
	opts ...Option,
) error {
	options := defaultOptions()
	for _, opt := range opts {
		opt(options)
	}
	return retry(ctx, options, operationFn, successFn)
}

func retry(
	ctx context.Context,
	opts *RetryOptions,
	operationFn func(args ...any) (any, error),
	successFn func(data any),
) error {
	delay := opts.RetryDelayBase
	var lastErr error

	for attempt := 1; attempt <= opts.MaxRetries; attempt++ {
		if opts.Logger != nil {
			opts.Logger("重试次数: %d, 时间: %s", attempt, time.Now().Format(time.DateTime))
		}

		select {
		case <-ctx.Done():
			return fmt.Errorf("重试已取消: %w", ctx.Err())
		default:
		}

		data, err := operationFn()
		if err == nil {
			successFn(data)
			return nil
		}
		lastErr = err
		if IsNonRetryable(err) {
			return err
		}

		if attempt < opts.MaxRetries {
			select {
			case <-ctx.Done():
				return fmt.Errorf("等待期间已超时: %w", ctx.Err())
			case <-time.After(delay):
			}
			delay = time.Duration(float64(delay) * opts.RetryDelayInterval)
		}
	}

	if lastErr != nil {
		return fmt.Errorf("已达最大重试次数，执行失败: %w", lastErr)
	}
	return errors.New("已达最大重试次数，执行失败")
}
