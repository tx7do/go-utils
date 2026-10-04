// Package eventbus_test 针对 pkg/eventbus 的 DefaultEventBus 做黑盒测试。
//
// 本文件覆盖总线核心生命周期：订阅/发布闭环（同步与异步）、
// 单次订阅的自动退订、取消订阅的精确移除、未知主题发布、
// 处理器错误不影响发布结果、订阅计数与事件类型枚举、
// 总线关闭后的订阅/发布拒绝与重复关闭报错。
//
// 所有等待都用 channel + select + time.After 的带超时模式，
// 杜绝裸 sleep 与死锁挂死。
package eventbus_test

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	bLogger "github.com/tx7do/kratos-bootstrap/logger"

	"github.com/tx7do/go-utils/eventbus"
)

// waitEvent 在超时窗口内等待事件抵达；超时则测试失败（防死锁挂死）。
func waitEvent(t *testing.T, ch chan *eventbus.Event) *eventbus.Event {
	t.Helper()
	select {
	case ev := <-ch:
		return ev
	case <-time.After(2 * time.Second):
		t.Fatal("超时窗口内未收到事件：处理器未执行或订阅链路断裂")
		return nil
	}
}

// asDefaultBus 把 EventBus 接口断言回具体类型，
// 以便调用仅存在于 DefaultEventBus 上的计数/枚举方法。
func asDefaultBus(t *testing.T, bus eventbus.EventBus) *eventbus.DefaultEventBus {
	t.Helper()
	db, ok := bus.(*eventbus.DefaultEventBus)
	if !ok {
		t.Fatal("测试总线必须是 *DefaultEventBus")
	}
	return db
}

// expectNoEvent 断言在短窗口内不再有事件抵达（用于退订/单次订阅移除后的负向验证）。
func expectNoEvent(t *testing.T, ch chan *eventbus.Event) {
	t.Helper()
	select {
	case ev := <-ch:
		t.Fatalf("不应再收到事件，却收到: %+v", ev)
	case <-time.After(200 * time.Millisecond):
		// 预期路径：静默通过。
	}
}

// TestSubscribeAndPublish_SyncDelivery 同步订阅-发布闭环：
// 处理器收到完整事件对象（类型、数据、来源均不丢失）。
func TestSubscribeAndPublish_SyncDelivery(t *testing.T) {
	bus := eventbus.NewEventBus(bLogger.NopLogger())
	defer func() { require.NoError(t, bus.Close()) }()

	received := make(chan *eventbus.Event, 1)
	require.NoError(t, bus.Subscribe("unit.test", eventbus.EventHandlerFunc(
		func(_ context.Context, event *eventbus.Event) error {
			received <- event
			return nil
		})))

	require.NoError(t, bus.Publish(context.Background(),
		eventbus.NewEvent("unit.test", map[string]any{"payload": 1}).WithSource("unit-source")))

	ev := waitEvent(t, received)
	assert.Equal(t, "unit.test", ev.Type)
	assert.Equal(t, "unit-source", ev.Source)
	assert.Equal(t, map[string]any{"payload": 1}, ev.Data)
}

// TestPublish_AllHandlersReceive 同一主题的多个订阅者都必须收到事件。
func TestPublish_AllHandlersReceive(t *testing.T) {
	bus := eventbus.NewEventBus(bLogger.NopLogger())
	defer func() { require.NoError(t, bus.Close()) }()

	const subs = 3
	chans := make([]chan *eventbus.Event, subs)
	for i := 0; i < subs; i++ {
		ch := make(chan *eventbus.Event, 1)
		chans[i] = ch
		require.NoError(t, bus.Subscribe("fanout", eventbus.EventHandlerFunc(
			func(_ context.Context, event *eventbus.Event) error {
				ch <- event
				return nil
			})))
	}

	require.NoError(t, bus.Publish(context.Background(), eventbus.NewEvent("fanout", nil)))

	for i := 0; i < subs; i++ {
		ev := waitEvent(t, chans[i])
		assert.Equal(t, "fanout", ev.Type)
	}
}

// TestPublish_UnknownTopicNoHandlers 未知主题（无订阅者）发布应返回 nil，
// 不报错、不触发任何处理器。
func TestPublish_UnknownTopicNoHandlers(t *testing.T) {
	bus := eventbus.NewEventBus(bLogger.NopLogger())
	defer func() { require.NoError(t, bus.Close()) }()

	assert.NoError(t, bus.Publish(context.Background(), eventbus.NewEvent("nobody-listens", nil)))
}

// TestPublish_HandlerErrorDoesNotFailPublish 处理器返回错误时发布本身仍返回 nil
// （错误仅记录日志），且其它处理器继续执行。
func TestPublish_HandlerErrorDoesNotFailPublish(t *testing.T) {
	bus := eventbus.NewEventBus(bLogger.NopLogger())
	defer func() { require.NoError(t, bus.Close()) }()

	require.NoError(t, bus.Subscribe("flaky", eventbus.EventHandlerFunc(
		func(_ context.Context, _ *eventbus.Event) error {
			return assert.AnError
		})))

	received := make(chan *eventbus.Event, 1)
	require.NoError(t, bus.Subscribe("flaky", eventbus.EventHandlerFunc(
		func(_ context.Context, event *eventbus.Event) error {
			received <- event
			return nil
		})))

	assert.NoError(t, bus.Publish(context.Background(), eventbus.NewEvent("flaky", nil)),
		"单个处理器失败不得导致发布失败")
	waitEvent(t, received) // 第二个处理器仍须执行
}

// TestUnsubscribe_StopsDelivery 取消订阅后不再收到该主题事件；
// 重复取消同一处理器报错。
func TestUnsubscribe_StopsDelivery(t *testing.T) {
	bus := eventbus.NewEventBus(bLogger.NopLogger())
	defer func() { require.NoError(t, bus.Close()) }()

	received := make(chan *eventbus.Event, 1)
	handler := eventbus.EventHandlerFunc(func(_ context.Context, event *eventbus.Event) error {
		received <- event
		return nil
	})
	require.NoError(t, bus.Subscribe("gone", handler))
	require.Equal(t, 1, asDefaultBus(t, bus).GetSubscriberCount("gone"))

	require.NoError(t, bus.Unsubscribe("gone", handler))
	require.Equal(t, 0, asDefaultBus(t, bus).GetSubscriberCount("gone"))

	require.NoError(t, bus.Publish(context.Background(), eventbus.NewEvent("gone", nil)))
	expectNoEvent(t, received)

	err := bus.Unsubscribe("gone", handler)
	require.Error(t, err, "已移除的处理器再次退订必须报错")
	assert.Contains(t, err.Error(), "handler not found")
}

// TestUnsubscribe_UnknownTopicOrHandler 未知主题、未注册处理器的退订均报错。
func TestUnsubscribe_UnknownTopicOrHandler(t *testing.T) {
	bus := eventbus.NewEventBus(bLogger.NopLogger())
	defer func() { require.NoError(t, bus.Close()) }()

	never := eventbus.EventHandlerFunc(func(context.Context, *eventbus.Event) error { return nil })

	err := bus.Unsubscribe("ghost-topic", never)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "handler not found")
}

// TestUnsubscribe_RemovesSingleRegistration 同一处理器注册两次时一次退订只移除一份，
// 剩余注册继续有效。
func TestUnsubscribe_RemovesSingleRegistration(t *testing.T) {
	bus := eventbus.NewEventBus(bLogger.NopLogger())
	defer func() { require.NoError(t, bus.Close()) }()

	handler := eventbus.EventHandlerFunc(func(context.Context, *eventbus.Event) error { return nil })
	require.NoError(t, bus.Subscribe("dup", handler))
	require.NoError(t, bus.Subscribe("dup", handler))
	require.Equal(t, 2, asDefaultBus(t, bus).GetSubscriberCount("dup"))

	require.NoError(t, bus.Unsubscribe("dup", handler))
	assert.Equal(t, 1, asDefaultBus(t, bus).GetSubscriberCount("dup"), "一次退订只移除一份注册")

	require.NoError(t, bus.Unsubscribe("dup", handler))
	assert.Equal(t, 0, asDefaultBus(t, bus).GetSubscriberCount("dup"))
}

// TestSubscribeAsync_AsyncDelivery SubscribeAsync 注册的处理器在独立 goroutine 中执行，
// 发布后仍能在超时窗口内收到事件。
func TestSubscribeAsync_AsyncDelivery(t *testing.T) {
	bus := eventbus.NewEventBus(bLogger.NopLogger())
	defer func() { require.NoError(t, bus.Close()) }()

	received := make(chan *eventbus.Event, 1)
	require.NoError(t, bus.SubscribeAsync("async.topic", eventbus.EventHandlerFunc(
		func(_ context.Context, event *eventbus.Event) error {
			received <- event
			return nil
		})))
	require.Equal(t, 1, asDefaultBus(t, bus).GetSubscriberCount("async.topic"))

	require.NoError(t, bus.Publish(context.Background(), eventbus.NewEvent("async.topic", nil)))
	ev := waitEvent(t, received)
	assert.Equal(t, "async.topic", ev.Type)
}

// TestSubscribeOnce_FiresOnceAndAutoRemoved 单次订阅只在首次发布时执行，
// 随后自动移除（计数归零、后续发布不再触达）。
func TestSubscribeOnce_FiresOnceAndAutoRemoved(t *testing.T) {
	bus := eventbus.NewEventBus(bLogger.NopLogger())
	defer func() { require.NoError(t, bus.Close()) }()

	received := make(chan *eventbus.Event, 1)
	require.NoError(t, bus.SubscribeOnce("once.topic", eventbus.EventHandlerFunc(
		func(_ context.Context, event *eventbus.Event) error {
			received <- event
			return nil
		})))
	require.Equal(t, 1, asDefaultBus(t, bus).GetSubscriberCount("once.topic"), "单次订阅在触发前计入订阅数")

	require.NoError(t, bus.Publish(context.Background(), eventbus.NewEvent("once.topic", nil)))
	waitEvent(t, received) // 首次发布必须触达

	assert.Equal(t, 0, asDefaultBus(t, bus).GetSubscriberCount("once.topic"),
		"单次订阅执行后必须被自动移除")

	require.NoError(t, bus.Publish(context.Background(), eventbus.NewEvent("once.topic", nil)))
	expectNoEvent(t, received) // 后续发布不得再触达
}

// TestSubscribeOnce_ErrorHandlerStillRemoved 单次订阅处理器返回错误时
// 同样在执行后移除，且发布结果仍为 nil。
func TestSubscribeOnce_ErrorHandlerStillRemoved(t *testing.T) {
	bus := eventbus.NewEventBus(bLogger.NopLogger())
	defer func() { require.NoError(t, bus.Close()) }()

	require.NoError(t, bus.SubscribeOnce("once.err", eventbus.EventHandlerFunc(
		func(context.Context, *eventbus.Event) error { return assert.AnError })))
	require.Equal(t, 1, asDefaultBus(t, bus).GetSubscriberCount("once.err"))

	assert.NoError(t, bus.Publish(context.Background(), eventbus.NewEvent("once.err", nil)))
	assert.Equal(t, 0, asDefaultBus(t, bus).GetSubscriberCount("once.err"),
		"出错的单次订阅也必须被执行后移除")
}

// TestSubscribeOnce_CannotBeUnsubscribed 退订 API 只作用于普通订阅，
// 单次订阅不在其列——必须报错。
func TestSubscribeOnce_CannotBeUnsubscribed(t *testing.T) {
	bus := eventbus.NewEventBus(bLogger.NopLogger())
	defer func() { require.NoError(t, bus.Close()) }()

	handler := eventbus.EventHandlerFunc(func(context.Context, *eventbus.Event) error { return nil })
	require.NoError(t, bus.SubscribeOnce("once.forever", handler))

	err := bus.Unsubscribe("once.forever", handler)
	require.Error(t, err, "退订不得移除单次订阅")
	assert.Equal(t, 1, asDefaultBus(t, bus).GetSubscriberCount("once.forever"))
}

// TestPublishAsync_DeliversInBackground PublishAsync 立即返回 nil，
// 事件随后在后台 goroutine 投递到同步订阅者。
func TestPublishAsync_DeliversInBackground(t *testing.T) {
	bus := eventbus.NewEventBus(bLogger.NopLogger())
	defer func() { require.NoError(t, bus.Close()) }()

	received := make(chan *eventbus.Event, 1)
	require.NoError(t, bus.Subscribe("bg.topic", eventbus.EventHandlerFunc(
		func(_ context.Context, event *eventbus.Event) error {
			received <- event
			return nil
		})))

	assert.NoError(t, bus.PublishAsync(context.Background(), eventbus.NewEvent("bg.topic", nil)),
		"异步发布必须立即成功返回")
	ev := waitEvent(t, received)
	assert.Equal(t, "bg.topic", ev.Type)
}

// TestGetSubscriberCount_PerType 订阅计数按主题独立统计，
// 普通订阅与单次订阅都计入。
func TestGetSubscriberCount_PerType(t *testing.T) {
	bus := eventbus.NewEventBus(bLogger.NopLogger())
	defer func() { require.NoError(t, bus.Close()) }()

	h := eventbus.EventHandlerFunc(func(context.Context, *eventbus.Event) error { return nil })
	require.NoError(t, bus.Subscribe("a", h))
	require.NoError(t, bus.Subscribe("a", h))
	require.NoError(t, bus.SubscribeOnce("b", h))

	assert.Equal(t, 2, asDefaultBus(t, bus).GetSubscriberCount("a"))
	assert.Equal(t, 1, asDefaultBus(t, bus).GetSubscriberCount("b"))
	assert.Equal(t, 0, asDefaultBus(t, bus).GetSubscriberCount("c"))
}

// TestGetEventTypes_UnionsRegularAndOnce GetEventTypes 必须合并普通与单次订阅的主题集合。
func TestGetEventTypes_UnionsRegularAndOnce(t *testing.T) {
	bus := eventbus.NewEventBus(bLogger.NopLogger())
	defer func() { require.NoError(t, bus.Close()) }()

	h := eventbus.EventHandlerFunc(func(context.Context, *eventbus.Event) error { return nil })
	require.NoError(t, bus.Subscribe("regular.topic", h))
	require.NoError(t, bus.SubscribeOnce("once.topic", h))

	types := asDefaultBus(t, bus).GetEventTypes()
	assert.ElementsMatch(t, []string{"regular.topic", "once.topic"}, types,
		"事件类型枚举应包含普通与单次订阅的主题")
}

// TestPublishAsync_OnClosedBusInnerErrorOnlyLogged 对已关闭总线的异步发布：
// 对外立即返回 nil，内部错误仅记录日志——后台 goroutine 在关闭的总线上
// 发布必然失败，这里短暂让步以确保该错误分支被执行。
func TestPublishAsync_OnClosedBusInnerErrorOnlyLogged(t *testing.T) {
	bus := eventbus.NewEventBus(bLogger.NopLogger())
	require.NoError(t, bus.Close())

	assert.NoError(t, bus.PublishAsync(context.Background(), eventbus.NewEvent("closed", nil)),
		"异步发布外壳必须无条件返回 nil")

	// 仅让步给后台 goroutine 走完内部错误路径，不承担断言时序职责。
	time.Sleep(50 * time.Millisecond)
}

// TestClose_RejectsFurtherOperations 关闭总线后订阅与发布都必须报错，
// 重复关闭也必须报错，订阅计数清零。
func TestClose_RejectsFurtherOperations(t *testing.T) {
	bus := eventbus.NewEventBus(bLogger.NopLogger())

	h := eventbus.EventHandlerFunc(func(context.Context, *eventbus.Event) error { return nil })
	require.NoError(t, bus.Subscribe("close.me", h))
	require.NoError(t, bus.SubscribeOnce("close.me", h))
	require.Equal(t, 2, asDefaultBus(t, bus).GetSubscriberCount("close.me"))

	require.NoError(t, bus.Close())

	assert.Equal(t, 0, asDefaultBus(t, bus).GetSubscriberCount("close.me"), "关闭后订阅表必须清空")
	assert.Empty(t, asDefaultBus(t, bus).GetEventTypes())

	err := bus.Subscribe("close.me", h)
	require.Error(t, err, "关闭后订阅必须被拒绝")
	assert.Contains(t, err.Error(), "event bus is closed")

	err = bus.SubscribeOnce("close.me", h)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "event bus is closed")

	err = bus.Publish(context.Background(), eventbus.NewEvent("close.me", nil))
	require.Error(t, err, "关闭后发布必须被拒绝")
	assert.Contains(t, err.Error(), "event bus is closed")

	err = bus.Close()
	require.Error(t, err, "重复关闭必须报错")
	assert.Contains(t, err.Error(), "already closed")
}
