package id

import (
	"sync"

	"github.com/sony/sonyflake"
)

var (
	sf   *sonyflake.Sonyflake
	sfMu sync.Mutex
)

func NewSonyflakeID() (uint64, error) {
	// 64 位 ID = 39 位时间戳 + 8 位机器 ID + 16 位序列号

	sfMu.Lock()
	defer sfMu.Unlock()

	if sf == nil {
		sf = sonyflake.NewSonyflake(sonyflake.Settings{})
	}

	return sf.NextID()
}
