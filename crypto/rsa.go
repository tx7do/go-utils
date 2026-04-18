package crypto

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/pem"
	"errors"
)

type RSACipher struct {
	privateKey *rsa.PrivateKey
	publicKey  *rsa.PublicKey
}

// NewRSACipher 生成新的RSA密钥对
func NewRSACipher(bits int) (*RSACipher, error) {
	priv, err := rsa.GenerateKey(rand.Reader, bits)
	if err != nil {
		return nil, err
	}
	return &RSACipher{
		privateKey: priv,
		publicKey:  &priv.PublicKey,
	}, nil
}

// NewRSACipherFromKey 用已有密钥初始化
func NewRSACipherFromKey(priv *rsa.PrivateKey, pub *rsa.PublicKey) *RSACipher {
	return &RSACipher{privateKey: priv, publicKey: pub}
}

// Encrypt 使用RSA公钥加密
func (r *RSACipher) Encrypt(plain []byte) ([]byte, error) {
	if r.publicKey == nil {
		return nil, errors.New("public key is nil")
	}
	return rsa.EncryptOAEP(sha256.New(), rand.Reader, r.publicKey, plain, nil)
}

// Decrypt 使用RSA私钥解密
func (r *RSACipher) Decrypt(ciphertext []byte) ([]byte, error) {
	if r.privateKey == nil {
		return nil, errors.New("private key is nil")
	}
	return rsa.DecryptOAEP(sha256.New(), rand.Reader, r.privateKey, ciphertext, nil)
}

// ExportPrivateKey 导出私钥PEM
func (r *RSACipher) ExportPrivateKey() (string, error) {
	if r.privateKey == nil {
		return "", errors.New("private key is nil")
	}
	privBytes := x509.MarshalPKCS1PrivateKey(r.privateKey)
	privPem := pem.EncodeToMemory(&pem.Block{Type: "RSA PRIVATE KEY", Bytes: privBytes})
	return string(privPem), nil
}

// ExportPublicKey 导出公钥PEM
func (r *RSACipher) ExportPublicKey() (string, error) {
	if r.publicKey == nil {
		return "", errors.New("public key is nil")
	}
	pubBytes, err := x509.MarshalPKIXPublicKey(r.publicKey)
	if err != nil {
		return "", err
	}
	pubPem := pem.EncodeToMemory(&pem.Block{Type: "RSA PUBLIC KEY", Bytes: pubBytes})
	return string(pubPem), nil
}

// PublicKey 获取公钥
func (r *RSACipher) PublicKey() *rsa.PublicKey {
	return r.publicKey
}

// PrivateKey 获取私钥
func (r *RSACipher) PrivateKey() *rsa.PrivateKey {
	return r.privateKey
}

func (r *RSACipher) Name() string {
	return "RSA"
}
