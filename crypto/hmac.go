package crypto

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
)

type HMAC struct {
	key []byte
}

// NewHMAC 创建HMAC实例
func NewHMAC(key []byte) *HMAC {
	return &HMAC{key: key}
}

// Sum 计算HMAC-SHA256
func (h *HMAC) Sum(data []byte) string {
	mac := hmac.New(sha256.New, h.key)
	mac.Write(data)
	return hex.EncodeToString(mac.Sum(nil))
}

// Verify 校验HMAC
func (h *HMAC) Verify(data []byte, expected string) bool {
	mac := hmac.New(sha256.New, h.key)
	mac.Write(data)
	actual := hex.EncodeToString(mac.Sum(nil))
	return hmac.Equal([]byte(actual), []byte(expected))
}

// SetKey 变更密钥
func (h *HMAC) SetKey(key []byte) {
	h.key = key
}
