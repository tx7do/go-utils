package rand

import (
	cryptoRand "crypto/rand"
	"encoding/binary"
	"golang.org/x/exp/rand"
	"hash/maphash"
	mathRand "math/rand"
	"runtime"
	"time"
)

type SeedType string

const (
	UnixNanoSeed     SeedType = "UnixNano"
	MapHashSeed      SeedType = "MapHash"
	CryptoRandSeed   SeedType = "CryptoRand"
	RandomStringSeed SeedType = "RandomString"
)

type Randomizer struct {
	seedType SeedType
}

func NewRandomizer(seedType SeedType) *Randomizer {
	return &Randomizer{
		seedType: seedType,
	}
}

func (r *Randomizer) UnixNano() int64 {
	// 获取当前时间戳
	timestamp := time.Now().UnixNano()

	// 生成一个随机数
	var randomBytes [8]byte
	_, err := rand.Read(randomBytes[:])
	if err != nil {
		panic("failed to generate random bytes")
	}
	randomPart := int64(binary.LittleEndian.Uint64(randomBytes[:]))

	// 获取 Goroutine ID（或其他唯一标识）
	goroutineID := int64(runtime.NumGoroutine())

	// 结合时间戳、随机数和 Goroutine ID
	seed := timestamp ^ randomPart ^ goroutineID

	return seed
}

func (r *Randomizer) MapHash() int64 {
	return int64(new(maphash.Hash).Sum64())
}

func (r *Randomizer) CryptoRand() int64 {
	var b [8]byte
	_, err := cryptoRand.Read(b[:])
	if err != nil {
		panic("cannot seed math/rand package with cryptographically secure random number generator")
	}
	seed := int64(binary.LittleEndian.Uint64(b[:]))
	return seed
}

var Alpha = "abcdefghijkmnpqrstuvwxyzABCDEFGHJKLMNPQRSTUVWXYZ23456789"

func (r *Randomizer) RandomString() int64 {
	const size = 8
	buf := make([]byte, size)
	for i := 0; i < size; i++ {
		buf[i] = Alpha[mathRand.Intn(len(Alpha))]
	}
	seed := int64(binary.LittleEndian.Uint64(buf[:]))

	return seed
}

func (r *Randomizer) Seed() int64 {
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
	randomizer := NewRandomizer(seedType)
	return randomizer.Seed()
}
