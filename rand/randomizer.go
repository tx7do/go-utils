package rand

import (
	"math/rand/v2"

	"github.com/tx7do/go-utils/math"
)

type Randomizer struct {
	rnd *rand.Rand
}

func NewRandomizer(seedType SeedType) *Randomizer {
	seed := Seed(seedType)
	source := rand.NewPCG(uint64(seed), 0)
	rnd := rand.New(source)
	return &Randomizer{
		rnd: rnd,
	}
}

func (r *Randomizer) Float32() float32 {
	return r.rnd.Float32()
}

func (r *Randomizer) Float64() float64 {
	return r.rnd.Float64()
}

func (r *Randomizer) Int() int {
	return r.rnd.Int()
}

func (r *Randomizer) Int32() int32 {
	return r.rnd.Int32()
}

func (r *Randomizer) Int64() int64 {
	return r.rnd.Int64()
}

func (r *Randomizer) Uint32() uint32 {
	return r.rnd.Uint32()
}

func (r *Randomizer) Uint64() uint64 {
	return r.rnd.Uint64()
}

func (r *Randomizer) IntN(n int) int {
	return r.rnd.IntN(n)
}

func (r *Randomizer) Int32N(n int32) int32 {
	return r.rnd.Int32N(n)
}

func (r *Randomizer) Int64N(n int64) int64 {
	return r.rnd.Int64N(n)
}

func (r *Randomizer) UintN(n uint) uint {
	return r.rnd.UintN(n)
}

func (r *Randomizer) Uint32N(n uint32) uint32 {
	return r.rnd.Uint32N(n)
}

func (r *Randomizer) Uint64N(n uint64) uint64 {
	return r.rnd.Uint64N(n)
}

// RangeInt 根据区间产生随机数
func (r *Randomizer) RangeInt(min, max int) int {
	if min >= max {
		return max
	}
	return min + r.IntN(max-min+1)
}

// RangeInt32 根据区间产生随机数
func (r *Randomizer) RangeInt32(min, max int32) int32 {
	if min >= max {
		return max
	}
	return min + r.Int32N(max-min+1)
}

// RangeInt64 根据区间产生随机数
func (r *Randomizer) RangeInt64(min, max int64) int64 {
	if min >= max {
		return max
	}
	return min + r.Int64N(max-min+1)
}

// RangeUint 根据区间产生随机数
func (r *Randomizer) RangeUint(min, max uint) uint {
	if min >= max {
		return max
	}
	return min + uint(r.IntN(int(max-min+1)))
}

// RangeUint32 根据区间产生随机数
func (r *Randomizer) RangeUint32(min, max uint32) uint32 {
	if min >= max {
		return max
	}
	return min + uint32(r.Int32N(int32(max-min+1)))
}

// RangeUint64 根据区间产生随机数
func (r *Randomizer) RangeUint64(min, max uint64) uint64 {
	if min >= max {
		return max
	}
	return min + uint64(r.Int64N(int64(max-min+1)))
}

// RangeFloat32 根据区间产生随机数
func (r *Randomizer) RangeFloat32(min, max float32) float32 {
	if min >= max {
		return max
	}
	return min + r.Float32()*(max-min)
}

// RangeFloat64 根据区间产生随机数
func (r *Randomizer) RangeFloat64(min, max float64) float64 {
	if min >= max {
		return max
	}
	return min + r.Float64()*(max-min)
}

// RandomString 随机字符串，包含大小写字母和数字
func (r *Randomizer) RandomString(l int) string {
	bytes := make([]byte, l)
	for i := 0; i < l; i++ {
		x := r.IntN(3)
		switch x {
		case 0:
			bytes[i] = byte(r.RangeInt(65, 90)) //大写字母
		case 1:
			bytes[i] = byte(r.RangeInt(97, 122))
		case 2:
			bytes[i] = byte(r.IntN(10))
		}
	}
	return string(bytes)
}

// WeightedChoice 根据权重随机，返回对应选项的索引，O(n)
func (r *Randomizer) WeightedChoice(weightArray []int) int {
	if len(weightArray) == 0 {
		return -1
	}

	total := math.SumInt(weightArray)
	if total <= 0 {
		return 0
	}

	rv := r.Int64N(total)
	for i, v := range weightArray {
		if rv < int64(v) {
			return i
		}
		rv -= int64(v)
	}

	return len(weightArray) - 1
}

// NonWeightedChoice 根据权重随机，返回对应选项的索引，O(n). 权重大于等于0
func (r *Randomizer) NonWeightedChoice(weightArray []int) int {
	if len(weightArray) == 0 {
		return -1
	}

	// 复制避免修改调用方传入的切片
	weights := make([]int, len(weightArray))
	copy(weights, weightArray)

	for i, weight := range weights {
		if weight < 0 {
			weights[i] = 0
		}
	}

	total := math.SumInt(weights)
	if total <= 0 {
		return 0
	}

	rv := r.Int64N(total)
	for i, v := range weights {
		if rv < int64(v) {
			return i
		}
		rv -= int64(v)
	}

	return len(weights) - 1
}
