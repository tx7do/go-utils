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
func AesEncrypt(origData, key []byte, iv []byte) ([]byte, error) {
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

	origData = PKCS5Padding(origData, blockSize)
	blockMode := cipher.NewCBCEncrypter(block, iv)
	crypted := make([]byte, len(origData))
	blockMode.CryptBlocks(crypted, origData)
	return crypted, nil
}

// AesDecrypt AES解密
func AesDecrypt(crypted, key []byte, iv []byte) ([]byte, error) {
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
	origData := make([]byte, len(crypted))
	blockMode.CryptBlocks(origData, crypted)
	origData = PKCS5UnPadding(origData)
	return origData, nil
}
