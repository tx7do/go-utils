package password

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
)

// RSACrypto 实现 RSA 加密和解密
type RSACrypto struct {
	privateKey *rsa.PrivateKey
	publicKey  *rsa.PublicKey
}

// NewRSACrypto 创建一个新的 RSACrypto 实例
func NewRSACrypto(keySize int) (*RSACrypto, error) {
	privateKey, err := rsa.GenerateKey(rand.Reader, keySize)
	if err != nil {
		return nil, err
	}
	return &RSACrypto{
		privateKey: privateKey,
		publicKey:  &privateKey.PublicKey,
	}, nil
}

// Encrypt 使用公钥加密数据
func (r *RSACrypto) Encrypt(data string) (string, error) {
	encryptedBytes, err := rsa.EncryptOAEP(sha256.New(), rand.Reader, r.publicKey, []byte(data), nil)
	if err != nil {
		return "", err
	}
	return base64.StdEncoding.EncodeToString(encryptedBytes), nil
}

// Decrypt 使用私钥解密数据
func (r *RSACrypto) Decrypt(encryptedData string) (string, error) {
	decodedData, err := base64.StdEncoding.DecodeString(encryptedData)
	if err != nil {
		return "", err
	}
	decryptedBytes, err := rsa.DecryptOAEP(sha256.New(), rand.Reader, r.privateKey, decodedData, nil)
	if err != nil {
		return "", err
	}
	return string(decryptedBytes), nil
}

// ExportPrivateKey 导出私钥为 PEM 格式
func (r *RSACrypto) ExportPrivateKey() (string, error) {
	privateKeyBytes := x509.MarshalPKCS1PrivateKey(r.privateKey)
	privateKeyPEM := pem.EncodeToMemory(&pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: privateKeyBytes,
	})
	return string(privateKeyPEM), nil
}

// ExportPublicKey 导出公钥为 PEM 格式
func (r *RSACrypto) ExportPublicKey() (string, error) {
	publicKeyBytes, err := x509.MarshalPKIXPublicKey(r.publicKey)
	if err != nil {
		return "", err
	}
	publicKeyPEM := pem.EncodeToMemory(&pem.Block{
		Type:  "RSA PUBLIC KEY",
		Bytes: publicKeyBytes,
	})
	return string(publicKeyPEM), nil
}
