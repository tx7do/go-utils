package math

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSign(t *testing.T) {
	assert.True(t, Sign(2) == 1)
	assert.True(t, Sign(-2) == -1)
	assert.True(t, Sign(0) == 0)

	assert.True(t, Sign(int64(2)) == 1)
	assert.True(t, Sign(int64(-2)) == -1)
	assert.True(t, Sign(int64(0)) == 0)

	assert.True(t, Sign(float32(2)) == 1)
	assert.True(t, Sign(float32(-2)) == -1)
	assert.True(t, Sign(float32(0)) == 0)

	assert.True(t, Sign(float64(2)) == 1)
	assert.True(t, Sign(float64(-2)) == -1)
	assert.True(t, Sign(float64(0)) == 0)
}

func TestStandardDeviation(t *testing.T) {
	assert.Equal(t, StandardDeviation([]float64{3, 5, 9, 1, 8, 6, 58, 9, 4, 10}), 15.8117045254457)
	assert.Equal(t, StandardDeviation([]float64{1, 3, 5, 7, 9, 11, 2, 4, 6, 8}), 3.0397368307141326)
}
