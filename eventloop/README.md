# 事件循环（eventloop）

## 简要说明

- 本包实现一个单 goroutine 顺序事件循环（EventLoop），按优先级调度事件并调用用户提供的 `EventProcessor` 进行处理。
- 设计目标：简单、可预测的顺序处理；入队并发安全；可通过 `EventProcessor` 实现并发或异步处理。

## 核心类型（概念）

- `EventLoop`：事件循环主体，负责从优先级队列读取事件并处理。
- `Event`：事件结构体，典型包含 `Priority`、`Payload`、`Callback`、`Ctx` 等字段。注意：包内字段已统一使用 `Payload`。
- `EventProcessor`：处理器接口，定义 `Process(Event) Result`，由调用方实现实际业务逻辑。
- `Result`：处理结果，通常含 `Err error` 字段，用于回调或上报处理状态。

## 主要 API（示例签名）

- `func NewEventLoop(bufferSize int, processor EventProcessor) *EventLoop`：创建实例。
- `func (el *EventLoop) Start() error`：启动事件循环（在单独 goroutine 中运行）。
- `func (el *EventLoop) Stop()`：优雅停止并等待循环结束。
- `func (el *EventLoop) Submit(ev Event) error`：将事件入队（行为详见下文）。
- `func (el *EventLoop) SetCallbackInline(inline bool, timeout time.Duration)`：切换回调投递模式。`inline=true` 表示在事件循环的 goroutine 内同步投递回调（可能阻塞事件循环）；`timeout` 控制同步投递的超时（`0` 表示无限等待）。注意建议在 `Start()` 之前设置以获得更确定的行为。

## 回调投递语义（重要）

- 两种回调模式：
  - 同步（inline）：在事件循环 goroutine 内直接将 `Result` 发送到 `Event.Callback`。可以设置超时以避免长期阻塞。该模式保证回调由事件循环逻辑线程投递，但可能对事件循环造成背压。
  - 异步（async，默认）：事件循环把回调项封装为 `callbackItem` 放入内部 `callbackCh`，由独立的 `callbackDispatcher` 负责可靠投递（带重试直到超时或取消）。在该模式下回调不在事件循环 goroutine 中执行。
- 实现细节：
  - `Start()` 仅在异步模式下启动 `callbackDispatcher`。切换模式时内部通过互斥锁保护 `cbInline`/`cbTimeout` 的读写以避免数据竞争。
  - 在异步模式且 `callbackCh` 满时，会启动后台重试 goroutine 去尝试入队直到超时或上下文取消；这可能在极端场景下产生较多短期 goroutine，生产环境可用固定 worker 池或限流替换该策略。
  - 建议回调接收方在独立 goroutine 中消费 `Callback` channel，以避免阻塞或丢失（除非使用 inline 模式并接受阻塞语义）。

## 优先级行为

- 支持至少三个优先级：`PriorityHigh`、`PriorityMedium`、`PriorityLow`。
- 事件循环在每个循环周期优先尝试读取高优先级，其次中、最后低，保证高优先级事件更早被处理。
- 低优先级读取可采用阻塞或被取消的方式（参见实现细节）。

## 并发与执行语义

- 所有事件的 `Process` 调用均在同一个 goroutine（事件循环 goroutine）中串行执行，保证顺序可预测，但不绑定到特定 OS 线程。
- `Submit`/`Dispatch` 可并发调用；它们将事件放入通道，入队是并发安全的。
- 回调行为：
  - 默认异步模式：事件循环不会在处理完成时阻塞于回调发送，采用内部队列并由分发器投递；当内部队列满且重试超时后，结果可能被丢弃并记录。
  - inline 模式：回调由事件循环 goroutine 直接投递，可能阻塞事件循环；可设置超时以限制阻塞时间。
- 若需要严格的 OS 线程绑定，请在业务中使用 `runtime.LockOSThread()` 并谨慎设计；大多数场景不需要。

## 使用示例

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
  fmt.Println("process:", ev.Payload)
  return eventloop.Result{}
}

func main() {
  el := eventloop.NewEventLoop(64, &myProc{})

  // 可选：在 Start 之前选择回调模式；inline=true 表示在事件循环 goroutine 内投递回调
  el.SetCallbackInline(false, time.Second) // 使用异步分发（默认）并设置超时

  _ = el.Start()
  defer el.Stop()

  // 若传回调，可用 buffered channel 接收结果
  cb := make(chan eventloop.Result, 1)
  el.Submit(eventloop.Event{Priority: eventloop.PriorityHigh, Payload: "hello", Ctx: context.Background(), Callback: cb})

  // 在独立 goroutine 中读取回调（避免阻塞事件循环）
  go func() {
    select {
    case r := <-cb:
      fmt.Println("callback result:", r)
    case <-time.After(2 * time.Second):
      fmt.Println("callback timeout")
    }
  }()

  time.Sleep(100 * time.Millisecond)
}
```

## 测试与验证

- 运行 `go test ./...` 以执行单元测试。
- 关键测试覆盖项（示例）：
  - 优先级调度（高/中/低 的处理顺序）。
  - nil 处理器时通过回调返回错误的场景。
  - 异步回调队列满时的重试/超时行为。
  - inline 模式下回调在事件循环 goroutine 中投递且超时控制生效。

## 注意事项

- 如果业务要求严格的线程亲和性（OS 线程绑定），需使用 `runtime.LockOSThread()` 并谨慎设计；大多数场景不需要。
- 处理器内部若执行耗时或阻塞操作，应自行并行化以避免阻塞整个循环。
- 回调策略为非阻塞写入，调用方若需保证回调可靠性，应使用有足够缓冲或外部同步机制。
