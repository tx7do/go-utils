package password

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"math/big"
	"strings"
)

// ECDSACrypto 实现基于 ECDSA 的加密和验证
type ECDSACrypto struct {
	privateKey *ecdsa.PrivateKey
	publicKey  *ecdsa.PublicKey
}

// NewECDSACrypto 创建一个新的 ECDSACrypto 实例
func NewECDSACrypto() (*ECDSACrypto, error) {
	privateKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		return nil, err
	}
	return &ECDSACrypto{
		privateKey: privateKey,
		publicKey:  &privateKey.PublicKey,
	}, nil
}

// Encrypt 使用 ECDSA 对消息进行签名
func (e *ECDSACrypto) Encrypt(plainPassword string) (string, error) {
	if plainPassword == "" {
		return "", errors.New("密码不能为空")
	}

	hash := sha256.Sum256([]byte(plainPassword))
	r, s, err := ecdsa.Sign(rand.Reader, e.privateKey, hash[:])
	if err != nil {
		return "", err
	}

	signature := r.String() + "$" + s.String()
	return "ecdsa$" + signature, nil
}

// Verify 验证消息的签名是否有效
func (e *ECDSACrypto) Verify(plainPassword, encrypted string) (bool, error) {
	if plainPassword == "" || encrypted == "" {
		return false, errors.New("密码或加密字符串不能为空")
	}

	parts := strings.SplitN(encrypted, "$", 3)
	if len(parts) != 3 || parts[0] != "ecdsa" {
		return false, errors.New("加密字符串格式无效")
	}

	r := new(big.Int)
	s := new(big.Int)
	r.SetString(parts[1], 10)
	s.SetString(parts[2], 10)

	hash := sha256.Sum256([]byte(plainPassword))
	return ecdsa.Verify(e.publicKey, hash[:], r, s), nil
}

// ECDHCrypto 实现基于 ECDH 的密钥交换
type ECDHCrypto struct {
	privateKey *ecdsa.PrivateKey
	publicKey  *ecdsa.PublicKey
}

// NewECDHCrypto 创建一个新的 ECDHCrypto 实例
func NewECDHCrypto() (*ECDHCrypto, error) {
	privateKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		return nil, err
	}
	return &ECDHCrypto{
		privateKey: privateKey,
		publicKey:  &privateKey.PublicKey,
	}, nil
}

// Encrypt 返回公钥作为加密结果
func (e *ECDHCrypto) Encrypt(plainPassword string) (string, error) {
	if plainPassword == "" {
		return "", errors.New("密码不能为空")
	}

	publicKeyBytes := elliptic.Marshal(e.privateKey.Curve, e.publicKey.X, e.publicKey.Y)
	return "ecdh$" + base64.StdEncoding.EncodeToString(publicKeyBytes), nil
}

// Verify 验证共享密钥是否一致
func (e *ECDHCrypto) Verify(plainPassword, encrypted string) (bool, error) {
	if plainPassword == "" || encrypted == "" {
		return false, errors.New("密码或加密字符串不能为空")
	}

	parts := strings.SplitN(encrypted, "$", 2)
	if len(parts) != 2 || parts[0] != "ecdh" {
		return false, errors.New("加密字符串格式无效")
	}

	publicKeyBytes, err := base64.StdEncoding.DecodeString(parts[1])
	if err != nil {
		return false, err
	}

	x, y := elliptic.Unmarshal(e.privateKey.Curve, publicKeyBytes)
	if x == nil || y == nil {
		return false, errors.New("无效的公钥")
	}

	sharedX, _ := e.privateKey.Curve.ScalarMult(x, y, e.privateKey.D.Bytes())
	expectedHash := sha256.Sum256(sharedX.Bytes())
	actualHash := sha256.Sum256([]byte(plainPassword))

	return expectedHash == actualHash, nil
}

func (e *ECDHCrypto) DeriveSharedSecret(publicKey string) ([]byte, error) {
	// 解码对方的公钥
	publicKeyBytes, err := base64.StdEncoding.DecodeString(publicKey)
	if err != nil {
		return nil, err
	}

	// 反序列化公钥
	x, y := elliptic.Unmarshal(e.privateKey.Curve, publicKeyBytes)
	if x == nil || y == nil {
		return nil, errors.New("无效的公钥")
	}

	// 计算共享密钥
	sharedX, _ := e.privateKey.Curve.ScalarMult(x, y, e.privateKey.D.Bytes())

	// 返回共享密钥的字节表示
	return sharedX.Bytes(), nil
}
