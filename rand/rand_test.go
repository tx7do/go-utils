package rand

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFloat32_GeneratesValueWithinRange(t *testing.T) {
	for i := 0; i < 1000; i++ {
		value := Float32()
		assert.True(t, value >= 0.0)
		assert.True(t, value < 1.0)
	}
}

func TestFloat32_GeneratesDifferentValues(t *testing.T) {
	values := make(map[float32]bool)
	for i := 0; i < 1000; i++ {
		value := Float32()
		values[value] = true
	}
	assert.Greater(t, len(values), 1)
}

func TestFloat64_GeneratesValueWithinRange(t *testing.T) {
	for i := 0; i < 1000; i++ {
		value := Float64()
		assert.True(t, value >= 0.0)
		assert.True(t, value < 1.0)
	}
}

func TestFloat64_GeneratesDifferentValues(t *testing.T) {
	values := make(map[float64]bool)
	for i := 0; i < 1000; i++ {
		value := Float64()
		values[value] = true
	}
	assert.Greater(t, len(values), 1)
}

func TestRandomInt(t *testing.T) {
	for i := 0; i < 1000; i++ {
		n := RandomInt(1, 10)
		fmt.Println(n)
		assert.True(t, n >= 1)
		assert.True(t, n <= 100)
	}
}

func TestRandomInt_MinEqualsMax(t *testing.T) {
	n := RandomInt(5, 5)
	assert.Equal(t, 5, n)
}

func TestRandomInt_MinGreaterThanMax(t *testing.T) {
	n := RandomInt(10, 5)
	assert.Equal(t, 5, n)
}

func TestRandomInt_NegativeRange(t *testing.T) {
	for i := 0; i < 1000; i++ {
		n := RandomInt(-10, -1)
		assert.True(t, n >= -10)
		assert.True(t, n <= -1)
	}
}

func TestRandomInt_ZeroRange(t *testing.T) {
	for i := 0; i < 1000; i++ {
		n := RandomInt(0, 0)
		assert.Equal(t, 0, n)
	}
}

func TestShuffle_EmptyArray(t *testing.T) {
	var array []int
	Shuffle(array)
	assert.Equal(t, array, array)
}

func TestShuffle_SingleElementArray(t *testing.T) {
	array := []int{1}
	Shuffle(array)
	assert.Equal(t, []int{1}, array)
}

func TestShuffle_MultipleElementsArray(t *testing.T) {
	array := []int{1, 2, 3, 4, 5}
	original := make([]int, len(array))
	copy(original, array)
	Shuffle(array)
	assert.ElementsMatch(t, original, array)
}

func TestShuffle_ArrayWithDuplicates(t *testing.T) {
	array := []int{1, 2, 2, 3, 3, 3}
	original := make([]int, len(array))
	copy(original, array)
	Shuffle(array)
	fmt.Println(array)
	assert.ElementsMatch(t, original, array)
}

func TestRandomChoice_EmptyArray(t *testing.T) {
	result := RandomChoice([]int{}, 3)
	assert.Nil(t, result)
}

func TestRandomChoice_NegativeN(t *testing.T) {
	array := []int{1, 2, 3}
	result := RandomChoice(array, -1)
	assert.Nil(t, result)
}

func TestRandomChoice_ZeroN(t *testing.T) {
	array := []int{1, 2, 3}
	result := RandomChoice(array, 0)
	assert.Nil(t, result)
}

func TestRandomChoice_NGreaterThanArrayLength(t *testing.T) {
	array := []int{1, 2, 3}
	result := RandomChoice(array, 5)
	assert.ElementsMatch(t, array, result)
}

func TestRandomChoice_NEqualToArrayLength(t *testing.T) {
	array := []int{1, 2, 3}
	result := RandomChoice(array, 3)
	assert.ElementsMatch(t, array, result)
}

func TestRandomChoice_NLessThanArrayLength(t *testing.T) {
	array := []int{1, 2, 3, 4, 5}
	result := RandomChoice(array, 3)
	assert.Len(t, result, 3)
}

func TestRandomInt32_MinEqualsMax(t *testing.T) {
	result := RandomInt32(5, 5)
	assert.Equal(t, int32(5), result)
}

func TestRandomInt32_MinGreaterThanMax(t *testing.T) {
	result := RandomInt32(10, 5)
	assert.Equal(t, int32(5), result)
}

func TestRandomInt32_PositiveRange(t *testing.T) {
	for i := 0; i < 1000; i++ {
		result := RandomInt32(1, 10)
		assert.True(t, result >= 1)
		assert.True(t, result <= 10)
	}
}

func TestRandomInt32_NegativeRange(t *testing.T) {
	for i := 0; i < 1000; i++ {
		result := RandomInt32(-10, -1)
		assert.True(t, result >= -10)
		assert.True(t, result <= -1)
	}
}

func TestRandomInt32_ZeroRange(t *testing.T) {
	for i := 0; i < 1000; i++ {
		result := RandomInt32(0, 0)
		assert.Equal(t, int32(0), result)
	}
}

func TestRandomInt64_MinEqualsMax(t *testing.T) {
	result := RandomInt64(5, 5)
	assert.Equal(t, int64(5), result)
}

func TestRandomInt64_MinGreaterThanMax(t *testing.T) {
	result := RandomInt64(10, 5)
	assert.Equal(t, int64(5), result)
}

func TestRandomInt64_PositiveRange(t *testing.T) {
	for i := 0; i < 1000; i++ {
		result := RandomInt64(1, 10)
		assert.True(t, result >= 1)
		assert.True(t, result <= 10)
	}
}

func TestRandomInt64_NegativeRange(t *testing.T) {
	for i := 0; i < 1000; i++ {
		result := RandomInt64(-10, -1)
		assert.True(t, result >= -10)
		assert.True(t, result <= -1)
	}
}

func TestRandomInt64_ZeroRange(t *testing.T) {
	for i := 0; i < 1000; i++ {
		result := RandomInt64(0, 0)
		assert.Equal(t, int64(0), result)
	}
}

func TestRandomString_LengthZero(t *testing.T) {
	result := RandomString(0)
	assert.Equal(t, "", result)
}

func TestRandomString_PositiveLength(t *testing.T) {
	result := RandomString(10)
	assert.Len(t, result, 10)
}

func TestRandomString_ContainsOnlyValidCharacters(t *testing.T) {
	result := RandomString(100)
	for _, char := range result {
		assert.True(t, (char >= 'A' && char <= 'Z') || (char >= 'a' && char <= 'z') || (char >= '0' && char <= '9'))
	}
}

func TestRandomString_NegativeLength(t *testing.T) {
	result := RandomString(-5)
	assert.Equal(t, "", result)
}

func TestWeightedChoice_EmptyArray(t *testing.T) {
	result := WeightedChoice([]int{})
	assert.Equal(t, -1, result)
}

func TestWeightedChoice_AllZeroWeights(t *testing.T) {
	result := WeightedChoice([]int{0, 0, 0})
	assert.Equal(t, 2, result)
}

func TestWeightedChoice_NegativeWeights(t *testing.T) {
	result := WeightedChoice([]int{-1, -2, -3})
	assert.Equal(t, 2, result)
}

func TestWeightedChoice_MixedWeights(t *testing.T) {
	weightArray := []int{1, 0, 3, 0, 2}
	counts := make([]int, len(weightArray))
	for i := 0; i < 1000; i++ {
		choice := WeightedChoice(weightArray)
		counts[choice]++
	}
	assert.Greater(t, counts[0], 0)
	assert.Greater(t, counts[2], 0)
	assert.Greater(t, counts[4], 0)
}

func TestWeightedChoice_SingleElement(t *testing.T) {
	result := WeightedChoice([]int{5})
	assert.Equal(t, 0, result)
}

func TestNonWeightedChoice_EmptyArray(t *testing.T) {
	result := NonWeightedChoice([]int{})
	assert.Equal(t, -1, result)
}

func TestNonWeightedChoice_AllZeroWeights(t *testing.T) {
	result := NonWeightedChoice([]int{0, 0, 0})
	assert.Equal(t, 2, result)
}

func TestNonWeightedChoice_NegativeWeights(t *testing.T) {
	result := NonWeightedChoice([]int{-1, -2, -3})
	assert.Equal(t, 2, result)
}

func TestNonWeightedChoice_MixedWeights(t *testing.T) {
	weightArray := []int{1, 0, 3, 0, 2}
	counts := make([]int, len(weightArray))
	for i := 0; i < 1000; i++ {
		choice := NonWeightedChoice(weightArray)
		counts[choice]++
	}
	assert.Greater(t, counts[0], 0)
	assert.Greater(t, counts[2], 0)
	assert.Greater(t, counts[4], 0)
}

func TestNonWeightedChoice_SingleElement(t *testing.T) {
	result := NonWeightedChoice([]int{5})
	assert.Equal(t, 0, result)
}
