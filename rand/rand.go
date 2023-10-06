package rand

import (
	"math/rand"
	"time"
)

var RANDOM = rand.New(rand.NewSource(time.Now().UnixNano()))

// RandomInt 根据区间产生随机数
func RandomInt(min, max int) int {
	if min >= max {
		return max
	}
	return RANDOM.Intn(max-min) + min
}

// RandomInt64 根据区间产生随机数
func RandomInt64(min, max int64) int64 {
	if min >= max {
		return max
	}
	return RANDOM.Int63n(max-min) + min
}
