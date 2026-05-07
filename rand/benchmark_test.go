package rand

import (
	"testing"
	"time"
)

var (
	benchInt    int
	benchInt64  int64
	benchUint64 uint64
	benchStr    string
	benchBool   bool
	benchSlice  []int
)

func BenchmarkIntN(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		benchInt = IntN(1000)
	}
}

func BenchmarkRandomInt64(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		benchInt64 = RandomInt64(1, 1_000_000)
	}
}

func BenchmarkRandomDuration(b *testing.B) {
	b.ReportAllocs()
	min := 100 * time.Millisecond
	max := 2 * time.Second
	for i := 0; i < b.N; i++ {
		benchInt64 = int64(RandomDuration(min, max))
	}
}

func BenchmarkRandomString16(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		benchStr = RandomString(16)
	}
}

func BenchmarkRandomChoice10of100(b *testing.B) {
	b.ReportAllocs()
	arr := make([]int, 100)
	for i := range arr {
		arr[i] = i
	}

	for i := 0; i < b.N; i++ {
		benchSlice = RandomChoice(arr, 10)
	}
}

func BenchmarkWeightedChoice(b *testing.B) {
	b.ReportAllocs()
	weights := []int{1, 3, 5, 7, 9, 2, 4, 6, 8, 10}
	for i := 0; i < b.N; i++ {
		benchInt = WeightedChoice(weights)
	}
}

func BenchmarkSHA256Value(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		benchUint64 = SHA256Value("server-seed", "client-seed", uint64(i))
	}
}

func BenchmarkRandomizerRangeUint64(b *testing.B) {
	b.ReportAllocs()
	r := NewRandomizer(PCGRandType, UnixNanoSeed)
	for i := 0; i < b.N; i++ {
		benchUint64 = r.RangeUint64(1, 1_000_000)
	}
}

func BenchmarkRandomizerWeightedChoice(b *testing.B) {
	b.ReportAllocs()
	r := NewRandomizer(PCGRandType, UnixNanoSeed)
	weights := []int{1, 3, 5, 7, 9, 2, 4, 6, 8, 10}
	for i := 0; i < b.N; i++ {
		benchInt = r.WeightedChoice(weights)
	}
}

func BenchmarkRandomizerAliasChoice(b *testing.B) {
	b.ReportAllocs()
	r := NewRandomizer(PCGRandType, UnixNanoSeed)
	weights := []int{1, 3, 5, 7, 9, 2, 4, 6, 8, 10}
	at := r.NewAliasTable(weights)
	for i := 0; i < b.N; i++ {
		benchInt = r.AliasChoice(at)
	}
}

func BenchmarkNewRandomizer(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		benchBool = NewRandomizer(PCGRandType, UnixNanoSeed) != nil
	}
}

func BenchmarkNewZipfRandomizer(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		benchBool = NewZipfRandomizer(UnixNanoSeed, 1.2, 1.0, 100) != nil
	}
}

func BenchmarkSeederSeed(b *testing.B) {
	b.ReportAllocs()

	cases := []struct {
		name     string
		seedType SeedType
	}{
		{name: "UnixNano", seedType: UnixNanoSeed},
		{name: "MapHash", seedType: MapHashSeed},
		{name: "CryptoRand", seedType: CryptoRandSeed},
		{name: "RandomString", seedType: RandomStringSeed},
	}

	for _, tc := range cases {
		b.Run(tc.name, func(b *testing.B) {
			s := NewSeeder(tc.seedType)
			b.ReportAllocs()
			for i := 0; i < b.N; i++ {
				benchInt64 = s.Seed()
			}
		})
	}
}
