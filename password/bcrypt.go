package password

import (
	"errors"

	"golang.org/x/crypto/bcrypt"
)

type BCryptCrypto struct {
	cost int
}

func NewBCryptCrypto(cost ...int) *BCryptCrypto {
	useCost := bcrypt.DefaultCost
	if len(cost) > 0 && cost[0] >= bcrypt.MinCost && cost[0] <= bcrypt.MaxCost {
		useCost = cost[0]
	}

	return &BCryptCrypto{
		cost: useCost,
	}
}

// Encrypt 使用 bcrypt 加密密码，返回加密后的字符串和空盐值
func (b *BCryptCrypto) Encrypt(password string) (encrypted string, err error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), b.cost)
	if err != nil {
		return "", err
	}
	return string(hash), nil
}

// Verify 验证密码是否匹配加密后的字符串
func (b *BCryptCrypto) Verify(password, encrypted string) (bool, error) {
	err := bcrypt.CompareHashAndPassword([]byte(encrypted), []byte(password))
	if err != nil {
		if errors.Is(err, bcrypt.ErrMismatchedHashAndPassword) {
			return false, nil
		}
		return false, err
	}
	return true, nil
}
