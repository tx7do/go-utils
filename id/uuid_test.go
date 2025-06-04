package id

import (
	"strings"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewGUIDv4(t *testing.T) {
	// 测试带有连字符的 GUID
	withHyphen := NewGUIDv4(true)
	assert.NotEmpty(t, withHyphen)
	assert.Equal(t, 4, strings.Count(withHyphen, "-"), "GUID 应包含 4 个连字符")

	// 测试不带连字符的 GUID
	withoutHyphen := NewGUIDv4(false)
	assert.NotEmpty(t, withoutHyphen)
	assert.Equal(t, 0, strings.Count(withoutHyphen, "-"), "GUID 不应包含连字符")

	// 验证 GUID 的长度
	assert.Equal(t, 36, len(withHyphen), "带连字符的 GUID 长度应为 36")
	assert.Equal(t, 32, len(withoutHyphen), "不带连字符的 GUID 长度应为 32")
}

func TestNewGUIDv4CollisionRate(t *testing.T) {
	const (
		testCount  = 100000 // 测试生成的GUID数量
		withHyphen = true   // 是否带连字符
	)

	ids := make(map[string]struct{})
	for i := 0; i < testCount; i++ {
		id := NewGUIDv4(withHyphen)
		if _, exists := ids[id]; exists {
			t.Errorf("碰撞发生: %s 已存在", id)
		}
		ids[id] = struct{}{}
	}

	t.Logf("生成了 %d 个GUID，无碰撞。", testCount)
}

func TestNewShortUUID(t *testing.T) {
	// 测试生成的 ShortUUID 是否非空
	id := NewShortUUID()
	assert.NotEmpty(t, id, "生成的 ShortUUID 应该非空")

	// 测试生成的 ShortUUID 的长度是否符合预期
	assert.True(t, len(id) > 0, "生成的 ShortUUID 长度应该大于 0")
}

func TestNewShortUUIDCollisionRate(t *testing.T) {
	const testCount = 100000 // 测试生成的ShortUUID数量

	ids := make(map[string]struct{})
	for i := 0; i < testCount; i++ {
		id := NewShortUUID()
		if _, exists := ids[id]; exists {
			t.Errorf("碰撞发生: %s 已存在", id)
		}
		ids[id] = struct{}{}
	}

	t.Logf("生成了 %d 个ShortUUID，无碰撞。", testCount)
}

func TestNewKSUID(t *testing.T) {
	// 测试生成的 KSUID 是否非空
	id := NewKSUID()
	assert.NotEmpty(t, id, "生成的 KSUID 应该非空")

	// 测试生成的 KSUID 的长度是否符合预期
	assert.Equal(t, 27, len(id), "生成的 KSUID 长度应该为 27")
}

func TestNewKSUIDCollisionRate(t *testing.T) {
	const testCount = 100000 // 测试生成的KSUID数量

	ids := make(map[string]struct{})
	for i := 0; i < testCount; i++ {
		id := NewKSUID()
		if _, exists := ids[id]; exists {
			t.Errorf("碰撞发生: %s 已存在", id)
		}
		ids[id] = struct{}{}
	}

	t.Logf("生成了 %d 个KSUID，无碰撞。", testCount)
}

func TestNewXID(t *testing.T) {
	// 测试生成的 XID 是否非空
	id := NewXID()
	assert.NotEmpty(t, id, "生成的 XID 应该非空")

	// 测试生成的 XID 的长度是否符合预期
	assert.Equal(t, 20, len(id), "生成的 XID 长度应该为 20")
}

func TestNewXIDCollisionRate(t *testing.T) {
	const testCount = 100000 // 测试生成的XID数量

	ids := make(map[string]struct{})
	for i := 0; i < testCount; i++ {
		id := NewXID()
		if _, exists := ids[id]; exists {
			t.Errorf("碰撞发生: %s 已存在", id)
		}
		ids[id] = struct{}{}
	}

	t.Logf("生成了 %d 个XID，无碰撞。", testCount)
}

func TestNewSnowflakeID(t *testing.T) {
	tests := []struct {
		workerId  int64
		expectErr bool
	}{
		{0, false},  // 有效的 workerId
		{31, false}, // 有效的 workerId
		{32, false}, // 有效的 workerId

		{-1, true}, // 无效的 workerId
	}

	for _, tt := range tests {
		id, err := NewSnowflakeID(tt.workerId)
		if (err != nil) != tt.expectErr {
			t.Errorf("NewSnowflakeID(%d) 错误状态不符合预期: %v", tt.workerId, err)
		}
		if err == nil && id <= 0 {
			t.Errorf("NewSnowflakeID(%d) 生成的ID无效: %d", tt.workerId, id)
		}
		t.Logf("NewSnowflakeID(%d) ID: %d", tt.workerId, id)
	}
}

func TestNewSnowflakeIDCollisionRate(t *testing.T) {
	const (
		workerId  = 0
		testCount = 100000 // 测试生成的ID数量
	)

	ids := make(map[int64]struct{})
	for i := 0; i < testCount; i++ {
		id, err := NewSnowflakeID(workerId)
		if err != nil {
			t.Errorf("生成ID时出现错误: %v", err)
			continue
		}
		if _, exists := ids[id]; exists {
			t.Errorf("碰撞发生: %d 已存在", id)
		}
		ids[id] = struct{}{}
	}

	t.Logf("生成了 %d 个Snowflake ID，无碰撞。", testCount)
}

func TestConcurrentNewSnowflakeIDCollisionRate(t *testing.T) {
	const (
		workerId    = 0      // Snowflake 工作节点 ID
		testCount   = 100000 // 测试生成的 ID 数量
		workerCount = 10     // 并发工作线程数
	)

	var mu sync.Mutex
	ids := make(map[int64]struct{})
	var wg sync.WaitGroup

	for w := 0; w < workerCount; w++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for i := 0; i < testCount/workerCount; i++ {
				id, err := NewSnowflakeID(workerId)
				if err != nil {
					t.Errorf("生成 Snowflake ID 时出现错误: %v", err)
					continue
				}
				mu.Lock()
				if _, exists := ids[id]; exists {
					t.Errorf("碰撞发生: %d 已存在", id)
				}
				ids[id] = struct{}{}
				mu.Unlock()
			}
		}()
	}

	wg.Wait()
	t.Logf("生成了 %d 个 Snowflake ID，无碰撞。", testCount)
}

func TestNewSonyflakeID(t *testing.T) {
	// 测试生成的 Sonyflake ID 是否有效
	id, err := NewSonyflakeID()
	t.Logf("sonyflake id: %v", id)
	assert.NoError(t, err, "生成 Sonyflake ID 时不应出现错误")
	assert.True(t, id > 0, "生成的 Sonyflake ID 应该是正数")
}

func TestNewSonyflakeIDCollisionRate(t *testing.T) {
	const testCount = 100000 // 测试生成的 Sonyflake ID 数量

	ids := make(map[uint64]struct{})
	for i := 0; i < testCount; i++ {
		id, err := NewSonyflakeID()
		if err != nil {
			t.Errorf("生成 Sonyflake ID 时出现错误: %v", err)
			continue
		}
		if _, exists := ids[id]; exists {
			t.Errorf("碰撞发生: %d 已存在", id)
		}
		ids[id] = struct{}{}
	}

	t.Logf("生成了 %d 个 Sonyflake ID，无碰撞。", testCount)
}

func TestConcurrentNewSonyflakeIDCollisionRate(t *testing.T) {
	const (
		testCount   = 100000 // 测试生成的 Sonyflake ID 数量
		workerCount = 10     // 并发工作线程数
	)

	var mu sync.Mutex
	ids := make(map[uint64]struct{})
	var wg sync.WaitGroup

	for w := 0; w < workerCount; w++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for i := 0; i < testCount/workerCount; i++ {
				id, err := NewSonyflakeID()
				if err != nil {
					t.Errorf("生成 Sonyflake ID 时出现错误: %v", err)
					continue
				}
				mu.Lock()
				if _, exists := ids[id]; exists {
					t.Errorf("碰撞发生: %d 已存在", id)
				}
				ids[id] = struct{}{}
				mu.Unlock()
			}
		}()
	}

	wg.Wait()
	t.Logf("生成了 %d 个 Sonyflake ID，无碰撞。", testCount)
}

func TestNewMongoObjectID(t *testing.T) {
	// 测试生成的 ObjectID 是否非空
	id := NewMongoObjectID()
	assert.NotEmpty(t, id, "生成的 Mongo ObjectID 应该非空")

	// 测试生成的 ObjectID 的长度是否符合预期
	assert.Equal(t, 36, len(id), "生成的 Mongo ObjectID 长度应该为 36")
}
