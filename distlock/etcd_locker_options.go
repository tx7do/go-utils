package distlock

import "time"

const (
	defaultEtcdDialTimeout    = 5 * time.Second
	defaultEtcdSessionTTL     = 5
	defaultEtcdSessionTimeout = 5 * time.Second
	defaultEtcdLockTimeout    = 0
)

// EtcdOptions 配置 etcd 分布式锁行为。
type EtcdOptions struct {
	// DialTimeout 是 etcd 连接超时；默认 5s。
	DialTimeout time.Duration
	// SessionTTL 是 session 续约租约秒数；默认 5s。
	SessionTTL int
	// SessionTimeout 是创建会话超时；默认 5s。
	SessionTimeout time.Duration
	// 锁超时，超过这个时间未释放锁会被自动释放；默认 0（不自动释放）
	LockTimeout time.Duration
}

func (o EtcdOptions) withDefaults() EtcdOptions {
	if o.DialTimeout <= 0 {
		o.DialTimeout = defaultEtcdDialTimeout
	}
	if o.SessionTTL <= 0 {
		o.SessionTTL = defaultEtcdSessionTTL
	}
	if o.SessionTimeout <= 0 {
		o.SessionTimeout = defaultEtcdSessionTimeout
	}
	if o.LockTimeout <= 0 {
		o.LockTimeout = defaultEtcdLockTimeout
	}

	return o
}
