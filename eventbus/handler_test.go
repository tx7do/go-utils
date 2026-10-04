// Package eventbus_test 针对 pkg/eventbus 的处理器组合原语做黑盒测试。
//
// 本文件覆盖 handler.go：
//   - EventHandlerFunc 函数适配器；
//   - AsyncHandler 异步包装（立即返回、后台执行）；
//   - ChainHandler 顺序执行、首个错误即短路；
//   - FilterHandler 按谓词放行/拦截。
package eventbus_test

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/tx7do/go-utils/eventbus"
)

// TestEventHandlerFunc_AdaptsFunction 适配器必须把函数调用原样转发，
// 返回值一并透传。
func TestEventHandlerFunc_AdaptsFunction(t *testing.T) {
	called := false
	f := eventbus.EventHandlerFunc(func(ctx context.Context, ev *eventbus.Event) error {
		called = true
		assert.Equal(t, "adapted", ev.Type)
		return assert.AnError
	})

	err := f.Handle(context.Background(), eventbus.NewEvent("adapted", nil))
	assert.True(t, called, "适配器必须转发到内层函数")
	assert.ErrorIs(t, err, assert.AnError, "返回值必须原样透传")
}

// TestAsyncHandler_ReturnsImmediatelyAndRunsInBackground 异步包装的 Handle
// 必须立即返回 nil，内层处理器在后台 goroutine 执行。
func TestAsyncHandler_ReturnsImmediatelyAndRunsInBackground(t *testing.T) {
	done := make(chan struct{}, 1)
	inner := eventbus.EventHandlerFunc(func(context.Context, *eventbus.Event) error {
		done <- struct{}{}
		return nil
	})
	async := eventbus.NewAsyncHandler(inner)

	start := time.Now()
	require.NoError(t, async.Handle(context.Background(), eventbus.NewEvent("async", nil)))
	assert.Less(t, time.Since(start), time.Second, "异步包装必须立即返回，不等待内层执行")

	select {
	case <-done:
		// 后台执行已观察到。
	case <-time.After(2 * time.Second):
		t.Fatal("内层处理器未在超时窗口内于后台执行")
	}
}

// TestChainHandler_RunsAllInOrder 链式处理器按序执行全部成员。
func TestChainHandler_RunsAllInOrder(t *testing.T) {
	events := make(chan string, 3)
	mk := func(tag string) eventbus.Handler {
		return eventbus.EventHandlerFunc(func(context.Context, *eventbus.Event) error {
			events <- tag
			return nil
		})
	}
	chain := eventbus.NewChainHandler(mk("first"), mk("second"), mk("third"))

	require.NoError(t, chain.Handle(context.Background(), eventbus.NewEvent("chain", nil)))

	var order []string
	for i := 0; i < 3; i++ {
		select {
		case tag := <-events:
			order = append(order, tag)
		case <-time.After(2 * time.Second):
			t.Fatal("链式成员未全部执行")
		}
	}
	assert.Equal(t, []string{"first", "second", "third"}, order, "链式处理器必须按声明顺序执行")
}

// TestChainHandler_ShortCircuitsOnFirstError 任一成员出错即短路：
// 后续成员不执行，错误向上透传。
func TestChainHandler_ShortCircuitsOnFirstError(t *testing.T) {
	secondCalled := false
	chain := eventbus.NewChainHandler(
		eventbus.EventHandlerFunc(func(context.Context, *eventbus.Event) error {
			return assert.AnError
		}),
		eventbus.EventHandlerFunc(func(context.Context, *eventbus.Event) error {
			secondCalled = true
			return nil
		}),
	)

	err := chain.Handle(context.Background(), eventbus.NewEvent("chain.err", nil))
	assert.ErrorIs(t, err, assert.AnError, "成员错误必须透传")
	assert.False(t, secondCalled, "首个错误后必须短路，后续成员不得执行")
}

// TestFilterHandler_FilterDecidesInvocation 谓词为真才调用内层处理器，
// 为假直接返回 nil 且不触达内层。
func TestFilterHandler_FilterDecidesInvocation(t *testing.T) {
	t.Run("filter passes", func(t *testing.T) {
		invoked := false
		f := eventbus.NewFilterHandler(
			func(*eventbus.Event) bool { return true },
			eventbus.EventHandlerFunc(func(context.Context, *eventbus.Event) error {
				invoked = true
				return nil
			}),
		)
		require.NoError(t, f.Handle(context.Background(), eventbus.NewEvent("pass", nil)))
		assert.True(t, invoked, "谓词为真时内层必须被调用")
	})

	t.Run("filter blocks", func(t *testing.T) {
		invoked := false
		f := eventbus.NewFilterHandler(
			func(*eventbus.Event) bool { return false },
			eventbus.EventHandlerFunc(func(context.Context, *eventbus.Event) error {
				invoked = true
				return nil
			}),
		)
		require.NoError(t, f.Handle(context.Background(), eventbus.NewEvent("blocked", nil)),
			"谓词为假时必须静默返回 nil")
		assert.False(t, invoked, "谓词为假时内层不得被调用")
	})
}
