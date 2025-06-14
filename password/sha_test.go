package password

import (
	"testing"
)

func TestSHACrypto_EncryptAndVerify_SHA256(t *testing.T) {
	crypto := NewSHA256Crypto()

	password := "securepassword"
	encrypted, err := crypto.Encrypt(password)
	if err != nil {
		t.Fatalf("加密失败: %v", err)
	}

	if encrypted == "" {
		t.Fatal("加密结果为空")
	}
	t.Log(encrypted)

	isValid, err := crypto.Verify(password, encrypted)
	if err != nil {
		t.Fatalf("验证失败: %v", err)
	}

	if !isValid {
		t.Fatal("验证结果不匹配")
	}
}

func TestSHACrypto_EncryptAndVerify_SHA512(t *testing.T) {
	crypto := NewSHA512Crypto()

	password := "securepassword"
	encrypted, err := crypto.Encrypt(password)
	if err != nil {
		t.Fatalf("加密失败: %v", err)
	}

	if encrypted == "" {
		t.Fatal("加密结果为空")
	}
	t.Log(encrypted)

	isValid, err := crypto.Verify(password, encrypted)
	if err != nil {
		t.Fatalf("验证失败: %v", err)
	}

	if !isValid {
		t.Fatal("验证结果不匹配")
	}
}
