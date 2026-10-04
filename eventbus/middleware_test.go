// Package eventbus_test 针对 pkg/eventbus 的中间件做黑盒测试。
//
// 本文件覆盖 middleware.go：
//   - LoggingMiddleware：透传结果（成功与失败两种路径）；
//   - RecoveryMiddleware：捕获内层 panic 并转为 PanicError；
//   - TimeoutMiddleware：内层超时返回 TimeoutError、及时完成则透传；
//   - RetryMiddleware：失败重试至上限、中途成功即返回；
//   - MetricsMiddleware：结果透传；
//   - Chain：按序组合多个中间件（首个参数最外层）。
//
// 中间件链在本文件中直接以 Handle 调用验证，不经过总线。
package eventbus_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	bLogger "github.com/tx7do/kratos-bootstrap/logger"

	"github.com/tx7do/go-utils/eventbus"
)

// quietLogger 返回静默日志助手，避免测试输出被中间件日志淹没。
func quietLogger() *bLogger.Helper {
	return bLogger.NewHelper(bLogger.NopLogger())
}

// TestLoggingMiddleware_PassesThroughResult 日志中间件不得改变内层结果：
// 成功路径返回 nil，失败路径原样透传错误。
func TestLoggingMiddleware_PassesThroughResult(t *testing.T) {
	mw := eventbus.LoggingMiddleware(quietLogger())
	ev := eventbus.NewEvent("log.me", nil)

	succeed := mw(eventbus.EventHandlerFunc(func(context.Context, *eventbus.Event) error { return nil }))
	require.NoError(t, succeed.Handle(context.Background(), ev))

	failing := mw(eventbus.EventHandlerFunc(func(context.Context, *eventbus.Event) error { return assert.AnError }))
	assert.ErrorIs(t, failing.Handle(context.Background(), ev), assert.AnError)
}

// TestRecoveryMiddleware_ConvertsPanicToError 内层 panic 必须被捕获并转换为
// PanicError，错误消息固定；无 panic 时结果透传。
func TestRecoveryMiddleware_ConvertsPanicToError(t *testing.T) {
	mw := eventbus.RecoveryMiddleware(quietLogger())
	ev := eventbus.NewEvent("panic.me", nil)

	panicking := mw(eventbus.EventHandlerFunc(func(context.Context, *eventbus.Event) error {
		panic("boom")
	}))
	err := panicking.Handle(context.Background(), ev)
	require.Error(t, err, "panic 必须转为错误返回")
	var panicErr *eventbus.PanicError
	require.True(t, errors.As(err, &panicErr), "必须是 PanicError 类型")
	assert.Equal(t, "panic in event handler", panicErr.Error())
	assert.Equal(t, "boom", panicErr.Value)

	calm := mw(eventbus.EventHandlerFunc(func(context.Context, *eventbus.Event) error { return assert.AnError }))
	assert.ErrorIs(t, calm.Handle(context.Background(), ev), assert.AnError,
		"无 panic 时原错误必须透传而非吞掉")
}

// TestTimeoutMiddleware_TimesOutSlowHandler 内层慢于阈值时返回 TimeoutError，
// 错误消息固定且携带事件 ID。
func TestTimeoutMiddleware_TimesOutSlowHandler(t *testing.T) {
	mw := eventbus.TimeoutMiddleware(50 * time.Millisecond)
	ev := eventbus.NewEvent("too.slow", nil)

	slow := mw(eventbus.EventHandlerFunc(func(context.Context, *eventbus.Event) error {
		time.Sleep(500 * time.Millisecond)
		return nil
	}))
	start := time.Now()
	err := slow.Handle(context.Background(), ev)
	assert.Less(t, time.Since(start), 400*time.Millisecond, "超时必须在阈值附近返回而非等满内层")
	require.Error(t, err)
	var timeoutErr *eventbus.TimeoutError
	require.True(t, errors.As(err, &timeoutErr), "必须是 TimeoutError 类型")
	assert.Equal(t, "event handling timeout", timeoutErr.Error())
	assert.Equal(t, ev.ID, timeoutErr.EventID)
	assert.Equal(t, "too.slow", timeoutErr.EventType)
	assert.Equal(t, 50*time.Millisecond, timeoutErr.Timeout)
}

// TestTimeoutMiddleware_PassesThroughFastHandler 内层快于阈值时结果透传。
func TestTimeoutMiddleware_PassesThroughFastHandler(t *testing.T) {
	mw := eventbus.TimeoutMiddleware(time.Second)
	fast := mw(eventbus.EventHandlerFunc(func(context.Context, *eventbus.Event) error { return nil }))
	require.NoError(t, fast.Handle(context.Background(), eventbus.NewEvent("fast", nil)))
}

// TestRetryMiddleware_ExhaustsRetries 持续失败时按上限重试并返回最后的错误。
func TestRetryMiddleware_ExhaustsRetries(t *testing.T) {
	attempts := 0
	var mw eventbus.Middleware = eventbus.RetryMiddleware(2, time.Millisecond)
	handler := mw(eventbus.EventHandlerFunc(func(context.Context, *eventbus.Event) error {
		attempts++
		return assert.AnError
	}))

	err := handler.Handle(context.Background(), eventbus.NewEvent("always.fails", nil))
	assert.ErrorIs(t, err, assert.AnError, "重试耗尽后必须返回错误")
	assert.Equal(t, 3, attempts, "maxRetries=2 意味着共尝试 1+2 次")
}

// TestRetryMiddleware_StopsOnSuccess 中途成功即停止重试，返回 nil。
func TestRetryMiddleware_StopsOnSuccess(t *testing.T) {
	attempts := 0
	mw := eventbus.RetryMiddleware(5, time.Millisecond)
	handler := mw(eventbus.EventHandlerFunc(func(context.Context, *eventbus.Event) error {
		attempts++
		if attempts < 2 {
			return assert.AnError
		}
		return nil
	}))

	require.NoError(t, handler.Handle(context.Background(), eventbus.NewEvent("succeeds.second", nil)))
	assert.Equal(t, 2, attempts, "第二次成功后必须停止重试")
}

// TestMetricsMiddleware_PassesThroughResult 指标中间件只观测不改写结果。
func TestMetricsMiddleware_PassesThroughResult(t *testing.T) {
	mw := eventbus.MetricsMiddleware(quietLogger())

	succeed := mw(eventbus.EventHandlerFunc(func(context.Context, *eventbus.Event) error { return nil }))
	require.NoError(t, succeed.Handle(context.Background(), eventbus.NewEvent("ok", nil)))

	failing := mw(eventbus.EventHandlerFunc(func(context.Context, *eventbus.Event) error { return assert.AnError }))
	assert.ErrorIs(t, failing.Handle(context.Background(), eventbus.NewEvent("err", nil)), assert.AnError)
}

// TestChain_ComposesInOrder Chain 按序组合：首参数为最外层。
// 此处 Recovery 首参在最外层，内层 panic 应被转换为 PanicError；
// 调换顺序后（Logging 最外、Recovery 内层）panic 将穿透到调用方。
func TestChain_ComposesInOrder(t *testing.T) {
	t.Run("recovery outermost catches panic", func(t *testing.T) {
		wrapped := eventbus.Chain(
			eventbus.RecoveryMiddleware(quietLogger()),
			eventbus.LoggingMiddleware(quietLogger()),
		)(eventbus.EventHandlerFunc(func(context.Context, *eventbus.Event) error {
			panic("inner boom")
		}))

		err := wrapped.Handle(context.Background(), eventbus.NewEvent("chain.panic", nil))
		var panicErr *eventbus.PanicError
		require.Error(t, err)
		require.True(t, errors.As(err, &panicErr), "Recovery 在最外层时必须兜住 panic")
	})
}
