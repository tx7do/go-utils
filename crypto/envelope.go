// envelope.go —— 口令传输混合信封（rsa-v1 格式）。
//
// 场景：前端以随机 AES-256-GCM 会话密钥加密口令，再用 RSA-2048
// OAEP(SHA-256) 包裹会话密钥，整体以信封串传输——公钥可静态内置前端
// （公开值不降低安全性），私钥仅服务端持有。信封格式：
//
//	"rsa-v1:<base64(wrappedKey)>:<base64(nonce)>:<base64(cipher)>"
//
// 前端（WebCrypto：RSA-OAEP SHA-256 包裹 + AES-256-GCM 加密）与服务端
// 各自独立实现同一格式，本包的编解码是两端的格式契约；改格式等于断掉
// 已部署前端，只能以新的版本前缀并行过渡。
//
// 密钥注入制：私钥的装载（env/配置文件/密管）由调用方完成，本包不感知
// 部署形态；开发默认密钥对属应用侧的部署契约，不进库。
package crypto

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
	"errors"
	"strings"
)

// EnvelopePrefix 信封格式版本前缀。
const EnvelopePrefix = "rsa-v1:"

// ErrNotEnvelope 输入不是 rsa-v1 信封（无前缀）。调用方可据此回落旧形态
// 解密（如过渡期的 base64+AES-CBC）。
var ErrNotEnvelope = errors.New("crypto: not an rsa-v1 envelope")

// ParsePrivateKeyPEM 解析 PEM 私钥：PKCS8（"PRIVATE KEY"）优先，
// PKCS1（"RSA PRIVATE KEY"）回落；env 单行形态的 \n 转义由调用方先行还原。
func ParsePrivateKeyPEM(pemStr string) (*rsa.PrivateKey, error) {
	pemStr = strings.TrimSpace(pemStr)
	block, _ := pem.Decode([]byte(pemStr))
	if block == nil {
		return nil, errors.New("crypto: pem decode failed")
	}
	if key, err := x509.ParsePKCS8PrivateKey(block.Bytes); err == nil {
		if rsaKey, ok := key.(*rsa.PrivateKey); ok {
			return rsaKey, nil
		}
		return nil, errors.New("crypto: pem is not an RSA private key")
	}
	rsaKey, err := x509.ParsePKCS1PrivateKey(block.Bytes)
	if err != nil {
		return nil, errors.New("crypto: unsupported private key format")
	}
	return rsaKey, nil
}

// EnvelopeCodec 持私钥的信封编解码器。
type EnvelopeCodec struct {
	privateKey *rsa.PrivateKey
}

// NewEnvelopeCodec 以 PEM 私钥构造信封编解码器。
func NewEnvelopeCodec(privateKeyPEM string) (*EnvelopeCodec, error) {
	key, err := ParsePrivateKeyPEM(privateKeyPEM)
	if err != nil {
		return nil, err
	}
	return &EnvelopeCodec{privateKey: key}, nil
}

// NewEnvelopeCodecFromKey 以已解析私钥构造（测试/运维工具场景）。
func NewEnvelopeCodecFromKey(key *rsa.PrivateKey) *EnvelopeCodec {
	if key == nil {
		return nil
	}
	return &EnvelopeCodec{privateKey: key}
}

// Encrypt 混合信封加密：随机 AES-256-GCM 会话密钥加密明文，
// RSA-OAEP(SHA-256) 包裹会话密钥，输出 EnvelopePrefix 形态。
// 服务端参考实现/单测/运维工具用；前端在 WebCrypto 里独立实现同格式。
func (c *EnvelopeCodec) Encrypt(plain string) (string, error) {
	key := c.privateKey
	if key == nil {
		return "", errors.New("crypto: private key not loaded")
	}

	aesKey := make([]byte, 32)
	if _, err := rand.Read(aesKey); err != nil {
		return "", err
	}
	block, err := aes.NewCipher(aesKey)
	if err != nil {
		return "", err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}
	nonce := make([]byte, gcm.NonceSize())
	if _, err := rand.Read(nonce); err != nil {
		return "", err
	}
	cipherText := gcm.Seal(nil, nonce, []byte(plain), nil)

	wrapped, err := rsa.EncryptOAEP(sha256.New(), rand.Reader, &key.PublicKey, aesKey, nil)
	if err != nil {
		return "", err
	}

	return EnvelopePrefix +
		base64.StdEncoding.EncodeToString(wrapped) + ":" +
		base64.StdEncoding.EncodeToString(nonce) + ":" +
		base64.StdEncoding.EncodeToString(cipherText), nil
}

// Decrypt 解开 rsa-v1 信封，返回明文口令。输入无前缀时返回
// ErrNotEnvelope（调用方决定是否回落旧形态）；格式/解密失败返回错误
// （调用方以 400 语义透传）。
func (c *EnvelopeCodec) Decrypt(v string) (string, error) {
	if !strings.HasPrefix(v, EnvelopePrefix) {
		return "", ErrNotEnvelope
	}
	key := c.privateKey
	if key == nil {
		return "", errors.New("crypto: private key not loaded")
	}

	parts := strings.Split(strings.TrimPrefix(v, EnvelopePrefix), ":")
	if len(parts) != 3 {
		return "", errors.New("crypto: malformed envelope")
	}
	wrapped, err1 := base64.StdEncoding.DecodeString(parts[0])
	nonce, err2 := base64.StdEncoding.DecodeString(parts[1])
	cipherText, err3 := base64.StdEncoding.DecodeString(parts[2])
	if err1 != nil || err2 != nil || err3 != nil {
		return "", errors.New("crypto: malformed envelope base64")
	}
	aesKey, err := rsa.DecryptOAEP(sha256.New(), rand.Reader, key, wrapped, nil)
	if err != nil {
		return "", errors.New("crypto: envelope unwrap failed")
	}
	block, err := aes.NewCipher(aesKey)
	if err != nil {
		return "", err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}
	plain, err := gcm.Open(nil, nonce, cipherText, nil)
	if err != nil {
		return "", errors.New("crypto: envelope open failed")
	}
	return string(plain), nil
}
