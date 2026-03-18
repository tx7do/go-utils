package crypto

import (
	"bytes"
	"fmt"

	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
)

// DefaultAESKey 默认AES密钥(16字节)
var DefaultAESKey = []byte("f51d66a73d8a0927")

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

// PKCS5Padding 填充明文
func PKCS5Padding(plaintext []byte, blockSize int) []byte {
	if blockSize <= 0 || blockSize > 255 {
		panic("blockSize must be in 1..255 for PKCS5Padding")
	}
	if len(plaintext) < 0 {
		panic("plaintext length invalid")
	}
	padding := blockSize - len(plaintext)%blockSize
	padText := bytes.Repeat([]byte{byte(padding)}, padding)
	return append(plaintext, padText...)
}

// PKCS5UnPadding 去除填充数据
func PKCS5UnPadding(origData []byte) []byte {
	length := len(origData)
	if length == 0 {
		return []byte{}
	}
	unpadding := int(origData[length-1])
	if unpadding == 0 || unpadding > length {
		return []byte{}
	}
	// 检查填充内容是否都等于 unpadding
	for i := length - unpadding; i < length; i++ {
		if int(origData[i]) != unpadding {
			return []byte{}
		}
	}
	return origData[:(length - unpadding)]
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
