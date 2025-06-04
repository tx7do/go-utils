package id

import (
	"fmt"
	"math/rand"
	"strings"
	"sync/atomic"
	"time"

	"github.com/tx7do/go-utils/trans"
)

type idCounter uint32

func (c *idCounter) Increase() uint32 {
	cur := *c
	atomic.AddUint32((*uint32)(c), 1)
	atomic.CompareAndSwapUint32((*uint32)(c), 1000, 0)
	return uint32(cur)
}

var orderIdIndex idCounter

// GenerateOrderIdWithRandom 生成20位订单号，前缀 + 时间戳 + 随机数
func GenerateOrderIdWithRandom(prefix string, tm *time.Time) string {
	// 前缀 + 时间戳（14位） + 随机数（4位）

	if tm == nil {
		tm = trans.Time(time.Now())
	}

	timestamp := tm.Format("20060102150405")

	randNum := rand.Intn(10000) // 生成0-9999之间的随机数

	return fmt.Sprintf("%s%s%d", prefix, timestamp, randNum)
}

// GenerateOrderIdWithIncreaseIndex 生成20位订单号，前缀+时间+自增长索引
func GenerateOrderIdWithIncreaseIndex(prefix string, tm *time.Time) string {
	if tm == nil {
		tm = trans.Time(time.Now())
	}

	timestamp := tm.Format("20060102150405")

	index := orderIdIndex.Increase()

	return fmt.Sprintf("%s%s%d", prefix, timestamp, index)
}

// GenerateOrderIdWithTenantId 带商户ID的订单ID生成器：202506041234567890123
func GenerateOrderIdWithTenantId(tenantID string) string {
	// 时间戳（14位） + 商户ID（固定 5 位） + 随机数（4位）

	// 时间戳部分（精确到毫秒）
	now := time.Now()
	timestamp := now.Format("20060102150405")

	// 商户ID部分（截取或补零到5位）
	tenantPart := tenantID
	if len(tenantPart) > 5 {
		tenantPart = tenantPart[:5]
	} else {
		tenantPart = fmt.Sprintf("%-5s", tenantPart)
		tenantPart = strings.ReplaceAll(tenantPart, " ", "0")
	}

	// 随机数部分（4位）
	n := rand.Int31n(10000)
	randomPart := fmt.Sprintf("%04d", n)

	return timestamp + tenantPart + randomPart
}

func GenerateOrderIdWithPrefixSonyflake(prefix string) string {
	id, _ := NewSonyflakeID()
	return fmt.Sprintf("%s%d", prefix, id)
}

func GenerateOrderIdWithPrefixSnowflake(workerId int64, prefix string) string {
	id, _ := NewSnowflakeID(workerId)
	return fmt.Sprintf("%s%d", prefix, id)
}
