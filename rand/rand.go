package rand

import (
	"crypto/sha256"
	"encoding/binary"
	"math/rand/v2"
	"time"

	"github.com/tx7do/go-utils/math"
)

// Float32 生成 [0, 1) 随机 float32
func Float32() float32 {
	return rand.Float32()
}

// Float64 生成 [0, 1) 随机 float64
func Float64() float64 {
	return rand.Float64()
}

// IntN 生成 [0, n) 随机 int
func IntN(n int) int {
	return rand.IntN(n)
}

// Int32N 生成 [0, n) 随机 int32
func Int32N(n int32) int32 {
	return rand.Int32N(n)
}

// Int64N 生成 [0, n) 随机 int64
func Int64N(n int64) int64 {
	return rand.Int64N(n)
}

// RandomInt 生成 [min, max] 范围内随机 int
func RandomInt(min, max int) int {
	if min >= max {
		return max
	}
	return min + IntN(max-min+1)
}

// RandomInt32 生成 [min, max] 范围内随机 int32
func RandomInt32(min, max int32) int32 {
	if min >= max {
		return max
	}
	return min + Int32N(max-min+1)
}

// RandomInt64 生成 [min, max] 范围内随机 int64
func RandomInt64(min, max int64) int64 {
	if min >= max {
		return max
	}
	return min + Int64N(max-min+1)
}

// RandomUint 生成 [min, max] 范围内随机 uint
func RandomUint(min, max uint) uint {
	if min >= max {
		return max
	}
	return min + rand.UintN(max-min+1)
}

// RandomUint32 生成 [min, max] 范围内随机 uint32
func RandomUint32(min, max uint32) uint32 {
	if min >= max {
		return max
	}
	return min + rand.Uint32N(max-min+1)
}

// RandomUint64 生成 [min, max] 范围内随机 uint64
func RandomUint64(min, max uint64) uint64 {
	if min >= max {
		return max
	}
	return min + rand.Uint64N(max-min+1)
}

// RandomDuration 生成 [min, max] 范围内随机 time.Duration
func RandomDuration(min, max time.Duration) time.Duration {
	if min >= max {
		return max
	}
	return min + time.Duration(Int64N(int64(max-min+1)))
}

// RandomString 生成指定长度随机字符串（大小写字母+数字）
func RandomString(l int) string {
	if l <= 0 {
		return ""
	}

	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

	bytes := make([]byte, l)
	for i := 0; i < l; i++ {
		bytes[i] = charset[IntN(len(charset))]
	}

	return string(bytes)
}

// RandomChoice 从数组随机选择 n 个元素（不修改原数组）
func RandomChoice[T any](array []T, n int) []T {
	if n <= 0 || len(array) == 0 {
		return nil
	}

	if n == 1 {
		return []T{array[rand.IntN(len(array))]}
	}

	tmp := make([]T, len(array))
	copy(tmp, array)

	if n >= len(tmp) {
		rand.Shuffle(len(tmp), func(i, j int) {
			tmp[i], tmp[j] = tmp[j], tmp[i]
		})
		return tmp
	}

	for i := 0; i < n; i++ {
		j := i + rand.IntN(len(tmp)-i)
		tmp[i], tmp[j] = tmp[j], tmp[i]
	}
	return tmp[:n]
}

// Shuffle 均匀打乱切片（标准 Fisher–Yates 算法）
func Shuffle[T any](array []T) {
	if len(array) < 2 {
		return
	}

	for i := len(array) - 1; i > 0; i-- {
		j := IntN(i + 1)
		array[i], array[j] = array[j], array[i]
	}
}

// WeightedChoice 根据权重随机，返回对应选项的索引，O(n)
func WeightedChoice(weightArray []int) int {
	n := len(weightArray)
	if n == 0 {
		return -1
	}
	if n == 1 {
		return 0
	}

	var total int64
	for _, w := range weightArray {
		if w > 0 {
			total += int64(w)
		}
	}

	if total <= 0 {
		return 0
	}

	rv := rand.Int64N(total)
	var cursor int64
	for i, v := range weightArray {
		if v <= 0 {
			continue
		}
		cursor += int64(v)
		if rv < cursor {
			return i
		}
	}
	return n - 1
}

// NonWeightedChoice 权重非负随机选择，返回对应选项的索引，O(n). 权重大于等于0
func NonWeightedChoice(weightArray []int) int {
	if weightArray == nil {
		return -1
	}

	// 复制避免修改原数组
	weights := make([]int, len(weightArray))
	copy(weights, weightArray)

	// 确保权重 >= 0
	for i := range weights {
		if weights[i] < 0 {
			weights[i] = 0
		}
	}

	total := math.SumInt(weights)
	if total <= 0 {
		return 0
	}

	rv := Int64N(total)
	for i, v := range weights {
		if rv < int64(v) {
			return i
		}
		rv -= int64(v)
	}

	return len(weights) - 1
}

// SHA256Value 生成基于 serverSeed、clientSeed 和 nonce 的 SHA256 哈希值，并返回前 8 字节作为 uint64。适用于需要基于多个输入生成随机数的场景，如游戏中的随机事件生成等。
func SHA256Value(serverSeed, clientSeed string, nonce uint64) uint64 {
	h := sha256.New()
	h.Write([]byte(serverSeed))
	h.Write([]byte(clientSeed))

	nBuf := make([]byte, 8)
	binary.LittleEndian.PutUint64(nBuf, nonce)
	h.Write(nBuf)

	res := h.Sum(nil)
	return binary.LittleEndian.Uint64(res[:8])
}
