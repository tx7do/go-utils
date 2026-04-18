package crypto

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"fmt"
)

// DefaultAESKey 默认AES密钥(16字节)
var DefaultAESKey = []byte("f51d66a73d8a0927")

type AESCipher struct {
	key []byte
	iv  []byte
}

func NewAESCipher(key, iv []byte) *AESCipher {
	if len(key) == 0 {
		key = DefaultAESKey
	}

	return &AESCipher{
		key: key,
		iv:  iv,
	}
}

func (a *AESCipher) Encrypt(plain []byte) ([]byte, error) {
	return AesEncrypt(plain, a.key, a.iv)
}

func (a *AESCipher) Decrypt(cipher []byte) ([]byte, error) {
	return AesDecrypt(cipher, a.key, a.iv)
}

func (a *AESCipher) Name() string {
	return "AES"
}

// GenerateAESKey 生成AES密钥
func GenerateAESKey(length int) ([]byte, error) {
	if length != 16 && length != 24 && length != 32 {
		return nil, fmt.Errorf("invalid key length: %d, must be 16, 24, or 32 bytes", length)
	}
	key := make([]byte, length)
	_, err := rand.Read(key)
	if err != nil {
		return nil, err
	}
	return key, nil
}

// AesEncrypt AES加密
func AesEncrypt(plainText, key, iv []byte) ([]byte, error) {
	if plainText == nil || len(plainText) == 0 {
		return nil, fmt.Errorf("plain text is nil or empty")
	}
	if key == nil {
		return nil, fmt.Errorf("key is nil")
	}
	if len(key) != 16 && len(key) != 24 && len(key) != 32 {
		return nil, fmt.Errorf("invalid key length: %d, must be 16, 24, or 32 bytes", len(key))
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	blockSize := block.BlockSize()

	if iv != nil && len(iv) != blockSize {
		return nil, fmt.Errorf("invalid iv length: %d, must be %d bytes", len(iv), blockSize)
	}
	if iv == nil {
		// 初始向量的长度必须等于块block的长度16字节
		iv = key[:blockSize]
	}

	plainText = PKCS5Padding(plainText, blockSize)

	blockMode := cipher.NewCBCEncrypter(block, iv)
	cryptedText := make([]byte, len(plainText))
	blockMode.CryptBlocks(cryptedText, plainText)
	return cryptedText, nil
}

// AesDecrypt AES解密
func AesDecrypt(cryptedText, key, iv []byte) ([]byte, error) {
	if cryptedText == nil || len(cryptedText) == 0 {
		return nil, fmt.Errorf("crypted text is nil or empty")
	}
	if key == nil {
		return nil, fmt.Errorf("key is nil")
	}
	if len(key) != 16 && len(key) != 24 && len(key) != 32 {
		return nil, fmt.Errorf("invalid key length: %d, must be 16, 24, or 32 bytes", len(key))
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	blockSize := block.BlockSize()

	if len(cryptedText)%blockSize != 0 {
		return nil, fmt.Errorf("invalid crypted text length: %d, must be multiple of block size %d", len(cryptedText), blockSize)
	}

	if iv != nil && len(iv) != blockSize {
		return nil, fmt.Errorf("invalid iv length: %d, must be %d bytes", len(iv), blockSize)
	}
	if iv == nil {
		// 初始向量的长度必须等于块block的长度16字节
		iv = key[:blockSize]
	}

	blockMode := cipher.NewCBCDecrypter(block, iv)

	plainText := make([]byte, len(cryptedText))
	blockMode.CryptBlocks(plainText, cryptedText)
	plainText = PKCS5UnPadding(plainText)
	return plainText, nil
}

type AESGCMCipher struct {
	key []byte
}

// NewAESGCMCipher 创建AES-GCM模式的Cipher
func NewAESGCMCipher(key []byte) (*AESGCMCipher, error) {
	if len(key) != 16 && len(key) != 24 && len(key) != 32 {
		return nil, fmt.Errorf("invalid key length: %d, must be 16, 24, or 32 bytes", len(key))
	}
	return &AESGCMCipher{key: key}, nil
}

// Encrypt 使用AES-GCM加密
func (a *AESGCMCipher) Encrypt(plain []byte) ([]byte, error) {
	block, err := aes.NewCipher(a.key)
	if err != nil {
		return nil, err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}
	// 生成随机nonce
	nonce := make([]byte, gcm.NonceSize())
	if _, err := rand.Read(nonce); err != nil {
		return nil, err
	}
	ciphertext := gcm.Seal(nil, nonce, plain, nil)
	// 输出格式: nonce|ciphertext
	return append(nonce, ciphertext...), nil
}

// Decrypt 使用AES-GCM解密
func (a *AESGCMCipher) Decrypt(ciphertext []byte) ([]byte, error) {
	block, err := aes.NewCipher(a.key)
	if err != nil {
		return nil, err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}
	if len(ciphertext) < gcm.NonceSize() {
		return nil, fmt.Errorf("ciphertext too short")
	}
	nonce := ciphertext[:gcm.NonceSize()]
	ct := ciphertext[gcm.NonceSize():]
	plain, err := gcm.Open(nil, nonce, ct, nil)
	if err != nil {
		return nil, err
	}
	return plain, nil
}

func (a *AESGCMCipher) Name() string {
	return "AES-GCM"
}
