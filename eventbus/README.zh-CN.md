# eventbus

[English](README.md) | 简体中文

进程内发布/订阅事件总线：同步/异步/一次性订阅、按事件类型分发、多命名总线管理与中间件组合。

## 功能特性

- **EventBus**（`eventbus.go`）：同步/异步/一次性（once）订阅，按事件类型分发；单个 handler 失败不阻断其余 handler；`PublishAsync` 用带 30 秒超时的独立后台上下文执行，不受调用方取消影响；
- **Handler 组合子**（`handler.go`）：`AsyncHandler`（异步执行）/ `ChainHandler`（顺序链）/ `FilterHandler`（条件过滤）；
- **Manager**（`manager.go`）：多命名总线管理，空名与 `"global"` 特判返回全局总线，避免"发布到 global"与"订阅在 Global()"分属两个实例的经典错误；
- **中间件**（`middleware.go`）：日志 / recover / 超时 / 重试 / 指标，`Chain` 组合；
- **事件自带身份**：`NewEvent` 自动生成唯一 ID（时间戳 + 8 位随机串）与 `Timestamp`。

## 安装

独立模块（日志接口依赖 `github.com/tx7do/kratos-bootstrap/logger`）：

```bash
go get github.com/tx7do/go-utils/eventbus@latest
```

```go
import (
    "github.com/tx7do/kratos-bootstrap/logger"
    "github.com/tx7do/go-utils/eventbus"
)
```

## 快速开始

```go
package main

import (
	"context"
	"fmt"

	"github.com/tx7do/kratos-bootstrap/logger"
	"github.com/tx7do/go-utils/eventbus"
)

func main() {
	bus := eventbus.NewEventBus(logger.NopLogger())
	defer bus.Close()

	_ = bus.Subscribe("order.created", eventbus.EventHandlerFunc(func(ctx context.Context, ev *eventbus.Event) error {
		var order map[string]any
		if err := ev.GetData(&order); err != nil {
			return err
		}
		fmt.Println("order created:", order["order_id"])
		return nil
	}))

	ctx := context.Background()
	_ = bus.Publish(ctx, eventbus.NewEvent("order.created", map[string]any{"order_id": "order-123"}))
}
```

> 注意：构造参数是 `logger.Logger` 接口。本库依赖的 kratos-bootstrap/logger 没有 `DefaultLogger` 变量，测试/裸进程用 `logger.NopLogger()`，接入框架日志用 `logger.GetLogger()`。

事件类型是任意字符串（约定 `"域.动作"` 风格），载荷 `Event.Data` 为 `any`，订阅侧用 `Event.GetData(&v)` 做 JSON 往返还原——发布侧传结构体、订阅侧还原成同形结构体即可。

## API 参考

### EventBus（`eventbus.go`）

```go
type EventBus interface {
    Subscribe(eventType string, handler Handler) error       // 同步订阅
    SubscribeAsync(eventType string, handler Handler) error  // 异步订阅（等价 Subscribe + AsyncHandler 包装）
    SubscribeOnce(eventType string, handler Handler) error   // 一次性订阅：触发一次后自动移除
    Unsubscribe(eventType string, handler Handler) error     // 按指针身份移除；找不到报 "handler not found"
    Publish(ctx context.Context, event *Event) error         // 同步发布：依次调用所有 handler
    PublishAsync(ctx context.Context, event *Event) error    // 异步发布：立即返回，后台 goroutine 内发布
    Close() error                                            // 关闭总线并清空全部订阅
}

func NewEventBus(logger logger.Logger) EventBus
```

补充方法（`*DefaultEventBus` 具体类型上）：

```go
func (eb *DefaultEventBus) GetSubscriberCount(eventType string) int // 普通 + once 订阅数
func (eb *DefaultEventBus) GetEventTypes() []string                 // 有订阅者的全部事件类型
```

### Event（`event.go`）

```go
type Event struct {
    ID        string            // NewEvent 自动生成：时间戳 + 8 位随机串
    Type      string            // 事件类型
    Source    string            // 来源标识
    Data      any               // 载荷
    Metadata  map[string]string // 附加元数据
    Timestamp time.Time         // NewEvent 自动置为当前时间
    Priority  int               // 优先级（配合 FilterHandler 使用）
}

func NewEvent(eventType string, data any) *Event          // 自动生成 ID/Timestamp，Priority 0
func (e *Event) WithSource(source string) *Event          // 链式设置来源
func (e *Event) WithPriority(priority int) *Event         // 链式设置优先级
func (e *Event) WithMetadata(key, value string) *Event    // 链式追加元数据
func (e *Event) GetData(v any) error                      // Data 经 JSON 往返还原进 v
func (e *Event) Clone() *Event                            // 深拷贝（Metadata 复制；Data 为浅引用）
```

### Manager（`manager.go`）

```go
func NewManager(logger logger.Logger) *Manager

func (m *Manager) GetBus(name string) EventBus       // ""/"global" 特判返回全局总线；未知名自动创建
func (m *Manager) Global() EventBus                  // 全局总线
func (m *Manager) Publish(ctx context.Context, busName string, event *Event) error
func (m *Manager) PublishGlobal(ctx context.Context, event *Event) error
func (m *Manager) Subscribe(busName, eventType string, handler Handler) error
func (m *Manager) SubscribeGlobal(eventType string, handler Handler) error
func (m *Manager) GetStats() map[string]any          // total_buses / buses{event_types} / global_bus{event_types}
func (m *Manager) Close() error                      // 关闭全部命名总线与全局总线
```

### Handler 组合子（`handler.go`）

```go
type Handler interface { Handle(ctx context.Context, event *Event) error }
type EventHandlerFunc func(ctx context.Context, event *Event) error // 适配函数为 Handler

func NewAsyncHandler(handler Handler) *AsyncHandler       // Handle 立即返回，handler 在新 goroutine 执行
func NewChainHandler(handlers ...Handler) *ChainHandler   // 顺序执行，遇错短路返回
func NewFilterHandler(filter func(*Event) bool, handler Handler) *FilterHandler // filter 为 false 时跳过
```

### 中间件（`middleware.go`）

```go
type Middleware func(Handler) Handler

func LoggingMiddleware(logger *logger.Helper) Middleware    // 记录处理开始/成功/失败
func RecoveryMiddleware(logger *logger.Helper) Middleware   // recover panic → *PanicError
func TimeoutMiddleware(timeout time.Duration) Middleware    // 超时 → *TimeoutError
func RetryMiddleware(maxRetries int, delay time.Duration) Middleware // 失败重试，间隔固定 delay
func MetricsMiddleware(logger *logger.Helper) Middleware    // 记录处理耗时
func Chain(middlewares ...Middleware) Middleware            // 组合：先列出的在最外层
```

中间件产出的两类错误：

```go
type PanicError struct{ Value any }                                    // Error() == "panic in event handler"
type TimeoutError struct{ EventID, EventType string; Timeout time.Duration } // Error() == "event handling timeout"
```

## Notes / 注意事项

- **handler 失败不阻断**：`Publish` 依次调用所有 handler，单个失败只记日志，其余照常执行；发布方拿不到 handler 错误（`Publish` 正常返回 nil）——需要感知错误就别用总线，或把错误兜进 handler 自己的逻辑里；
- **无人订阅不是错误**：`Publish` 到没有订阅者的事件类型直接返回 nil；
- **once 语义**：`SubscribeOnce` 的 handler 在第一次 `Publish` 后整组移除（按事件类型）；
- **PublishAsync 的超时独立**：后台 goroutine 用 `context.Background()` + 30 秒超时，调用方 ctx 取消不影响已入队的异步发布；发布本身仍是同步遍历 handler（handler 内部的异步靠 `SubscribeAsync`/`AsyncHandler`）；
- **Close 后复用报错**：`Subscribe` / `Publish` 在 `Close` 之后返回 "event bus is closed"；重复 `Close` 返回 "event bus already closed"；
- **`Unsubscribe` 只对普通订阅生效**，且按 handler 指针身份匹配（传同一个包装函数没用，要留好原引用）；once 订阅无法单独移除；
- **`Chain` 的顺序**：`Chain(a, b)(h)` 执行顺序为 a → b → h，先列出的中间件在最外层；
- **超时中间件不中断 handler**：`TimeoutMiddleware` 超时后返回 `*TimeoutError`，但被超时的 handler goroutine 仍在跑（channel 语义决定的）；需要硬中断请在 handler 内部响应 ctx。
