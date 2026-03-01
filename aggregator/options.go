package aggregator

import "time"

const (
	defaultLimit   = 20
	defaultTimeout = 0
	defaultRetry   = 0
)

type options struct {
	limit   int
	timeout time.Duration
	retry   int
}

type Option func(*options)

// WithLimit 设置最大并发数
func WithLimit(limit int) Option {
	return func(o *options) { o.limit = limit }
}

// WithTimeout 设置整体执行的超时时间
func WithTimeout(timeout time.Duration) Option {
	return func(o *options) { o.timeout = timeout }
}

// WithRetry 设置任务失败后的重试次数
func WithRetry(retry int) Option {
	return func(o *options) { o.retry = retry }
}
