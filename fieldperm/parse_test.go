// Package fieldperm_test 针对 pkg/fieldperm 的纯字符串解析原语做黑盒测试。
//
// 本文件覆盖 ParseTokenEntries / HiddenFieldsOf / IsEmpty：
// 令牌隐藏字段条目（"资源.字段"）的解析、非法条目跳过、多资源聚合，
// 以及按资源取隐藏字段集与空集判定。
package fieldperm_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/tx7do/go-utils/fieldperm"
)

// TestParseTokenEntries_ValidEntriesAggregated 合法条目按资源聚合为
// 资源名 → 字段名集合的结构，重复条目被集合去重。
func TestParseTokenEntries_ValidEntriesAggregated(t *testing.T) {
	tests := []struct {
		name     string
		entries  []string
		expected map[string]map[string]struct{}
	}{
		{
			name:    "single entry",
			entries: []string{"resourceA.fieldX"},
			expected: map[string]map[string]struct{}{
				"resourceA": {"fieldX": {}},
			},
		},
		{
			name:    "multiple resources and fields aggregated",
			entries: []string{"resourceA.field1", "resourceB.field2", "resourceA.field3"},
			expected: map[string]map[string]struct{}{
				"resourceA": {"field1": {}, "field3": {}},
				"resourceB": {"field2": {}},
			},
		},
		{
			name:    "duplicate entries deduplicated by set",
			entries: []string{"resourceA.field1", "resourceA.field1"},
			expected: map[string]map[string]struct{}{
				"resourceA": {"field1": {}},
			},
		},
		{
			name:     "empty input yields empty map",
			entries:  nil,
			expected: map[string]map[string]struct{}{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := fieldperm.ParseTokenEntries(tt.entries)
			require.NotNil(t, got, "解析结果必须为非 nil 空结构而非 nil")
			assert.Equal(t, tt.expected, got)
		})
	}
}

// TestParseTokenEntries_InvalidEntriesSkipped 缺资源、缺字段或无点号的条目
// 必须整条跳过，不得产生半截聚合，也不得污染其它合法条目。
func TestParseTokenEntries_InvalidEntriesSkipped(t *testing.T) {
	tests := []struct {
		name     string
		entries  []string
		expected map[string]map[string]struct{}
	}{
		{
			name:     "no dot separator",
			entries:  []string{"nodot"},
			expected: map[string]map[string]struct{}{},
		},
		{
			name:     "missing resource",
			entries:  []string{".fieldOnly"},
			expected: map[string]map[string]struct{}{},
		},
		{
			name:     "missing field",
			entries:  []string{"resourceOnly."},
			expected: map[string]map[string]struct{}{},
		},
		{
			name:     "empty string entry",
			entries:  []string{""},
			expected: map[string]map[string]struct{}{},
		},
		{
			name:    "invalid entries mixed with valid ones",
			entries: []string{"goodRes.goodField", "bad-no-dot", ".noRes", "noField.", ""},
			expected: map[string]map[string]struct{}{
				"goodRes": {"goodField": {}},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := fieldperm.ParseTokenEntries(tt.entries)
			require.NotNil(t, got)
			assert.Equal(t, tt.expected, got)
		})
	}
}

// TestParseTokenEntries_CutsAtFirstDot strings.Cut 在第一个点号处切分：
// "a.b.c" 解析为资源 "a"、字段 "b.c"——记录该切分语义。
func TestParseTokenEntries_CutsAtFirstDot(t *testing.T) {
	got := fieldperm.ParseTokenEntries([]string{"a.b.c"})
	require.Len(t, got, 1)
	assert.Contains(t, got, "a")
	assert.Equal(t, map[string]struct{}{"b.c": {}}, got["a"])
}

// TestHiddenFieldsOf_HitAndMiss 命中资源返回其字段集合；
// 未命中资源（含从未出现的资源）返回 nil。
func TestHiddenFieldsOf_HitAndMiss(t *testing.T) {
	entries := []string{
		"resourceA.field1",
		"resourceA.field2",
		"resourceB.field3",
	}

	hit := fieldperm.HiddenFieldsOf(entries, "resourceA")
	require.NotNil(t, hit, "命中的资源必须返回字段集合")
	assert.Equal(t, map[string]struct{}{"field1": {}, "field2": {}}, hit)

	for _, missing := range []string{"resourceB.typo", "resourceC", ""} {
		t.Run("miss/"+missing, func(t *testing.T) {
			assert.Nil(t, fieldperm.HiddenFieldsOf(entries, missing), "未命中资源必须返回 nil")
		})
	}
}

// TestIsEmpty 空集判定：nil 与空 map 均为空（无需裁剪），非空集合不为空。
func TestIsEmpty(t *testing.T) {
	tests := []struct {
		name   string
		hidden map[string]struct{}
		want   bool
	}{
		{name: "nil map", hidden: nil, want: true},
		{name: "empty map", hidden: map[string]struct{}{}, want: true},
		{name: "non-empty map", hidden: map[string]struct{}{"x": {}}, want: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, fieldperm.IsEmpty(tt.hidden))
		})
	}
}
