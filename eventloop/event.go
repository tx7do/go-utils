package eventloop

import (
	"context"
	"time"
)

// Priority 事件优先级
type Priority int

const (
	PriorityHigh Priority = iota
	PriorityMedium
	PriorityLow
)

const (
	FrameInterval = 50 * time.Millisecond // 20Hz 逻辑帧
	FrameBudget   = 25 * time.Millisecond // 预留50%缓冲
	MaxLowTime    = 2 * time.Millisecond  // 低优先级最多占用2ms/帧
)

// Result 回调结果载体
type Result struct {
	Data any
	Err  error
}

// Event 统一事件结构
type Event struct {
	Priority Priority // 事件优先级
	Type     string   // 事件类型

	Data any // 事件数据

	Callback chan Result // 回调
	Ctx      context.Context

	TS time.Time // 事件发送时间
}

type EventOption func(*Event)

// WithPriority 设置优先级
func WithPriority(p Priority) EventOption {
	return func(e *Event) { e.Priority = p }
}

// WithContext 设置上下文
func WithContext(ctx context.Context) EventOption {
	return func(e *Event) { e.Ctx = ctx }
}

// WithCallback 设置回调通道
func WithCallback(cb chan Result) EventOption {
	return func(e *Event) { e.Callback = cb }
}

// NewEvent 创建 Event，支持可选参数；默认 PriorityLow，TS 自动设置
func NewEvent(typ string, payload any, opts ...EventOption) Event {
	ev := Event{
		Priority: PriorityLow,
		Type:     typ,
		Data:     payload,
		TS:       time.Now(),
	}
	for _, opt := range opts {
		opt(&ev)
	}
	return ev
}

// NewRequestEvent 创建带 reply channel 的请求事件（返回 Event 与 reply chan）
func NewRequestEvent(typ string, payload any) (Event, chan Result) {
	reply := make(chan Result, 1)
	ev := NewEvent(typ, payload, WithCallback(reply))
	return ev, reply
}
