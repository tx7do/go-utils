package crypto

import (
	"crypto/ecdh"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/sha256"
	"encoding/asn1"
	"encoding/base64"
	"errors"
	"fmt"
	"math/big"
	"strings"
)

// ECDSACipher 实现ECDSA签名/验签
// 适合数字签名场景

type ECDSACipher struct {
	privateKey *ecdsa.PrivateKey
	publicKey  *ecdsa.PublicKey
}

// NewECDSACipher 生成新的ECDSA密钥对
func NewECDSACipher() (*ECDSACipher, error) {
	priv, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		return nil, err
	}
	return &ECDSACipher{
		privateKey: priv,
		publicKey:  &priv.PublicKey,
	}, nil
}

// Sign 签名
func (e *ECDSACipher) Sign(data []byte) (string, error) {
	hash := sha256.Sum256(data)
	r, s, err := ecdsa.Sign(rand.Reader, e.privateKey, hash[:])
	if err != nil {
		return "", err
	}
	return base64.StdEncoding.EncodeToString(r.Bytes()) + "$" + base64.StdEncoding.EncodeToString(s.Bytes()), nil
}

// Verify 验签
func (e *ECDSACipher) Verify(data []byte, signature string) (bool, error) {
	parts := strings.Split(signature, "$")
	if len(parts) != 2 {
		return false, errors.New("invalid signature format")
	}

	rBytes, err := base64.StdEncoding.DecodeString(parts[0])
	if err != nil {
		return false, err
	}
	sBytes, err := base64.StdEncoding.DecodeString(parts[1])
	if err != nil {
		return false, err
	}

	r := new(big.Int).SetBytes(rBytes)
	s := new(big.Int).SetBytes(sBytes)
	hash := sha256.Sum256(data)

	return ecdsa.Verify(e.publicKey, hash[:], r, s), nil
}

// PublicKeyBytes 公钥/私钥导出
func (e *ECDSACipher) PublicKeyBytes() []byte {
	// 使用ASN.1编码公钥
	pubBytes, _ := asn1.Marshal(*e.publicKey)
	return pubBytes
}
func (e *ECDSACipher) PrivateKey() *ecdsa.PrivateKey { return e.privateKey }

// Name 返回算法名称
func (e *ECDSACipher) Name() string {
	return "ECDSA"
}

// ECDHCipher 实现ECDH密钥协商
// 适合安全通道密钥交换

type ECDHCipher struct {
	privateKey *ecdh.PrivateKey
	publicKey  *ecdh.PublicKey
}

// NewECDHCipher 生成新的ECDH密钥对
func NewECDHCipher() (*ECDHCipher, error) {
	priv, err := ecdh.P256().GenerateKey(rand.Reader)
	if err != nil {
		return nil, err
	}
	return &ECDHCipher{
		privateKey: priv,
		publicKey:  priv.PublicKey(),
	}, nil
}

// PublicKeyBytes 获取公钥字节
func (e *ECDHCipher) PublicKeyBytes() []byte {
	return e.publicKey.Bytes()
}

// DeriveSharedSecret 计算共享密钥
func (e *ECDHCipher) DeriveSharedSecret(peerPubBytes []byte) ([]byte, error) {
	peerPub, err := ecdh.P256().NewPublicKey(peerPubBytes)
	if err != nil {
		return nil, fmt.Errorf("ecdh public key: %w", err)
	}
	secret, err := e.privateKey.ECDH(peerPub)
	if err != nil {
		return nil, fmt.Errorf("ecdh derive: %w", err)
	}
	return secret, nil
}

// Name 返回算法名称
func (e *ECDHCipher) Name() string {
	return "ECDH"
}

// PublicKey 获取公钥（用于导出、传输给对方）
func (e *ECDHCipher) PublicKey() *ecdh.PublicKey { return e.publicKey }

// PrivateKey 获取私钥（仅用于安全存储，禁止对外泄露）
func (e *ECDHCipher) PrivateKey() *ecdh.PrivateKey { return e.privateKey }
