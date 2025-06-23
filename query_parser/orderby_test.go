package query_parser

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParseOrderByString(t *testing.T) {
	var results []struct {
		Field string
		Desc  bool
	}

	handler := func(field string, desc bool) {
		results = append(results, struct {
			Field string
			Desc  bool
		}{Field: field, Desc: desc})
	}

	// 测试正常解析
	err := ParseOrderByString("name,-age,+created_at", handler)
	assert.NoError(t, err)
	assert.Equal(t, 3, len(results))
	assert.Equal(t, "name", results[0].Field)
	assert.False(t, results[0].Desc)
	assert.Equal(t, "age", results[1].Field)
	assert.True(t, results[1].Desc)
	assert.Equal(t, "created_at", results[2].Field)
	assert.False(t, results[2].Desc)

	// 测试空字符串
	results = nil
	err = ParseOrderByString("", handler)
	assert.NoError(t, err)
	assert.Equal(t, 0, len(results))

	// 测试只有空格的字符串
	results = nil
	err = ParseOrderByString("   ", handler)
	assert.NoError(t, err)
	assert.Equal(t, 0, len(results))
}

func TestParseOrderByStrings(t *testing.T) {
	var results []struct {
		Field string
		Desc  bool
	}

	handler := func(field string, desc bool) {
		results = append(results, struct {
			Field string
			Desc  bool
		}{Field: field, Desc: desc})
	}

	// 测试正常解析
	err := ParseOrderByStrings([]string{"name", "-age", "+created_at"}, handler)
	assert.NoError(t, err)
	assert.Equal(t, 3, len(results))
	assert.Equal(t, "name", results[0].Field)
	assert.False(t, results[0].Desc)
	assert.Equal(t, "age", results[1].Field)
	assert.True(t, results[1].Desc)
	assert.Equal(t, "created_at", results[2].Field)
	assert.False(t, results[2].Desc)

	// 测试空字符串数组
	results = nil
	err = ParseOrderByStrings([]string{}, handler)
	assert.NoError(t, err)
	assert.Equal(t, 0, len(results))

	// 测试包含空字符串的数组
	results = nil
	err = ParseOrderByStrings([]string{"", "   "}, handler)
	assert.NoError(t, err)
	assert.Equal(t, 0, len(results))
}

func TestParseOrderByField(t *testing.T) {
	var results []struct {
		Field string
		Desc  bool
	}

	handler := func(field string, desc bool) {
		results = append(results, struct {
			Field string
			Desc  bool
		}{Field: field, Desc: desc})
	}

	// 测试升序解析
	results = nil
	ParseOrderByField("name", handler)
	assert.Equal(t, 1, len(results))
	assert.Equal(t, "name", results[0].Field)
	assert.False(t, results[0].Desc)

	// 测试降序解析
	results = nil
	ParseOrderByField("-age", handler)
	assert.Equal(t, 1, len(results))
	assert.Equal(t, "age", results[0].Field)
	assert.True(t, results[0].Desc)

	// 测试带+的升序解析
	results = nil
	ParseOrderByField("+created_at", handler)
	assert.Equal(t, 1, len(results))
	assert.Equal(t, "created_at", results[0].Field)
	assert.False(t, results[0].Desc)

	// 测试空字符串
	results = nil
	ParseOrderByField("", handler)
	assert.Equal(t, 0, len(results))

	// 测试只有空格的字符串
	results = nil
	ParseOrderByField("   ", handler)
	assert.Equal(t, 0, len(results))
}
