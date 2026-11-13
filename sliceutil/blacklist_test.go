package sliceutil

import (
	"strconv"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFilterBlacklist(t *testing.T) {
	out := FilterBlacklist([]string{"a", "b", "c", "d"}, []string{"b", "d"})
	assert.Equal(t, []string{"a", "c"}, out)

	assert.Empty(t, FilterBlacklist([]string{"a"}, []string{"a"}))

	// 空/不命中黑名单：原样返回全部
	assert.Equal(t, []string{"x"}, FilterBlacklist([]string{"x"}, nil))
}

func TestNumberSliceToString(t *testing.T) {
	assert.Equal(t, "1,2,3", NumberSliceToString([]uint32{1, 2, 3}))
	assert.Equal(t, "", NumberSliceToString[uint32](nil))

	// 与手工拼接一致
	want := strings.Join([]string{strconv.FormatUint(9, 10), strconv.FormatUint(8, 10)}, ",")
	assert.Equal(t, want, NumberSliceToString([]uint32{9, 8}))

	// 有符号/无符号各位宽口径均正确（含 int64 与负数）
	assert.Equal(t, "4,5,6", NumberSliceToString([]int64{4, 5, 6}))
	assert.Equal(t, "7,8", NumberSliceToString([]int{7, 8}))
	assert.Equal(t, "-1,2", NumberSliceToString([]int64{-1, 2}))
	assert.Equal(t, "18446744073709551615", NumberSliceToString([]uint64{^uint64(0)}))
}
