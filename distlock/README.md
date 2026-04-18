# distlock

`go-utils/distlock` 提供统一的分布式锁抽象，屏蔽底层实现差异（Redis/Etcd），让业务层只依赖一套接口。

本包已从原有业务项目独立，现为通用分布式锁中间件，适用于多项目/多环境。

---

## 1. 设计目标

- 统一接口：业务只依赖 `distlock.Locker` / `distlock.Lock`
- 可切换后端：Redis 与 Etcd 可按环境切换
- 生命周期安全：提供 `StartRefresh` + 幂等 `stop()`，避免 refresh-after-release 竞态
- 一致错误语义：锁抢占失败统一映射为 `distlock.ErrNotObtained`
- 轻依赖：仅依赖 redis/etcd 官方库，无额外侵入
- 适合云原生/微服务/多实例场景

---

## 2. 安装

```shell
go get github.com/tx7do/go-utils/distlock
```

---

## 3. 核心接口

```go
type Locker interface {
    Obtain(ctx context.Context, key string) (Lock, error)
    Close() error
}

type Lock interface {
    Key() string
    Release(ctx context.Context) error
    Refresh(ctx context.Context) error
    StartRefresh(ctx context.Context, onError func(err error)) (stop func())
}

var ErrNotObtained = errors.New("distlock: lock not obtained")
```

---

## 4. Redis 后端

依赖 [github.com/bsm/redislock](https://github.com/bsm/redislock)

```go
locker := distlock.New(rdb, distlock.Options{
    TTL:             30 * time.Second,
    MaxRetries:      10,
    RetryDelay:      100 * time.Millisecond,
    RefreshInterval: 10 * time.Second, // 默认 TTL/3
})
```

---

## 5. Etcd 后端

依赖 [go.etcd.io/etcd/client/v3/concurrency](https://pkg.go.dev/go.etcd.io/etcd/client/v3/concurrency)

```go
locker, err := distlock.NewEtcd([]string{"127.0.0.1:2379"}, distlock.EtcdOptions{})
if err != nil {
    return err
}
defer locker.Close()
```

---

## 6. 推荐用法

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

---

## 7. 迁移说明

本包原为某业务项目的分布式锁实现，现已独立为 go-utils/distlock 公共库。
- 代码结构更清晰，接口更通用
- 推荐所有新项目直接依赖本库
- 老项目可平滑迁移，接口兼容

---

## 8. 注意事项

- `Release` 不是幂等保证接口：重复释放可能返回后端错误（按需忽略或记录）
- `StartRefresh` 返回的 `stop()` 是幂等的，可以安全多次调用
- 长任务务必持有并调用 `stop()`，避免 goroutine 泄漏
- 生产环境建议：
  - 合理设置 `TTL` 与 `RefreshInterval`
  - 给 `onError` 接告警（锁丢失通常意味着实例异常或网络分区）

---

## 9. License

MIT
