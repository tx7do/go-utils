// Package eventbus_test 针对 pkg/eventbus 的 Manager 做黑盒测试。
//
// 本文件覆盖 manager.go：
//   - 全局总线特判：GetBus("")/GetBus("global") 与 Global() 必须是同一实例，
//     否则跨入口的订阅与发布会永远错开；
//   - 具名总线的创建与缓存复用；
//   - 经 Manager 的订阅/发布转发到正确总线（含全局入口）；
//   - GetStats 的统计结构；
//   - Close 关闭全部总线（含已单独关闭的总线的容错）与关闭后的拒绝行为。
package eventbus_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	bLogger "github.com/tx7do/kratos-bootstrap/logger"

	"github.com/tx7do/go-utils/eventbus"
)

// TestManager_GlobalBusAliases "" 与 "global" 都是 Global() 的别名，
// 三者必须指向同一总线实例。
func TestManager_GlobalBusAliases(t *testing.T) {
	m := eventbus.NewManager(bLogger.NopLogger())
	defer func() { require.NoError(t, m.Close()) }()

	global := m.Global()
	require.NotNil(t, global)
	assert.True(t, m.GetBus("") == global, "空名必须是全局总线的别名")
	assert.True(t, m.GetBus("global") == global, "\"global\" 必须是全局总线的别名")
}

// TestManager_GetBusCreatesAndCaches 具名总线首次创建、之后缓存复用同一实例，
// 且与全局总线相互独立。
func TestManager_GetBusCreatesAndCaches(t *testing.T) {
	m := eventbus.NewManager(bLogger.NopLogger())
	defer func() { require.NoError(t, m.Close()) }()

	first := m.GetBus("metrics")
	require.NotNil(t, first)
	second := m.GetBus("metrics")
	assert.True(t, first == second, "同名总线必须缓存复用同一实例")
	assert.False(t, first == m.Global(), "具名总线不得与全局总线混同")

	stats := m.GetStats()
	require.Len(t, stats, 3)
	assert.Equal(t, 1, stats["total_buses"], "仅应创建一个具名总线")
}

// TestManager_SubscribePublishRoundTripOnNamedBus 经 Manager 的订阅/发布必须
// 落到同一具名总线上并完成投递。
func TestManager_SubscribePublishRoundTripOnNamedBus(t *testing.T) {
	m := eventbus.NewManager(bLogger.NopLogger())
	defer func() { require.NoError(t, m.Close()) }()

	received := make(chan *eventbus.Event, 1)
	require.NoError(t, m.Subscribe("jobs", "job.done", eventbus.EventHandlerFunc(
		func(_ context.Context, event *eventbus.Event) error {
			received <- event
			return nil
		})))

	require.NoError(t, m.Publish(context.Background(), "jobs", eventbus.NewEvent("job.done", nil)))

	ev := waitEvent(t, received)
	assert.Equal(t, "job.done", ev.Type)
}

// TestManager_GlobalRoundTrip 全局订阅与全局发布必须互达；
// 空名发布也必须落到全局总线。
func TestManager_GlobalRoundTrip(t *testing.T) {
	m := eventbus.NewManager(bLogger.NopLogger())
	defer func() { require.NoError(t, m.Close()) }()

	// 每个通道将收到两次发布（全局入口 + 空名别名入口），容量须 >= 2，
	// 否则同步发布会在第二次投递时阻塞在 chan send 上。
	viaAlias := make(chan *eventbus.Event, 2)
	viaGlobal := make(chan *eventbus.Event, 2)

	require.NoError(t, m.Subscribe("", "sys.echo", eventbus.EventHandlerFunc(
		func(_ context.Context, event *eventbus.Event) error {
			viaAlias <- event
			return nil
		})))
	require.NoError(t, m.SubscribeGlobal("sys.echo", eventbus.EventHandlerFunc(
		func(_ context.Context, event *eventbus.Event) error {
			viaGlobal <- event
			return nil
		})))

	require.NoError(t, m.PublishGlobal(context.Background(), eventbus.NewEvent("sys.echo", nil)))
	require.NoError(t, m.Publish(context.Background(), "", eventbus.NewEvent("sys.echo", nil)))

	assert.Equal(t, "sys.echo", waitEvent(t, viaAlias).Type)
	assert.Equal(t, "sys.echo", waitEvent(t, viaGlobal).Type)
}

// TestManager_PublishToSubscriberlessBus 发布到无订阅者的总线应返回 nil
// （该总线可能因此被惰性创建）。
func TestManager_PublishToSubscriberlessBus(t *testing.T) {
	m := eventbus.NewManager(bLogger.NopLogger())
	defer func() { require.NoError(t, m.Close()) }()

	assert.NoError(t, m.Publish(context.Background(), "lazy.bus", eventbus.NewEvent("nobody", nil)))
}

// TestManager_GetStats_Structure GetStats 应包含总线总数、具名总线明细
// 与全局总线明细，明细里的 event_types 反映各总线的订阅主题。
func TestManager_GetStats_Structure(t *testing.T) {
	m := eventbus.NewManager(bLogger.NopLogger())
	defer func() { require.NoError(t, m.Close()) }()

	named := m.GetBus("telemetry")
	require.NotNil(t, named)
	require.NoError(t, m.Subscribe("telemetry", "telemetry.beat", eventbus.EventHandlerFunc(
		func(context.Context, *eventbus.Event) error { return nil })))
	require.NoError(t, m.SubscribeGlobal("global.beat", eventbus.EventHandlerFunc(
		func(context.Context, *eventbus.Event) error { return nil })))

	stats := m.GetStats()
	require.Len(t, stats, 3, "统计应含 total_buses/buses/global_bus 键")
	assert.Equal(t, 1, stats["total_buses"])

	buses, ok := stats["buses"].(map[string]any)
	require.True(t, ok, "buses 必须是 map[string]any")
	namedStats, ok := buses["telemetry"].(map[string]any)
	require.True(t, ok, "telemetry 总线必须有条目")
	topics, ok := namedStats["event_types"].([]string)
	require.True(t, ok, "event_types 必须是 []string")
	assert.ElementsMatch(t, []string{"telemetry.beat"}, topics)

	globalStats, ok := stats["global_bus"].(map[string]any)
	require.True(t, ok, "global_bus 必须存在")
	globalTopics, ok := globalStats["event_types"].([]string)
	require.True(t, ok)
	assert.ElementsMatch(t, []string{"global.beat"}, globalTopics)
}

// TestManager_Close_ClosesAllBuses Close 必须关闭全部总线（含全局）；
// 关闭后经 Manager 的订阅被拒绝，统计中的具名总线清零；
// 已被单独关闭的总线在 Close 时报错但整体仍返回 nil。
func TestManager_Close_ClosesAllBuses(t *testing.T) {
	m := eventbus.NewManager(bLogger.NopLogger())

	stale := m.GetBus("stale")
	require.NoError(t, stale.Close(), "预先单独关闭一个具名总线")

	// 全局总线也预先关闭：Manager.Close 对已关闭总线的错误必须逐一容忍并记日志。
	require.NoError(t, m.Global().Close())

	require.NoError(t, m.Close(), "已关闭总线应被容忍，Close 整体仍返回 nil")

	stats := m.GetStats()
	require.Len(t, stats, 3)
	assert.Equal(t, 0, stats["total_buses"], "关闭后具名总线统计必须归零")
	buses, ok := stats["buses"].(map[string]any)
	require.True(t, ok)
	assert.Empty(t, buses)

	err := m.SubscribeGlobal("after.close", eventbus.EventHandlerFunc(
		func(context.Context, *eventbus.Event) error { return nil }))
	require.Error(t, err, "全局总线已关闭，订阅必须被拒绝")
	assert.Contains(t, err.Error(), "event bus is closed")
}
