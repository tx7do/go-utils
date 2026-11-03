# eventloop

English | [简体中文](README.zh-CN.md)

A single-goroutine sequential event loop: three-priority scheduling, an optional frame-driven mode, request-reply callbacks, with injectable logging and metrics interfaces.

## Features

- **Three-priority scheduling**: `PriorityHigh` / `PriorityMedium` / `PriorityLow` — every cycle drains High first, then Medium, and defers Low into a local buffer (re-checking high/medium before each deferred item), so high-priority events are always handled earlier;
- **Single-goroutine sequential processing**: every `Process` call runs serially on the event-loop goroutine, giving predictable ordering; submission is concurrency-safe;
- **Frame-driven mode** (`frameDriven=true`): processes events against a fixed frame rate and per-frame time budget, suited to game main loops, real-time rendering and other time-sliced scenarios;
- **Request-reply pattern**: `NewRequestEvent` hands you the event and its reply channel in one call;
- **Two callback delivery modes**: inline (synchronously on the event-loop goroutine) and async (a dedicated dispatcher, the default);
- **Injectable logging and metrics**: `SetLogger` / `SetMetrics`, defaulting to `NoopLogger` / `NoopMetrics`;
- **Graceful stop**: `Stop` cancels the context and waits for the loop goroutines to exit.

## Installation

Standalone module with zero third-party dependencies:

```bash
go get github.com/tx7do/go-utils/eventloop@latest
```

```go
import "github.com/tx7do/go-utils/eventloop"
```

## Quick Start

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
	fmt.Println("process:", ev.Data) // the field is named Data
	return eventloop.Result{}
}

func main() {
	// NewEventLoop(bufferSize, processor, frameDriven): bufferSize<=0 falls back to the default 100
	el := eventloop.NewEventLoop(64, &myProc{}, false)

	// Optional: callback mode. false = async dispatch (default), 1s timeout
	el.SetCallbackInline(false, time.Second)

	if err := el.Start(); err != nil {
		panic(err)
	}
	defer el.Stop()

	// Request-reply: one call yields the event and its reply channel
	ev, reply := eventloop.NewRequestEvent("calc.add", "hello")
	ev.Priority = eventloop.PriorityHigh // Event is returned by value; fields can be set before Submit
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

## Core API

### Creation & Lifecycle

```go
// bufferSize<=0 falls back to the default 100; processor may be nil (callback-carrying events
// then receive ErrNoEventProcessor via their callback); frameDriven=true selects the frame loop.
func NewEventLoop(bufferSize int, processor EventProcessor, frameDriven bool) *EventLoop

func (el *EventLoop) Start() error          // idempotent: returns nil if already running; waits for the readiness signal (up to 50ms, still returns nil on timeout)
func (el *EventLoop) Stop()                 // graceful stop: cancels the context and waits for all goroutines
func (el *EventLoop) IsRunning() bool
```

### Submitting Events

```go
// Non-blocking submit: a full queue returns the matching Err*QueueFull immediately; an
// unrecognized priority returns ErrUnknownPriority; a stopped loop returns ErrEventLoopNotRunning.
func (el *EventLoop) Submit(event Event) error

// Blocking submit: waits for queue space, the caller's ctx cancellation (returns ctx.Err()),
// or the loop stopping (returns ErrEventLoopStopped).
func (el *EventLoop) SubmitBlocking(ctx context.Context, event Event) error
```

### Events & Request-Reply (`event.go`)

```go
type Event struct {
    Priority Priority        // priority (default PriorityLow)
    Type     string          // event type
    Data     any             // event payload (note: the field is Data, not Payload)
    Callback chan Result     // reply channel (may be nil)
    Ctx      context.Context // event context (may be nil; falls back to the loop's ctx)
    TS       time.Time       // set to time.Now() by NewEvent
}

type Result struct {
    Data any
    Err  error
}

// Build an event: default PriorityLow, TS set automatically, options override via chaining.
func NewEvent(typ string, payload any, opts ...EventOption) Event

// Request-reply: returns the event and a capacity-1 reply channel (WithCallback applied internally).
func NewRequestEvent(typ string, payload any) (Event, chan Result)

func WithPriority(p Priority) EventOption
func WithContext(ctx context.Context) EventOption
func WithCallback(cb chan Result) EventOption
```

### Configuration Injection

```go
// Inject a Logger (NoopLogger / NewStdLogger or your own); nil resets to NoopLogger.
func (el *EventLoop) SetLogger(l Logger)

// Inject a Metrics implementation (e.g. NewSimpleMetrics()); nil resets to NoopMetrics.
func (el *EventLoop) SetMetrics(m Metrics)

// Switch callback delivery mode: inline=true delivers synchronously on the event-loop
// goroutine (may block the loop); timeout bounds that delivery, 0 meaning wait forever.
// Prefer setting this before Start().
func (el *EventLoop) SetCallbackInline(inline bool, timeout time.Duration)

// Set frame-driven parameters. Non-positive interval/budget keep the current value;
// a negative maxLow keeps the current value (0 is accepted). Prefer calling before Start().
func (el *EventLoop) SetFrameParameters(interval, budget, maxLow time.Duration)
```

### Observation

```go
// Instantaneous snapshot of the internal channel lengths (non-blocking): High / Medium / Low / Callback.
func (el *EventLoop) QueueLengths() QueueLengths
```

## Frame-Driven Mode

With `NewEventLoop(…, true)`, the loop beats at `frameInterval` and processes per frame:

1. Each frame drains all High events first, then all Medium, then processes Low within the remaining budget;
2. When the frame budget (`frameBudget`) is exhausted it returns immediately, leaving the rest for the next frame; Low is additionally bounded by `maxLowTime`;
3. Default constants (`event.go`): `FrameInterval = 50ms` (20Hz), `FrameBudget = 25ms`, `MaxLowTime = 2ms`.

Tuning goes through `SetFrameParameters` (before `Start()` or at any time) — the three fields are unexported and cannot be assigned from outside the package:

```go
el := eventloop.NewEventLoop(64, myProcessor, true)
el.SetFrameParameters(16*time.Millisecond, 10*time.Millisecond, 2*time.Millisecond)
_ = el.Start()
```

Keep `frameDriven=false` (free-running: events are processed as they arrive, no frame cadence) for non-realtime scenarios.

## Callback Delivery Semantics

Two modes (switch with `SetCallbackInline`, ideally before `Start()`):

- **Async (default)**: callback items enter an internal `callbackCh`; a dedicated `callbackDispatcher` goroutine delivers them reliably (retrying until delivered, the callback side cancels, or `cbTimeout` expires and the result is discarded with a log). Callbacks do not run on the event-loop goroutine; `Start()` launches the dispatcher only in this mode.
- **Synchronous (inline)**: the event-loop goroutine sends the `Result` straight to `Event.Callback`; `cbTimeout=0` waits forever (potentially blocking the loop for a long time), otherwise it gives up and logs after the timeout. Guarantees delivery on the loop thread but applies backpressure to the loop.

Queue-full fallback: when `callbackCh` is full, a background goroutine retries the enqueue until delivered, the event's ctx is cancelled, the loop stops, or the timeout gives up — under extreme pressure this can spawn many short-lived goroutines, so production deployments may swap in a fixed worker pool.

Consumers should drain the `Callback` channel from their own goroutine (the `NewRequestEvent` reply channel has capacity 1 — read it promptly after `Submit`).

## Metrics

The `Metrics` interface counts submitted/dropped/processed per priority, callback discards and inline timeouts, and observes processing durations; `Snapshot()` returns a read-only snapshot (including per-priority average processing nanoseconds). Built-in implementations:

- `NoopMetrics{}`: the default, zero overhead;
- `NewSimpleMetrics()`: atomic-counter based, suitable for tests and lightweight monitoring.

```go
m := eventloop.NewSimpleMetrics()
el.SetMetrics(m)
// after running for a while:
snap := m.Snapshot()
```

`el.QueueLengths()` additionally observes instantaneous per-queue backlog.

## Error Sentinels (`errors.go`)

```go
var (
    ErrEventLoopNotRunning           // loop not started, from Submit/SubmitBlocking
    ErrEventLoopStopped              // the loop was Stop()ed while SubmitBlocking waited
    ErrNoEventProcessor              // processor is nil: callback-carrying events get this via their callback
    ErrHighQueueFull                 // high-priority queue full
    ErrMediumQueueFull               // medium-priority queue full
    ErrLowQueueFull                  // low-priority queue full
    ErrUnknownPriority               // priority not one of High/Medium/Low
)
```

## Priority & Concurrency Semantics

- Every `Process` runs serially on the event-loop goroutine — predictable ordering, but not pinned to a specific OS thread; if you need strict thread affinity, use `runtime.LockOSThread()` in your business code and design carefully;
- `Submit` / `SubmitBlocking` are safe for concurrent callers; enqueueing is concurrency-safe;
- In free-running mode Low events are buffered locally (deferredLow), and high/medium are re-drained before each deferred item — a sustained high-priority flood can starve Low;
- An event whose `Ctx` is already cancelled or expired never reaches the processor: callback-carrying events receive `ctx.Err()` through the callback; those without a callback are dropped silently.

## Notes / Caveats

- **Do not block for long inside `Process`**: the whole loop stalls behind that one call (and the frame budget is blown in frame-driven mode); parallelize slow work inside the processor itself;
- **Frame-loop logging uses the standard log package**: the frame loop's start/stop and "frame budget exceeded" messages go through plain `log.Println`, not the Logger injected via `SetLogger` (the injected Logger covers callback discard/timeout warnings);
- `Submit` is non-blocking — use `SubmitBlocking` or add your own buffering when you need backpressure;
- Events still sitting in the channels when `Stop` returns were never processed; to "drain then stop", wait for `QueueLengths()` to reach zero in your own layer before calling `Stop`.
