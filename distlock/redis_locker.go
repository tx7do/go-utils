package distlock

import (
	"context"
	"errors"
	"sync"
	"time"

	"github.com/bsm/redislock"
	"github.com/redis/go-redis/v9"
)

type redisLock struct {
	inner    *redislock.Lock
	ttl      time.Duration
	interval time.Duration
}

func (l *redisLock) Key() string { return l.inner.Key() }

func (l *redisLock) Release(ctx context.Context) error {
	return l.inner.Release(ctx)
}

func (l *redisLock) Refresh(ctx context.Context) error {
	return l.inner.Refresh(ctx, l.ttl, nil)
}

// StartRefresh 启动后台 goroutine，按 interval 定期续期锁。
// 续期失败（包括 context 取消）时协程自行退出。
// 返回的 stop 函数会取消协程并等待其完全退出，可安全多次调用。
func (l *redisLock) StartRefresh(ctx context.Context, onError func(err error)) (stop func()) {
	refreshCtx, cancel := context.WithCancel(ctx)
	done := make(chan struct{})

	go func() {
		defer close(done)
		defer cancel() // 确保 context 资源始终释放

		ticker := time.NewTicker(l.interval)
		defer ticker.Stop()

		for {
			select {
			case <-refreshCtx.Done():
				return
			case <-ticker.C:
				if err := l.Refresh(refreshCtx); err != nil {
					if onError != nil {
						onError(err)
					}
					return
				}
			}
		}
	}()

	var once sync.Once
	return func() {
		once.Do(func() {
			cancel()
			<-done
		})
	}
}

// RedisLocker 是基于 redislock 的分布式锁实现。
type RedisLocker struct {
	client *redislock.Client
	opts   Options
}

// NewRedisLocker 用给定的 Redis 客户端创建 Locker。
// opts 零值字段自动补全默认值。
func NewRedisLocker(rdb *redis.Client, opts Options) Locker {
	return &RedisLocker{
		client: redislock.New(rdb),
		opts:   opts.withDefaults(),
	}
}

// Obtain 尝试获取 key 对应的锁。
//   - 成功：返回 Lock，调用方必须在完成后调用 Release。
//   - 锁已被持有：返回 ErrNotObtained（可用 errors.Is 判断）。
//   - 其他错误：返回底层 Redis 错误。
func (l *RedisLocker) Obtain(ctx context.Context, key string, opts ...LockOption) (Lock, error) {
	inner, err := l.client.Obtain(ctx, key, l.opts.TTL, &redislock.Options{
		RetryStrategy: redislock.LimitRetry(
			redislock.LinearBackoff(l.opts.RetryDelay),
			l.opts.MaxRetries,
		),
	})
	if err != nil {
		if errors.Is(err, redislock.ErrNotObtained) {
			return nil, ErrNotObtained
		}
		return nil, err
	}
	return &redisLock{inner: inner, ttl: l.opts.TTL, interval: l.opts.RefreshInterval}, nil
}

// IsLocked 检查 key 是否已被锁定；仅供监控/调试使用，不能替代 Obtain 的原子性保证。
func (l *RedisLocker) IsLocked(ctx context.Context, key string) (bool, error) {
	lock, err := l.Obtain(ctx, key)
	if err != nil {
		if errors.Is(err, ErrNotObtained) {
			return true, nil // 已被加锁
		}
		return false, err // 其它错误
	}
	// 获取到锁，说明未加锁，需立即释放
	_ = lock.Release(ctx)
	return false, nil
}

// Close 对 redislock 为 no-op，保留接口一致性。
func (l *RedisLocker) Close() error {
	_ = l
	return nil
}
