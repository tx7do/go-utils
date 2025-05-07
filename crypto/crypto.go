package crypto

import (
	"crypto/rand"
	"crypto/sha256"

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
