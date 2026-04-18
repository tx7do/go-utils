package crypto

import (
	"crypto/rand"
	"errors"

	"github.com/tjfoc/gmsm/sm2"
)

// SM2Cipher 实现 Cipher 接口，支持国密SM2加解密
// 仅适用于小数据块加密（如密钥交换），不适合大数据流
// 生产环境请妥善管理私钥
type SM2Cipher struct {
	privateKey *sm2.PrivateKey
	publicKey  *sm2.PublicKey
}

// NewSM2Cipher 生成新的SM2密钥对
func NewSM2Cipher() (*SM2Cipher, error) {
	priv, err := sm2.GenerateKey(rand.Reader)
	if err != nil {
		return nil, err
	}
	return &SM2Cipher{
		privateKey: priv,
		publicKey:  &priv.PublicKey,
	}, nil
}

// NewSM2CipherFromKey 用已有密钥初始化
func NewSM2CipherFromKey(priv *sm2.PrivateKey, pub *sm2.PublicKey) *SM2Cipher {
	return &SM2Cipher{privateKey: priv, publicKey: pub}
}

// Encrypt 使用SM2公钥加密
func (s *SM2Cipher) Encrypt(plain []byte) ([]byte, error) {
	if s.publicKey == nil {
		return nil, errors.New("public key is nil")
	}

	return sm2.Encrypt(s.publicKey, plain, rand.Reader, sm2.C1C3C2)
}

// Decrypt 使用SM2私钥解密
func (s *SM2Cipher) Decrypt(ciphertext []byte) ([]byte, error) {
	if s.privateKey == nil {
		return nil, errors.New("private key is nil")
	}

	return sm2.Decrypt(s.privateKey, ciphertext, sm2.C1C3C2)
}

// EncryptAsn1 加密并输出ASN.1格式（更通用，兼容多数平台）
func (s *SM2Cipher) EncryptAsn1(plain []byte) ([]byte, error) {
	if s.publicKey == nil {
		return nil, errors.New("sm2: public key not initialized")
	}
	return sm2.EncryptAsn1(s.publicKey, plain, rand.Reader)
}

// DecryptAsn1 解密ASN.1格式密文
func (s *SM2Cipher) DecryptAsn1(ciphertext []byte) ([]byte, error) {
	if s.privateKey == nil {
		return nil, errors.New("sm2: private key not initialized")
	}
	return sm2.DecryptAsn1(s.privateKey, ciphertext)
}

// PublicKey 获取公钥（用于导出、传输给对方）
func (s *SM2Cipher) PublicKey() *sm2.PublicKey {
	return s.publicKey
}

// PrivateKey 获取私钥（仅用于安全存储，禁止对外泄露）
func (s *SM2Cipher) PrivateKey() *sm2.PrivateKey {
	return s.privateKey
}

// Name 返回加密算法名称
func (s *SM2Cipher) Name() string {
	return "SM2"
}

// Sign 使用SM2私钥签名（ASN.1编码）
func (s *SM2Cipher) Sign(digest []byte) (string, error) {
	if s.privateKey == nil {
		return "", errors.New("private key is nil")
	}
	// uid为nil时，gmsm默认使用"1234567812345678"
	r, sInt, err := sm2.Sm2Sign(s.privateKey, digest, nil, rand.Reader)
	if err != nil {
		return "", err
	}
	sign, err := sm2.SignDigitToSignData(r, sInt)
	if err != nil {
		return "", err
	}
	return string(sign), nil
}

// Verify 使用SM2公钥验证签名（ASN.1编码）
func (s *SM2Cipher) Verify(data []byte, signature string) (bool, error) {
	if s.publicKey == nil {
		return false, errors.New("public key is nil")
	}
	r, sInt, err := sm2.SignDataToSignDigit([]byte(signature))
	if err != nil {
		return false, err
	}
	ok := sm2.Sm2Verify(s.publicKey, data, nil, r, sInt)
	return ok, nil
}
