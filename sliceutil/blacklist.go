package sliceutil

import (
	"strconv"
	"strings"
)

// FilterBlacklist 就地剔除 data 中命中 blacklist 的元素（黑名单过滤）。
// 返回过滤后的切片（与入参共享底层数组，容量可能大于长度）。
func FilterBlacklist(data []string, blacklist []string) []string {
	bm := make(map[string]struct{}, len(blacklist))
	for _, s := range blacklist {
		bm[s] = struct{}{}
	}

	n := 0
	for _, x := range data {
		if _, found := bm[x]; !found {
			data[n] = x
			n++
		}
	}
	return data[:n]
}

// NumberSliceToString 把数字切片格式化为逗号分隔的十进制字符串。
func NumberSliceToString(numbers []uint32) string {
	return strings.Join(
		Map(numbers, func(value uint32, _ int, _ []uint32) string { return strconv.FormatUint(uint64(value), 10) }),
		",",
	)
}
