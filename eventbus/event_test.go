// Package eventbus_test 针对 pkg/eventbus 做黑盒测试。
//
// 本文件覆盖 Event 值对象：NewEvent 的默认字段、链式构造器、
// GetData 的 JSON 往返（含 nil 数据、不可序列化数据、类型不匹配三类路径）
// 以及 Clone 的深拷贝语义。
package eventbus_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/tx7do/go-utils/eventbus"
)

// TestNewEvent_DefaultFields NewEvent 必须填充事件元信息：
// 非空 ID、指定类型与数据、当前时间戳、零优先级、非 nil 空 Metadata。
func TestNewEvent_DefaultFields(t *testing.T) {
	ev := eventbus.NewEvent("user.created", map[string]any{"k": "v"})

	assert.NotEmpty(t, ev.ID, "事件 ID 必须自动生成")
	assert.Equal(t, "user.created", ev.Type)
	assert.Equal(t, map[string]any{"k": "v"}, ev.Data)
	assert.False(t, ev.Timestamp.IsZero(), "时间戳必须自动填充")
	assert.Equal(t, 0, ev.Priority)
	require.NotNil(t, ev.Metadata, "Metadata 必须初始化为空 map")
	assert.Empty(t, ev.Metadata)
}

// TestEvent_BuilderChaining 链式构造器必须返回自身并落盘对应字段；
// WithMetadata 在 Metadata 为 nil 时自动建表。
func TestEvent_BuilderChaining(t *testing.T) {
	ev := &eventbus.Event{}
	assert.Same(t, ev, ev.WithSource("test-source"))
	assert.Same(t, ev, ev.WithPriority(7))
	assert.Same(t, ev, ev.WithMetadata("trace", "abc"))

	assert.Equal(t, "test-source", ev.Source)
	assert.Equal(t, 7, ev.Priority)
	assert.Equal(t, map[string]string{"trace": "abc"}, ev.Metadata)
}

// testEmailPayload 是测试专用的类型化载荷（替代库内业务事件类型）。
type testEmailPayload struct {
	EmailID   string `json:"emailId"`
	From      string `json:"from"`
	To        string `json:"to"`
	Subject   string `json:"subject"`
	Mailbox   string `json:"mailbox"`
	AccountID string `json:"accountId"`
}

// TestEvent_GetData_RoundTrip 类型化载荷经 JSON 往返后无损还原。
func TestEvent_GetData_RoundTrip(t *testing.T) {
	payload := testEmailPayload{
		EmailID:   "email-1",
		From:      "a@example.com",
		To:        "b@example.com",
		Subject:   "hi",
		Mailbox:   "INBOX",
		AccountID: "acct-1",
	}
	ev := eventbus.NewEvent("email.received", payload)

	var got testEmailPayload
	require.NoError(t, ev.GetData(&got))
	assert.Equal(t, payload, got)
}

// TestEvent_GetData_NilPayloadReturnsNil nil 数据直接返回 nil，不进入 JSON 往返。
func TestEvent_GetData_NilPayloadReturnsNil(t *testing.T) {
	ev := eventbus.NewEvent("x", nil)
	var out struct{}
	assert.NoError(t, ev.GetData(&out))
}

// TestEvent_GetData_ErrorPaths 不可序列化载荷（marshal 失败）与
// 类型不匹配目标（unmarshal 失败）都必须显式报错。
func TestEvent_GetData_ErrorPaths(t *testing.T) {
	t.Run("marshal failure on unserializable payload", func(t *testing.T) {
		ev := eventbus.NewEvent("x", make(chan int))
		var out map[string]any
		err := ev.GetData(&out)
		assert.Error(t, err, "chan 载荷不可 JSON 序列化，必须报错")
	})

	t.Run("unmarshal failure on type mismatch", func(t *testing.T) {
		ev := eventbus.NewEvent("x", "definitely-not-a-number")
		var out int
		err := ev.GetData(&out)
		assert.Error(t, err, "字符串载荷不可解入 int，必须报错")
	})
}

// TestEvent_Clone_DeepCopyMetadata Clone 必须深拷贝 Metadata：
// 修改克隆的元数据不影响原件，其余字段值逐一相等。
func TestEvent_Clone_DeepCopyMetadata(t *testing.T) {
	origin := eventbus.NewEvent("clone.me", "payload").
		WithSource("src").
		WithPriority(3).
		WithMetadata("k1", "v1")

	clone := origin.Clone()
	require.NotNil(t, clone)

	// 元数据独立：改克隆不动原件。
	clone.Metadata["k2"] = "v2"
	delete(clone.Metadata, "k1")
	assert.Equal(t, map[string]string{"k1": "v1"}, origin.Metadata, "原件元数据不得被克隆篡改")

	// 其余字段逐一相等（含时间戳与优先级）。
	assert.Equal(t, origin.ID, clone.ID)
	assert.Equal(t, origin.Type, clone.Type)
	assert.Equal(t, origin.Source, clone.Source)
	assert.Equal(t, origin.Data, clone.Data)
	assert.Equal(t, origin.Priority, clone.Priority)
	assert.True(t, origin.Timestamp.Equal(clone.Timestamp))
}

// TestEvent_Clone_NilMetadataBecomesEmptyMap 原件 Metadata 为 nil 时，
// 克隆应得到独立的空 map 而非共享 nil。
func TestEvent_Clone_NilMetadataBecomesEmptyMap(t *testing.T) {
	origin := &eventbus.Event{ID: "id", Type: "t", Timestamp: time.Now()}
	clone := origin.Clone()
	require.NotNil(t, clone.Metadata)
	assert.Empty(t, clone.Metadata)
	assert.NotSame(t, &origin.Metadata, &clone.Metadata)
}
