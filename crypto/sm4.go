package crypto

import (
	"crypto/rand"
	"errors"

	"github.com/tjfoc/gmsm/sm4"
)

// SM4Cipher 实现国密SM4对称加密算法（ECB模式，适合小数据块）
type SM4Cipher struct {
	key []byte
}

// NewSM4Cipher 创建SM4加密器
func NewSM4Cipher(key []byte) (*SM4Cipher, error) {
	if len(key) != 16 {
		return nil, errors.New("SM4 key length must be 16 bytes")
	}
	return &SM4Cipher{key: key}, nil
}

// Encrypt ECB模式加密
func (s *SM4Cipher) Encrypt(plain []byte) ([]byte, error) {
	padded := PKCS5Padding(plain, 16)
	return sm4.Sm4Ecb(s.key, padded, true)
}

// Decrypt ECB模式解密
func (s *SM4Cipher) Decrypt(ciphertext []byte) ([]byte, error) {
	decrypted, err := sm4.Sm4Ecb(s.key, ciphertext, false)
	if err != nil {
		return nil, err
	}
	return PKCS5UnPadding(decrypted), nil
}

// GenerateSM4Key 生成随机SM4密钥
func GenerateSM4Key() ([]byte, error) {
	key := make([]byte, 16)
	_, err := rand.Read(key)
	return key, err
}

func (s *SM4Cipher) Name() string {
	return "SM4"
}
