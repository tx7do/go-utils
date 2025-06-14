package password

import (
	"crypto/rand"
	"crypto/subtle"
	"encoding/base64"
	"errors"
	"fmt"
	"strings"

	"golang.org/x/crypto/argon2"
)

// Argon2Crypto 实现 Argon2id 密码哈希算法
type Argon2Crypto struct {
	// 参数可配置，默认使用推荐值
	Memory      uint32
	Iterations  uint32
	Parallelism uint8
	SaltLength  uint32
	KeyLength   uint32
}

// NewArgon2Crypto 创建带默认参数的 Argon2 加密器
func NewArgon2Crypto() *Argon2Crypto {
	return &Argon2Crypto{
		Memory:      64 * 1024, // 64MB
		Iterations:  3,
		Parallelism: 2,
		SaltLength:  16,
		KeyLength:   32,
	}
}

// Encrypt 实现密码加密
func (a *Argon2Crypto) Encrypt(password string) (string, error) {
	// 生成随机盐值
	salt := make([]byte, a.SaltLength)
	if _, err := rand.Read(salt); err != nil {
		return "", err
	}

	// 生成哈希
	hash := argon2.IDKey(
		[]byte(password),
		salt,
		a.Iterations,
		a.Memory,
		a.Parallelism,
		a.KeyLength,
	)

	// 格式化输出（兼容标准格式）
	return fmt.Sprintf(
		"$argon2id$v=%d$m=%d,t=%d,p=%d$%s$%s",
		argon2.Version,
		a.Memory,
		a.Iterations,
		a.Parallelism,
		base64.RawStdEncoding.EncodeToString(salt),
		base64.RawStdEncoding.EncodeToString(hash),
	), nil
}

// Verify 验证密码
func (a *Argon2Crypto) Verify(password, encrypted string) (bool, error) {
	// 解析哈希字符串
	parts := strings.Split(encrypted, "$")
	if len(parts) != 6 {
		return false, errors.New("无效的 Argon2 哈希格式")
	}

	// 解析参数
	var version int
	var memory, iterations uint32
	var parallelism uint8
	_, err := fmt.Sscanf(parts[2], "v=%d", &version)
	if err != nil || version != argon2.Version {
		return false, errors.New("不支持的 Argon2 版本")
	}

	_, err = fmt.Sscanf(parts[3], "m=%d,t=%d,p=%d", &memory, &iterations, &parallelism)
	if err != nil {
		return false, errors.New("无效的 Argon2 参数")
	}

	// 解码盐值和哈希
	salt, err := base64.RawStdEncoding.DecodeString(parts[4])
	if err != nil {
		return false, err
	}

	decodedHash, err := base64.RawStdEncoding.DecodeString(parts[5])
	if err != nil {
		return false, err
	}
	keyLength := uint32(len(decodedHash))

	// 使用相同参数生成新哈希
	newHash := argon2.IDKey(
		[]byte(password),
		salt,
		iterations,
		memory,
		parallelism,
		keyLength,
	)

	// 安全比较
	return subtle.ConstantTimeCompare(newHash, decodedHash) == 1, nil
}
