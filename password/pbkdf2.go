package password

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"crypto/sha512"
	"encoding/base64"
	"encoding/binary"
	"errors"
	"fmt"
	"hash"
	"strconv"
	"strings"
)

// PBKDF2Crypto 实现 PBKDF2-HMAC 密码哈希算法
type PBKDF2Crypto struct {
	// 可配置参数，默认使用推荐值
	Iterations int
	KeyLength  int
	Hash       func() hash.Hash
	HashName   string
}

// NewPBKDF2Crypto 创建带默认参数的 PBKDF2 加密器 (SHA256)
func NewPBKDF2Crypto() *PBKDF2Crypto {
	return &PBKDF2Crypto{
		Iterations: 310000, // NIST 推荐最小值
		KeyLength:  32,     // 256-bit
		Hash:       sha256.New,
		HashName:   "sha256",
	}
}

// NewPBKDF2WithSHA512 创建使用 SHA512 的 PBKDF2 加密器
func NewPBKDF2WithSHA512() *PBKDF2Crypto {
	return &PBKDF2Crypto{
		Iterations: 600000, // SHA512 需要更多迭代
		KeyLength:  64,     // 512-bit
		Hash:       sha512.New,
		HashName:   "sha512",
	}
}

// Encrypt 实现密码加密
func (p *PBKDF2Crypto) Encrypt(password string) (string, error) {
	// 生成随机盐值 (16 bytes 推荐最小值)
	salt := make([]byte, 16)
	if _, err := rand.Read(salt); err != nil {
		return "", err
	}

	// 生成密钥
	key := pbkdf2Key([]byte(password), salt, p.Iterations, p.KeyLength, p.Hash)

	// 格式: pbkdf2:<hash>:<iterations>:<base64-salt>:<base64-key>
	return fmt.Sprintf(
		"pbkdf2:%s:%d:%s:%s",
		p.HashName,
		p.Iterations,
		base64.RawStdEncoding.EncodeToString(salt),
		base64.RawStdEncoding.EncodeToString(key),
	), nil
}

// Verify 验证密码
func (p *PBKDF2Crypto) Verify(password, encrypted string) (bool, error) {
	// 解析哈希字符串
	parts := strings.Split(encrypted, ":")
	if len(parts) != 5 || parts[0] != "pbkdf2" {
		return false, errors.New("无效的 PBKDF2 哈希格式")
	}

	// 解析参数
	hashName := parts[1]
	iterations, err := strconv.Atoi(parts[2])
	if err != nil {
		return false, errors.New("无效的迭代次数")
	}

	salt, err := base64.RawStdEncoding.DecodeString(parts[3])
	if err != nil {
		return false, err
	}

	expectedKey, err := base64.RawStdEncoding.DecodeString(parts[4])
	if err != nil {
		return false, err
	}

	// 根据哈希名称选择哈希函数
	hashFunc, ok := getHashFunction(hashName)
	if !ok {
		return false, fmt.Errorf("不支持的哈希算法: %s", hashName)
	}

	// 生成新密钥
	keyLength := len(expectedKey)
	newKey := pbkdf2Key([]byte(password), salt, iterations, keyLength, hashFunc)

	// 安全比较
	return hmac.Equal(newKey, expectedKey), nil
}

// pbkdf2Key 实现 PBKDF2 核心算法
func pbkdf2Key(password, salt []byte, iterations, keyLength int, hashFunc func() hash.Hash) []byte {
	prf := hmac.New(hashFunc, password)
	hashLength := prf.Size()
	blockCount := (keyLength + hashLength - 1) / hashLength

	output := make([]byte, 0, blockCount*hashLength)
	for i := 1; i <= blockCount; i++ {
		// U1 = PRF(password, salt || INT(i))
		prf.Reset()
		prf.Write(salt)
		binary.BigEndian.PutUint32(make([]byte, 4), uint32(i))
		prf.Write([]byte{byte(i >> 24), byte(i >> 16), byte(i >> 8), byte(i)})
		u := prf.Sum(nil)

		// F = U1 ⊕ U2 ⊕ ... ⊕ U_iterations
		f := make([]byte, len(u))
		copy(f, u)

		for j := 1; j < iterations; j++ {
			prf.Reset()
			prf.Write(u)
			u = prf.Sum(nil)
			for k := 0; k < len(f); k++ {
				f[k] ^= u[k]
			}
		}

		output = append(output, f...)
	}

	return output[:keyLength]
}

// getHashFunction 根据名称获取哈希函数
func getHashFunction(name string) (func() hash.Hash, bool) {
	switch name {
	case "sha256":
		return sha256.New, true
	case "sha512":
		return sha512.New, true
	default:
		return nil, false
	}
}
