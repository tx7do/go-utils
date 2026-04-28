package distlock

import "time"

// LockOption 控制锁获取行为的函数选项
type LockOption func(*lockConfig)

type lockConfig struct {
	blockWait   bool          // 是否阻塞等待
	maxWaitTime time.Duration // 最大等待时间（0 表示无限等待）
	retryDelay  time.Duration // 重试基础间隔（默认 100ms）
}

// WithBlockWait 启用阻塞等待模式
func WithBlockWait(maxWait time.Duration) LockOption {
	return func(cfg *lockConfig) {
		cfg.blockWait = true
		cfg.maxWaitTime = maxWait
	}
}

// WithRetryDelay 设置重试退避间隔（仅阻塞模式生效）
func WithRetryDelay(delay time.Duration) LockOption {
	return func(cfg *lockConfig) {
		cfg.retryDelay = delay
	}
}

// 默认配置
func defaultLockConfig() *lockConfig {
	return &lockConfig{
		blockWait:   false,
		maxWaitTime: 0,
		retryDelay:  100 * time.Millisecond,
	}
}
