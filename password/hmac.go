package password

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"errors"
)

// HMACCrypto 实现基于 HMAC 的加密和验证
type HMACCrypto struct {
	secretKey []byte
}

// NewHMACCrypto 创建一个新的 HMACCrypto 实例
func NewHMACCrypto(secretKey string) *HMACCrypto {
	return &HMACCrypto{
		secretKey: []byte(secretKey),
	}
}

// Encrypt 使用 HMAC-SHA256 对数据进行加密
func (h *HMACCrypto) Encrypt(data string) (string, error) {
	if data == "" {
		return "", errors.New("数据不能为空")
	}

	mac := hmac.New(sha256.New, h.secretKey)
	_, err := mac.Write([]byte(data))
	if err != nil {
		return "", err
	}

	hash := mac.Sum(nil)
	return hex.EncodeToString(hash), nil
}

// Verify 验证数据的 HMAC 值是否匹配
func (h *HMACCrypto) Verify(data, encrypted string) (bool, error) {
	if data == "" || encrypted == "" {
		return false, errors.New("数据或加密字符串不能为空")
	}

	expectedHash, err := h.Encrypt(data)
	if err != nil {
		return false, err
	}

	return hmac.Equal([]byte(expectedHash), []byte(encrypted)), nil
}
