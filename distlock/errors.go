package distlock

import "errors"

// ErrNotObtained 当锁已被其他节点持有、无法获取时返回。
// 调用方用 errors.Is(err, distlock.ErrNotObtained) 判断。
var ErrNotObtained = errors.New("distlock: lock not obtained")
