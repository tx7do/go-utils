package order_id

import (
	"fmt"
	"math/rand"
	"sync/atomic"
	"time"

	"github.com/tx7do/kratos-utils/trans"
)

type idCounter uint32

func (c *idCounter) Increase() uint32 {
	cur := *c
	atomic.AddUint32((*uint32)(c), 1)
	atomic.CompareAndSwapUint32((*uint32)(c), 1000, 0)
	return uint32(cur)
}

var orderIdIndex idCounter

// GenerateOrderIdWithRandom 生成20位订单号，前缀+时间+随机数
func GenerateOrderIdWithRandom(prefix string, split string, tm *time.Time) string {
	if tm == nil {
		tm = trans.Time(time.Now())
	}

	index := rand.Intn(1000)

	return fmt.Sprintf("%s%s%.4d%s%.2d%s%.2d%s%.2d%s%.2d%s%.2d%s%.4d", prefix, split,
		tm.Year(), split, tm.Month(), split, tm.Day(), split,
		tm.Hour(), split, tm.Minute(), split, tm.Second(), split, index)
}

// GenerateOrderIdWithIncreaseIndex 生成20位订单号，前缀+时间+自增长索引
func GenerateOrderIdWithIncreaseIndex(prefix string, split string, tm *time.Time) string {
	if tm == nil {
		tm = trans.Time(time.Now())
	}

	index := orderIdIndex.Increase()

	return fmt.Sprintf("%s%s%.4d%s%.2d%s%.2d%s%.2d%s%.2d%s%.2d%s%.4d", prefix, split,
		tm.Year(), split, tm.Month(), split, tm.Day(), split,
		tm.Hour(), split, tm.Minute(), split, tm.Second(), split, index)
}
