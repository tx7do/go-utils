package crypto

import (
	"fmt"

	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"

	"encoding/base64"
	"encoding/hex"

	"golang.org/x/crypto/bcrypt"
)

// DefaultCost 最小值=4 最大值=31 默认值=10
var DefaultCost = 10

// HashPassword 加密密码
func HashPassword(password string) (string, error) {
	// Prefix + Cost + Salt + Hashed Text
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), DefaultCost)
	return string(bytes), err
}

// HashPasswordWithSalt 对密码进行加盐哈希处理
func HashPasswordWithSalt(password, salt string) (string, error) {
	// 将密码和盐组合
	combined := []byte(password + salt)

	// 计算哈希值
	hash := sha256.Sum256(combined)

	// 将哈希值转换为十六进制字符串
	return hex.EncodeToString(hash[:]), nil
}

// VerifyPassword 验证密码是否正确
func VerifyPassword(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}

// VerifyPasswordWithSalt 验证密码是否正确
func VerifyPasswordWithSalt(password, salt, hashedPassword string) bool {
	// 对输入的密码和盐进行哈希处理
	newHash, _ := HashPasswordWithSalt(password, salt)
	// 比较哈希值是否相同
	return newHash == hashedPassword
}

// GenerateSalt 生成指定长度的盐
func GenerateSalt(length int) (string, error) {
	salt := make([]byte, length)
	_, err := rand.Read(salt)
	if err != nil {
		return "", err
	}
	return hex.EncodeToString(salt), nil
}

// DecryptAES AES解密函数
func DecryptAES(cipherText, key string) (string, error) {
	// 将密文从Base64解码
	cipherData, err := base64.StdEncoding.DecodeString(cipherText)
	if err != nil {
		return "", err
	}

	// 创建AES块
	block, err := aes.NewCipher([]byte(key))
	if err != nil {
		return "", err
	}

	// 检查密文长度是否为块大小的倍数
	if len(cipherData)%aes.BlockSize != 0 {
		return "", fmt.Errorf("ciphertext is not a multiple of the block size")
	}

	// 创建CBC解密器
	iv := cipherData[:aes.BlockSize] // 提取IV
	cipherData = cipherData[aes.BlockSize:]
	mode := cipher.NewCBCDecrypter(block, iv)

	// 解密
	plainText := make([]byte, len(cipherData))
	mode.CryptBlocks(plainText, cipherData)

	// 去除填充
	plainText = pkcs7Unpad(plainText)

	return string(plainText), nil
}

// pkcs7Unpad PKCS7填充去除
func pkcs7Unpad(data []byte) []byte {
	length := len(data)
	padding := int(data[length-1])
	return data[:length-padding]
}
