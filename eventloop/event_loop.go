package eventloop

import (
	"context"
	"sync"
	"sync/atomic"
	"time"
)

const (
	defaultBufferSize = 100
	defaultTimeout    = time.Second
)

// EventLoop 是单线程优先级事件循环
type EventLoop struct {
	highChan   chan Event // 高
	mediumChan chan Event // 中
	lowChan    chan Event // 低

	processor EventProcessor // 事件处理器

	ctx    context.Context    // 上下文
	cancel context.CancelFunc // 取消函数

	wg sync.WaitGroup // 等待组

	running atomic.Bool // 运行状态
	started atomic.Pointer[chan struct{}]

	cbInline   bool              // 回调是否内联执行
	mu         sync.Mutex        // 互斥锁（用于保护 cbInline/cbTimeout 等可变配置）
	callbackCh chan callbackItem // 回调通道
	cbTimeout  time.Duration     // 回调超时设置

	logger Logger // 注入式日志接口（默认 NoopLogger）
}

// NewEventLoop 创建并返回一个 EventLoop 实例
func NewEventLoop(bufferSize int, processor EventProcessor) *EventLoop {
	if bufferSize <= 0 {
		bufferSize = defaultBufferSize
	}
	ctx, cancel := context.WithCancel(context.Background())
	return &EventLoop{
		ctx:     ctx,
		cancel:  cancel,
		running: atomic.Bool{},

		highChan:   make(chan Event, bufferSize),
		mediumChan: make(chan Event, bufferSize),
		lowChan:    make(chan Event, bufferSize),
		processor:  processor,

		callbackCh: make(chan callbackItem, bufferSize),
		cbTimeout:  defaultTimeout,

		logger: NoopLogger{}, // 默认无操作日志
	}
}

// SetLogger 在运行时注入自定义 Logger（可传 NoopLogger / StdLogger）
func (el *EventLoop) SetLogger(l Logger) {
	if l == nil {
		el.logger = NoopLogger{}
		return
	}
	el.logger = l
}

// SetCallbackInline 切换回调投递模式；inline=true 表示在事件循环内同步投递，timeout 控制同步投递的超时（0 表示无限等待）。
func (el *EventLoop) SetCallbackInline(inline bool, timeout time.Duration) {
	el.mu.Lock()
	defer el.mu.Unlock()

	el.cbInline = inline
	if timeout >= 0 {
		el.cbTimeout = timeout
	}
}

// Start 启动逻辑引擎，开始处理事件。
func (el *EventLoop) Start() error {
	if !el.running.CompareAndSwap(false, true) {
		return nil
	}

	startedCh := make(chan struct{})
	el.mu.Lock()
	el.started.Store(&startedCh)
	inline := el.cbInline
	el.mu.Unlock()

	el.wg.Add(1)
	go el.eventLoop()

	// 仅在异步回调模式启动回调分发器
	if !inline {
		el.wg.Add(1)
		go el.callbackDispatcher()
	}

	select {
	case <-startedCh:
		return nil
	case <-time.After(50 * time.Millisecond):
		// 仍然返回 nil，但打印日志帮助排查启动延迟
		el.logger.Warnf("warning: eventLoop start timeout waiting for readiness")
		return nil
	}
}

// Stop 停止逻辑引擎，等待所有事件处理完成。
func (el *EventLoop) Stop() {
	if !el.running.CompareAndSwap(true, false) {
		return
	}

	el.cancel()

	el.wg.Wait()
}

func (el *EventLoop) IsRunning() bool {
	return el.running.Load()
}

// Submit 提交事件到事件循环
func (el *EventLoop) Submit(event Event) error {
	if !el.running.Load() {
		return ErrEventLoopNotRunning
	}
	switch event.Priority {
	case PriorityHigh:
		select {
		case el.highChan <- event:
			return nil
		default:
			return ErrHighQueueFull
		}

	case PriorityMedium:
		select {
		case el.mediumChan <- event:
			return nil
		default:
			return ErrMediumQueueFull
		}

	case PriorityLow:
		select {
		case el.lowChan <- event:
			return nil
		default:
			return ErrLowQueueFull
		}

	default:
		return ErrUnknownPriority
	}
}

// SubmitBlocking 提交事件到事件循环，若队列满则阻塞等待或取消。
func (el *EventLoop) SubmitBlocking(ctx context.Context, event Event) error {
	if !el.running.Load() {
		return ErrEventLoopNotRunning
	}
	switch event.Priority {
	case PriorityHigh:
		select {
		case el.highChan <- event:
			return nil
		case <-ctx.Done():
			return ctx.Err()
		case <-el.ctx.Done():
			return ErrEventLoopStopped
		}
	case PriorityMedium:
		select {
		case el.mediumChan <- event:
			return nil
		case <-ctx.Done():
			return ctx.Err()
		case <-el.ctx.Done():
			return ErrEventLoopStopped
		}
	case PriorityLow:
		select {
		case el.lowChan <- event:
			return nil
		case <-ctx.Done():
			return ctx.Err()
		case <-el.ctx.Done():
			return ErrEventLoopStopped
		}
	default:
		return ErrUnknownPriority
	}
}

// eventLoop 事件循环，持续监听事件通道并处理事件。
func (el *EventLoop) eventLoop() {
	defer el.wg.Done()

	// 通知 Start 就绪（保持原有 started 实现）
	if p := el.started.Load(); p != nil {
		el.mu.Lock()
		startedPtr := el.started.Load()
		if startedPtr != nil {
			close(*startedPtr)
			el.started.Store(nil)
		}
		el.mu.Unlock()
	}

	//el.logger.Infof("EventLoop started")

	// 本地缓冲低优先级事件，保证在无高/中优先级事件时再处理
	var deferredLow []Event

	for {
		// 快速检查取消
		select {
		case <-el.ctx.Done():
			return
		default:
		}

		// 1. 先尽可能清空所有高优先级事件
		for {
			select {
			case ev := <-el.highChan:
				el.handleEvent(ev)
			default:
				goto drainMedium
			}
		}

	drainMedium:
		// 2. 然后尽可能清空所有中优先级事件
		for {
			select {
			case ev := <-el.mediumChan:
				el.handleEvent(ev)
			default:
				goto handleDeferred
			}
		}

	handleDeferred:
		// 3. 若有 deferred low，优先处理（在处理前会再次清空高/中）
		if len(deferredLow) > 0 {
			ev := deferredLow[0]
			deferredLow = deferredLow[1:]
			el.handleEvent(ev)
			continue
		}

		// 4. 阻塞等待任一个事件到达：如果收到 low，先放入 deferred 再回到顶部继续优先级检查
		select {
		case ev := <-el.highChan:
			el.handleEvent(ev)
			continue
		case ev := <-el.mediumChan:
			el.handleEvent(ev)
			continue
		case ev := <-el.lowChan:
			// 不立即处理，先缓冲，回到循环顶部会优先尝试清空 high/medium
			deferredLow = append(deferredLow, ev)
			continue
		case <-el.ctx.Done():
			return
		}
	}
}

// handleEvent 事件处理：内置上下文超时/取消判断
func (el *EventLoop) handleEvent(event Event) {
	evCtx := event.Ctx
	if evCtx == nil {
		evCtx = el.ctx
	}

	// 1. 优先检查上下文是否已取消/超时
	select {
	case <-evCtx.Done():
		if event.Callback != nil {
			el.deliverResult(event, Result{Err: evCtx.Err()})
		}
		return
	default:
	}

	// 2. 调用业务处理器处理事件（保护 processor 为空）
	if el.processor == nil {
		// 没有处理器，直接回调错误结果
		if event.Callback != nil {
			el.deliverResult(event, Result{Err: ErrNoEventProcessor})
		}
		return
	}

	// 3. 执行处理（由业务负责不阻塞太久）
	result := el.processor.Process(event)

	// 4. 统一回调投递（根据模式选择 inline 或异步）
	if event.Callback != nil {
		el.deliverResult(event, result)
	}
}

// 供 eventLoop 调用的辅助：将回调项入队，不在事件循环中直接阻塞很久
func (el *EventLoop) enqueueCallback(item callbackItem) {
	el.mu.Lock()
	timeout := el.cbTimeout
	el.mu.Unlock()

	// 尝试快速入队
	select {
	case el.callbackCh <- item:
		return
	default:
	}

	// 队列已满时，启动后台尝试（避免阻塞主事件循环）
	// 该 goroutine 会尝试在一定时间内把回调放入分发器队列
	go func() {
		timer := time.NewTimer(timeout)
		defer timer.Stop()
		for {
			select {
			case el.callbackCh <- item:
				return
			case <-item.ctx.Done():
				// 回调上下文已取消，直接放弃
				return
			case <-el.ctx.Done():
				return
			case <-timer.C:
				// 超时后放弃并记录
				el.logger.Warnf("enqueue callback timeout, discard result")
				return
			}
		}
	}()
}

// 回调分发器：负责将 result 可靠地送到目标回调 channel，允许重试直到超时或取消。
func (el *EventLoop) callbackDispatcher() {
	defer el.wg.Done()
	for {
		select {
		case <-el.ctx.Done():
			return
		case item := <-el.callbackCh:
			el.mu.Lock()
			timeout := el.cbTimeout
			el.mu.Unlock()

			deadline := time.After(timeout)
			for {
				select {
				case item.cb <- item.res:
					goto next
				case <-item.ctx.Done():
					// 回调方取消
					goto next
				case <-el.ctx.Done():
					goto next
				case <-deadline:
					// 超时后放弃并记录
					el.logger.Warnf("callback deliver timeout, discard result")
					goto next
				}
			}
		next:
			// continue to next callback
		}
	}
}

// handleEvent 中回调投递部分替换为下列逻辑，保证在 inline 模式下在当前事件循环 goroutine 内投递。
func (el *EventLoop) deliverResult(event Event, result Result) {
	if event.Callback == nil {
		return
	}

	el.mu.Lock()
	inline := el.cbInline
	timeout := el.cbTimeout
	el.mu.Unlock()

	// 选择同步（inline）或异步（enqueue）投递
	if inline {
		// 在事件循环 goroutine 内进行投递，使用 event.Ctx 或 el.ctx 作为控制上下文
		deliverCtx := event.Ctx
		if deliverCtx == nil {
			deliverCtx = el.ctx
		}

		// 如果 cbTimeout == 0，表示无限等待（可能阻塞事件循环）
		if timeout == 0 {
			select {
			case event.Callback <- result:
				// delivered
			case <-deliverCtx.Done():
				// callback receiver canceled
			case <-el.ctx.Done():
				// loop stopped
			}
			return
		}

		// 带超时的同步投递，超时后记录并放弃（可根据需求改为返回错误以触发上游重试）
		timer := time.NewTimer(timeout)
		defer timer.Stop()
		select {
		case event.Callback <- result:
			// delivered
		case <-deliverCtx.Done():
			// receiver canceled
		case <-el.ctx.Done():
			// loop stopped
		case <-timer.C:
			el.logger.Warnf("inline callback deliver timeout, discard result for event priority: %v", event.Priority)
		}
		return
	}

	// 异步模式：使用现有的 enqueueCallback 来可靠投递（可能重试直到超时）
	item := callbackItem{
		cb:  event.Callback,
		res: result,
		ctx: event.Ctx,
	}
	el.enqueueCallback(item)
}

// callbackItem 用于回调分发队列项
type callbackItem struct {
	cb  chan Result
	res Result
	ctx context.Context
}
