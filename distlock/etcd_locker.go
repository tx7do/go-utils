package distlock

import (
	"context"
	"errors"
	"sync"

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
	return &EtcdLocker{client: cli, opts: o, ownClient: true}, nil
}

// NewEtcdWithClient 使用已有 etcd client 创建 locker（便于复用连接）。
func NewEtcdWithClient(cli *clientv3.Client, opts EtcdOptions) Locker {
	return &EtcdLocker{client: cli, opts: opts.withDefaults(), ownClient: false}
}

// EtcdLocker 基于 etcd mutex 的分布式锁实现。
type EtcdLocker struct {
	client    *clientv3.Client
	opts      EtcdOptions
	ownClient bool
}

func (l *EtcdLocker) Obtain(ctx context.Context, key string) (Lock, error) {
	session, err := concurrency.NewSession(l.client,
		concurrency.WithTTL(l.opts.SessionTTL),
		// Session 需要绑定任务上下文，不能在创建后立刻 cancel；
		// 否则 keepalive 会停止，锁会很快失效。
		concurrency.WithContext(ctx),
	)
	if err != nil {
		return nil, err
	}

	mu := concurrency.NewMutex(session, key)
	if err = mu.TryLock(ctx); err != nil {
		_ = session.Close()
		if errors.Is(err, concurrency.ErrLocked) {
			return nil, ErrNotObtained
		}
		if errors.Is(err, concurrency.ErrSessionExpired) {
			return nil, ErrNotObtained
		}
		return nil, err
	}

	return &etcdLock{session: session, mu: mu, key: key}, nil
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
	unlockErr := l.mu.Unlock(ctx)
	sessionErr := l.session.Close()
	if unlockErr != nil && !errors.Is(unlockErr, concurrency.ErrLockReleased) {
		return unlockErr
	}
	return sessionErr
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
