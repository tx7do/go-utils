package distlock

import (
	"context"
	"errors"
	"fmt"
	"log"
	"sync"
	"time"

	clientv3 "go.etcd.io/etcd/client/v3"
	"go.etcd.io/etcd/client/v3/concurrency"
)

// NewEtcd 创建基于 etcd concurrency.Session + Mutex 的锁实现。
func NewEtcd(endpoints []string, opts EtcdOptions) (Locker, error) {
	o := opts.withDefaults()
	cli, err := clientv3.New(clientv3.Config{
		Endpoints:   endpoints,
		DialTimeout: o.DialTimeout,
	})
	if err != nil {
		return nil, err
	}
	return &EtcdLocker{
		client:    cli,
		opts:      o,
		ownClient: true,
	}, nil
}

// NewEtcdWithClient 使用已有 etcd client 创建 locker（便于复用连接）。
func NewEtcdWithClient(cli *clientv3.Client, opts EtcdOptions) Locker {
	return &EtcdLocker{
		client:    cli,
		opts:      opts.withDefaults(),
		ownClient: false,
	}
}

// EtcdLocker 基于 etcd mutex 的分布式锁实现。
type EtcdLocker struct {
	client    *clientv3.Client
	opts      EtcdOptions
	ownClient bool
}

func (l *EtcdLocker) Obtain(ctx context.Context, key string, opts ...LockOption) (Lock, error) {
	cfg := defaultLockConfig()
	for _, opt := range opts {
		opt(cfg)
	}

	// 非阻塞模式：保持原有行为（快速失败）
	if !cfg.blockWait {
		return l.obtainTryLock(ctx, key)
	}

	// 阻塞等待模式
	return l.obtainBlockLock(ctx, key, cfg)
}

// obtainTryLock 原有逻辑提取，保持向后兼容
func (l *EtcdLocker) obtainTryLock(ctx context.Context, key string) (Lock, error) {
	session, err := concurrency.NewSession(l.client,
		concurrency.WithTTL(l.opts.SessionTTL),
		concurrency.WithContext(context.Background()),
	)
	if err != nil {
		return nil, err
	}

	mu := concurrency.NewMutex(session, key)
	if err = mu.TryLock(ctx); err != nil {
		_ = session.Close()
		if errors.Is(err, concurrency.ErrLocked) || errors.Is(err, concurrency.ErrSessionExpired) {
			return nil, ErrNotObtained
		}
		return nil, err
	}

	return &etcdLock{session: session, mu: mu, key: key}, nil
}

// obtainBlockLock 阻塞等待实现（带指数退避）
func (l *EtcdLocker) obtainBlockLock(ctx context.Context, key string, cfg *lockConfig) (Lock, error) {
	var (
		startTime  = time.Now()
		attempts   = 0
		retryDelay = cfg.retryDelay
	)

	for {
		// 检查整体超时
		if cfg.maxWaitTime > 0 && time.Since(startTime) >= cfg.maxWaitTime {
			return nil, fmt.Errorf("lock wait timeout: %w", ErrNotObtained)
		}

		// 创建 session（每次尝试新建，避免复用过期 session）
		session, err := concurrency.NewSession(l.client,
			concurrency.WithTTL(l.opts.SessionTTL),
			concurrency.WithContext(context.Background()),
		)
		if err != nil {
			return nil, fmt.Errorf("create session failed: %w", err)
		}

		mu := concurrency.NewMutex(session, key)

		// 使用带超时的子上下文尝试获取锁
		lockCtx, cancel := context.WithTimeout(ctx, l.opts.LockTimeout)
		err = mu.Lock(lockCtx) // 注意：使用 Lock 而非 TryLock
		cancel()               // 及时释放资源

		if err == nil {
			// 成功获取锁
			return &etcdLock{session: session, mu: mu, key: key}, nil
		}

		// 获取失败，清理资源
		_ = session.Close()

		// 区分错误类型：可重试 vs 不可重试
		if errors.Is(err, concurrency.ErrLocked) {
			// 锁被占用，按策略重试
			attempts++
			select {
			case <-ctx.Done():
				return nil, ctx.Err()
			case <-time.After(retryDelay):
				// 指数退避：上限 2s
				if retryDelay < 2*time.Second {
					retryDelay *= 2
				}
				continue
			}
		}

		// 其他错误（网络/权限等）直接返回
		if !errors.Is(err, context.DeadlineExceeded) && !errors.Is(err, context.Canceled) {
			return nil, fmt.Errorf("acquire lock failed: %w", err)
		}

		// 上下文取消/超时
		return nil, err
	}
}

func (l *EtcdLocker) Close() error {
	if !l.ownClient || l.client == nil {
		return nil
	}
	return l.client.Close()
}

type etcdLock struct {
	session *concurrency.Session
	mu      *concurrency.Mutex
	key     string
}

func (l *etcdLock) Key() string { return l.key }

func (l *etcdLock) Release(ctx context.Context) error {
	if err := l.mu.Unlock(ctx); err != nil && !errors.Is(err, concurrency.ErrLockReleased) {
		return fmt.Errorf("unlock failed: %w", err)
	}

	if err := l.session.Close(); err != nil {
		log.Printf("session close error (lock already released): %v", err)
	}

	return nil
}

// Refresh 对 etcd 会话锁是 no-op：续租由 session keepalive 自动完成。
func (l *etcdLock) Refresh(ctx context.Context) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
		return nil
	}
}

// StartRefresh 对 etcd 锁无需主动续租，这里只监听 session 失效事件并回调 onError。
func (l *etcdLock) StartRefresh(ctx context.Context, onError func(err error)) (stop func()) {
	watchCtx, cancel := context.WithCancel(ctx)
	done := make(chan struct{})

	go func() {
		defer close(done)
		select {
		case <-watchCtx.Done():
			return
		case <-l.session.Done():
			if onError != nil {
				onError(ErrNotObtained)
			}
			return
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

// IsLocked 检查 key 是否已被锁定；仅供监控/调试使用，不能替代 Obtain 的原子性保证。
func (l *EtcdLocker) IsLocked(ctx context.Context, key string) (bool, error) {
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
