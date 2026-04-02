package k

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"
)

// -------------------- NonRetryableError --------------------

func TestNonRetryable_ErrorMessage(t *testing.T) {
	inner := errors.New("权限不足")
	err := NonRetryable(inner)
	want := "不可重试错误: 权限不足"
	if err.Error() != want {
		t.Errorf("got %q, want %q", err.Error(), want)
	}
}

func TestNonRetryable_Unwrap(t *testing.T) {
	inner := errors.New("原始错误")
	err := NonRetryable(inner)
	if !errors.Is(err, inner) {
		t.Error("errors.Is 应能透过 NonRetryableError 找到内层错误")
	}
}

func TestIsNonRetryable_True(t *testing.T) {
	err := NonRetryable(errors.New("x"))
	if !IsNonRetryable(err) {
		t.Error("应识别为不可重试错误")
	}
}

func TestIsNonRetryable_False(t *testing.T) {
	err := errors.New("普通错误")
	if IsNonRetryable(err) {
		t.Error("普通错误不应识别为不可重试错误")
	}
}

func TestIsNonRetryable_Nil(t *testing.T) {
	if IsNonRetryable(nil) {
		t.Error("nil 不应识别为不可重试错误")
	}
}

// -------------------- Retry 成功路径 --------------------

func TestRetry_SuccessOnFirstAttempt(t *testing.T) {
	called := 0
	successCalled := false

	err := Retry(
		func(args ...any) (any, error) {
			called++
			return "ok", nil
		},
		func(data any) { successCalled = true },
	)

	fmt.Println("错误", err, "called数字", called, "状态", successCalled)
	if err != nil {
		t.Fatalf("期望无错误，got %v", err)
	}
	if called != 1 {
		t.Errorf("期望调用 1 次，got %d", called)
	}
	if !successCalled {
		t.Error("successFn 应被调用")
	}
}

func TestRetry_SuccessOnSecondAttempt(t *testing.T) {
	called := 0

	err := Retry(
		func(args ...any) (any, error) {
			called++
			if called < 2 {
				return nil, errors.New("暂时失败")
			}
			return "ok", nil
		},
		func(data any) {},
		WithMaxRetries(3),
		WithRetryDelay(10*time.Millisecond),
	)

	if err != nil {
		t.Fatalf("期望成功，got %v", err)
	}
	if called != 2 {
		t.Errorf("期望调用 2 次，got %d", called)
	}
}

func TestRetry_SuccessDataPassedToSuccessFn(t *testing.T) {
	var received any

	_ = Retry(
		func(args ...any) (any, error) { return 42, nil },
		func(data any) { received = data },
	)

	if received != 42 {
		t.Errorf("successFn 应收到 42，got %v", received)
	}
}

// -------------------- Retry 失败路径 --------------------

func TestRetry_ExhaustsMaxRetries(t *testing.T) {
	called := 0

	err := Retry(
		func(args ...any) (any, error) {
			called++
			return nil, errors.New("一直失败")
		},
		func(data any) {},
		WithMaxRetries(3),
		WithRetryDelay(10*time.Millisecond),
	)

	if err == nil {
		t.Fatal("期望返回错误")
	}
	if called != 3 {
		t.Errorf("期望调用 3 次，got %d", called)
	}
}

func TestRetry_MaxRetriesOne(t *testing.T) {
	called := 0

	err := Retry(
		func(args ...any) (any, error) {
			called++
			return nil, errors.New("失败")
		},
		func(data any) {},
		WithMaxRetries(1),
	)

	if err == nil {
		t.Fatal("期望返回错误")
	}
	if called != 1 {
		t.Errorf("期望调用 1 次，got %d", called)
	}
}

// -------------------- 不可重试错误 --------------------

func TestRetry_StopsOnNonRetryableError(t *testing.T) {
	called := 0

	err := Retry(
		func(args ...any) (any, error) {
			called++
			return nil, NonRetryable(errors.New("认证失败"))
		},
		func(data any) {},
		WithMaxRetries(5),
		WithRetryDelay(10*time.Millisecond),
	)

	if err == nil {
		t.Fatal("期望返回错误")
	}
	if called != 1 {
		t.Errorf("遇到不可重试错误应立即终止，期望调用 1 次，got %d", called)
	}
	if !IsNonRetryable(err) {
		t.Error("返回的错误应为 NonRetryableError")
	}
}

func TestRetry_NonRetryablePreservesInnerError(t *testing.T) {
	inner := errors.New("原始原因")

	err := Retry(
		func(args ...any) (any, error) {
			return nil, NonRetryable(inner)
		},
		func(data any) {},
		WithMaxRetries(3),
	)

	if !errors.Is(err, inner) {
		t.Error("应能通过 errors.Is 找到内层原始错误")
	}
}

func TestRetry_NonRetryableOnSecondAttempt(t *testing.T) {
	called := 0

	err := Retry(
		func(args ...any) (any, error) {
			called++
			if called == 1 {
				return nil, errors.New("可重试错误")
			}
			return nil, NonRetryable(errors.New("第二次遇到不可重试"))
		},
		func(data any) {},
		WithMaxRetries(5),
		WithRetryDelay(10*time.Millisecond),
	)

	if err == nil {
		t.Fatal("期望返回错误")
	}
	if called != 2 {
		t.Errorf("期望调用 2 次后终止，got %d", called)
	}
}

// -------------------- 延迟与退避 --------------------

func TestRetry_ExponentialBackoff(t *testing.T) {
	const base = 50 * time.Millisecond
	attempts := 0
	timestamps := make([]time.Time, 0, 3)

	_ = Retry(
		func(args ...any) (any, error) {
			timestamps = append(timestamps, time.Now())
			attempts++
			return nil, errors.New("失败")
		},
		func(data any) {},
		WithMaxRetries(3),
		WithRetryDelay(base),
		WithDelayMultiplier(2.0),
	)

	// 检查间隔大致符合指数退避（允许 20ms 误差）
	if len(timestamps) < 3 {
		t.Fatalf("期望 3 次调用，got %d", len(timestamps))
	}

	gap1 := timestamps[1].Sub(timestamps[0])
	gap2 := timestamps[2].Sub(timestamps[1])

	if gap1 < base-20*time.Millisecond {
		t.Errorf("第一次间隔 %v 小于预期 %v", gap1, base)
	}
	if gap2 < base*2-20*time.Millisecond {
		t.Errorf("第二次间隔 %v 小于预期 %v（应约为 2x base）", gap2, base*2)
	}
}

func TestRetry_NoGrowthWithMultiplierOne(t *testing.T) {
	const base = 30 * time.Millisecond
	timestamps := make([]time.Time, 0, 3)

	_ = Retry(
		func(args ...any) (any, error) {
			timestamps = append(timestamps, time.Now())
			return nil, errors.New("失败")
		},
		func(data any) {},
		WithMaxRetries(3),
		WithRetryDelay(base),
		WithDelayMultiplier(1.0),
	)

	gap1 := timestamps[1].Sub(timestamps[0])
	gap2 := timestamps[2].Sub(timestamps[1])
	diff := gap2 - gap1
	if diff < 0 {
		diff = -diff
	}
	if diff > 20*time.Millisecond {
		t.Errorf("间隔增长倍数为 1.0 时，两次间隔应相近，gap1=%v gap2=%v", gap1, gap2)
	}
}

// -------------------- Context 超时 --------------------

func TestRetry_MaxTimeExceeded(t *testing.T) {
	err := Retry(
		func(args ...any) (any, error) {
			time.Sleep(50 * time.Millisecond)
			return nil, errors.New("失败")
		},
		func(data any) {},
		WithMaxRetries(10),
		WithRetryDelay(10*time.Millisecond),
		WithMaxTime(80*time.Millisecond),
	)

	if err == nil {
		t.Fatal("期望因超时返回错误")
	}
	if !errors.Is(err, context.DeadlineExceeded) {
		t.Errorf("期望 DeadlineExceeded，got %v", err)
	}
}

func TestRetryWithContext_CancelledContext(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // 立即取消

	called := 0
	err := RetryWithContext(
		ctx,
		func(args ...any) (any, error) {
			called++
			return nil, errors.New("失败")
		},
		func(data any) {},
		WithMaxRetries(5),
	)

	if err == nil {
		t.Fatal("期望返回错误")
	}
	if !errors.Is(err, context.Canceled) {
		t.Errorf("期望 context.Canceled，got %v", err)
	}
	if called > 1 {
		t.Errorf("context 已取消时最多执行 1 次，got %d", called)
	}
}

func TestRetryWithContext_CancelDuringWait(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())

	called := 0
	err := RetryWithContext(
		ctx,
		func(args ...any) (any, error) {
			called++
			if called == 1 {
				// 第一次失败后，在等待期间取消
				go func() { time.Sleep(20 * time.Millisecond); cancel() }()
			}
			return nil, errors.New("失败")
		},
		func(data any) {},
		WithMaxRetries(5),
		WithRetryDelay(500*time.Millisecond),
	)

	if err == nil {
		t.Fatal("期望返回错误")
	}
	if !errors.Is(err, context.Canceled) {
		t.Errorf("期望等待期间捕获 context.Canceled，got %v", err)
	}
}

func TestRetryWithContext_SuccessWithValidContext(t *testing.T) {
	ctx := context.Background()
	successCalled := false

	err := RetryWithContext(
		ctx,
		func(args ...any) (any, error) { return "data", nil },
		func(data any) { successCalled = true },
	)

	if err != nil {
		t.Fatalf("期望成功，got %v", err)
	}
	if !successCalled {
		t.Error("successFn 应被调用")
	}
}

// -------------------- Option 配置 --------------------

func TestWithMaxRetries(t *testing.T) {
	opts := defaultOptions()
	WithMaxRetries(7)(opts)
	if opts.MaxRetries != 7 {
		t.Errorf("got %d, want 7", opts.MaxRetries)
	}
}

func TestWithRetryDelay(t *testing.T) {
	opts := defaultOptions()
	WithRetryDelay(3 * time.Second)(opts)
	if opts.RetryDelayBase != 3*time.Second {
		t.Errorf("got %v, want 3s", opts.RetryDelayBase)
	}
}

func TestWithDelayMultiplier(t *testing.T) {
	opts := defaultOptions()
	WithDelayMultiplier(1.5)(opts)
	if opts.RetryDelayInterval != 1.5 {
		t.Errorf("got %v, want 1.5", opts.RetryDelayInterval)
	}
}

func TestWithMaxTime(t *testing.T) {
	opts := defaultOptions()
	WithMaxTime(30 * time.Second)(opts)
	if opts.MaxTime != 30*time.Second {
		t.Errorf("got %v, want 30s", opts.MaxTime)
	}
}

func TestDefaultOptions(t *testing.T) {
	opts := defaultOptions()
	if opts.MaxRetries != 5 {
		t.Errorf("MaxRetries: got %d, want 5", opts.MaxRetries)
	}
	if opts.RetryDelayBase != time.Second {
		t.Errorf("RetryDelayBase: got %v, want 1s", opts.RetryDelayBase)
	}
	if opts.RetryDelayInterval != 1.0 {
		t.Errorf("RetryDelayInterval: got %v, want 1.0", opts.RetryDelayInterval)
	}
	if opts.MaxTime != 10*time.Minute {
		t.Errorf("MaxTime: got %v, want 10m", opts.MaxTime)
	}
}

// -------------------- 边界与组合 --------------------

func TestRetry_SuccessFnReceivesCorrectType(t *testing.T) {
	type Result struct{ Value string }
	var got any

	_ = Retry(
		func(args ...any) (any, error) { return Result{"hello"}, nil },
		func(data any) { got = data },
	)

	r, ok := got.(Result)
	if !ok || r.Value != "hello" {
		t.Errorf("successFn 收到的值不符，got %v", got)
	}
}

func TestRetry_ErrorMessageOnExhaustion(t *testing.T) {
	err := Retry(
		func(args ...any) (any, error) { return nil, errors.New("x") },
		func(data any) {},
		WithMaxRetries(2),
		WithRetryDelay(5*time.Millisecond),
	)
	want := "已达最大重试次数，执行失败"
	if err == nil || err.Error() != want {
		t.Errorf("got %v, want %q", err, want)
	}
}

func TestRetry_MultipleOptions(t *testing.T) {
	called := 0

	start := time.Now()
	_ = Retry(
		func(args ...any) (any, error) {
			called++
			return nil, fmt.Errorf("失败 %d", called)
		},
		func(data any) {},
		WithMaxRetries(3),
		WithRetryDelay(20*time.Millisecond),
		WithDelayMultiplier(1.0),
		WithMaxTime(5*time.Second),
	)
	elapsed := time.Since(start)

	if called != 3 {
		t.Errorf("期望 3 次，got %d", called)
	}
	// 3 次调用，2 次等待，每次 20ms，总耗时应在 40ms~500ms 之间
	if elapsed < 40*time.Millisecond || elapsed > 500*time.Millisecond {
		t.Errorf("耗时异常: %v", elapsed)
	}
}
