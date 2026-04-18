package crypto

import (
	"encoding/hex"
	"testing"
)

func TestSM3Hasher_Sum(t *testing.T) {
	hasher := NewSM3Hasher()
	data := []byte("hello, sm3 hasher test!")
	hash := hasher.Sum(data)
	if len(hash) != 32 {
		t.Fatalf("SM3 hash length should be 32 bytes, got: %d", len(hash))
	}

	// 验证一致性
	hash2 := hasher.Sum(data)
	if hex.EncodeToString(hash) != hex.EncodeToString(hash2) {
		t.Fatalf("SM3 hash not deterministic, got: %x, want: %x", hash2, hash)
	}
}

func TestSM3Hasher_Empty(t *testing.T) {
	hasher := NewSM3Hasher()
	hash := hasher.Sum([]byte{})
	if len(hash) != 32 {
		t.Fatalf("SM3 hash of empty should be 32 bytes, got: %d", len(hash))
	}
}
