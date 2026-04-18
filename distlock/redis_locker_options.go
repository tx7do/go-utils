package distlock

import "time"

const (
	defaultTTL        = 30 * time.Second
	defaultMaxRetries = 10
	defaultRetryDelay = 100 * time.Millisecond
)

// Options 配置锁的获取与续期行为。零值字段自动使用内置默认值。
type Options struct {
	// TTL 是锁的有效期；默认 30s。
	TTL time.Duration

	// MaxRetries 是争抢锁时的最大重试次数；默认 10 次。
	MaxRetries int

	// RetryDelay 是线性退避的步长；默认 100ms。
	RetryDelay time.Duration

	// RefreshInterval 是 StartRefresh 的续期间隔；默认 TTL/3。
	RefreshInterval time.Duration
}

func (o Options) withDefaults() Options {
	if o.TTL <= 0 {
		o.TTL = defaultTTL
	}
	if o.MaxRetries <= 0 {
		o.MaxRetries = defaultMaxRetries
	}
	if o.RetryDelay <= 0 {
		o.RetryDelay = defaultRetryDelay
	}
	if o.RefreshInterval <= 0 {
		o.RefreshInterval = o.TTL / 3
	}
	return o
}
