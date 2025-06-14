package password

import (
	"testing"
)

func TestPBKDF2Crypto_EncryptAndVerify(t *testing.T) {
	crypto := NewPBKDF2Crypto()

	// 测试加密
	password := "securepassword"
	encrypted, err := crypto.Encrypt(password)
	if err != nil {
		t.Fatalf("加密失败: %v", err)
	}

	if encrypted == "" {
		t.Fatal("加密结果为空")
	}
	t.Log(encrypted)

	// 测试验证成功
	isValid, err := crypto.Verify(password, encrypted)
	if err != nil {
		t.Fatalf("验证失败: %v", err)
	}

	if !isValid {
		t.Fatal("验证未通过，密码应匹配")
	}

	// 测试验证失败
	isValid, err = crypto.Verify("wrongpassword", encrypted)
	if err != nil {
		t.Fatalf("验证失败: %v", err)
	}

	if isValid {
		t.Fatal("验证通过，但密码不应匹配")
	}
}

func TestPBKDF2WithSHA512_EncryptAndVerify(t *testing.T) {
	crypto := NewPBKDF2WithSHA512()

	// 测试加密
	password := "securepassword"
	encrypted, err := crypto.Encrypt(password)
	if err != nil {
		t.Fatalf("加密失败: %v", err)
	}

	if encrypted == "" {
		t.Fatal("加密结果为空")
	}
	t.Log(encrypted)

	// 测试验证成功
	isValid, err := crypto.Verify(password, encrypted)
	if err != nil {
		t.Fatalf("验证失败: %v", err)
	}

	if !isValid {
		t.Fatal("验证未通过，密码应匹配")
	}

	// 测试验证失败
	isValid, err = crypto.Verify("wrongpassword", encrypted)
	if err != nil {
		t.Fatalf("验证失败: %v", err)
	}

	if isValid {
		t.Fatal("验证通过，但密码不应匹配")
	}
}
