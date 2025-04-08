package rand

import (
	"fmt"
	"testing"
)

func TestUnixNanoSeed(t *testing.T) {
	seeder := NewSeeder(UnixNanoSeed)

	var seeds = make(map[int64]bool)
	for i := 0; i < 100000; i++ {
		seed := seeder.Seed()
		seeds[seed] = true
	}
	fmt.Println("UnixNano Seed", len(seeds))
}

func TestMapHashSeed(t *testing.T) {
	seeder := NewSeeder(MapHashSeed)

	var seeds = make(map[int64]bool)
	for i := 0; i < 100000; i++ {
		seed := seeder.Seed()
		seeds[seed] = true
	}
	fmt.Println("MapHash Seed", len(seeds))
}

func TestCryptoRandSeed(t *testing.T) {
	seeder := NewSeeder(CryptoRandSeed)

	var seeds = make(map[int64]bool)
	for i := 0; i < 100000; i++ {
		seed := seeder.Seed()
		seeds[seed] = true
	}
	fmt.Println("CryptoRand Seed", len(seeds))
}

func TestRandomStringSeed(t *testing.T) {
	seeder := NewSeeder(RandomStringSeed)

	var seeds = make(map[int64]bool)
	for i := 0; i < 100000; i++ {
		seed := seeder.Seed()
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
