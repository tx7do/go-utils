package rand

import (
	"crypto/rand"
	"encoding/binary"
	mathRand "math/rand/v2"

	"github.com/tx7do/go-utils/math"
)

// AliasTable 用于实现基于别名方法的加权随机选择算法，提供 O(1) 时间复杂度的随机选择
type AliasTable struct {
	prob  []float32 // 每个选项的概率
	alias []int     // 矩形中选项的索引
}

// RandType 随机数生成器类型枚举
type RandType string

const (
	PCGRandType     RandType = "PCG"     // PCG 随机数生成器
	ChaCha8RandType RandType = "ChaCha8" // ChaCha8 随机数生成器
	ZipfRandType    RandType = "Zipf"    // Zipf 随机数生成器
	SHA256RandType  RandType = "SHA256"  // 基于 SHA-256 哈希的随机数生成器（测试用，非高性能）
)

type Randomizer struct {
	rnd  *mathRand.Rand
	zipf *mathRand.Zipf
}

func NewRandomizer(randType RandType, seedType SeedType) *Randomizer {
	seeder := NewSeeder(seedType)

	switch randType {
	default:
		fallthrough
	case PCGRandType:
		s := uint64(seeder.Seed())
		source := mathRand.NewPCG(s, s^0x55AA55AA55AA55AA)
		return &Randomizer{rnd: mathRand.New(source)}

	case ChaCha8RandType:
		var seed [32]byte
		s := seeder.Seed()
		binary.LittleEndian.PutUint64(seed[0:8], uint64(s))
		_, _ = rand.Read(seed[8:32])

		source := mathRand.NewChaCha8(seed)
		return &Randomizer{rnd: mathRand.New(source)}

	case SHA256RandType:
		s := seeder.Seed()

		seedBuf := make([]byte, 8)
		binary.LittleEndian.PutUint64(seedBuf, uint64(s))

		source := NewSha256Source(seedBuf)
		return &Randomizer{rnd: mathRand.New(source)}

	case ZipfRandType:
		return nil
	}
}

// NewZipfRandomizer 创建 Zipf 随机数生成器，参数 s > 1，q >= 1，v > 0
// s: 控制分布的陡峭程度，s 越大，分布越陡峭；
// q: 控制分布的偏移量，q 越大，分布越偏向于较小的整数；
// v: 定义了生成随机数的范围，即生成的随机数将位于 [0, v) 之间。
func NewZipfRandomizer(seedType SeedType, s, q float64, v uint64) *Randomizer {
	seeder := NewSeeder(seedType)
	seed := uint64(seeder.Seed())

	source := mathRand.NewPCG(seed, seed^0x55AA)
	r := mathRand.New(source)

	z := mathRand.NewZipf(r, s, q, v)

	return &Randomizer{
		rnd:  r,
		zipf: z,
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
	if r.zipf != nil {
		return r.zipf.Uint64()
	}

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
		return min
	}
	return min + r.rnd.IntN(max-min+1)
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
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	b := make([]byte, l)
	for i := range b {
		b[i] = charset[r.rnd.IntN(len(charset))]
	}
	return string(b)
}

// WeightedChoice 根据权重随机，返回对应选项的索引，O(n)
func (r *Randomizer) WeightedChoice(weightArray []int) int {
	n := len(weightArray)
	if n == 0 {
		return -1
	}

	if n == 1 {
		if weightArray[0] > 0 {
			return 0
		}
		return -1
	}

	var total int64
	for _, w := range weightArray {
		if w > 0 {
			total += int64(w)
		}
	}

	if total <= 0 {
		return -1
	}

	rv := r.rnd.Int64N(total)
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

// Shuffle 洗牌算法，随机打乱元素顺序
func (r *Randomizer) Shuffle(n int, swap func(i, j int)) {
	r.rnd.Shuffle(n, swap)
}

// Normal 生成符合正态分布的随机数，参数 mean 是均值，stdDev 是标准差
func (r *Randomizer) Normal(mean, stdDev float64) float64 {
	return mean + r.rnd.NormFloat64()*stdDev
}

// Exp 生成符合指数分布的随机数，参数 lambda 是速率参数（lambda > 0）
func (r *Randomizer) Exp(lambda float64) float64 {
	return r.rnd.ExpFloat64() / lambda
}

// Perm 生成一个长度为 n 的随机排列，返回一个包含 0 到 n-1 的切片，顺序被随机打乱
func (r *Randomizer) Perm(n int) []int {
	return r.rnd.Perm(n)
}

// ZipfUint64 生成符合 Zipf 分布的随机数，如果未初始化 Zipf 生成器，则退回到普通随机数生成
func (r *Randomizer) ZipfUint64() uint64 {
	if r.zipf == nil {
		return r.rnd.Uint64()
	}
	return r.zipf.Uint64()
}

// NewAliasTable 根据权重数组构建别名表，用于 O(1) 时间复杂度的加权随机选择
func (r *Randomizer) NewAliasTable(weights []int) *AliasTable {
	n := len(weights)
	if n == 0 {
		return nil
	}

	sum := 0
	for _, w := range weights {
		sum += w
	}

	prob := make([]float32, n)
	alias := make([]int, n)

	// 计算平均权重，并将权重缩放到平均值的倍数
	avg := float32(sum) / float32(n)
	scaledWeights := make([]float32, n)

	var small, large []int
	for i, w := range weights {
		sw := float32(w) / avg
		scaledWeights[i] = sw
		if sw < 1.0 {
			small = append(small, i)
		} else {
			large = append(large, i)
		}
	}

	// 构建别名表
	for len(small) > 0 && len(large) > 0 {
		s := small[len(small)-1]
		small = small[:len(small)-1]
		l := large[len(large)-1]
		large = large[:len(large)-1]

		prob[s] = scaledWeights[s]
		alias[s] = l

		// 更新大权重的剩余权重
		scaledWeights[l] = (scaledWeights[l] + scaledWeights[s]) - 1.0
		if scaledWeights[l] < 1.0 {
			small = append(small, l)
		} else {
			large = append(large, l)
		}
	}

	// 处理剩余的权重，确保它们的概率为 1.0
	for len(large) > 0 {
		l := large[len(large)-1]
		large = large[:len(large)-1]
		prob[l] = 1.0
	}
	for len(small) > 0 {
		s := small[len(small)-1]
		small = small[:len(small)-1]
		prob[s] = 1.0
	}

	return &AliasTable{prob: prob, alias: alias}
}

// AliasChoice 使用别名表进行加权随机选择，返回选项的索引，O(1) 时间复杂度
func (r *Randomizer) AliasChoice(at *AliasTable) int {
	if at == nil || len(at.prob) == 0 {
		return -1
	}

	n := len(at.prob)
	// 1. 随机选择一个索引
	idx := r.IntN(n)

	// 2. 根据概率和别名表进行选择 (0.0 ~ 1.0)
	if r.Float32() < at.prob[idx] {
		return idx
	}
	return at.alias[idx]
}
