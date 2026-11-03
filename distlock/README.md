# distlock

English | [简体中文](README.zh-CN.md)

`distlock` is a standalone Go module (`github.com/tx7do/go-utils/distlock`) providing a unified distributed-lock abstraction over Redis and etcd backends, so business code depends on a single interface.

## Features

- Unified interface: business code only depends on `distlock.Locker` / `distlock.Lock`
- Swappable backends: Redis (based on [github.com/bsm/redislock](https://github.com/bsm/redislock)) and etcd (based on [go.etcd.io/etcd/client/v3/concurrency](https://pkg.go.dev/go.etcd.io/etcd/client/v3/concurrency))
- Safe lifecycle: `StartRefresh` returns an idempotent `stop()`, avoiding refresh-after-release races
- Consistent error semantics: failed acquisitions map to `distlock.ErrNotObtained` (check with `errors.Is`)
- Monitoring helper: `Locker.IsLocked` checks whether a key is locked
- etcd backend supports a blocking-wait acquisition mode (exponential backoff, capped at 2s)

## Installation

This package is a standalone module (it has its own `go.mod`), not a subpackage of the root module `github.com/tx7do/go-utils`:

```bash
go get github.com/tx7do/go-utils/distlock
```

```go
import "github.com/tx7do/go-utils/distlock"
```

## Core Interfaces

```go
type Locker interface {
	// Obtain tries to acquire the lock for key.
	//   - Success: returns a Lock; the caller must Release when done.
	//   - Already held: returns ErrNotObtained (check with errors.Is).
	//   - Other errors: the underlying error is returned.
	Obtain(ctx context.Context, key string, opts ...LockOption) (Lock, error)

	// Close releases backend resources; call when no longer needed.
	Close() error

	// IsLocked reports whether key is currently locked; for monitoring/debugging
	// only — it does not replace the atomicity guarantees of Obtain.
	IsLocked(ctx context.Context, key string) (bool, error)
}

type Lock interface {
	Key() string
	Release(ctx context.Context) error                       // returns the underlying error if the lock already expired/released
	Refresh(ctx context.Context) error                       // extends the lock once with the original TTL
	StartRefresh(ctx context.Context, onError func(err error)) (stop func())
}
```

`StartRefresh` behavior: a background goroutine calls `Refresh` every `RefreshInterval`; on a refresh failure (including context cancellation) it invokes `onError` (may be nil) and the goroutine exits automatically; the returned `stop()` cancels the goroutine and blocks until it exits, and is safe to call multiple times (idempotent).

## Quick Start

```go
package main

import (
	"context"
	"errors"
	"log"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/tx7do/go-utils/distlock"
)

func main() {
	rdb := redis.NewClient(&redis.Options{Addr: "localhost:6379"})
	locker := distlock.NewRedisLocker(rdb, distlock.Options{TTL: 30 * time.Second})
	defer locker.Close()

	ctx := context.Background()
	lock, err := locker.Obtain(ctx, "job:nightly-report")
	if err != nil {
		if errors.Is(err, distlock.ErrNotObtained) {
			log.Println("lock held by another node")
			return
		}
		log.Fatal(err)
	}

	stop := lock.StartRefresh(ctx, func(err error) { log.Printf("refresh failed: %v", err) })

	// ... business logic ...

	stop()                                 // stop refreshing first
	_ = lock.Release(context.Background()) // then release
}
```

## Redis Backend

```go
rdb := redis.NewClient(&redis.Options{Addr: "localhost:6379"})

locker := distlock.NewRedisLocker(rdb, distlock.Options{
	TTL:             30 * time.Second,
	MaxRetries:      10,
	RetryDelay:      100 * time.Millisecond,
	RefreshInterval: 10 * time.Second, // default is TTL/3
})
defer locker.Close() // the Redis backend's Close is a no-op
```

`Options` fields (zero/negative values fall back to defaults):

| Field | Description | Default |
|-------|-------------|---------|
| TTL | Lock validity period | 30s |
| MaxRetries | Max retries when contending for the lock (linear backoff) | 10 |
| RetryDelay | Linear backoff step | 100ms |
| RefreshInterval | Renewal interval for `StartRefresh` | TTL/3 |

The acquisition retry strategy is `redislock.LimitRetry(LinearBackoff(RetryDelay), MaxRetries)`.

## etcd Backend

```go
// Option 1: let the package create the etcd client (Close also closes it)
locker, err := distlock.NewEtcd([]string{"127.0.0.1:2379"}, distlock.EtcdOptions{})
if err != nil {
	return err
}
defer locker.Close()

// Option 2: reuse an existing etcd client (Close does not close it)
locker = distlock.NewEtcdWithClient(cli, distlock.EtcdOptions{})
```

`EtcdOptions` fields (zero/negative values fall back to defaults):

| Field | Description | Default | Currently effective |
|-------|-------------|---------|---------------------|
| DialTimeout | etcd connection timeout | 5s | Yes |
| SessionTTL | Session lease TTL (seconds) | 5 | Yes |
| SessionTimeout | Session creation timeout | 5s | **Reserved, not read yet** (session creation does not apply this timeout) |
| LockTimeout | Per-attempt lock timeout; 0 means no auto-release | 0 | Yes (blocking-wait mode only) |

### LockOption (effective on the etcd backend only)

`Obtain` accepts per-call acquisition options; the Redis backend currently ignores them:

```go
lock, err := locker.Obtain(ctx, key,
	distlock.WithBlockWait(5*time.Second),          // blocking-wait mode: wait up to 5s (0 means wait indefinitely)
	distlock.WithRetryDelay(200*time.Millisecond), // base retry interval (blocking mode only)
)
```

- Default is non-blocking: a held lock returns `ErrNotObtained` immediately
- `WithBlockWait(maxWait)`: blocking-wait mode; each attempt creates a fresh session and locks with `mu.Lock`, retrying with exponential backoff (step from `WithRetryDelay`, default 100ms, capped at 2s); on overall timeout it returns a timeout error wrapping `ErrNotObtained`

## Recommended Usage

```go
lock, err := locker.Obtain(ctx, key)
if err != nil {
	if errors.Is(err, distlock.ErrNotObtained) {
		return nil // held by another node
	}
	return err
}

stopRefresh := lock.StartRefresh(ctx, func(err error) {
	// log / alert
})

// ... business logic ...

// Recommended order: stop refreshing first, then release
stopRefresh()
_ = lock.Release(context.Background())
```

## Notes

- `Release` is not guaranteed idempotent: releasing twice may return a backend error (ignore or log as appropriate)
- The `stop()` returned by `StartRefresh` is idempotent and safe to call multiple times
- Long-running tasks must keep and call `stop()` to avoid goroutine leaks
- **The etcd backend's `Refresh` is a no-op**: renewal is handled automatically by the `concurrency.Session` keepalive (`Refresh` only checks ctx); `StartRefresh` does not actively renew either — it only watches for session loss and invokes `onError(ErrNotObtained)` when the session dies
- **The Redis backend ignores `LockOption`**; blocking-wait and retry-delay options only apply to the etcd backend
- `IsLocked` works by trying to acquire and immediately releasing; use it for monitoring/debugging only
- Production recommendations:
  - Tune `TTL` and `RefreshInterval` (Redis backend)
  - Wire `onError` to alerting — a lost lock usually means an unhealthy instance or a network partition

## License

MIT
