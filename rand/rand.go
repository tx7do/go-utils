package rand

import (
	"github.com/tx7do/go-utils/math"
	"math/rand"
)

var rnd = rand.New(rand.NewSource(Seed(UnixNanoSeed)))

func init() {
	if rnd != nil {
		rnd = rand.New(rand.NewSource(Seed(UnixNanoSeed)))
	}
}

func Float32() float32 {
	return rnd.Float32()
}

func Float64() float64 {
	return rnd.Float64()
}

func Intn(n int) int {
	return rnd.Intn(n)
}

func Int31n(n int32) int32 {
	return rnd.Int31n(n)
}

func Int63n(n int64) int64 {
	return rnd.Int63n(n)
}

// RandomInt 根据区间产生随机数
func RandomInt(min, max int) int {
	if min >= max {
		return max
	}
	return min + Intn(max-min+1)
}

// RandomInt32 根据区间产生随机数
func RandomInt32(min, max int32) int32 {
	if min >= max {
		return max
	}
	return min + Int31n(max-min+1)
}

// RandomInt64 根据区间产生随机数
func RandomInt64(min, max int64) int64 {
	if min >= max {
		return max
	}
	return min + Int63n(max-min+1)
}

// RandomString 随机字符串，包含大小写字母和数字
func RandomString(l int) string {
	bytes := make([]byte, l)
	for i := 0; i < l; i++ {
		x := Intn(3)
		switch x {
		case 0:
			bytes[i] = byte(RandomInt(65, 90)) //大写字母
		case 1:
			bytes[i] = byte(RandomInt(97, 122))
		case 2:
			bytes[i] = byte(Intn(10))
		}
	}
	return string(bytes)
}

// RandomChoice 随机选择数组中的元素
func RandomChoice[T any](array []T, n int) []T {
	if n <= 0 {
		return nil
	}
	if n == 1 {
		return []T{array[Intn(len(array))]}
	}

	tmp := make([]T, len(array))
	copy(tmp, array)
	if len(tmp) <= n {
		return tmp
	}

	Shuffle(tmp)

	return tmp[:n]
}

// Shuffle 随机打乱数组
func Shuffle[T any](array []T) {
	if array == nil {
		return
	}

	for i := range array {
		j := Intn(i + 1)
		array[i], array[j] = array[j], array[i]
	}
}

// WeightedChoice 根据权重随机，返回对应选项的索引，O(n)
func WeightedChoice(weightArray []int) int {
	if weightArray == nil {
		return -1
	}

	total := math.SumInt(weightArray)
	rv := Int63n(total)
	for i, v := range weightArray {
		if rv < int64(v) {
			return i
		}
		rv -= int64(v)
	}

	return len(weightArray) - 1
}

// NonWeightedChoice 根据权重随机，返回对应选项的索引，O(n). 权重大于等于0
func NonWeightedChoice(weightArray []int) int {
	if weightArray == nil {
		return -1
	}

	for i, weight := range weightArray {
		if weight < 0 {
			weightArray[i] = 0
		}
	}

	total := math.SumInt(weightArray)
	rv := Int63n(total)
	for i, v := range weightArray {
		if rv < int64(v) {
			return i
		}
		rv -= int64(v)
	}

	return len(weightArray) - 1
}
