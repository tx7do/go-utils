package crypto

import (
	"crypto/ecdsa"
	"crypto/x509"
	"encoding/pem"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestECDSACipher_SignVerify(t *testing.T) {
	cipher, err := NewECDSACipher()
	assert.NoError(t, err)
	data := []byte("ecdsa sign test data")

	sig, err := cipher.Sign(data)
	assert.NoError(t, err)
	ok, err := cipher.Verify(data, sig)
	assert.NoError(t, err)
	assert.True(t, ok, "ECDSA signature should verify")

	// tampered data
	ok, err = cipher.Verify([]byte("tampered"), sig)
	assert.NoError(t, err)
	assert.False(t, ok, "ECDSA signature should not verify for tampered data")
}

func TestECDSACipher_SignVerify_Fail(t *testing.T) {
	cipher, _ := NewECDSACipher()
	data := []byte("ecdsa fail test")
	sig, _ := cipher.Sign(data)
	ok, err := cipher.Verify([]byte("other data"), sig)
	if err != nil {
		t.Fatalf("Verify error: %v", err)
	}
	if ok {
		t.Fatalf("Verify should fail for wrong data")
	}
}

func TestECDHCipher_KeyExchange(t *testing.T) {
	alice, err := NewECDHCipher()
	assert.NoError(t, err)
	bob, err := NewECDHCipher()
	assert.NoError(t, err)

	alicePub := alice.PublicKeyBytes()
	bobPub := bob.PublicKeyBytes()

	aliceSecret, err := alice.DeriveSharedSecret(bobPub)
	assert.NoError(t, err)
	bobSecret, err := bob.DeriveSharedSecret(alicePub)
	assert.NoError(t, err)

	assert.Equal(t, aliceSecret, bobSecret, "ECDH shared secrets should match")
}

func TestECDHCipher_InvalidPeerKey(t *testing.T) {
	alice, _ := NewECDHCipher()
	_, err := alice.DeriveSharedSecret([]byte("invalid"))
	if err == nil {
		t.Fatalf("Should fail for invalid peer public key")
	}
}

// TestECDSACipher_PublicKeyBytes_RoundTrip 回归测试：
// PublicKeyBytes 必须返回非空的 PKIX/DER 编码公钥，
// 且解析回的公钥能用于验签（曾因 asn1.Marshal(ecdsa.PublicKey) 失败被吞错而恒返回 nil）
func TestECDSACipher_PublicKeyBytes_RoundTrip(t *testing.T) {
	cipher, err := NewECDSACipher()
	assert.NoError(t, err)

	pubBytes := cipher.PublicKeyBytes()
	assert.NotEmpty(t, pubBytes, "PublicKeyBytes should not return nil/empty bytes")

	// 用包内解析助手还原公钥，并以还原的公钥验签
	parsed, err := ParseECDSAPublicKey(pubBytes)
	assert.NoError(t, err)
	assert.Equal(t, cipher.publicKey.Curve, parsed.Curve, "curve should roundtrip")
	assert.Equal(t, cipher.publicKey.X, parsed.X, "X coordinate should roundtrip")
	assert.Equal(t, cipher.publicKey.Y, parsed.Y, "Y coordinate should roundtrip")

	verifier := NewECDSACipherFromKey(nil, parsed)
	data := []byte("ecdsa pubkey roundtrip test")
	sig, err := cipher.Sign(data)
	assert.NoError(t, err)

	ok, err := verifier.Verify(data, sig)
	assert.NoError(t, err)
	assert.True(t, ok, "signature should verify with the parsed public key")

	ok, err = verifier.Verify([]byte("tampered"), sig)
	assert.NoError(t, err)
	assert.False(t, ok, "signature should not verify for tampered data")
}

// TestECDSACipher_PublicKeyBytes_StdParse 交叉验证编码格式为标准 X.509 PKIX，
// 且标准库解析出的公钥与签名验证语义等价
func TestECDSACipher_PublicKeyBytes_StdParse(t *testing.T) {
	cipher, _ := NewECDSACipher()

	pubBytes := cipher.PublicKeyBytes()
	assert.NotEmpty(t, pubBytes)

	stdPub, err := x509.ParsePKIXPublicKey(pubBytes)
	assert.NoError(t, err)

	ecdsaPub, ok := stdPub.(*ecdsa.PublicKey)
	assert.True(t, ok, "parsed key should be *ecdsa.PublicKey, got %T", stdPub)
	assert.Equal(t, cipher.publicKey.X, ecdsaPub.X)
	assert.Equal(t, cipher.publicKey.Y, ecdsaPub.Y)

	// 用标准库解析的公钥构造验签器完成一次完整验签
	verifier := NewECDSACipherFromKey(nil, ecdsaPub)
	data := []byte("ecdsa stdlib parse test")
	sig, err := cipher.Sign(data)
	assert.NoError(t, err)
	ok, err = verifier.Verify(data, sig)
	assert.NoError(t, err)
	assert.True(t, ok, "signature should verify with stdlib-parsed public key")
}

func TestParseECDSAPublicKey_Invalid(t *testing.T) {
	_, err := ParseECDSAPublicKey([]byte("invalid"))
	assert.Error(t, err)

	// 非ECDSA公钥（RSA）应被拒绝
	rsaCipher, err := NewRSACipher(2048)
	assert.NoError(t, err)
	pemStr, err := rsaCipher.ExportPublicKey()
	assert.NoError(t, err)
	block, _ := pem.Decode([]byte(pemStr))
	assert.NotNil(t, block, "RSA public key PEM should decode")
	_, err = ParseECDSAPublicKey(block.Bytes)
	assert.Error(t, err, "RSA public key should not parse as ECDSA")
}
