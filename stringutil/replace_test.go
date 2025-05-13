package stringutil

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestReplaceJSONField(t *testing.T) {
	// 测试替换单个字段
	jsonStr := `{"tenantId": "123", "name": "test"}`
	result := ReplaceJSONField("tenantId", "456", jsonStr)
	expected := `{"tenantId": "456", "name": "test"}`
	assert.Equal(t, expected, result)

	// 测试替换多个字段
	jsonStr = `{"tenantId": "123", "tenant_id": "789", "name": "test"}`
	result = ReplaceJSONField("tenantId|tenant_id", "456", jsonStr)
	expected = `{"tenantId": "456", "tenant_id": "456", "name": "test"}`
	assert.Equal(t, expected, result)

	// 测试字段不存在
	jsonStr = `{"name": "test"}`
	result = ReplaceJSONField("tenantId", "456", jsonStr)
	expected = `{"name": "test"}`
	assert.Equal(t, expected, result)

	// 测试空 JSON 字符串
	jsonStr = ``
	result = ReplaceJSONField("tenantId", "456", jsonStr)
	expected = ``
	assert.Equal(t, expected, result)

	// 测试空字段名
	jsonStr = `{"tenantId": "123"}`
	result = ReplaceJSONField("", "456", jsonStr)
	expected = `{"tenantId": "123"}`
	assert.Equal(t, expected, result)
}
