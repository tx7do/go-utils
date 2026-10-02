package crypto

import (
	"encoding/hex"
	"testing"
)

func TestSM3Hasher_Sum(t *testing.T) {
	hasher := NewSM3Hasher()
	data := []byte("hello, sm3 hasher test!")
	hash, err := hasher.Sum(data)
	if err != nil {
		t.Fatalf("Sum error: %v", err)
	}
	if len(hash) != 32 {
		t.Fatalf("SM3 hash length should be 32 bytes, got: %d", len(hash))
	}

	// 验证一致性
	hash2, err := hasher.Sum(data)
	if err != nil {
		t.Fatalf("Sum error: %v", err)
	}
	if hex.EncodeToString(hash) != hex.EncodeToString(hash2) {
		t.Fatalf("SM3 hash not deterministic, got: %x, want: %x", hash2, hash)
	}
}

func TestSM3Hasher_Empty(t *testing.T) {
	hasher := NewSM3Hasher()
	hash, err := hasher.Sum([]byte{})
	if err != nil {
		t.Fatalf("Sum error: %v", err)
	}
	if len(hash) != 32 {
		t.Fatalf("SM3 hash of empty should be 32 bytes, got: %d", len(hash))
	}
}

// TestSM3Hasher_KnownVector 通过 Hasher 接口验证标准测试向量
// SM3("abc") = 66c7f0f462eeedd9d1f2d46bdc10e4e24167c4875cf2f7a2297da02b8f4ba8e0 (GB/T 32905-2016)
func TestSM3Hasher_KnownVector(t *testing.T) {
	var hasher Hasher = NewSM3Hasher() // 编译期验证接口兼容性
	hash, err := hasher.Sum([]byte("abc"))
	if err != nil {
		t.Fatalf("Sum error: %v", err)
	}
	want := "66c7f0f462eeedd9d1f2d46bdc10e4e24167c4875cf2f7a2297da02b8f4ba8e0"
	if got := hex.EncodeToString(hash); got != want {
		t.Fatalf("SM3 known vector mismatch, got: %s, want: %s", got, want)
	}
}

// TestSM3Hasher_Name 验证 Hasher 接口的 Name 方法
func TestSM3Hasher_Name(t *testing.T) {
	var hasher Hasher = NewSM3Hasher()
	if hasher.Name() != "SM3" {
		t.Fatalf("Name should be SM3, got: %s", hasher.Name())
	}
}
