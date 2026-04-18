package crypto

import (
	"encoding/hex"
	"testing"
)

func TestHMAC_SumAndVerify(t *testing.T) {
	key := []byte("supersecretkey")
	h := NewHMAC(key)
	data := []byte("hello, hmac test!")
	hash := h.Sum(data)
	if len(hash) != 64 {
		t.Fatalf("HMAC-SHA256 hex length should be 64, got: %d", len(hash))
	}
	if !h.Verify(data, hash) {
		t.Fatalf("HMAC verify failed, hash: %s", hash)
	}
}

func TestHMAC_Consistency(t *testing.T) {
	key := []byte("anotherkey")
	h := NewHMAC(key)
	data := []byte("consistency test")
	hash1 := h.Sum(data)
	hash2 := h.Sum(data)
	if hash1 != hash2 {
		t.Fatalf("HMAC not deterministic, got: %s, want: %s", hash2, hash1)
	}
}

func TestHMAC_SetKey(t *testing.T) {
	key1 := []byte("key1")
	key2 := []byte("key2")
	h := NewHMAC(key1)
	data := []byte("key switch test")
	hash1 := h.Sum(data)
	h.SetKey(key2)
	hash2 := h.Sum(data)
	if hash1 == hash2 {
		t.Fatalf("HMAC hash should change after key update")
	}
}

func TestHMAC_VerifyFail(t *testing.T) {
	key := []byte("failkey")
	h := NewHMAC(key)
	data := []byte("fail test")
	wrongHash := hex.EncodeToString([]byte("wrong"))
	if h.Verify(data, wrongHash) {
		t.Fatalf("HMAC verify should fail with wrong hash")
	}
}
