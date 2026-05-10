package rand

import (
	"testing"
	"time"
)

var (
	benchInt     int
	benchInt32   int32
	benchInt64   int64
	benchUint    uint
	benchUint32  uint32
	benchUint64  uint64
	benchFloat32 float32
	benchFloat64 float64
	benchStr     string
	benchBool    bool
	benchSlice   []int
	benchAny     any
)

// ─────────────────────────────────────────────
// Global functions
// ─────────────────────────────────────────────

func BenchmarkFloat32(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		benchFloat32 = Float32()
	}
}

func BenchmarkFloat64(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		benchFloat64 = Float64()
	}
}

func BenchmarkIntN(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		benchInt = IntN(1000)
	}
}

func BenchmarkInt32N(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		benchInt32 = Int32N(1000)
	}
}

func BenchmarkInt64N(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		benchInt64 = Int64N(1000)
	}
}

func BenchmarkRandomInt(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		benchInt = RandomInt(1, 1000)
	}
}

func BenchmarkRandomInt32(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		benchInt32 = RandomInt32(1, 1000)
	}
}

func BenchmarkRandomInt64(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		benchInt64 = RandomInt64(1, 1_000_000)
	}
}

func BenchmarkRandomUint(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		benchUint = RandomUint(0, 1000)
	}
}

func BenchmarkRandomUint32(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		benchUint32 = RandomUint32(0, 1000)
	}
}

func BenchmarkRandomUint64(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		benchUint64 = RandomUint64(0, 1_000_000)
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

func BenchmarkRandomString64(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		benchStr = RandomString(64)
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

func BenchmarkShuffle(b *testing.B) {
	b.ReportAllocs()
	arr := make([]int, 100)
	for i := range arr {
		arr[i] = i
	}
	for i := 0; i < b.N; i++ {
		Shuffle(arr)
	}
}

func BenchmarkWeightedChoice(b *testing.B) {
	b.ReportAllocs()
	weights := []int{1, 3, 5, 7, 9, 2, 4, 6, 8, 10}
	for i := 0; i < b.N; i++ {
		benchInt = WeightedChoice(weights)
	}
}

func BenchmarkNonWeightedChoice(b *testing.B) {
	b.ReportAllocs()
	weights := []int{1, 3, 5, 7, 9, 2, 4, 6, 8, 10}
	for i := 0; i < b.N; i++ {
		benchInt = NonWeightedChoice(weights)
	}
}

func BenchmarkSHA256Value(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		benchUint64 = SHA256Value("server-seed", "client-seed", uint64(i))
	}
}

// ─────────────────────────────────────────────
// Construction benchmarks
// ─────────────────────────────────────────────

func BenchmarkNewRandomizer(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		benchBool = NewRandomizer(PCGRandType, UnixNanoSeed) != nil
	}
}

func BenchmarkNewRandomizer_ChaCha8(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		benchBool = NewRandomizer(ChaCha8RandType, UnixNanoSeed) != nil
	}
}

func BenchmarkNewRandomizer_SHA256(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		benchBool = NewRandomizer(SHA256RandType, UnixNanoSeed) != nil
	}
}

func BenchmarkNewZipfRandomizer(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		benchBool = NewZipfRandomizer(UnixNanoSeed, 1.2, 1.0, 100) != nil
	}
}

// ─────────────────────────────────────────────
// Seeder
// ─────────────────────────────────────────────

func BenchmarkSeederSeed(b *testing.B) {
	cases := []struct {
		name     string
		seedType SeedType
	}{
		{"UnixNano", UnixNanoSeed},
		{"MapHash", MapHashSeed},
		{"CryptoRand", CryptoRandSeed},
		{"RandomString", RandomStringSeed},
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

// ─────────────────────────────────────────────
// Randomizer — primitives
// ─────────────────────────────────────────────

func BenchmarkRandomizerFloat32(b *testing.B) {
	b.ReportAllocs()
	r := NewRandomizer(PCGRandType, UnixNanoSeed)
	for i := 0; i < b.N; i++ {
		benchFloat32 = r.Float32()
	}
}

func BenchmarkRandomizerFloat64(b *testing.B) {
	b.ReportAllocs()
	r := NewRandomizer(PCGRandType, UnixNanoSeed)
	for i := 0; i < b.N; i++ {
		benchFloat64 = r.Float64()
	}
}

func BenchmarkRandomizerInt(b *testing.B) {
	b.ReportAllocs()
	r := NewRandomizer(PCGRandType, UnixNanoSeed)
	for i := 0; i < b.N; i++ {
		benchInt = r.Int()
	}
}

func BenchmarkRandomizerInt32(b *testing.B) {
	b.ReportAllocs()
	r := NewRandomizer(PCGRandType, UnixNanoSeed)
	for i := 0; i < b.N; i++ {
		benchInt32 = r.Int32()
	}
}

func BenchmarkRandomizerInt64(b *testing.B) {
	b.ReportAllocs()
	r := NewRandomizer(PCGRandType, UnixNanoSeed)
	for i := 0; i < b.N; i++ {
		benchInt64 = r.Int64()
	}
}

func BenchmarkRandomizerUint32(b *testing.B) {
	b.ReportAllocs()
	r := NewRandomizer(PCGRandType, UnixNanoSeed)
	for i := 0; i < b.N; i++ {
		benchUint32 = r.Uint32()
	}
}

func BenchmarkRandomizerUint64(b *testing.B) {
	b.ReportAllocs()
	r := NewRandomizer(PCGRandType, UnixNanoSeed)
	for i := 0; i < b.N; i++ {
		benchUint64 = r.Uint64()
	}
}

func BenchmarkRandomizerUint64_Zipf(b *testing.B) {
	b.ReportAllocs()
	r := NewZipfRandomizer(UnixNanoSeed, 1.2, 1.0, 100)
	for i := 0; i < b.N; i++ {
		benchUint64 = r.Uint64()
	}
}

func BenchmarkRandomizerIntN(b *testing.B) {
	b.ReportAllocs()
	r := NewRandomizer(PCGRandType, UnixNanoSeed)
	for i := 0; i < b.N; i++ {
		benchInt = r.IntN(1000)
	}
}

func BenchmarkRandomizerUint32N(b *testing.B) {
	b.ReportAllocs()
	r := NewRandomizer(PCGRandType, UnixNanoSeed)
	for i := 0; i < b.N; i++ {
		benchUint32 = r.Uint32N(1000)
	}
}

func BenchmarkRandomizerUint64N(b *testing.B) {
	b.ReportAllocs()
	r := NewRandomizer(PCGRandType, UnixNanoSeed)
	for i := 0; i < b.N; i++ {
		benchUint64 = r.Uint64N(1000)
	}
}

func BenchmarkRandomizerBool(b *testing.B) {
	b.ReportAllocs()
	r := NewRandomizer(PCGRandType, UnixNanoSeed)
	for i := 0; i < b.N; i++ {
		benchBool = r.Bool()
	}
}

func BenchmarkRandomizerSecureInt64(b *testing.B) {
	b.ReportAllocs()
	r := NewRandomizer(PCGRandType, UnixNanoSeed)
	for i := 0; i < b.N; i++ {
		benchInt64 = r.SecureInt64()
	}
}

// ─────────────────────────────────────────────
// Randomizer — range methods
// ───────────────────────────────���─────────────

func BenchmarkRandomizerRangeInt(b *testing.B) {
	b.ReportAllocs()
	r := NewRandomizer(PCGRandType, UnixNanoSeed)
	for i := 0; i < b.N; i++ {
		benchInt = r.RangeInt(1, 1000)
	}
}

func BenchmarkRandomizerRangeInt32(b *testing.B) {
	b.ReportAllocs()
	r := NewRandomizer(PCGRandType, UnixNanoSeed)
	for i := 0; i < b.N; i++ {
		benchInt32 = r.RangeInt32(1, 1000)
	}
}

func BenchmarkRandomizerRangeInt64(b *testing.B) {
	b.ReportAllocs()
	r := NewRandomizer(PCGRandType, UnixNanoSeed)
	for i := 0; i < b.N; i++ {
		benchInt64 = r.RangeInt64(1, 1_000_000)
	}
}

func BenchmarkRandomizerRangeUint(b *testing.B) {
	b.ReportAllocs()
	r := NewRandomizer(PCGRandType, UnixNanoSeed)
	for i := 0; i < b.N; i++ {
		benchUint = r.RangeUint(0, 1000)
	}
}

func BenchmarkRandomizerRangeUint32(b *testing.B) {
	b.ReportAllocs()
	r := NewRandomizer(PCGRandType, UnixNanoSeed)
	for i := 0; i < b.N; i++ {
		benchUint32 = r.RangeUint32(0, 1000)
	}
}

func BenchmarkRandomizerRangeUint64(b *testing.B) {
	b.ReportAllocs()
	r := NewRandomizer(PCGRandType, UnixNanoSeed)
	for i := 0; i < b.N; i++ {
		benchUint64 = r.RangeUint64(1, 1_000_000)
	}
}

func BenchmarkRandomizerRangeFloat32(b *testing.B) {
	b.ReportAllocs()
	r := NewRandomizer(PCGRandType, UnixNanoSeed)
	for i := 0; i < b.N; i++ {
		benchFloat32 = r.RangeFloat32(0.0, 1.0)
	}
}

func BenchmarkRandomizerRangeFloat64(b *testing.B) {
	b.ReportAllocs()
	r := NewRandomizer(PCGRandType, UnixNanoSeed)
	for i := 0; i < b.N; i++ {
		benchFloat64 = r.RangeFloat64(0.0, 1.0)
	}
}

// ─────────────────────────────────────────────
// Randomizer — string generation
// ─────────────────────────────────────────────

func BenchmarkRandomizerRandomString16(b *testing.B) {
	b.ReportAllocs()
	r := NewRandomizer(PCGRandType, UnixNanoSeed)
	for i := 0; i < b.N; i++ {
		benchStr = r.RandomString(16)
	}
}

func BenchmarkRandomizerRandomString64(b *testing.B) {
	b.ReportAllocs()
	r := NewRandomizer(PCGRandType, UnixNanoSeed)
	for i := 0; i < b.N; i++ {
		benchStr = r.RandomString(64)
	}
}

func BenchmarkRandomizerHex(b *testing.B) {
	b.ReportAllocs()
	r := NewRandomizer(PCGRandType, UnixNanoSeed)
	for i := 0; i < b.N; i++ {
		benchStr = r.Hex(32)
	}
}

func BenchmarkRandomizerLetterString(b *testing.B) {
	b.ReportAllocs()
	r := NewRandomizer(PCGRandType, UnixNanoSeed)
	for i := 0; i < b.N; i++ {
		benchStr = r.LetterString(16)
	}
}

func BenchmarkRandomizerUUID(b *testing.B) {
	b.ReportAllocs()
	r := NewRandomizer(PCGRandType, UnixNanoSeed)
	for i := 0; i < b.N; i++ {
		benchStr = r.UUID()
	}
}

// ─────────────────────────────────────────────
// Randomizer — bool / pick / shuffle
// ─────────────────────────────────────────────

func BenchmarkRandomizerWeightedBool(b *testing.B) {
	b.ReportAllocs()
	r := NewRandomizer(PCGRandType, UnixNanoSeed)
	for i := 0; i < b.N; i++ {
		benchBool = r.WeightedBool(3, 7)
	}
}

func BenchmarkRandomizerPick(b *testing.B) {
	b.ReportAllocs()
	r := NewRandomizer(PCGRandType, UnixNanoSeed)
	list := []any{1, "two", 3.0, true, nil}
	for i := 0; i < b.N; i++ {
		benchAny = r.Pick(list)
	}
}

func BenchmarkRandomizerPickString(b *testing.B) {
	b.ReportAllocs()
	r := NewRandomizer(PCGRandType, UnixNanoSeed)
	list := []string{"alpha", "beta", "gamma", "delta", "epsilon"}
	for i := 0; i < b.N; i++ {
		benchStr = r.PickString(list)
	}
}

func BenchmarkRandomizerPickInt(b *testing.B) {
	b.ReportAllocs()
	r := NewRandomizer(PCGRandType, UnixNanoSeed)
	list := []int{10, 20, 30, 40, 50}
	for i := 0; i < b.N; i++ {
		benchInt = r.PickInt(list)
	}
}

func BenchmarkRandomizerShuffle100(b *testing.B) {
	b.ReportAllocs()
	r := NewRandomizer(PCGRandType, UnixNanoSeed)
	arr := make([]int, 100)
	for i := range arr {
		arr[i] = i
	}
	for i := 0; i < b.N; i++ {
		r.Shuffle(len(arr), func(a, c int) { arr[a], arr[c] = arr[c], arr[a] })
	}
}

func BenchmarkRandomizerRandomPickUniqueStr(b *testing.B) {
	b.ReportAllocs()
	r := NewRandomizer(PCGRandType, UnixNanoSeed)
	list := []string{"a", "b", "c", "d", "e", "f", "g", "h", "i", "j"}
	for i := 0; i < b.N; i++ {
		_ = r.RandomPickUniqueStr(list, 5)
	}
}

func BenchmarkRandomizerRandomPickUniqueInt(b *testing.B) {
	b.ReportAllocs()
	r := NewRandomizer(PCGRandType, UnixNanoSeed)
	list := []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}
	for i := 0; i < b.N; i++ {
		_ = r.RandomPickUniqueInt(list, 5)
	}
}

// ─────────────────────────────────────────────
// Randomizer — weighted / distribution
// ─────────────────────────────────────────────

func BenchmarkRandomizerWeightedChoice(b *testing.B) {
	b.ReportAllocs()
	r := NewRandomizer(PCGRandType, UnixNanoSeed)
	weights := []int{1, 3, 5, 7, 9, 2, 4, 6, 8, 10}
	for i := 0; i < b.N; i++ {
		benchInt = r.WeightedChoice(weights)
	}
}

func BenchmarkRandomizerNonWeightedChoice(b *testing.B) {
	b.ReportAllocs()
	r := NewRandomizer(PCGRandType, UnixNanoSeed)
	weights := []int{1, 3, 5, 7, 9, 2, 4, 6, 8, 10}
	for i := 0; i < b.N; i++ {
		benchInt = r.NonWeightedChoice(weights)
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

func BenchmarkRandomizerNormal(b *testing.B) {
	b.ReportAllocs()
	r := NewRandomizer(PCGRandType, UnixNanoSeed)
	for i := 0; i < b.N; i++ {
		benchFloat64 = r.Normal(0, 1)
	}
}

func BenchmarkRandomizerExp(b *testing.B) {
	b.ReportAllocs()
	r := NewRandomizer(PCGRandType, UnixNanoSeed)
	for i := 0; i < b.N; i++ {
		benchFloat64 = r.Exp(1.5)
	}
}

func BenchmarkRandomizerPerm(b *testing.B) {
	b.ReportAllocs()
	r := NewRandomizer(PCGRandType, UnixNanoSeed)
	for i := 0; i < b.N; i++ {
		benchSlice = r.Perm(20)
	}
}

func BenchmarkRandomizerZipfUint64(b *testing.B) {
	b.ReportAllocs()
	r := NewZipfRandomizer(UnixNanoSeed, 1.2, 1.0, 100)
	for i := 0; i < b.N; i++ {
		benchUint64 = r.ZipfUint64()
	}
}

func BenchmarkRandomizerLogNormal(b *testing.B) {
	b.ReportAllocs()
	r := NewRandomizer(PCGRandType, UnixNanoSeed)
	for i := 0; i < b.N; i++ {
		benchFloat64 = r.LogNormal(0, 0.5)
	}
}

func BenchmarkRandomizerTruncatedNormal(b *testing.B) {
	b.ReportAllocs()
	r := NewRandomizer(PCGRandType, UnixNanoSeed)
	for i := 0; i < b.N; i++ {
		benchFloat64 = r.TruncatedNormal(100, 20, 50, 200)
	}
}

// ─────────────────────────────────────────────
// Randomizer — probability
// ─────────────────────────────────────────────

func BenchmarkRandomizerProbabilityHit(b *testing.B) {
	b.ReportAllocs()
	r := NewRandomizer(PCGRandType, UnixNanoSeed)
	for i := 0; i < b.N; i++ {
		benchBool = r.ProbabilityHit(30)
	}
}

func BenchmarkRandomizerProbPermilleHit(b *testing.B) {
	b.ReportAllocs()
	r := NewRandomizer(PCGRandType, UnixNanoSeed)
	for i := 0; i < b.N; i++ {
		benchBool = r.ProbPermilleHit(300)
	}
}

func BenchmarkRandomizerJitterDuration(b *testing.B) {
	b.ReportAllocs()
	r := NewRandomizer(PCGRandType, UnixNanoSeed)
	base := time.Second
	for i := 0; i < b.N; i++ {
		benchInt64 = int64(r.JitterDuration(base, 20))
	}
}

func BenchmarkRandomizerJitterInt(b *testing.B) {
	b.ReportAllocs()
	r := NewRandomizer(PCGRandType, UnixNanoSeed)
	for i := 0; i < b.N; i++ {
		benchInt = r.JitterInt(100, 10)
	}
}

func BenchmarkRandomizerGuaranteedProb(b *testing.B) {
	b.ReportAllocs()
	r := NewRandomizer(PCGRandType, UnixNanoSeed)
	for i := 0; i < b.N; i++ {
		benchBool = r.GuaranteedProb(10, 5, 50)
	}
}

func BenchmarkRandomizerProbWithCooldown(b *testing.B) {
	b.ReportAllocs()
	r := NewRandomizer(PCGRandType, UnixNanoSeed)
	for i := 0; i < b.N; i++ {
		benchBool = r.ProbWithCooldown(30, 3, 5)
	}
}

// ─────────────────────────────────────────────
// Randomizer — game domain
// ─────────────────────────────────────────────

func BenchmarkRandomizerD6(b *testing.B) {
	b.ReportAllocs()
	r := NewRandomizer(PCGRandType, UnixNanoSeed)
	for i := 0; i < b.N; i++ {
		benchInt = r.D6()
	}
}

func BenchmarkRandomizerD20(b *testing.B) {
	b.ReportAllocs()
	r := NewRandomizer(PCGRandType, UnixNanoSeed)
	for i := 0; i < b.N; i++ {
		benchInt = r.D20()
	}
}

func BenchmarkRandomizerD100(b *testing.B) {
	b.ReportAllocs()
	r := NewRandomizer(PCGRandType, UnixNanoSeed)
	for i := 0; i < b.N; i++ {
		benchInt = r.D100()
	}
}

func BenchmarkRandomizerCritHit(b *testing.B) {
	b.ReportAllocs()
	r := NewRandomizer(PCGRandType, UnixNanoSeed)
	for i := 0; i < b.N; i++ {
		benchBool = r.CritHit(15)
	}
}

func BenchmarkRandomizerCritHitWithLucky(b *testing.B) {
	b.ReportAllocs()
	r := NewRandomizer(PCGRandType, UnixNanoSeed)
	for i := 0; i < b.N; i++ {
		benchBool = r.CritHitWithLucky(10, 50, 0.1)
	}
}

func BenchmarkRandomizerRandomGrade(b *testing.B) {
	b.ReportAllocs()
	r := NewRandomizer(PCGRandType, UnixNanoSeed)
	weights := []int{50, 30, 15, 4, 1}
	for i := 0; i < b.N; i++ {
		benchInt = r.RandomGrade(weights)
	}
}

func BenchmarkRandomizerRandomTwoChoice(b *testing.B) {
	b.ReportAllocs()
	r := NewRandomizer(PCGRandType, UnixNanoSeed)
	for i := 0; i < b.N; i++ {
		benchBool = r.RandomTwoChoice(3, 7)
	}
}

func BenchmarkRandomizerRandomAttrValue(b *testing.B) {
	b.ReportAllocs()
	r := NewRandomizer(PCGRandType, UnixNanoSeed)
	for i := 0; i < b.N; i++ {
		benchInt = r.RandomAttrValue(100, 0.8, 1.2)
	}
}

func BenchmarkRandomizerRandomInterval(b *testing.B) {
	b.ReportAllocs()
	r := NewRandomizer(PCGRandType, UnixNanoSeed)
	for i := 0; i < b.N; i++ {
		benchInt = r.RandomInterval(30, 5)
	}
}

func BenchmarkRandomizerLuckyValue(b *testing.B) {
	b.ReportAllocs()
	r := NewRandomizer(PCGRandType, UnixNanoSeed)
	for i := 0; i < b.N; i++ {
		benchInt = r.LuckyValue()
	}
}

func BenchmarkRandomizerAttrAssign(b *testing.B) {
	b.ReportAllocs()
	r := NewRandomizer(PCGRandType, UnixNanoSeed)
	for i := 0; i < b.N; i++ {
		benchSlice = r.AttrAssign(100, 5)
	}
}

func BenchmarkRandomizerShrinkIntSlice(b *testing.B) {
	b.ReportAllocs()
	r := NewRandomizer(PCGRandType, UnixNanoSeed)
	arr := []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}
	for i := 0; i < b.N; i++ {
		benchSlice = r.ShrinkIntSlice(arr, 3)
	}
}

func BenchmarkRandomizerShuffleGrid(b *testing.B) {
	b.ReportAllocs()
	r := NewRandomizer(PCGRandType, UnixNanoSeed)
	for i := 0; i < b.N; i++ {
		_ = r.ShuffleGrid(8, 8)
	}
}

// ─────────────────────────────────────────────
// Randomizer — finance / statistics
// ─────────────────────────────────────────────

func BenchmarkRandomizerSlippagePrice(b *testing.B) {
	b.ReportAllocs()
	r := NewRandomizer(PCGRandType, UnixNanoSeed)
	for i := 0; i < b.N; i++ {
		benchFloat64 = r.SlippagePrice(100.0, 0.001)
	}
}

func BenchmarkRandomizerRandomPositionRate(b *testing.B) {
	b.ReportAllocs()
	r := NewRandomizer(PCGRandType, UnixNanoSeed)
	for i := 0; i < b.N; i++ {
		benchFloat64 = r.RandomPositionRate()
	}
}

func BenchmarkRandomizerRandomVolatility(b *testing.B) {
	b.ReportAllocs()
	r := NewRandomizer(PCGRandType, UnixNanoSeed)
	for i := 0; i < b.N; i++ {
		benchFloat64 = r.RandomVolatility(0.01, 0.05)
	}
}

func BenchmarkRandomizerRandomReturn(b *testing.B) {
	b.ReportAllocs()
	r := NewRandomizer(PCGRandType, UnixNanoSeed)
	for i := 0; i < b.N; i++ {
		benchFloat64 = r.RandomReturn(0.001, 0.02)
	}
}

func BenchmarkRandomizerRoundFloat(b *testing.B) {
	b.ReportAllocs()
	r := NewRandomizer(PCGRandType, UnixNanoSeed)
	for i := 0; i < b.N; i++ {
		benchFloat64 = r.RoundFloat(1.0, 100.0, 2)
	}
}

// ─────────────────────────────────────────────
// Randomizer — geometry
// ─────────────────────────────────────────────

func BenchmarkRandomizerRandomAngle(b *testing.B) {
	b.ReportAllocs()
	r := NewRandomizer(PCGRandType, UnixNanoSeed)
	for i := 0; i < b.N; i++ {
		benchFloat64 = r.RandomAngle()
	}
}

func BenchmarkRandomizerCircleRandomPoint(b *testing.B) {
	b.ReportAllocs()
	r := NewRandomizer(PCGRandType, UnixNanoSeed)
	var x, y float64
	for i := 0; i < b.N; i++ {
		x, y = r.CircleRandomPoint(0, 0, 10)
	}
	_, _ = x, y
}

func BenchmarkRandomizerRectRandomPoint(b *testing.B) {
	b.ReportAllocs()
	r := NewRandomizer(PCGRandType, UnixNanoSeed)
	var x, y float64
	for i := 0; i < b.N; i++ {
		x, y = r.RectRandomPoint(0, 0, 100, 100)
	}
	_, _ = x, y
}

// ─────────────────────────────────────────────
// Randomizer — fake / personal data
// ─────────────────────────────────────────────

func BenchmarkRandomizerPhoneNumber(b *testing.B) {
	b.ReportAllocs()
	r := NewRandomizer(PCGRandType, UnixNanoSeed)
	for i := 0; i < b.N; i++ {
		benchStr = r.PhoneNumber()
	}
}

func BenchmarkRandomizerEmail(b *testing.B) {
	b.ReportAllocs()
	r := NewRandomizer(PCGRandType, UnixNanoSeed)
	for i := 0; i < b.N; i++ {
		benchStr = r.Email()
	}
}

func BenchmarkRandomizerIDCard(b *testing.B) {
	b.ReportAllocs()
	r := NewRandomizer(PCGRandType, UnixNanoSeed)
	for i := 0; i < b.N; i++ {
		benchStr = r.IDCard()
	}
}

func BenchmarkRandomizerRandomColorHex(b *testing.B) {
	b.ReportAllocs()
	r := NewRandomizer(PCGRandType, UnixNanoSeed)
	for i := 0; i < b.N; i++ {
		benchStr = r.RandomColorHex()
	}
}

func BenchmarkRandomizerRandomIPv4(b *testing.B) {
	b.ReportAllocs()
	r := NewRandomizer(PCGRandType, UnixNanoSeed)
	for i := 0; i < b.N; i++ {
		benchStr = r.RandomIPv4()
	}
}

func BenchmarkRandomizerRandomIPv6(b *testing.B) {
	b.ReportAllocs()
	r := NewRandomizer(PCGRandType, UnixNanoSeed)
	for i := 0; i < b.N; i++ {
		benchStr = r.RandomIPv6()
	}
}

func BenchmarkRandomizerRandomLatLng(b *testing.B) {
	b.ReportAllocs()
	r := NewRandomizer(PCGRandType, UnixNanoSeed)
	var lat, lng float64
	for i := 0; i < b.N; i++ {
		lat, lng = r.RandomLatLng()
	}
	_, _ = lat, lng
}

func BenchmarkRandomizerRandomFilePath(b *testing.B) {
	b.ReportAllocs()
	r := NewRandomizer(PCGRandType, UnixNanoSeed)
	for i := 0; i < b.N; i++ {
		benchStr = r.RandomFilePath(false)
	}
}

// ─────────────────────────────────────────────
// Randomizer — RandType comparison
// ─────────────────────────────────────────────

func BenchmarkRandType_IntN(b *testing.B) {
	for _, tc := range []struct {
		name     string
		randType RandType
	}{
		{"PCG", PCGRandType},
		{"ChaCha8", ChaCha8RandType},
		{"SHA256", SHA256RandType},
		{"Zipf", ZipfRandType},
	} {
		b.Run(tc.name, func(b *testing.B) {
			r := NewRandomizer(tc.randType, UnixNanoSeed)
			b.ReportAllocs()
			for i := 0; i < b.N; i++ {
				benchInt = r.IntN(1000)
			}
		})
	}
}

func BenchmarkRandType_RandomString16(b *testing.B) {
	for _, tc := range []struct {
		name     string
		randType RandType
	}{
		{"PCG", PCGRandType},
		{"ChaCha8", ChaCha8RandType},
		{"SHA256", SHA256RandType},
	} {
		b.Run(tc.name, func(b *testing.B) {
			r := NewRandomizer(tc.randType, UnixNanoSeed)
			b.ReportAllocs()
			for i := 0; i < b.N; i++ {
				benchStr = r.RandomString(16)
			}
		})
	}
}

// ─────────────────────────────────────────────
// Sha256Source
// ─────────────────────────────────────────────

func BenchmarkSha256Source_Uint64(b *testing.B) {
	b.ReportAllocs()
	src := NewSha256Source([]byte("bench-seed"))
	for i := 0; i < b.N; i++ {
		benchUint64 = src.Uint64()
	}
}

func BenchmarkSHA256ValueVsRandomizer(b *testing.B) {
	b.Run("SHA256Value", func(b *testing.B) {
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			benchUint64 = SHA256Value("srv", "cli", uint64(i))
		}
	})
	b.Run("SHA256Randomizer_Uint64", func(b *testing.B) {
		b.ReportAllocs()
		r := NewRandomizer(SHA256RandType, UnixNanoSeed)
		for i := 0; i < b.N; i++ {
			benchUint64 = r.Uint64()
		}
	})
}

// ─────────────────────────────────────────────
// Performance comparison: Global functions vs defaultRandomizer
// ─────────────────────────────────────────────

// 对比全局函数 vs defaultRandomizer 的性能差异
// 说明设计意图：全局函数是轻量级的，Randomizer 是可配置的
func BenchmarkGlobalVsDefault_IntN(b *testing.B) {
	b.Run("Global_IntN", func(b *testing.B) {
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			benchInt = IntN(1000)
		}
	})
	b.Run("Default_IntN", func(b *testing.B) {
		b.ReportAllocs()
		d := Default()
		for i := 0; i < b.N; i++ {
			benchInt = d.IntN(1000)
		}
	})
}

func BenchmarkGlobalVsDefault_RandomString(b *testing.B) {
	b.Run("Global_RandomString16", func(b *testing.B) {
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			benchStr = RandomString(16)
		}
	})
	b.Run("Default_RandomString16", func(b *testing.B) {
		b.ReportAllocs()
		d := Default()
		for i := 0; i < b.N; i++ {
			benchStr = d.RandomString(16)
		}
	})
}

func BenchmarkGlobalVsDefault_WeightedChoice(b *testing.B) {
	weights := []int{1, 3, 5, 7, 9, 2, 4, 6, 8, 10}
	b.Run("Global_WeightedChoice", func(b *testing.B) {
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			benchInt = WeightedChoice(weights)
		}
	})
	b.Run("Default_WeightedChoice", func(b *testing.B) {
		b.ReportAllocs()
		d := Default()
		for i := 0; i < b.N; i++ {
			benchInt = d.WeightedChoice(weights)
		}
	})
}

func BenchmarkGlobalVsDefault_Shuffle(b *testing.B) {
	b.Run("Global_Shuffle100", func(b *testing.B) {
		b.ReportAllocs()
		arr := make([]int, 100)
		for i := range arr {
			arr[i] = i
		}
		for i := 0; i < b.N; i++ {
			Shuffle(arr)
		}
	})
	b.Run("Default_Shuffle100", func(b *testing.B) {
		b.ReportAllocs()
		d := Default()
		arr := make([]int, 100)
		for i := range arr {
			arr[i] = i
		}
		for i := 0; i < b.N; i++ {
			d.Shuffle(len(arr), func(a, c int) { arr[a], arr[c] = arr[c], arr[a] })
		}
	})
}
