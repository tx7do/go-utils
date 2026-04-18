package distlock

import "context"

// Lock 代表一把已持有的分布式锁。
type Lock interface {
	// Key 返回锁的键名。
	Key() string

	// Release 立即释放锁。
	// 若锁已过期或已被释放，返回底层错误。
	Release(ctx context.Context) error

	// Refresh 以原始 TTL 续期一次。
	Refresh(ctx context.Context) error

	// StartRefresh 启动后台续期协程，每隔配置的 RefreshInterval 调用一次 Refresh。
	//
	//   - onError 在每次续期失败时调用（可为 nil）；无论 onError 返回什么，失败后协程自动退出。
	//   - 返回的 stop 函数取消续期协程并阻塞等待其退出；可安全多次调用（幂等）。
	//
	// 典型用法：
	//
	//	stop := lock.StartRefresh(ctx, func(err error) { log.Warn(err) })
	//	defer stop()
	StartRefresh(ctx context.Context, onError func(err error)) (stop func())
}

// Locker 抽象分布式锁后端（Redis/Etcd 等），供上层业务统一依赖。
type Locker interface {
	// Obtain 尝试获取 key 对应的锁。
	//   - 成功：返回 Lock，调用方必须在完成后调用 Release。
	//   - 锁已被持有：返回 ErrNotObtained（可用 errors.Is 判断）。
	//   - 其他错误：返回底层错误。
	Obtain(ctx context.Context, key string) (Lock, error)

	// Close 关闭后端资源；不再使用时调用。
	Close() error
}
