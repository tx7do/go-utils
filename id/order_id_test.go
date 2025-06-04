package id

import (
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestGenerateOrderIdWithRandom(t *testing.T) {
	prefix := "PT"

	// 测试生成的订单号是否包含前缀
	orderID := GenerateOrderIdWithRandom(prefix, nil)
	assert.Contains(t, orderID, prefix, "订单号应包含前缀")
	t.Logf("GenerateOrderIdWithRandom: %s", orderID)

	// 测试生成的订单号长度是否正确
	assert.Equal(t, len(prefix)+14+4, len(orderID), "订单号长度应为前缀+时间戳+随机数")
}

func TestGenerateOrderIdWithIndex(t *testing.T) {
	prefix := "PT"

	tm := time.Now()

	fmt.Println(GenerateOrderIdWithIncreaseIndex(prefix, &(tm)))

	ids := make(map[string]bool)
	count := 100
	for i := 0; i < count; i++ {
		ids[GenerateOrderIdWithIncreaseIndex(prefix, &(tm))] = true
	}
	assert.Equal(t, count, len(ids))
}

func TestGenerateOrderIdWithIndexThread(t *testing.T) {
	tm := time.Now()

	var wg sync.WaitGroup
	var ids sync.Map
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			for i := 0; i < 100; i++ {
				id := GenerateOrderIdWithIncreaseIndex("PT", &(tm))
				ids.Store(id, true)
			}
			wg.Done()
		}()
	}
	wg.Wait()

	aLen := 0
	ids.Range(func(k, v interface{}) bool {
		aLen++
		return true
	})
	assert.Equal(t, 1000, aLen)
}

func TestGenerateOrderIdWithTenantId(t *testing.T) {
	tenantID := "M9876"
	orderID := GenerateOrderIdWithTenantId(tenantID)

	t.Logf(orderID)

	// 验证订单号长度是否正确
	assert.Equal(t, 14+5+4, len(orderID))

	// 验证时间戳部分是否正确
	timestamp := time.Now().Format("20060102150405")
	assert.Contains(t, orderID, timestamp)
	t.Logf("timestamp %d", len(timestamp))

	// 验证商户ID部分是否正确
	assert.Contains(t, orderID, tenantID)

	// 验证随机数部分是否为4位数字
	randomPart := orderID[len(orderID)-4:]
	assert.Regexp(t, `^\d{4}$`, randomPart)
}

func TestGenerateOrderIdWithTenantIdCollision(t *testing.T) {
	tenantID := "M9876"
	count := 1000 // 生成订单号的数量
	ids := make(map[string]bool)

	for i := 0; i < count; i++ {
		orderID := GenerateOrderIdWithTenantId(tenantID)
		if ids[orderID] {
			t.Errorf("碰撞的订单号: %s", orderID)
		}
		ids[orderID] = true
	}

	t.Logf("生成了 %d 个订单号，没有发生碰撞", count)
}

func TestGenerateOrderIdWithPrefixSonyflake(t *testing.T) {
	prefix := "ORD"
	orderID := GenerateOrderIdWithPrefixSonyflake(prefix)
	t.Logf("order id with prefix sonyflake: %s [%d]", orderID, len(orderID))

	// 验证订单号是否包含前缀
	assert.Contains(t, orderID, prefix, "订单号应包含前缀")

	// 验证订单号是否为有效的数字字符串
	assert.Regexp(t, `^ORD\d+$`, orderID, "订单号格式应为前缀加数字")
}

func TestGenerateOrderIdWithPrefixSonyflakeCollision(t *testing.T) {
	prefix := "ORD"
	count := 100000 // 生成订单号的数量
	ids := make(map[string]bool)

	for i := 0; i < count; i++ {
		orderID := GenerateOrderIdWithPrefixSonyflake(prefix)
		if ids[orderID] {
			t.Errorf("碰撞的订单号: %s", orderID)
		}
		ids[orderID] = true
	}

	t.Logf("生成了 %d 个订单号，没有发生碰撞", count)
}

func TestGenerateOrderIdWithPrefixSnowflake(t *testing.T) {
	workerId := int64(1) // 假设使用的 workerId
	prefix := "ORD"
	orderID := GenerateOrderIdWithPrefixSnowflake(workerId, prefix)
	t.Logf("order id with prefix snowflake: %s [%d]", orderID, len(orderID))

	// 验证订单号是否包含前缀
	assert.Contains(t, orderID, prefix, "订单号应包含前缀")

	// 验证订单号是否为有效的数字字符串
	assert.Regexp(t, `^ORD\d+$`, orderID, "订单号格式应为前缀加数字")
}

func TestGenerateOrderIdWithPrefixSnowflakeCollision(t *testing.T) {
	workerId := int64(1) // 假设使用的 workerId
	prefix := "ORD"
	count := 1000000 // 生成订单号的数量
	ids := make(map[string]bool)

	for i := 0; i < count; i++ {
		orderID := GenerateOrderIdWithPrefixSnowflake(workerId, prefix)
		if ids[orderID] {
			t.Errorf("碰撞的订单号: %s", orderID)
		}
		ids[orderID] = true
	}

	t.Logf("生成了 %d 个订单号，没有发生碰撞", count)
}
