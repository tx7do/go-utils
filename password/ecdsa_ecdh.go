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

	r, ok := new(big.Int).SetString(parts[1], 10)
	if !ok {
		return false, errors.New("签名 r 值无效")
	}
	s, ok := new(big.Int).SetString(parts[2], 10)
	if !ok {
		return false, errors.New("签名 s 值无效")
	}

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

// Verify 校验对端（Encrypt 返回的）ECDH 公钥有效，并确认能与其推导出共享密钥。
// plainPassword 仅作非空确认位，不参与密钥推导；成功返回 true 表示握手可继续，
// 后续共享密钥请使用 DeriveSharedSecret（双方对该结果一致）。
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

	// 对端公钥必须落在本方曲线上，且能完成一次标量乘（推导共享密钥）
	if !e.privateKey.Curve.IsOnCurve(x, y) {
		return false, errors.New("对端公钥不在本方曲线上")
	}
	if _, err := e.deriveSharedX(x, y); err != nil {
		return false, err
	}
	return true, nil
}

func (e *ECDHCrypto) DeriveSharedSecret(publicKey string) ([]byte, error) {
	// 兼容 Encrypt 返回的 "ecdh$<base64>" 形态与裸 base64 两种输入
	if prefix, rest, found := strings.Cut(publicKey, "$"); found {
		if prefix != "ecdh" {
			return nil, errors.New("加密字符串格式无效")
		}
		publicKey = rest
	}

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
	sharedX, err := e.deriveSharedX(x, y)
	if err != nil {
		return nil, err
	}
	return sharedX, nil
}

// deriveSharedX 计算本方私钥与对端公钥 (x, y) 的 ECDH 共享密钥，
// 并按曲线字节长度左侧补零，避免大整数序列化丢失前导零导致双方结果不一致。
func (e *ECDHCrypto) deriveSharedX(x, y *big.Int) ([]byte, error) {
	sharedX, _ := e.privateKey.Curve.ScalarMult(x, y, e.privateKey.D.Bytes())
	size := (e.privateKey.Curve.Params().BitSize + 7) / 8
	padded := make([]byte, size)
	sharedX.FillBytes(padded)
	return padded, nil
}
