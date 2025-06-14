package password

import (
	"errors"
	"strings"
)

// Crypto 密码加解密接口
type Crypto interface {
	// Encrypt 加密密码，返回加密后的字符串（包含算法标识和盐值）
	Encrypt(plainPassword string) (encrypted string, err error)

	// Verify 验证密码是否匹配
	Verify(plainPassword, encrypted string) (bool, error)
}

func CreateCrypto(algorithm string) (Crypto, error) {
	algorithm = strings.ToLower(algorithm)
	switch algorithm {
	case "bcrypt":
		return NewBCryptCrypto(), nil
	case "pbkdf2":
		return NewPBKDF2Crypto(), nil
	case "argon2":
		return NewArgon2Crypto(), nil
	default:
		return nil, errors.New("不支持的加密算法")
	}
}
