package rand

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestRandomizer_RangeUint32_MinEqualsMax(t *testing.T) {
	r := NewRandomizer(UnixNanoSeed)
	result := r.RangeUint32(5, 5)
	assert.Equal(t, uint32(5), result)
}

func TestRandomizer_RangeUint32_MinGreaterThanMax(t *testing.T) {
	r := NewRandomizer(UnixNanoSeed)
	result := r.RangeUint32(10, 5)
	assert.Equal(t, uint32(5), result)
}

func TestRandomizer_RangeUint32_PositiveRange(t *testing.T) {
	r := NewRandomizer(UnixNanoSeed)
	for i := 0; i < 1000; i++ {
		result := r.RangeUint32(1, 10)
		assert.True(t, result >= 1)
		assert.True(t, result <= 10)
	}
}

func TestRandomizer_RangeUint64_MinEqualsMax(t *testing.T) {
	r := NewRandomizer(UnixNanoSeed)
	result := r.RangeUint64(5, 5)
	assert.Equal(t, uint64(5), result)
}

func TestRandomizer_RangeUint64_MinGreaterThanMax(t *testing.T) {
	r := NewRandomizer(UnixNanoSeed)
	result := r.RangeUint64(10, 5)
	assert.Equal(t, uint64(5), result)
}

func TestRandomizer_RangeUint64_PositiveRange(t *testing.T) {
	r := NewRandomizer(UnixNanoSeed)
	for i := 0; i < 1000; i++ {
		result := r.RangeUint64(1, 10)
		assert.True(t, result >= 1)
		assert.True(t, result <= 10)
	}
}

func TestRandomizer_WeightedChoice_EmptyArray(t *testing.T) {
	r := NewRandomizer(UnixNanoSeed)
	result := r.WeightedChoice([]int{})
	assert.Equal(t, -1, result)
}

func TestRandomizer_WeightedChoice_AllZeroWeights(t *testing.T) {
	r := NewRandomizer(UnixNanoSeed)
	result := r.WeightedChoice([]int{0, 0, 0})
	assert.Equal(t, 2, result)
}

func TestRandomizer_WeightedChoice_MixedWeights(t *testing.T) {
	r := NewRandomizer(UnixNanoSeed)
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

func TestRandomizer_RandomString_LengthZero(t *testing.T) {
	r := NewRandomizer(UnixNanoSeed)
	result := r.RandomString(0)
	assert.Equal(t, "", result)
}

func TestRandomizer_RandomString_CorrectLength(t *testing.T) {
	r := NewRandomizer(UnixNanoSeed)
	length := 50
	result := r.RandomString(length)
	assert.Equal(t, length, len(result))
}
