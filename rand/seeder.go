package rand

import (
	cryptoRand "crypto/rand"
	"encoding/binary"
	"hash/maphash"
	mathRand "math/rand/v2"
	"runtime"
	"time"
)

// SeedType 种子类型枚举
type SeedType string

const (
	UnixNanoSeed     SeedType = "UnixNano"     // 时间戳 + 随机扰动
	MapHashSeed      SeedType = "MapHash"      // 快速哈希种子
	CryptoRandSeed   SeedType = "CryptoRand"   // 密码学安全种子（最高安全）
	RandomStringSeed SeedType = "RandomString" // 字符串生成种子（测试用）
)

// Seeder 种子生成器
type Seeder struct {
	seedType SeedType
}

// NewSeeder 创建种子生成器
func NewSeeder(seedType SeedType) *Seeder {
	return &Seeder{
		seedType: seedType,
	}
}

// UnixNano 使用时间戳 + 系统随机 + GoroutineID 混合生成种子
func (r *Seeder) UnixNano() int64 {
	b := make([]byte, 8)
	_, err := cryptoRand.Read(b)
	if err != nil {
		return time.Now().UnixNano()
	}
	rnd := int64(binary.LittleEndian.Uint64(b))
	ts := time.Now().UnixNano()
	goid := int64(runtime.NumGoroutine())
	return ts ^ rnd ^ goid
}

// MapHash 使用 maphash 快速生成种子
func (r *Seeder) MapHash() int64 {
	return int64(new(maphash.Hash).Sum64())
}

// CryptoRand 使用密码学安全随机数生成种子
func (r *Seeder) CryptoRand() int64 {
	var b [8]byte
	_, err := cryptoRand.Read(b[:])
	if err != nil {
		panic("cannot seed math/rand package with cryptographically secure random number generator")
	}
	seed := int64(binary.LittleEndian.Uint64(b[:]))
	return seed
}

var Alpha = "abcdefghijkmnpqrstuvwxyzABCDEFGHJKLMNPQRSTUVWXYZ23456789"

func (r *Seeder) RandomString() int64 {
	const size = 8
	buf := make([]byte, size)
	for i := 0; i < size; i++ {
		idx := mathRand.IntN(len(Alpha))
		buf[i] = Alpha[idx]
	}
	seed := int64(binary.LittleEndian.Uint64(buf[:]))

	return seed
}

// Seed 根据类型生成最终 int64 种子
func (r *Seeder) Seed() int64 {
	switch r.seedType {
	default:
		fallthrough
	case UnixNanoSeed:
		return r.UnixNano()
	case MapHashSeed:
		return r.MapHash()
	case CryptoRandSeed:
		return r.CryptoRand()
	case RandomStringSeed:
		return r.RandomString()
	}
}

// Seed generates a seed based on the specified SeedType.
func Seed(seedType SeedType) int64 {
	randomizer := NewSeeder(seedType)
	return randomizer.Seed()
}
