package rand

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestFloat_GenerateWithinRangeAndDiversity(t *testing.T) {
	tests := []struct {
		name string
		next func() float64
	}{
		{name: "float32", next: func() float64 { return float64(Float32()) }},
		{name: "float64", next: func() float64 { return Float64() }},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			values := make(map[float64]struct{}, 1000)
			for i := 0; i < 1000; i++ {
				v := tt.next()
				assert.GreaterOrEqual(t, v, 0.0)
				assert.Less(t, v, 1.0)
				values[v] = struct{}{}
			}
			assert.Greater(t, len(values), 1)
		})
	}
}

func TestRandomInt_Bounds(t *testing.T) {
	tests := []struct {
		name string
		min  int
		max  int
	}{
		{name: "positive", min: 1, max: 10},
		{name: "negative", min: -10, max: -1},
		{name: "equal", min: 5, max: 5},
		{name: "reversed", min: 10, max: 5},
		{name: "zero", min: 0, max: 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			for i := 0; i < 1000; i++ {
				v := RandomInt(tt.min, tt.max)
				if tt.min >= tt.max {
					assert.Equal(t, tt.max, v)
					continue
				}
				assert.GreaterOrEqual(t, v, tt.min)
				assert.LessOrEqual(t, v, tt.max)
			}
		})
	}
}

func TestRandomInt32_Bounds(t *testing.T) {
	tests := []struct {
		name string
		min  int32
		max  int32
	}{
		{name: "positive", min: 1, max: 10},
		{name: "negative", min: -10, max: -1},
		{name: "equal", min: 5, max: 5},
		{name: "reversed", min: 10, max: 5},
		{name: "zero", min: 0, max: 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			for i := 0; i < 1000; i++ {
				v := RandomInt32(tt.min, tt.max)
				if tt.min >= tt.max {
					assert.Equal(t, tt.max, v)
					continue
				}
				assert.GreaterOrEqual(t, v, tt.min)
				assert.LessOrEqual(t, v, tt.max)
			}
		})
	}
}

func TestRandomInt64_Bounds(t *testing.T) {
	tests := []struct {
		name string
		min  int64
		max  int64
	}{
		{name: "positive", min: 1, max: 10},
		{name: "negative", min: -10, max: -1},
		{name: "equal", min: 5, max: 5},
		{name: "reversed", min: 10, max: 5},
		{name: "zero", min: 0, max: 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			for i := 0; i < 1000; i++ {
				v := RandomInt64(tt.min, tt.max)
				if tt.min >= tt.max {
					assert.Equal(t, tt.max, v)
					continue
				}
				assert.GreaterOrEqual(t, v, tt.min)
				assert.LessOrEqual(t, v, tt.max)
			}
		})
	}
}

func TestRandomUint_Bounds(t *testing.T) {
	tests := []struct {
		name string
		min  uint
		max  uint
	}{
		{name: "positive", min: 1, max: 10},
		{name: "equal", min: 5, max: 5},
		{name: "reversed", min: 10, max: 5},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			for i := 0; i < 1000; i++ {
				v := RandomUint(tt.min, tt.max)
				if tt.min >= tt.max {
					assert.Equal(t, tt.max, v)
					continue
				}
				assert.GreaterOrEqual(t, v, tt.min)
				assert.LessOrEqual(t, v, tt.max)
			}
		})
	}
}

func TestRandomUint32_Bounds(t *testing.T) {
	tests := []struct {
		name string
		min  uint32
		max  uint32
	}{
		{name: "positive", min: 1, max: 10},
		{name: "equal", min: 5, max: 5},
		{name: "reversed", min: 10, max: 5},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			for i := 0; i < 1000; i++ {
				v := RandomUint32(tt.min, tt.max)
				if tt.min >= tt.max {
					assert.Equal(t, tt.max, v)
					continue
				}
				assert.GreaterOrEqual(t, v, tt.min)
				assert.LessOrEqual(t, v, tt.max)
			}
		})
	}
}

func TestRandomUint64_Bounds(t *testing.T) {
	tests := []struct {
		name string
		min  uint64
		max  uint64
	}{
		{name: "positive", min: 1, max: 10},
		{name: "equal", min: 5, max: 5},
		{name: "reversed", min: 10, max: 5},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			for i := 0; i < 1000; i++ {
				v := RandomUint64(tt.min, tt.max)
				if tt.min >= tt.max {
					assert.Equal(t, tt.max, v)
					continue
				}
				assert.GreaterOrEqual(t, v, tt.min)
				assert.LessOrEqual(t, v, tt.max)
			}
		})
	}
}

func TestRandomDuration_Bounds(t *testing.T) {
	tests := []struct {
		name string
		min  time.Duration
		max  time.Duration
	}{
		{name: "positive", min: 100 * time.Millisecond, max: 2 * time.Second},
		{name: "negative", min: -2 * time.Second, max: -100 * time.Millisecond},
		{name: "equal", min: time.Second, max: time.Second},
		{name: "reversed", min: 2 * time.Second, max: time.Second},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			for i := 0; i < 1000; i++ {
				v := RandomDuration(tt.min, tt.max)
				if tt.min >= tt.max {
					assert.Equal(t, tt.max, v)
					continue
				}
				assert.GreaterOrEqual(t, v, tt.min)
				assert.LessOrEqual(t, v, tt.max)
			}
		})
	}
}

func TestShuffle_Scenarios(t *testing.T) {
	tests := []struct {
		name  string
		array []int
	}{
		{name: "empty", array: nil},
		{name: "single", array: []int{1}},
		{name: "multiple", array: []int{1, 2, 3, 4, 5}},
		{name: "duplicates", array: []int{1, 2, 2, 3, 3, 3}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			arr := append([]int(nil), tt.array...)
			orig := append([]int(nil), tt.array...)
			Shuffle(arr)
			assert.ElementsMatch(t, orig, arr)
		})
	}
}

func TestRandomChoice_Scenarios(t *testing.T) {
	tests := []struct {
		name          string
		array         []int
		n             int
		expectNil     bool
		expectLen     int
		expectAllElem bool
	}{
		{name: "empty", array: []int{}, n: 3, expectNil: true},
		{name: "negative n", array: []int{1, 2, 3}, n: -1, expectNil: true},
		{name: "zero n", array: []int{1, 2, 3}, n: 0, expectNil: true},
		{name: "n greater", array: []int{1, 2, 3}, n: 5, expectLen: 3, expectAllElem: true},
		{name: "n equal", array: []int{1, 2, 3}, n: 3, expectLen: 3, expectAllElem: true},
		{name: "n less", array: []int{1, 2, 3, 4, 5}, n: 3, expectLen: 3},
		{name: "n one", array: []int{1, 2, 3, 4, 5}, n: 1, expectLen: 1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			original := append([]int(nil), tt.array...)
			result := RandomChoice(tt.array, tt.n)

			if tt.expectNil {
				assert.Nil(t, result)
			} else {
				assert.Len(t, result, tt.expectLen)
				for _, v := range result {
					assert.Contains(t, tt.array, v)
				}
				if tt.expectAllElem {
					assert.ElementsMatch(t, tt.array, result)
				}
			}

			assert.ElementsMatch(t, original, tt.array)
		})
	}
}

func TestRandomChoice_NoDuplicatesWhenInputUnique(t *testing.T) {
	array := []int{1, 2, 3, 4, 5}
	result := RandomChoice(array, 3)
	seen := make(map[int]struct{}, len(result))
	for _, v := range result {
		seen[v] = struct{}{}
	}
	assert.Len(t, seen, len(result))
}

func TestRandomString_Scenarios(t *testing.T) {
	tests := []struct {
		name   string
		length int
		want   int
	}{
		{name: "zero", length: 0, want: 0},
		{name: "negative", length: -5, want: 0},
		{name: "positive", length: 10, want: 10},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := RandomString(tt.length)
			assert.Len(t, result, tt.want)
			for _, char := range result {
				assert.True(t, (char >= 'A' && char <= 'Z') || (char >= 'a' && char <= 'z') || (char >= '0' && char <= '9'))
			}
		})
	}
}

func TestWeightedChoice_EdgeCases(t *testing.T) {
	tests := []struct {
		name    string
		weights []int
		want    int
	}{
		{name: "empty", weights: []int{}, want: -1},
		{name: "all zero", weights: []int{0, 0, 0}, want: 0},
		{name: "negative", weights: []int{-1, -2, -3}, want: 0},
		{name: "single positive", weights: []int{5}, want: 0},
		{name: "single zero", weights: []int{0}, want: 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, WeightedChoice(tt.weights))
		})
	}
}

func TestWeightedChoice_MixedWeights(t *testing.T) {
	weightArray := []int{1, 0, 3, 0, 2}
	counts := make([]int, len(weightArray))
	for i := 0; i < 1000; i++ {
		choice := WeightedChoice(weightArray)
		counts[choice]++
	}
	assert.Greater(t, counts[0], 0)
	assert.Equal(t, 0, counts[1])
	assert.Greater(t, counts[2], 0)
	assert.Equal(t, 0, counts[3])
	assert.Greater(t, counts[4], 0)
}

func TestNonWeightedChoice_EdgeCases(t *testing.T) {
	tests := []struct {
		name    string
		weights []int
		want    int
	}{
		{name: "nil", weights: nil, want: -1},
		{name: "empty", weights: []int{}, want: 0},
		{name: "all zero", weights: []int{0, 0, 0}, want: 0},
		{name: "negative", weights: []int{-1, -2, -3}, want: 0},
		{name: "single positive", weights: []int{5}, want: 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, NonWeightedChoice(tt.weights))
		})
	}
}

func TestNonWeightedChoice_MixedWeights(t *testing.T) {
	weightArray := []int{1, 0, 3, 0, 2}
	counts := make([]int, len(weightArray))
	for i := 0; i < 1000; i++ {
		choice := NonWeightedChoice(weightArray)
		counts[choice]++
	}
	assert.Greater(t, counts[0], 0)
	assert.Equal(t, 0, counts[1])
	assert.Greater(t, counts[2], 0)
	assert.Equal(t, 0, counts[3])
	assert.Greater(t, counts[4], 0)
}

func TestSHA256Value(t *testing.T) {
	tests := []struct {
		name       string
		serverSeed string
		clientSeed string
		nonce      uint64
	}{
		{name: "base", serverSeed: "server", clientSeed: "client", nonce: 1},
		{name: "nonce changed", serverSeed: "server", clientSeed: "client", nonce: 2},
		{name: "server changed", serverSeed: "server-a", clientSeed: "client", nonce: 1},
		{name: "client changed", serverSeed: "server", clientSeed: "client-b", nonce: 1},
	}

	vBase := SHA256Value(tests[0].serverSeed, tests[0].clientSeed, tests[0].nonce)
	vBaseAgain := SHA256Value(tests[0].serverSeed, tests[0].clientSeed, tests[0].nonce)
	assert.Equal(t, vBase, vBaseAgain)

	for _, tt := range tests[1:] {
		t.Run(tt.name, func(t *testing.T) {
			v := SHA256Value(tt.serverSeed, tt.clientSeed, tt.nonce)
			assert.NotEqual(t, vBase, v)
		})
	}
}
