package crypto

import (
	"crypto/sha256"
	"crypto/sha512"
)

// SHA256Sum 计算SHA-256哈希
func SHA256Sum(data []byte) []byte {
	h := sha256.New()
	h.Write(data)
	return h.Sum(nil)
}

// SHA512Sum 计算SHA-512哈希
func SHA512Sum(data []byte) []byte {
	h := sha512.New()
	h.Write(data)
	return h.Sum(nil)
}

type SHA256Hasher struct{}

func (h *SHA256Hasher) Sum(data []byte) ([]byte, error) {
	return SHA256Sum(data), nil
}
func (h *SHA256Hasher) Name() string { return "SHA256" }

type SHA512Hasher struct{}

func (h *SHA512Hasher) Sum(data []byte) ([]byte, error) {
	return SHA512Sum(data), nil
}
func (h *SHA512Hasher) Name() string { return "SHA512" }
