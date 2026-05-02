package rand

import (
	"encoding/binary"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSeeder_UnixNano_FallbackOnReadError(t *testing.T) {
	oldCryptoRead := cryptoRead
	oldNowUnixNano := nowUnixNano
	defer func() {
		cryptoRead = oldCryptoRead
		nowUnixNano = oldNowUnixNano
	}()

	const fallbackSeed int64 = 123456789
	cryptoRead = func(_ []byte) (int, error) {
		return 0, errors.New("read failed")
	}
	nowUnixNano = func() int64 { return fallbackSeed }

	seeder := NewSeeder(UnixNanoSeed)
	assert.Equal(t, fallbackSeed, seeder.UnixNano())
}

func TestSeeder_CryptoRand_NoPanicAndFallbackOnReadError(t *testing.T) {
	oldCryptoRead := cryptoRead
	oldNowUnixNano := nowUnixNano
	defer func() {
		cryptoRead = oldCryptoRead
		nowUnixNano = oldNowUnixNano
	}()

	const fallbackSeed int64 = 987654321
	cryptoRead = func(_ []byte) (int, error) {
		return 0, errors.New("read failed")
	}
	nowUnixNano = func() int64 { return fallbackSeed }

	seeder := NewSeeder(CryptoRandSeed)
	assert.NotPanics(t, func() {
		assert.Equal(t, fallbackSeed, seeder.CryptoRand())
	})
}

func TestSeeder_CryptoRand_UsesEntropyBytesWhenReadSucceeds(t *testing.T) {
	oldCryptoRead := cryptoRead
	defer func() {
		cryptoRead = oldCryptoRead
	}()

	bytes := []byte{1, 2, 3, 4, 5, 6, 7, 8}
	cryptoRead = func(dst []byte) (int, error) {
		copy(dst, bytes)
		return len(dst), nil
	}

	seeder := NewSeeder(CryptoRandSeed)
	expected := int64(binary.LittleEndian.Uint64(bytes))
	assert.Equal(t, expected, seeder.CryptoRand())
}

func TestSeeder_BasicDistribution_UniqueCount(t *testing.T) {
	tests := []struct {
		name      string
		seedType  SeedType
		minUnique int
	}{
		{name: "UnixNano", seedType: UnixNanoSeed, minUnique: 950},
		{name: "MapHash", seedType: MapHashSeed, minUnique: 950},
		{name: "CryptoRand", seedType: CryptoRandSeed, minUnique: 950},
		{name: "RandomString", seedType: RandomStringSeed, minUnique: 900},
	}

	const sampleSize = 1000
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			seeder := NewSeeder(tt.seedType)
			seeds := make(map[int64]struct{}, sampleSize)
			for i := 0; i < sampleSize; i++ {
				seeds[seeder.Seed()] = struct{}{}
			}
			assert.GreaterOrEqual(t, len(seeds), tt.minUnique)
		})
	}
}

func TestSeed_Function_BasicDistribution(t *testing.T) {
	types := []SeedType{UnixNanoSeed, MapHashSeed, CryptoRandSeed, RandomStringSeed}

	const sampleSize = 1000
	for _, seedType := range types {
		seeds := make(map[int64]struct{}, sampleSize)
		for i := 0; i < sampleSize; i++ {
			seeds[Seed(seedType)] = struct{}{}
		}
		assert.GreaterOrEqual(t, len(seeds), 900)
	}
}
