package query_parser

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestParseFilterJSONString(t *testing.T) {
	var results []struct {
		Field    string
		Operator string
		Value    string
	}

	handler := func(field, operator, value string) {
		results = append(results, struct {
			Field    string
			Operator string
			Value    string
		}{Field: field, Operator: operator, Value: value})
	}

	// 测试解析单个过滤条件
	results = nil
	err := ParseFilterJSONString(`{"name__exact":"John"}`, handler)
	assert.NoError(t, err)
	assert.Equal(t, 1, len(results))
	assert.Equal(t, "name", results[0].Field)
	assert.Equal(t, "exact", results[0].Operator)
	assert.Equal(t, "John", results[0].Value)

	// 测试解析多个过滤条件
	results = nil
	err = ParseFilterJSONString(`[{"age__gte":"30"},{"status__exact":"active"}]`, handler)
	assert.NoError(t, err)
	assert.Equal(t, 2, len(results))
	assert.Equal(t, "age", results[0].Field)
	assert.Equal(t, "gte", results[0].Operator)
	assert.Equal(t, "30", results[0].Value)
	assert.Equal(t, "status", results[1].Field)
	assert.Equal(t, "exact", results[1].Operator)
	assert.Equal(t, "active", results[1].Value)

	// 测试空字符串
	results = nil
	err = ParseFilterJSONString("", handler)
	assert.NoError(t, err)
	assert.Equal(t, 0, len(results))

	// 测试无效的JSON字符串
	results = nil
	err = ParseFilterJSONString(`invalid_json`, handler)
	assert.Error(t, err)
	assert.Equal(t, 0, len(results))

	// 测试包含特殊字符的字段和值
	results = nil
	err = ParseFilterJSONString(`{"na!me__exact":"Jo@hn"}`, handler)
	assert.NoError(t, err)
	assert.Equal(t, 1, len(results))
	assert.Equal(t, "na!me", results[0].Field)
	assert.Equal(t, "exact", results[0].Operator)
	assert.Equal(t, "Jo@hn", results[0].Value)

	// 测试包含特殊字符的多个过滤条件
	results = nil
	err = ParseFilterJSONString(`[{"ag#e__gte":"30"},{"sta$tus__exact":"ac^tive"}]`, handler)
	assert.NoError(t, err)
	assert.Equal(t, 2, len(results))
	assert.Equal(t, "ag#e", results[0].Field)
	assert.Equal(t, "gte", results[0].Operator)
	assert.Equal(t, "30", results[0].Value)
	assert.Equal(t, "sta$tus", results[1].Field)
	assert.Equal(t, "exact", results[1].Operator)
	assert.Equal(t, "ac^tive", results[1].Value)

	// 测试包含特殊字符的无效 JSON 字符串
	results = nil
	err = ParseFilterJSONString(`{"na!me__exact":Jo@hn}`, handler)
	assert.Error(t, err)
	assert.Equal(t, 0, len(results))
}

func TestParseFilterField(t *testing.T) {
	var results []struct {
		Field    string
		Operator string
		Value    string
	}

	handler := func(field, operator, value string) {
		results = append(results, struct {
			Field    string
			Operator string
			Value    string
		}{Field: field, Operator: operator, Value: value})
	}

	// 测试正常解析
	results = nil
	ParseFilterField("name__exact", "John", handler)
	assert.Equal(t, 1, len(results))
	assert.Equal(t, "name", results[0].Field)
	assert.Equal(t, "exact", results[0].Operator)
	assert.Equal(t, "John", results[0].Value)

	// 测试无操作符解析
	results = nil
	ParseFilterField("name", "John", handler)
	assert.Equal(t, 1, len(results))
	assert.Equal(t, "name", results[0].Field)
	assert.Equal(t, "", results[0].Operator)
	assert.Equal(t, "John", results[0].Value)

	// 测试空字段
	results = nil
	ParseFilterField("", "John", handler)
	assert.Equal(t, 0, len(results))

	// 测试空值
	results = nil
	ParseFilterField("name__exact", "", handler)
	assert.Equal(t, 0, len(results))

	// 测试无效字段格式
	results = nil
	ParseFilterField("__exact", "John", handler)
	assert.Equal(t, 0, len(results))
}

func TestSplitFieldAndOperator(t *testing.T) {
	// 测试正常分割
	result := SplitFieldAndOperator("name__exact")
	assert.Equal(t, 2, len(result))
	assert.Equal(t, "name", result[0])
	assert.Equal(t, "exact", result[1])

	// 测试无操作符
	result = SplitFieldAndOperator("name")
	assert.Equal(t, 1, len(result))
	assert.Equal(t, "name", result[0])

	// 测试空字符串
	result = SplitFieldAndOperator("")
	assert.Equal(t, 1, len(result))
	assert.Equal(t, "", result[0])

	// 测试多个分隔符
	result = SplitFieldAndOperator("name__exact__extra")
	assert.Equal(t, 3, len(result))
	assert.Equal(t, "name", result[0])
	assert.Equal(t, "exact", result[1])
	assert.Equal(t, "extra", result[2])
}

func TestSplitJSONField(t *testing.T) {
	// 测试正常分割
	result := SplitJSONField("user.address.city")
	assert.Equal(t, 3, len(result))
	assert.Equal(t, "user", result[0])
	assert.Equal(t, "address", result[1])
	assert.Equal(t, "city", result[2])

	// 测试单个字段
	result = SplitJSONField("user")
	assert.Equal(t, 1, len(result))
	assert.Equal(t, "user", result[0])

	// 测试空字符串
	result = SplitJSONField("")
	assert.Equal(t, 1, len(result))
	assert.Equal(t, "", result[0])

	// 测试多个分隔符
	result = SplitJSONField("user..address.city")
	assert.Equal(t, 4, len(result))
	assert.Equal(t, "user", result[0])
	assert.Equal(t, "", result[1])
	assert.Equal(t, "address", result[2])
	assert.Equal(t, "city", result[3])
}
