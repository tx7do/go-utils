package rand

import (
	"fmt"
	"testing"
)

func TestUnixNano(t *testing.T) {
	randomizer := NewRandomizer(UnixNanoSeed)

	var seeds = make(map[int64]bool)
	for i := 0; i < 100000; i++ {
		seed := randomizer.Seed()
		seeds[seed] = true
	}
	fmt.Println("UnixNano Seed", len(seeds))
}

func TestMapHash(t *testing.T) {
	randomizer := NewRandomizer(MapHashSeed)

	var seeds = make(map[int64]bool)
	for i := 0; i < 100000; i++ {
		seed := randomizer.Seed()
		seeds[seed] = true
	}
	fmt.Println("MapHash Seed", len(seeds))
}

func TestCryptoRand(t *testing.T) {
	randomizer := NewRandomizer(CryptoRandSeed)

	var seeds = make(map[int64]bool)
	for i := 0; i < 100000; i++ {
		seed := randomizer.Seed()
		seeds[seed] = true
	}
	fmt.Println("CryptoRand Seed", len(seeds))
}

func TestRandomString(t *testing.T) {
	randomizer := NewRandomizer(RandomStringSeed)

	var seeds = make(map[int64]bool)
	for i := 0; i < 100000; i++ {
		seed := randomizer.Seed()
		seeds[seed] = true
	}
	fmt.Println("RandomString Seed", len(seeds))
}

func TestSeed(t *testing.T) {
	var seeds = make(map[int64]bool)
	for i := 0; i < 100000; i++ {
		seed := Seed(UnixNanoSeed)
		seeds[seed] = true
	}
	fmt.Println("UnixNano Seed", len(seeds))

	seeds = make(map[int64]bool)
	for i := 0; i < 100000; i++ {
		seed := Seed(MapHashSeed)
		seeds[seed] = true
	}
	fmt.Println("MapHash Seed", len(seeds))

	seeds = make(map[int64]bool)
	for i := 0; i < 100000; i++ {
		seed := Seed(CryptoRandSeed)
		seeds[seed] = true
	}
	fmt.Println("CryptoRand Seed", len(seeds))

	seeds = make(map[int64]bool)
	for i := 0; i < 100000; i++ {
		seed := Seed(RandomStringSeed)
		seeds[seed] = true
	}
	fmt.Println("RandomString Seed", len(seeds))
}
