package rand

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSha256Source_Uint64_Deterministic(t *testing.T) {
	seed := []byte("test-seed-1234")
	s1 := NewSha256Source(seed)
	s2 := NewSha256Source(seed)

	// 相同种子相同调用顺序，结果应一致
	for i := 0; i < 20; i++ {
		assert.Equal(t, s1.Uint64(), s2.Uint64(), "第 %d 次调用结果不一致", i)
	}
}

func TestSha256Source_Uint64_NonceProgresses(t *testing.T) {
	seed := []byte("nonce-test")
	s := NewSha256Source(seed)

	// 每次调用结果不同（nonce 自增）
	seen := make(map[uint64]struct{}, 50)
	for i := 0; i < 50; i++ {
		v := s.Uint64()
		seen[v] = struct{}{}
	}
	assert.Equal(t, 50, len(seen), "nonce 自增应使每次结果不同")
}

func TestSha256Source_Uint64_DifferentSeedsProduceDifferentValues(t *testing.T) {
	s1 := NewSha256Source([]byte("seed-A"))
	s2 := NewSha256Source([]byte("seed-B"))
	assert.NotEqual(t, s1.Uint64(), s2.Uint64())
}

func TestSha256Source_EmptySeed(t *testing.T) {
	s := NewSha256Source([]byte{})
	assert.NotPanics(t, func() {
		v := s.Uint64()
		_ = v
	})
}

func TestSha256Source_NilSeed(t *testing.T) {
	s := NewSha256Source(nil)
	assert.NotPanics(t, func() {
		v := s.Uint64()
		_ = v
	})
}

func TestNewSha256Source_Initializes(t *testing.T) {
	seed := []byte("hello")
	s := NewSha256Source(seed)
	assert.NotNil(t, s)
}
