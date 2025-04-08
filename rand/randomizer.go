package rand

import (
	"github.com/tx7do/go-utils/math"
	"math/rand"
)

type Randomizer struct {
	rnd *rand.Rand
}

func NewRandomizer(seedType SeedType) *Randomizer {
	return &Randomizer{
		rnd: rand.New(rand.NewSource(Seed(seedType))),
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

func (r *Randomizer) Int31() int32 {
	return r.rnd.Int31()
}

func (r *Randomizer) Int63() int64 {
	return r.rnd.Int63()
}

func (r *Randomizer) Uint32() uint32 {
	return r.rnd.Uint32()
}

func (r *Randomizer) Uint64() uint64 {
	return r.rnd.Uint64()
}

func (r *Randomizer) Intn(n int) int {
	return r.rnd.Intn(n)
}

func (r *Randomizer) Int31n(n int32) int32 {
	return r.rnd.Int31n(n)
}

func (r *Randomizer) Int63n(n int64) int64 {
	return r.rnd.Int63n(n)
}

// RangeInt 根据区间产生随机数
func (r *Randomizer) RangeInt(min, max int) int {
	if min >= max {
		return max
	}
	return min + r.Intn(max-min+1)
}

// RangeInt32 根据区间产生随机数
func (r *Randomizer) RangeInt32(min, max int32) int32 {
	if min >= max {
		return max
	}
	return min + r.Int31n(max-min+1)
}

// RangeInt64 根据区间产生随机数
func (r *Randomizer) RangeInt64(min, max int64) int64 {
	if min >= max {
		return max
	}
	return min + r.Int63n(max-min+1)
}

// RangeUint 根据区间产生随机数
func (r *Randomizer) RangeUint(min, max uint) uint {
	if min >= max {
		return max
	}
	return min + uint(r.Intn(int(max-min+1)))
}

// RangeUint32 根据区间产生随机数
func (r *Randomizer) RangeUint32(min, max uint32) uint32 {
	if min >= max {
		return max
	}
	return min + uint32(r.Int31n(int32(max-min+1)))
}

// RangeUint64 根据区间产生随机数
func (r *Randomizer) RangeUint64(min, max uint64) uint64 {
	if min >= max {
		return max
	}
	return min + uint64(r.Int63n(int64(max-min+1)))
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
		x := r.Intn(3)
		switch x {
		case 0:
			bytes[i] = byte(r.RangeInt(65, 90)) //大写字母
		case 1:
			bytes[i] = byte(r.RangeInt(97, 122))
		case 2:
			bytes[i] = byte(r.Intn(10))
		}
	}
	return string(bytes)
}

// WeightedChoice 根据权重随机，返回对应选项的索引，O(n)
func (r *Randomizer) WeightedChoice(weightArray []int) int {
	if weightArray == nil {
		return -1
	}

	total := math.SumInt(weightArray)
	rv := r.Int63n(total)
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
	if weightArray == nil {
		return -1
	}

	for i, weight := range weightArray {
		if weight < 0 {
			weightArray[i] = 0
		}
	}

	total := math.SumInt(weightArray)
	rv := r.Int63n(total)
	for i, v := range weightArray {
		if rv < int64(v) {
			return i
		}
		rv -= int64(v)
	}

	return len(weightArray) - 1
}
