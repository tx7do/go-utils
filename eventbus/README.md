# eventbus

进程内发布/订阅事件总线：

- **EventBus**（`eventbus.go`）：同步/异步/一次性（once）订阅，按事件类型分发；
  单个 handler 失败不阻断其余 handler；`PublishAsync` 用带超时的独立上下文执行，
  不受调用方取消影响。
- **Handler 组合子**（`handler.go`）：`AsyncHandler` / `ChainHandler` / `FilterHandler`。
- **Manager**（`manager.go`）：多命名总线管理，空名与 `"global"` 特判返回全局总线，
  避免"发布到 global"与"订阅在 Global()"分属两个实例的经典错误。
- **中间件**（`middleware.go`）：日志 / recover / 超时 / 重试 / 指标，`Chain` 组合。

```go
import (
	"github.com/tx7do/kratos-bootstrap/logger"
	"github.com/tx7do/go-utils/eventbus"
)

bus := eventbus.NewEventBus(logger.DefaultLogger)
defer bus.Close()

bus.Subscribe("order.created", eventbus.EventHandlerFunc(func(ctx context.Context, ev *eventbus.Event) error {
	return nil
}))

_ = bus.Publish(ctx, eventbus.NewEvent("order.created", order))
```

事件类型是任意字符串（约定 `"域.动作"` 风格），载荷 `Event.Data` 为 `any`，
订阅侧用 `Event.GetData(&v)` 做 JSON 往返还原。
