package rand

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRandomInt(t *testing.T) {
	for i := 0; i < 1000; i++ {
		n := RandomInt(1, 100)
		fmt.Println(n)
		assert.True(t, n >= 1)
		assert.True(t, n < 100)
	}
}
