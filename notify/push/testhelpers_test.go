package push

import (
	"crypto/md5"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/base64"
	"encoding/hex"
	"encoding/pem"
	"fmt"
	"net/url"
	"sync"
	"testing"
)

// 测试共享件：RSA 测试私钥（PKCS#8 PEM，进程内生成一次）与 base64url 段解码。

var (
	testRSAKeyOnce sync.Once
	testRSAKeyPEM  string
	testRSAKeyErr  error
)

// mustGenerateTestRSAKeyPEM 进程内生成一把 2048 位测试私钥（仅测试用，不是生产凭据）。
func mustGenerateTestRSAKeyPEM() string {
	testRSAKeyOnce.Do(func() {
		key, err := rsa.GenerateKey(rand.Reader, 2048)
		if err != nil {
			testRSAKeyErr = fmt.Errorf("generate test rsa key: %w", err)
			return
		}
		der, err := x509.MarshalPKCS8PrivateKey(key)
		if err != nil {
			testRSAKeyErr = fmt.Errorf("marshal test rsa key: %w", err)
			return
		}
		testRSAKeyPEM = string(pem.EncodeToMemory(&pem.Block{Type: "PRIVATE KEY", Bytes: der}))
	})
	if testRSAKeyErr != nil {
		panic(testRSAKeyErr)
	}
	return testRSAKeyPEM
}

// base64DecodeSegment 解 base64url（无填充）段（JWT 的三段都是这个形态）。
func base64DecodeSegment(s string) ([]byte, error) {
	return base64.RawURLEncoding.DecodeString(s)
}

// parseForm 解析 x-www-form-urlencoded 请求体为键值表（测试断言用）。
func parseForm(raw string) (map[string]string, error) {
	values, err := url.ParseQuery(raw)
	if err != nil {
		return nil, err
	}
	out := make(map[string]string, len(values))
	for k, v := range values {
		out[k] = v[0]
	}
	return out, nil
}

// requireTestKeyPEM 测试里取 RSA PEM，生成失败当场炸出原因。
func requireTestKeyPEM(t *testing.T) string {
	t.Helper()
	pemContent := mustGenerateTestRSAKeyPEM()
	if pemContent == "" {
		t.Fatal("empty test rsa key pem")
	}
	return pemContent
}

// md5hex MD5 小写 hex（vivo 签名独立复算用）。
func md5hex(s string) string {
	sum := md5.Sum([]byte(s))
	return hex.EncodeToString(sum[:])
}
