# distlock

[English](README.md) | 简体中文

`distlock` 是一个独立 Go 模块（`github.com/tx7do/go-utils/distlock`），提供统一的分布式锁抽象，屏蔽 Redis 与 etcd 后端差异，业务层只依赖一套接口。

## 功能特性

- 统一接口：业务只依赖 `distlock.Locker` / `distlock.Lock`
- 可切换后端：Redis（基于 [github.com/bsm/redislock](https://github.com/bsm/redislock)）与 etcd（基于 [go.etcd.io/etcd/client/v3/concurrency](https://pkg.go.dev/go.etcd.io/etcd/client/v3/concurrency)）按环境切换
- 生命周期安全：`StartRefresh` 返回幂等的 `stop()`，避免 refresh-after-release 竞态
- 一致错误语义：抢锁失败统一映射为 `distlock.ErrNotObtained`（用 `errors.Is` 判断）
- 监控辅助：`Locker.IsLocked` 检查 key 是否已被锁定
- etcd 后端支持阻塞等待模式（指数退避，上限 2s）

## 安装

本包是独立模块（有自己的 `go.mod`），不是根模块 `github.com/tx7do/go-utils` 的子包：

```bash
go get github.com/tx7do/go-utils/distlock
```

```go
import "github.com/tx7do/go-utils/distlock"
```

## 核心接口

```go
type Locker interface {
	// Obtain 尝试获取 key 对应的锁。
	//   - 成功：返回 Lock，调用方完成后必须 Release。
	//   - 锁已被持有：返回 ErrNotObtained（可用 errors.Is 判断）。
	//   - 其他错误：返回底层错误。
	Obtain(ctx context.Context, key string, opts ...LockOption) (Lock, error)

	// Close 关闭后端资源；不再使用时调用。
	Close() error

	// IsLocked 检查 key 是否已被锁定；仅供监控/调试，不能替代 Obtain 的原子性保证。
	IsLocked(ctx context.Context, key string) (bool, error)
}

type Lock interface {
	Key() string
	Release(ctx context.Context) error                       // 锁已过期/已释放时返回底层错误
	Refresh(ctx context.Context) error                       // 以原始 TTL 续期一次
	StartRefresh(ctx context.Context, onError func(err error)) (stop func())
}
```

`StartRefresh` 行为：后台协程每隔 `RefreshInterval` 调用一次 `Refresh`；续期失败（含 context 取消）时先回调 `onError`（可为 nil），协程自动退出；返回的 `stop()` 取消协程并阻塞等待其退出，可安全多次调用（幂等）。

## 快速开始

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
			log.Println("其他节点已持有锁")
			return
		}
		log.Fatal(err)
	}

	stop := lock.StartRefresh(ctx, func(err error) { log.Printf("续期失败: %v", err) })

	// ... 业务逻辑 ...

	stop()                                        // 先停续期
	_ = lock.Release(context.Background())        // 再释放锁
}
```

## Redis 后端

```go
rdb := redis.NewClient(&redis.Options{Addr: "localhost:6379"})

locker := distlock.NewRedisLocker(rdb, distlock.Options{
	TTL:             30 * time.Second,
	MaxRetries:      10,
	RetryDelay:      100 * time.Millisecond,
	RefreshInterval: 10 * time.Second, // 默认 TTL/3
})
defer locker.Close() // Redis 后端的 Close 是 no-op
```

`Options` 字段（零值/负值自动补默认值）：

| 字段 | 说明 | 默认值 |
|------|------|--------|
| TTL | 锁的有效期 | 30s |
| MaxRetries | 争抢锁时的最大重试次数（线性退避） | 10 |
| RetryDelay | 线性退避步长 | 100ms |
| RefreshInterval | `StartRefresh` 的续期间隔 | TTL/3 |

获取锁的重试策略为 `redislock.LimitRetry(LinearBackoff(RetryDelay), MaxRetries)`。

## etcd 后端

```go
// 方式一：由本包创建 etcd client（Close 时会一并关闭）
locker, err := distlock.NewEtcd([]string{"127.0.0.1:2379"}, distlock.EtcdOptions{})
if err != nil {
	return err
}
defer locker.Close()

// 方式二：复用已有 etcd client（Close 不会关闭该 client）
locker = distlock.NewEtcdWithClient(cli, distlock.EtcdOptions{})
```

`EtcdOptions` 字段（零值/负值自动补默认值）：

| 字段 | 说明 | 默认值 | 当前是否生效 |
|------|------|--------|--------------|
| DialTimeout | etcd 连接超时 | 5s | 是 |
| SessionTTL | session 租约 TTL（秒） | 5 | 是 |
| SessionTimeout | 创建会话超时 | 5s | **预留字段，当前未被读取**（创建 session 未应用该超时） |
| LockTimeout | 单次加锁尝试的超时；0 表示不自动释放 | 0 | 是（仅阻塞等待模式） |

### LockOption（仅 etcd 后端生效）

`Obtain` 支持按次传入学选项；Redis 后端当前不读取这些选项：

```go
lock, err := locker.Obtain(ctx, key,
	distlock.WithBlockWait(5*time.Second), // 阻塞等待模式：最长等待 5s（0 表示无限等待）
	distlock.WithRetryDelay(200*time.Millisecond), // 重试基础间隔（仅阻塞模式生效）
)
```

- 默认非阻塞：锁被占用立即返回 `ErrNotObtained`
- `WithBlockWait(maxWait)`：阻塞等待模式，内部每次尝试新建 session 并以 `mu.Lock` 抢锁，指数退避重试（步长为 `WithRetryDelay` 的值，默认 100ms，上限 2s）；整体超时返回包装了 `ErrNotObtained` 的超时错误

## 推荐用法

```go
lock, err := locker.Obtain(ctx, key)
if err != nil {
	if errors.Is(err, distlock.ErrNotObtained) {
		return nil // 其他节点已持有
	}
	return err
}

stopRefresh := lock.StartRefresh(ctx, func(err error) {
	// 记录日志/告警
})

// ... 业务逻辑 ...

// 顺序建议：先停续期，再释放锁
stopRefresh()
_ = lock.Release(context.Background())
```

## 注意事项

- `Release` 不保证幂等：重复释放可能返回后端错误（按需忽略或记录）
- `StartRefresh` 返回的 `stop()` 是幂等的，可安全多次调用
- 长任务务必持有并调用 `stop()`，避免 goroutine 泄漏
- **etcd 后端的 `Refresh` 是 no-op**：续租由 `concurrency.Session` 的 keepalive 自动完成（`etcd_locker.go` 中 `Refresh` 仅检查 ctx）；`StartRefresh` 也不主动续租，只监听 session 失效事件并在失效时回调 `onError(ErrNotObtained)`
- **Redis 后端不读取 `LockOption`**；阻塞等待与重试间隔选项仅对 etcd 后端生效
- `IsLocked` 通过"尝试加锁再立即释放"实现，仅供监控/调试
- 生产环境建议：
  - 合理设置 `TTL` 与 `RefreshInterval`（Redis 后端）
  - 给 `onError` 接告警——锁丢失通常意味着实例异常或网络分区

## 许可证

MIT
