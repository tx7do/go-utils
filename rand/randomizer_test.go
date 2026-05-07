package rand

import (
	"math"
	"testing"

	"github.com/stretchr/testify/assert"
)

func newTestRandomizer() *Randomizer {
	return NewRandomizer(PCGRandType, UnixNanoSeed)
}

func TestNewRandomizer_SupportedTypes(t *testing.T) {
	types := []RandType{PCGRandType, ChaCha8RandType, SHA256RandType}
	for _, typ := range types {
		t.Run(string(typ), func(t *testing.T) {
			r := NewRandomizer(typ, FixedSeed)
			if assert.NotNil(t, r, "type=%s", typ) {
				v := r.IntN(10)
				assert.GreaterOrEqual(t, v, 0)
				assert.Less(t, v, 10)
			}
		})
	}
}

func TestNewRandomizer_ZipfTypeReturnsNil(t *testing.T) {
	r := NewRandomizer(ZipfRandType, UnixNanoSeed)
	assert.Nil(t, r)
}

func TestNewRandomizer_DefaultFallback(t *testing.T) {
	r := NewRandomizer(RandType("UNKNOWN"), FixedSeed)
	assert.NotNil(t, r)
}

func TestRandomizer_RangeUint32_Boundaries(t *testing.T) {
	r := newTestRandomizer()
	tests := []struct {
		name string
		min  uint32
		max  uint32
	}{
		{name: "equal", min: 5, max: 5},
		{name: "reversed", min: 10, max: 5},
		{name: "positive", min: 1, max: 10},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			for i := 0; i < 1000; i++ {
				v := r.RangeUint32(tt.min, tt.max)
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

func TestRandomizer_RangeUint64_Boundaries(t *testing.T) {
	r := newTestRandomizer()
	tests := []struct {
		name string
		min  uint64
		max  uint64
	}{
		{name: "equal", min: 5, max: 5},
		{name: "reversed", min: 10, max: 5},
		{name: "positive", min: 1, max: 10},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			for i := 0; i < 1000; i++ {
				v := r.RangeUint64(tt.min, tt.max)
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

func TestRandomizer_RangeInt_Boundaries(t *testing.T) {
	r := newTestRandomizer()
	tests := []struct {
		name string
		min  int
		max  int
		want int
	}{
		{name: "reversed", min: 10, max: 5, want: 10},
		{name: "equal", min: 7, max: 7, want: 7},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, r.RangeInt(tt.min, tt.max))
		})
	}

	for i := 0; i < 1000; i++ {
		v := r.RangeInt(1, 10)
		assert.GreaterOrEqual(t, v, 1)
		assert.LessOrEqual(t, v, 10)
	}
}

func TestRandomizer_RangeInt32_Boundaries(t *testing.T) {
	r := newTestRandomizer()
	tests := []struct {
		name string
		min  int32
		max  int32
		want int32
	}{
		{name: "reversed", min: 10, max: 5, want: 5},
		{name: "equal", min: 7, max: 7, want: 7},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, r.RangeInt32(tt.min, tt.max))
		})
	}
}

func TestRandomizer_RangeInt64_Boundaries(t *testing.T) {
	r := newTestRandomizer()
	tests := []struct {
		name string
		min  int64
		max  int64
		want int64
	}{
		{name: "reversed", min: 10, max: 5, want: 5},
		{name: "equal", min: 7, max: 7, want: 7},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, r.RangeInt64(tt.min, tt.max))
		})
	}
}

func TestRandomizer_RangeUint_Boundaries(t *testing.T) {
	r := newTestRandomizer()
	tests := []struct {
		name string
		min  uint
		max  uint
		want uint
	}{
		{name: "reversed", min: 10, max: 5, want: 5},
		{name: "equal", min: 7, max: 7, want: 7},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, r.RangeUint(tt.min, tt.max))
		})
	}
}

func TestRandomizer_RangeFloat32_Boundaries(t *testing.T) {
	r := newTestRandomizer()
	tests := []struct {
		name string
		min  float32
		max  float32
		want float32
	}{
		{name: "reversed", min: 10, max: 5, want: 5},
		{name: "equal", min: 7, max: 7, want: 7},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, r.RangeFloat32(tt.min, tt.max))
		})
	}

	for i := 0; i < 1000; i++ {
		v := r.RangeFloat32(1.5, 3.5)
		assert.GreaterOrEqual(t, v, float32(1.5))
		assert.Less(t, v, float32(3.5))
	}
}

func TestRandomizer_RangeFloat64_Boundaries(t *testing.T) {
	r := newTestRandomizer()
	tests := []struct {
		name string
		min  float64
		max  float64
		want float64
	}{
		{name: "reversed", min: 10, max: 5, want: 5},
		{name: "equal", min: 7, max: 7, want: 7},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, r.RangeFloat64(tt.min, tt.max))
		})
	}

	for i := 0; i < 1000; i++ {
		v := r.RangeFloat64(1.5, 3.5)
		assert.GreaterOrEqual(t, v, float64(1.5))
		assert.Less(t, v, float64(3.5))
	}
}

func TestRandomizer_WeightedChoice_EdgeCases(t *testing.T) {
	r := newTestRandomizer()
	tests := []struct {
		name    string
		weights []int
		want    int
	}{
		{name: "empty", weights: []int{}, want: -1},
		{name: "all zero", weights: []int{0, 0, 0}, want: -1},
		{name: "single positive", weights: []int{1}, want: 0},
		{name: "single zero", weights: []int{0}, want: -1},
		{name: "single negative", weights: []int{-1}, want: -1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, r.WeightedChoice(tt.weights))
		})
	}
}

func TestRandomizer_WeightedChoice_MixedWeights(t *testing.T) {
	r := newTestRandomizer()
	weightArray := []int{1, 0, 3, 0, 2}
	counts := make([]int, len(weightArray))
	for i := 0; i < 1000; i++ {
		choice := r.WeightedChoice(weightArray)
		counts[choice]++
	}
	assert.Greater(t, counts[0], 0)
	assert.Greater(t, counts[2], 0)
	assert.Greater(t, counts[4], 0)
}

func TestRandomizer_RandomString(t *testing.T) {
	r := newTestRandomizer()
	tests := []struct {
		name   string
		length int
		want   int
	}{
		{name: "zero", length: 0, want: 0},
		{name: "positive", length: 50, want: 50},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := r.RandomString(tt.length)
			assert.Equal(t, tt.want, len(result))
		})
	}
}

func TestRandomizer_WeightedChoice_SingleElement(t *testing.T) {
	r := newTestRandomizer()
	assert.Equal(t, 0, r.WeightedChoice([]int{1}))
	assert.Equal(t, -1, r.WeightedChoice([]int{0}))
	assert.Equal(t, -1, r.WeightedChoice([]int{-1}))
}

func TestRandomizer_NonWeightedChoice_EdgeCases(t *testing.T) {
	r := newTestRandomizer()
	tests := []struct {
		name    string
		weights []int
		want    int
	}{
		{name: "nil", weights: nil, want: -1},
		{name: "empty", weights: []int{}, want: -1},
		{name: "all zero", weights: []int{0, 0, 0}, want: 0},
		{name: "negative", weights: []int{-1, -2, -3}, want: 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, r.NonWeightedChoice(tt.weights))
		})
	}
}

func TestRandomizer_NormalAndExp(t *testing.T) {
	r := newTestRandomizer()

	for i := 0; i < 1000; i++ {
		v := r.Exp(1.5)
		assert.GreaterOrEqual(t, v, 0.0)
	}

	// 正态分布值应为有限数
	v := r.Normal(10, 2)
	assert.False(t, math.IsNaN(v))
	assert.False(t, math.IsInf(v, 0))
}

func TestRandomizer_Perm(t *testing.T) {
	r := newTestRandomizer()
	p := r.Perm(10)
	assert.Len(t, p, 10)

	seen := make(map[int]struct{}, 10)
	for _, v := range p {
		assert.GreaterOrEqual(t, v, 0)
		assert.Less(t, v, 10)
		seen[v] = struct{}{}
	}
	assert.Len(t, seen, 10)
}

func TestNewZipfRandomizer_AndZipfUint64(t *testing.T) {
	r := NewZipfRandomizer(FixedSeed, 1.2, 1.0, 100)
	if assert.NotNil(t, r) {
		for i := 0; i < 1000; i++ {
			v := r.ZipfUint64()
			assert.LessOrEqual(t, v, uint64(100))
		}
	}
}

func TestRandomizer_ZipfUint64_Fallback(t *testing.T) {
	r := newTestRandomizer()
	v := r.ZipfUint64()
	assert.GreaterOrEqual(t, v, uint64(0))
}

func TestRandomizer_AliasTableAndChoice(t *testing.T) {
	r := newTestRandomizer()
	assert.Nil(t, r.NewAliasTable(nil))
	assert.Nil(t, r.NewAliasTable([]int{}))

	weights := []int{1, 3, 5, 7, 9}
	at := r.NewAliasTable(weights)
	if assert.NotNil(t, at) {
		assert.Len(t, at.prob, len(weights))
		assert.Len(t, at.alias, len(weights))

		for i := 0; i < 1000; i++ {
			idx := r.AliasChoice(at)
			assert.GreaterOrEqual(t, idx, 0)
			assert.Less(t, idx, len(weights))
		}
	}

	assert.Equal(t, -1, r.AliasChoice(nil))
	assert.Equal(t, -1, r.AliasChoice(&AliasTable{}))
}
