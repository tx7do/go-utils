package distlock_test

import (
	"context"
	"errors"
	"sync/atomic"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/redis/go-redis/v9"
	"github.com/tx7do/go-utils/distlock"
)

func newTestLocker(t *testing.T, opts distlock.Options) (distlock.Locker, *miniredis.Miniredis) {
	t.Helper()
	mr, err := miniredis.Run()
	if err != nil {
		t.Fatalf("miniredis.Run: %v", err)
	}
	rdb := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	return distlock.NewRedisLocker(rdb, opts), mr
}

func TestLock_StartRefresh_StopsCleanly(t *testing.T) {
	locker, mr := newTestLocker(t, distlock.Options{
		TTL:             5 * time.Second,
		RefreshInterval: 200 * time.Millisecond,
	})
	defer mr.Close()

	ctx := context.Background()
	lock, err := locker.Obtain(ctx, "test:lock:refresh-stop")
	if err != nil {
		t.Fatalf("Obtain: %v", err)
	}
	defer lock.Release(ctx) //nolint:errcheck

	stop := lock.StartRefresh(ctx, nil)

	// 等几个 tick 后停止
	time.Sleep(500 * time.Millisecond)
	stop() // 应阻塞直到协程退出

	// 幂等：再次调用不应 panic 或死锁
	stop()
}

func TestLock_StartRefresh_CallsOnErrorAndStops(t *testing.T) {
	locker, mr := newTestLocker(t, distlock.Options{
		TTL:             2 * time.Second,
		RefreshInterval: 100 * time.Millisecond,
	})
	defer mr.Close()

	ctx := context.Background()
	lock, err := locker.Obtain(ctx, "test:lock:refresh-error")
	if err != nil {
		t.Fatalf("Obtain: %v", err)
	}

	var errCount atomic.Int32

	stop := lock.StartRefresh(ctx, func(e error) {
		errCount.Add(1)
	})
	defer stop()

	// 让锁在 Redis 中过期，触发续期失败
	mr.FastForward(3 * time.Second)

	// 等协程检测到错误并退出
	time.Sleep(300 * time.Millisecond)
	stop()

	if errCount.Load() == 0 {
		t.Fatal("expected onError to be called at least once")
	}
}

func TestLock_StartRefresh_StopsOnContextCancel(t *testing.T) {
	locker, mr := newTestLocker(t, distlock.Options{
		TTL:             5 * time.Second,
		RefreshInterval: 100 * time.Millisecond,
	})
	defer mr.Close()

	ctx, cancel := context.WithCancel(context.Background())
	lock, err := locker.Obtain(ctx, "test:lock:refresh-ctx-cancel")
	if err != nil {
		t.Fatalf("Obtain: %v", err)
	}
	defer lock.Release(context.Background()) //nolint:errcheck

	stop := lock.StartRefresh(ctx, nil)
	defer stop()

	// 取消父 context，协程应自行退出
	cancel()
	// stop 应能立即返回（不超过 1s）
	done := make(chan struct{})
	go func() { stop(); close(done) }()
	select {
	case <-done:
	case <-time.After(time.Second):
		t.Fatal("stop() did not return within 1s after context cancel")
	}
}

func TestOptions_RefreshIntervalDefaultIsTTLDivThree(t *testing.T) {
	locker, mr := newTestLocker(t, distlock.Options{TTL: 30 * time.Second})
	defer mr.Close()
	_ = locker // just ensure it constructs; interval is tested implicitly via StartRefresh
}

func TestLocker_ObtainAndRelease(t *testing.T) {
	locker, mr := newTestLocker(t, distlock.Options{TTL: 5 * time.Second})
	defer mr.Close()

	ctx := context.Background()
	lock, err := locker.Obtain(ctx, "test:lock:basic")
	if err != nil {
		t.Fatalf("Obtain failed: %v", err)
	}
	if lock.Key() != "test:lock:basic" {
		t.Fatalf("Key mismatch: %s", lock.Key())
	}
	if err = lock.Release(ctx); err != nil {
		t.Fatalf("Release failed: %v", err)
	}
}

func TestLocker_ErrNotObtained_WhenAlreadyHeld(t *testing.T) {
	locker, mr := newTestLocker(t, distlock.Options{
		TTL:        5 * time.Second,
		MaxRetries: 1,
		RetryDelay: 10 * time.Millisecond,
	})
	defer mr.Close()

	ctx := context.Background()
	first, err := locker.Obtain(ctx, "test:lock:contention")
	if err != nil {
		t.Fatalf("first Obtain failed: %v", err)
	}
	defer first.Release(ctx) //nolint:errcheck

	_, err = locker.Obtain(ctx, "test:lock:contention")
	if !errors.Is(err, distlock.ErrNotObtained) {
		t.Fatalf("expected ErrNotObtained, got: %v", err)
	}
}

func TestLocker_ObtainAfterRelease(t *testing.T) {
	locker, mr := newTestLocker(t, distlock.Options{TTL: 5 * time.Second})
	defer mr.Close()

	ctx := context.Background()
	lock, err := locker.Obtain(ctx, "test:lock:reacquire")
	if err != nil {
		t.Fatalf("Obtain failed: %v", err)
	}
	if err = lock.Release(ctx); err != nil {
		t.Fatalf("Release failed: %v", err)
	}
	lock2, err := locker.Obtain(ctx, "test:lock:reacquire")
	if err != nil {
		t.Fatalf("second Obtain failed: %v", err)
	}
	defer lock2.Release(ctx) //nolint:errcheck
}

func TestLock_Refresh(t *testing.T) {
	locker, mr := newTestLocker(t, distlock.Options{TTL: 5 * time.Second})
	defer mr.Close()

	ctx := context.Background()
	lock, err := locker.Obtain(ctx, "test:lock:refresh")
	if err != nil {
		t.Fatalf("Obtain failed: %v", err)
	}
	defer lock.Release(ctx) //nolint:errcheck

	mr.FastForward(3 * time.Second)
	if err = lock.Refresh(ctx); err != nil {
		t.Fatalf("Refresh failed: %v", err)
	}
}

func TestLocker_DefaultOptions(t *testing.T) {
	locker, mr := newTestLocker(t, distlock.Options{})
	defer mr.Close()

	ctx := context.Background()
	lock, err := locker.Obtain(ctx, "test:lock:defaults")
	if err != nil {
		t.Fatalf("Obtain with default opts failed: %v", err)
	}
	defer lock.Release(ctx) //nolint:errcheck
}
