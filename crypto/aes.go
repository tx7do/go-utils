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
	padding := blockSize - len(plaintext)%blockSize
	padText := bytes.Repeat([]byte{byte(padding)}, padding)
	return append(plaintext, padText...)
}

// PKCS5UnPadding 去除填充数据
func PKCS5UnPadding(origData []byte) []byte {
	length := len(origData)
	unpadding := int(origData[length-1])
	return origData[:(length - unpadding)]
}

// AesEncrypt AES加密
func AesEncrypt(plainText, key, iv []byte) ([]byte, error) {
	if plainText == nil {
		return nil, fmt.Errorf("plain text is nil")
	}
	if key == nil {
		return nil, fmt.Errorf("key is nil")
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	// AES分组长度为128位，所以blockSize=16，单位字节
	blockSize := block.BlockSize()

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
	if cryptedText == nil {
		return nil, fmt.Errorf("crypted text is nil")
	}
	if key == nil {
		return nil, fmt.Errorf("key is nil")
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	//AES分组长度为128位，所以blockSize=16，单位字节
	blockSize := block.BlockSize()

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
