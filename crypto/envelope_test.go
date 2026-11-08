package crypto

// envelope_test.go —— 口令传输混合信封（rsa-v1）的语义锁定。
//
// 重点：用裸 crypto 原语（RSA-OAEP 解包裹 + GCM 打开）复刻解密路径，
// 逐段验证信封串的形状与算法绑定——库内实现若漂移（换算法/换分段/换前缀），
// 与前端 WebCrypto 侧的格式契约即告破裂，这类回归必须在这里被拦住。

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// newTestKey 生成 2048 位测试密钥。
func newTestKey(t *testing.T) *rsa.PrivateKey {
	t.Helper()
	key, err := rsa.GenerateKey(rand.Reader, 2048)
	require.NoError(t, err)
	return key
}

// TestEnvelopeRoundTrip 加密→解密还原明文（含中文/符号/单字符口令）。
func TestEnvelopeRoundTrip(t *testing.T) {
	codec := NewEnvelopeCodecFromKey(newTestKey(t))
	for _, pw := range []string{"Abcd@1234", "中文密码+符号!@#", "x"} {
		env, err := codec.Encrypt(pw)
		require.NoError(t, err, "encrypt %q", pw)
		got, err := codec.Decrypt(env)
		require.NoError(t, err, "decrypt %q", pw)
		assert.Equal(t, pw, got)
	}
}

// TestEnvelopeFormatPinned 用裸原语逐段解密，锁定信封形状与算法绑定：
// rsa-v1 前缀 + 三段 base64 + OAEP-SHA256 包裹 AES-256-GCM。
func TestEnvelopeFormatPinned(t *testing.T) {
	key := newTestKey(t)
	codec := NewEnvelopeCodecFromKey(key)

	env, err := codec.Encrypt("s3cret-V@lue")
	require.NoError(t, err)

	require.True(t, strings.HasPrefix(env, EnvelopePrefix), "信封必须带 rsa-v1 前缀")
	parts := strings.Split(strings.TrimPrefix(env, EnvelopePrefix), ":")
	require.Len(t, parts, 3, "信封必须恰为 wrapped:nonce:cipher 三段")

	wrapped, err1 := base64.StdEncoding.DecodeString(parts[0])
	nonce, err2 := base64.StdEncoding.DecodeString(parts[1])
	cipherText, err3 := base64.StdEncoding.DecodeString(parts[2])
	require.NoError(t, err1)
	require.NoError(t, err2)
	require.NoError(t, err3)

	// 裸 RSA-OAEP(SHA-256) 解包裹出 32 字节会话密钥
	aesKey, err := rsa.DecryptOAEP(sha256.New(), rand.Reader, key, wrapped, nil)
	require.NoError(t, err, "wrappedKey 必须是 RSA-OAEP(SHA-256) 形态")
	assert.Len(t, aesKey, 32, "会话密钥必须为 AES-256")

	// 裸 GCM 打开密文
	block, err := aes.NewCipher(aesKey)
	require.NoError(t, err)
	gcm, err := cipher.NewGCM(block)
	require.NoError(t, err)
	plain, err := gcm.Open(nil, nonce, cipherText, nil)
	require.NoError(t, err, "cipher 必须是 AES-256-GCM 形态")
	assert.Equal(t, "s3cret-V@lue", string(plain))

	// nonce 长度锁定（WebCrypto 侧固定 12 字节）
	assert.Len(t, nonce, 12, "GCM nonce 必须为 12 字节")
}

// TestDecryptRejectsMalformed 各畸形输入一律报错不恐慌。
func TestDecryptRejectsMalformed(t *testing.T) {
	codec := NewEnvelopeCodecFromKey(newTestKey(t))
	for _, bad := range []string{
		EnvelopePrefix + "!!!:???:**:",    // 非 base64
		EnvelopePrefix + "aAAAbbb:",       // 段数不足
		EnvelopePrefix + "a:b:c:d",        // 段数超限
		EnvelopePrefix + "AAAA:BBBB:CCCC", // 非 base64 字符
	} {
		_, err := codec.Decrypt(bad)
		assert.Error(t, err, "garbage should fail: %q", bad)
	}
}

// TestDecryptWrongKeyWrongEnvelope 用另一把私钥解不出；无前缀返回 ErrNotEnvelope。
func TestDecryptWrongKeyWrongEnvelope(t *testing.T) {
	codec := NewEnvelopeCodecFromKey(newTestKey(t))
	env, err := codec.Encrypt("hello")
	require.NoError(t, err)

	other := NewEnvelopeCodecFromKey(newTestKey(t))
	_, err = other.Decrypt(env)
	assert.Error(t, err, "错误私钥必须解不开信封")

	_, err = codec.Decrypt("bm90LWFuLWVudmVsb3Bl") // 无前缀的 base64 串
	assert.ErrorIs(t, err, ErrNotEnvelope, "无前缀必须返回 ErrNotEnvelope 供调用方回落")
}

// TestParsePrivateKeyPEM PKCS8（"PRIVATE KEY"）与 PKCS1（"RSA PRIVATE KEY"）
// 两种 PEM 形态都能解析，垃圾输入报错。
func TestParsePrivateKeyPEM(t *testing.T) {
	key := newTestKey(t)

	pkcs1 := pem.EncodeToMemory(&pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(key)})
	got, err := ParsePrivateKeyPEM(string(pkcs1))
	require.NoError(t, err)
	assert.Equal(t, 0, key.N.Cmp(got.N), "PKCS1 形态应解析出同一把密钥")

	pkcs8, err := x509.MarshalPKCS8PrivateKey(key)
	require.NoError(t, err)
	pkcs8Pem := pem.EncodeToMemory(&pem.Block{Type: "PRIVATE KEY", Bytes: pkcs8})
	got, err = ParsePrivateKeyPEM(string(pkcs8Pem))
	require.NoError(t, err)
	assert.Equal(t, 0, key.N.Cmp(got.N), "PKCS8 形态应解析出同一把密钥")

	_, err = ParsePrivateKeyPEM("not a pem")
	assert.Error(t, err)
}

// TestNilKeyCodec 空私钥的编解码器必须报错而非恐慌。
func TestNilKeyCodec(t *testing.T) {
	assert.Nil(t, NewEnvelopeCodecFromKey(nil), "nil 私钥构造应返回 nil codec")

	empty := &EnvelopeCodec{}
	_, err := empty.Encrypt("x")
	assert.Error(t, err)
	_, err = empty.Decrypt(EnvelopePrefix + "AAAA:BBBB:CCCC")
	assert.Error(t, err)
}
