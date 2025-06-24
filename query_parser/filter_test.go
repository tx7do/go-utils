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
	assert.Equal(t, "na_me", results[0].Field)
	assert.Equal(t, "exact", results[0].Operator)
	assert.Equal(t, "Jo@hn", results[0].Value)

	// 测试包含特殊字符的多个过滤条件
	results = nil
	err = ParseFilterJSONString(`[{"ag#e__gte":"30"},{"sta$tus__exact":"ac^tive"}]`, handler)
	assert.NoError(t, err)
	assert.Equal(t, 2, len(results))
	assert.Equal(t, "ag_e", results[0].Field)
	assert.Equal(t, "gte", results[0].Operator)
	assert.Equal(t, "30", results[0].Value)
	assert.Equal(t, "sta_tus", results[1].Field)
	assert.Equal(t, "exact", results[1].Operator)
	assert.Equal(t, "ac^tive", results[1].Value)

	// 测试包含特殊字符的无效 JSON 字符串
	results = nil
	err = ParseFilterJSONString(`{"na!me__exact":Jo@hn}`, handler)
	assert.Error(t, err)
	assert.Equal(t, 0, len(results))
}

func TestParseFilterQueryString(t *testing.T) {
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
	err := ParseFilterQueryString("name__exact:John", handler)
	assert.NoError(t, err)
	assert.Equal(t, 1, len(results))
	assert.Equal(t, "name", results[0].Field)
	assert.Equal(t, "exact", results[0].Operator)
	assert.Equal(t, "John", results[0].Value)

	// 测试解析多个过滤条件
	results = nil
	err = ParseFilterQueryString("age__gte:30,status__exact:active", handler)
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
	err = ParseFilterQueryString("", handler)
	assert.NoError(t, err)
	assert.Equal(t, 0, len(results))

	// 测试无效的查询字符串
	results = nil
	err = ParseFilterQueryString("invalid_query", handler)
	assert.NoError(t, err)
	assert.Equal(t, 0, len(results))

	// 测试包含特殊字符的字段和值
	results = nil
	err = ParseFilterQueryString("na!me__exact:Jo@hn", handler)
	assert.NoError(t, err)
	assert.Equal(t, 1, len(results))
	assert.Equal(t, "na_me", results[0].Field)
	assert.Equal(t, "exact", results[0].Operator)
	assert.Equal(t, "Jo@hn", results[0].Value)

	// 测试包含特殊字符的多个过滤条件
	results = nil
	err = ParseFilterQueryString("ag#e__gte:30,sta$tus__exact:ac^tive", handler)
	assert.NoError(t, err)
	assert.Equal(t, 2, len(results))
	assert.Equal(t, "ag_e", results[0].Field)
	assert.Equal(t, "gte", results[0].Operator)
	assert.Equal(t, "30", results[0].Value)
	assert.Equal(t, "sta_tus", results[1].Field)
	assert.Equal(t, "exact", results[1].Operator)
	assert.Equal(t, "ac^tive", results[1].Value)

	// 测试 Field 中包含分隔符
	results = nil
	err = ParseFilterQueryString("na:me__exact:John", handler)
	assert.NoError(t, err)
	assert.Equal(t, 0, len(results))

	// 测试 Operator 中包含分隔符
	results = nil
	err = ParseFilterQueryString("name__ex:act:John", handler)
	assert.NoError(t, err)
	assert.Equal(t, 0, len(results))

	// 测试 Value 中包含分隔符
	results = nil
	err = ParseFilterQueryString("name__exact:Jo|hn", handler)
	assert.NoError(t, err)
	assert.Equal(t, 1, len(results))
	assert.Equal(t, "name", results[0].Field)
	assert.Equal(t, "exact", results[0].Operator)
	assert.Equal(t, "Jo|hn", results[0].Value)

	// 测试多个过滤条件中包含分隔符
	results = nil
	err = ParseFilterQueryString("ag:e__gte:30,sta|tus__exact:ac|tive", handler)
	assert.NoError(t, err)
	assert.Equal(t, 1, len(results))

	// 测试 Field 中包含编码后的分隔符
	results = nil
	encodedField := EncodeSpecialCharacters("na:me")
	err = ParseFilterQueryString(encodedField+"__exact:John", handler)
	assert.NoError(t, err)
	assert.Equal(t, 1, len(results))
	assert.Equal(t, "na_me", results[0].Field) // 注意：这里的字段名会被转换为 snake_case
	assert.Equal(t, "exact", results[0].Operator)
	assert.Equal(t, "John", results[0].Value)

	// 测试 Operator 中包含编码后的分隔符
	results = nil
	encodedOperator := EncodeSpecialCharacters("ex:act")
	err = ParseFilterQueryString("name__"+encodedOperator+":John", handler)
	assert.NoError(t, err)
	assert.Equal(t, 1, len(results))
	assert.Equal(t, "name", results[0].Field)
	assert.Equal(t, "ex:act", results[0].Operator)
	assert.Equal(t, "John", results[0].Value)

	// 测试 Value 中包含编码后的分隔符
	results = nil
	encodedValue := EncodeSpecialCharacters("Jo|hn")
	err = ParseFilterQueryString("name__exact:"+encodedValue, handler)
	assert.NoError(t, err)
	assert.Equal(t, 1, len(results))
	assert.Equal(t, "name", results[0].Field)
	assert.Equal(t, "exact", results[0].Operator)
	assert.Equal(t, "Jo|hn", results[0].Value)

	// 测试多个过滤条件中包含编码后的分隔符
	results = nil
	encodedField1 := EncodeSpecialCharacters("ag:e")
	encodedValue1 := EncodeSpecialCharacters("30")
	encodedField2 := EncodeSpecialCharacters("sta|tus")
	encodedValue2 := EncodeSpecialCharacters("ac|tive")
	err = ParseFilterQueryString(encodedField1+"__gte:"+encodedValue1+","+encodedField2+"__exact:"+encodedValue2, handler)
	assert.NoError(t, err)
	assert.Equal(t, 2, len(results))
	assert.Equal(t, "ag_e", results[0].Field) // 注意：这里的字段名会被转换为 snake_case
	assert.Equal(t, "gte", results[0].Operator)
	assert.Equal(t, "30", results[0].Value)
	assert.Equal(t, "sta_tus", results[1].Field) // 注意：这里的字段名会被转换为 snake_case
	assert.Equal(t, "exact", results[1].Operator)
	assert.Equal(t, "ac|tive", results[1].Value)
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

func TestSplitJsonFieldAndOperator(t *testing.T) {
	// 测试正常分割
	result := SplitJsonFieldAndOperator("name__exact")
	assert.Equal(t, 2, len(result))
	assert.Equal(t, "name", result[0])
	assert.Equal(t, "exact", result[1])

	// 测试无操作符
	result = SplitJsonFieldAndOperator("name")
	assert.Equal(t, 1, len(result))
	assert.Equal(t, "name", result[0])

	// 测试空字符串
	result = SplitJsonFieldAndOperator("")
	assert.Equal(t, 1, len(result))
	assert.Equal(t, "", result[0])

	// 测试多个分隔符
	result = SplitJsonFieldAndOperator("name__exact__extra")
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

func TestSplitQueryFieldAndOperator(t *testing.T) {
	// 测试正常分割
	result := SplitQueryFieldAndOperator("name:exact")
	assert.Equal(t, 2, len(result))
	assert.Equal(t, "name", result[0])
	assert.Equal(t, "exact", result[1])

	// 测试无操作符
	result = SplitQueryFieldAndOperator("name")
	assert.Equal(t, 1, len(result))
	assert.Equal(t, "name", result[0])

	// 测试空字符串
	result = SplitQueryFieldAndOperator("")
	assert.Equal(t, 1, len(result))
	assert.Equal(t, "", result[0])

	// 测试多个分隔符
	result = SplitQueryFieldAndOperator("name:exact:extra")
	assert.Equal(t, 3, len(result))
	assert.Equal(t, "name", result[0])
	assert.Equal(t, "exact", result[1])
	assert.Equal(t, "extra", result[2])
}

func TestSplitQueryQueries(t *testing.T) {
	// 测试正常分割多个键值对
	result := SplitQueryQueries("name:John,age:30,status:active")
	assert.Equal(t, 3, len(result))
	assert.Equal(t, "name:John", result[0])
	assert.Equal(t, "age:30", result[1])
	assert.Equal(t, "status:active", result[2])

	// 测试单个键值对
	result = SplitQueryQueries("name:John")
	assert.Equal(t, 1, len(result))
	assert.Equal(t, "name:John", result[0])

	// 测试空字符串
	result = SplitQueryQueries("")
	assert.Equal(t, 1, len(result))
	assert.Equal(t, "", result[0])

	// 测试多个分隔符
	result = SplitQueryQueries("name:John,,age:30")
	assert.Equal(t, 3, len(result))
	assert.Equal(t, "name:John", result[0])
	assert.Equal(t, "", result[1])
	assert.Equal(t, "age:30", result[2])
}

func TestSplitQueryValues(t *testing.T) {
	// 测试正常分割多个值
	result := SplitQueryValues("value1|value2|value3")
	assert.Equal(t, 3, len(result))
	assert.Equal(t, "value1", result[0])
	assert.Equal(t, "value2", result[1])
	assert.Equal(t, "value3", result[2])

	// 测试单个值
	result = SplitQueryValues("value1")
	assert.Equal(t, 1, len(result))
	assert.Equal(t, "value1", result[0])

	// 测试空字符串
	result = SplitQueryValues("")
	assert.Equal(t, 1, len(result))
	assert.Equal(t, "", result[0])

	// 测试多个分隔符
	result = SplitQueryValues("value1||value2")
	assert.Equal(t, 3, len(result))
	assert.Equal(t, "value1", result[0])
	assert.Equal(t, "", result[1])
	assert.Equal(t, "value2", result[2])
}
