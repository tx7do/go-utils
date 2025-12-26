package eventloop

import "errors"

var (
	ErrEventLoopNotRunning = errors.New("event loop not running")
	ErrEventLoopStopped    = errors.New("event loop stopped")
	ErrNoEventProcessor    = errors.New("no event processor")

	ErrHighQueueFull   = errors.New("high priority queue full")
	ErrMediumQueueFull = errors.New("medium priority queue full")
	ErrLowQueueFull    = errors.New("low priority queue full")

	ErrUnknownPriority = errors.New("unknown priority")
)
