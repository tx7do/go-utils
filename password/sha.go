package password

import (
	"errors"
	"fmt"
	"hash"
	"strings"

	"crypto/rand"
	"crypto/sha256"
	"crypto/sha512"

	"encoding/base64"
	"encoding/hex"
)

// SHACrypto 实现 SHA-256/SHA-512 密码哈希算法（带盐值）
type SHACrypto struct {
	Hash       func() hash.Hash
	HashName   string
	SaltLength int
}

// NewSHA256Crypto 创建 SHA-256 加密器
func NewSHA256Crypto() *SHACrypto {
	return &SHACrypto{
		Hash:       sha256.New,
		HashName:   "sha256",
		SaltLength: 16, // 16 字节盐值
	}
}

// NewSHA512Crypto 创建 SHA-512 加密器
func NewSHA512Crypto() *SHACrypto {
	return &SHACrypto{
		Hash:       sha512.New,
		HashName:   "sha512",
		SaltLength: 16, // 16 字节盐值
	}
}

// Encrypt 实现密码加密（带盐值）
func (s *SHACrypto) Encrypt(password string) (string, error) {
	// 生成随机盐值
	salt := make([]byte, s.SaltLength)
	if _, err := rand.Read(salt); err != nil {
		return "", err
	}

	// 计算哈希
	hashValue := s.Hash()
	hashValue.Write(salt)
	hashValue.Write([]byte(password))
	hashBytes := hashValue.Sum(nil)

	// 格式: sha256:$salt$hash 或 sha512:$salt$hash
	return fmt.Sprintf(
		"%s$%s$%s",
		s.HashName,
		base64.RawStdEncoding.EncodeToString(salt),
		hex.EncodeToString(hashBytes),
	), nil
}

// Verify 验证密码
func (s *SHACrypto) Verify(password, encrypted string) (bool, error) {
	// 解析哈希字符串
	parts := strings.Split(encrypted, "$")
	if len(parts) != 3 {
		return false, errors.New("无效的 SHA 哈希格式")
	}

	hashName := parts[0]
	if hashName != s.HashName {
		return false, fmt.Errorf("哈希算法不匹配: 期望 %s, 实际 %s", s.HashName, hashName)
	}

	// 解码盐值
	salt, err := base64.RawStdEncoding.DecodeString(parts[1])
	if err != nil {
		return false, err
	}

	// 解码原始哈希值
	originalHash, err := hex.DecodeString(parts[2])
	if err != nil {
		return false, err
	}

	// 计算新哈希
	hashValue := s.Hash()
	hashValue.Write(salt)
	hashValue.Write([]byte(password))
	newHash := hashValue.Sum(nil)

	// 安全比较
	return compareHash(newHash, originalHash), nil
}

// compareHash 安全比较两个哈希值
func compareHash(h1, h2 []byte) bool {
	if len(h1) != len(h2) {
		return false
	}
	result := 0
	for i := 0; i < len(h1); i++ {
		result |= int(h1[i] ^ h2[i])
	}
	return result == 0
}
