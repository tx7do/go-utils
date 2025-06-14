package password

import (
	"testing"
)

func TestHMACCrypto_EncryptAndVerify(t *testing.T) {
	secretKey := "mysecretkey"
	crypto := NewHMACCrypto(secretKey)

	data := "testdata"

	// 测试加密
	encrypted, err := crypto.Encrypt(data)
	if err != nil {
		t.Fatalf("加密失败: %v", err)
	}

	if encrypted == "" {
		t.Fatal("加密结果为空")
	}
	t.Log(encrypted)

	// 测试验证
	isValid, err := crypto.Verify(data, encrypted)
	if err != nil {
		t.Fatalf("验证失败: %v", err)
	}

	if !isValid {
		t.Fatal("验证结果不匹配")
	}

	// 测试验证失败的情况
	isValid, err = crypto.Verify("wrongdata", encrypted)
	if err != nil {
		t.Fatalf("验证失败: %v", err)
	}

	if isValid {
		t.Fatal("验证结果错误，预期验证失败")
	}
}
