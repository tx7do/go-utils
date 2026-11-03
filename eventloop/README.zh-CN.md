# eventloop

[English](README.md) | 简体中文

单 goroutine 顺序事件循环：三优先级调度、可选帧驱动模式、请求-响应回调，带可注入的日志与统计接口。

## 功能特性

- **三优先级调度**：`PriorityHigh` / `PriorityMedium` / `PriorityLow`，每个循环周期先清空 High、再清空 Medium、Low 缓存延后处理（处理前还会再清一遍高/中）——高优先级事件总是更早被处理；
- **单 goroutine 顺序处理**：所有 `Process` 调用在同一个事件循环 goroutine 中串行执行，顺序可预测；入队（Submit）并发安全；
- **帧驱动模式**（`frameDriven=true`）：按固定帧率与时间预算处理事件，适合游戏主循环、实时渲染等时间片控制场景；
- **请求-响应模式**：`NewRequestEvent` 一行拿到事件与回执 channel；
- **回调两种投递模式**：inline（事件循环 goroutine 内同步投递）与 async（独立分发器，默认）；
- **可注入日志与统计**：`SetLogger` / `SetMetrics`，默认 `NoopLogger` / `NoopMetrics`；
- **优雅停止**：`Stop` 取消上下文并等待循环 goroutine 退出。

## 安装

独立模块，零第三方依赖：

```bash
go get github.com/tx7do/go-utils/eventloop@latest
```

```go
import "github.com/tx7do/go-utils/eventloop"
```

## 快速开始

```go
package main

import (
	"context"
	"fmt"
	"time"

	"github.com/tx7do/go-utils/eventloop"
)

type myProc struct{}

func (p *myProc) Process(ev eventloop.Event) eventloop.Result {
	fmt.Println("process:", ev.Data) // 字段名是 Data
	return eventloop.Result{}
}

func main() {
	// NewEventLoop(bufferSize, processor, frameDriven)：bufferSize<=0 用默认 100
	el := eventloop.NewEventLoop(64, &myProc{}, false)

	// 可选：回调模式。false=异步分发（默认），超时 1s
	el.SetCallbackInline(false, time.Second)

	if err := el.Start(); err != nil {
		panic(err)
	}
	defer el.Stop()

	// 请求-响应：一行拿到事件与回执 channel（容量 1）
	ev, reply := eventloop.NewRequestEvent("calc.add", "hello")
	ev.Priority = eventloop.PriorityHigh // Event 按值返回，Submit 前可直接改字段
	ev.Ctx = context.Background()
	if err := el.Submit(ev); err != nil {
		panic(err)
	}

	select {
	case r := <-reply:
		fmt.Println("callback result:", r)
	case <-time.After(2 * time.Second):
		fmt.Println("callback timeout")
	}
}
```

## 核心 API

### 创建与生命周期

```go
// bufferSize<=0 用默认 100；processor 可为 nil（此时带回调的事件经回调收到 ErrNoEventProcessor）；
// frameDriven=true 启用帧驱动循环。
func NewEventLoop(bufferSize int, processor EventProcessor, frameDriven bool) *EventLoop

func (el *EventLoop) Start() error          // 幂等：已在跑直接返回 nil；等待就绪信号（最长 50ms，超时也返回 nil）
func (el *EventLoop) Stop()                 // 优雅停止：cancel 上下文并等待全部 goroutine 退出
func (el *EventLoop) IsRunning() bool
```

### 提交事件

```go
// 非阻塞提交：队列满立即返回对应 Err*QueueFull；优先级不认识返回 ErrUnknownPriority；
// 未启动返回 ErrEventLoopNotRunning。
func (el *EventLoop) Submit(event Event) error

// 阻塞提交：等队列有空位、调用方 ctx 取消（返回 ctx.Err()）或循环停止（返回 ErrEventLoopStopped）。
func (el *EventLoop) SubmitBlocking(ctx context.Context, event Event) error
```

### 事件与请求-响应（`event.go`）

```go
type Event struct {
    Priority Priority        // 优先级（默认 PriorityLow）
    Type     string          // 事件类型
    Data     any             // 事件数据（注意：字段名是 Data，不是 Payload）
    Callback chan Result     // 回执通道（可 nil）
    Ctx      context.Context // 事件上下文（可 nil，兜底用循环的 ctx）
    TS       time.Time       // NewEvent 自动置为当前时间
}

type Result struct {
    Data any
    Err  error
}

// 构造事件：默认 PriorityLow、TS 自动设置，选项链式覆盖。
func NewEvent(typ string, payload any, opts ...EventOption) Event

// 请求-响应：返回事件与容量 1 的回执 channel（自动 WithCallback）。
func NewRequestEvent(typ string, payload any) (Event, chan Result)

func WithPriority(p Priority) EventOption
func WithContext(ctx context.Context) EventOption
func WithCallback(cb chan Result) EventOption
```

### 配置注入

```go
// 注入日志接口（NoopLogger / NewStdLogger 或自定义实现）；传 nil 复位为 NoopLogger。
func (el *EventLoop) SetLogger(l Logger)

// 注入统计实现（如 NewSimpleMetrics()）；传 nil 复位为 NoopMetrics。
func (el *EventLoop) SetMetrics(m Metrics)

// 切换回调投递模式：inline=true 在事件循环 goroutine 内同步投递（可能阻塞循环）；
// timeout 控制同步投递超时，0 表示无限等待。建议 Start() 之前设置。
func (el *EventLoop) SetCallbackInline(inline bool, timeout time.Duration)

// 设置帧驱动参数。interval/budget 传非正值保持原值，maxLow 传负值保持原值（允许 0）。
// 建议在 Start() 之前调用。
func (el *EventLoop) SetFrameParameters(interval, budget, maxLow time.Duration)
```

### 观测

```go
// 各内部通道的瞬时长度快照（非阻塞）：High / Medium / Low / Callback。
func (el *EventLoop) QueueLengths() QueueLengths
```

## 帧驱动模式

`NewEventLoop(…, true)` 后事件循环按 `frameInterval` 心跳逐帧处理：

1. 单帧内先处理完全部 High，再处理完全部 Medium，最后在剩余预算内处理 Low；
2. 帧预算（`frameBudget`）耗尽立即返回，剩余事件留给下一帧；Low 额外受 `maxLowTime` 限制；
3. 默认参数常量（`event.go`）：`FrameInterval = 50ms`（20Hz）、`FrameBudget = 25ms`、`MaxLowTime = 2ms`。

微调必须在 `Start()` 之前（或任何时刻）经 `SetFrameParameters` 进行——三个字段未导出，包外无法直接赋值：

```go
el := eventloop.NewEventLoop(64, myProcessor, true)
el.SetFrameParameters(16*time.Millisecond, 10*time.Millisecond, 2*time.Millisecond)
_ = el.Start()
```

非实时场景保持 `frameDriven=false`（自由驱动：事件到达即处理，无帧节拍）。

## 回调投递语义

两种模式（`SetCallbackInline` 切换，建议 `Start()` 前设置）：

- **异步（默认）**：回调项进内部 `callbackCh`，由独立 `callbackDispatcher` goroutine 可靠投递（持续重试直到送达、回调方取消或 `cbTimeout` 超时后放弃并记日志）。回调不在事件循环 goroutine 中执行；`Start()` 仅在此模式启动分发器。
- **同步（inline）**：在事件循环 goroutine 内直接把 `Result` 发到 `Event.Callback`；`cbTimeout=0` 无限等待（可能长时间阻塞循环），非 0 超时后放弃并记日志。保证回调在循环线程投递，但对循环形成背压。

队列满时的兜底：`callbackCh` 满时启动后台 goroutine 重试入队，直到送达、事件 ctx 取消、循环停止或超时放弃——极端场景可能产生较多短期 goroutine，生产可按需替换为固定 worker 池。

接收方应在独立 goroutine 中消费 `Callback` channel（`NewRequestEvent` 的回执 channel 容量为 1，`Submit` 后尽快读）。

## Metrics

`Metrics` 接口统计提交/丢弃/处理计数（按优先级）、回调丢弃与 inline 超时计数、处理耗时；`Snapshot()` 返回只读快照（含各优先级平均处理耗时纳秒）。内置实现：

- `NoopMetrics{}`：默认，零开销；
- `NewSimpleMetrics()`：原子计数实现，适合测试与轻量监控。

```go
m := eventloop.NewSimpleMetrics()
el.SetMetrics(m)
// 运行一段时间后：
snap := m.Snapshot()
```

另可用 `el.QueueLengths()` 观测各队列瞬时积压。

## 错误哨兵（`errors.go`）

```go
var (
    ErrEventLoopNotRunning           // 循环未启动时 Submit/SubmitBlocking
    ErrEventLoopStopped              // SubmitBlocking 期间循环被 Stop
    ErrNoEventProcessor              // processor 为 nil 时，带回调的事件经回调收到该错误
    ErrHighQueueFull                 // 高优先级队列满
    ErrMediumQueueFull               // 中优先级队列满
    ErrLowQueueFull                  // 低优先级队列满
    ErrUnknownPriority               // 优先级不在 High/Medium/Low
)
```

## 优先级与并发语义

- 每个 `Process` 都在事件循环 goroutine 中串行执行，顺序可预测但不绑定特定 OS 线程；需要严格线程亲和时在业务侧用 `runtime.LockOSThread()` 并谨慎设计；
- `Submit` / `SubmitBlocking` 可并发调用，入队并发安全；
- 自由驱动下 Low 事件会被本地缓存（deferredLow），处理前先再清一遍 High/Medium——持续高优先级洪流可能让 Low 饥饿；
- 事件自带 `Ctx` 已取消/超时时不进处理器：带回调的事件经回调收到 `ctx.Err()`，不带回调的静默丢弃。

## Notes / 注意事项

- **不要在 `Process` 里长时间阻塞**：整个循环被这一个调用挡住（包括帧驱动模式的帧预算也会被冲破）；耗时操作请在处理器内部自行并行化；
- **帧驱动日志走标准库 log**：帧循环的启动/停止与"帧预算超支"消息直接 `log.Println` 输出，不经过 `SetLogger` 注入的 Logger（注入的 Logger 用于回调丢弃/超时等告警）；
- `Submit` 是非阻塞的——需要背压时用 `SubmitBlocking` 或在上层做缓冲；
- `Stop` 之后通道里未处理的事件不会被处理；需要"排空后再停"请在业务层等待 `QueueLengths()` 归零后再调 `Stop`。
