package password

import (
	"testing"
)

func TestECDSACrypto_EncryptAndVerify(t *testing.T) {
	crypto, err := NewECDSACrypto()
	if err != nil {
		t.Fatalf("创建 ECDSACrypto 实例失败: %v", err)
	}

	message := "test message"

	// 签名消息
	encrypted, err := crypto.Encrypt(message)
	if err != nil {
		t.Fatalf("签名失败: %v", err)
	}

	// 验证签名
	isValid, err := crypto.Verify(message, encrypted)
	if err != nil {
		t.Fatalf("验证失败: %v", err)
	}

	if !isValid {
		t.Fatal("签名验证未通过")
	}
}

func TestECDHCrypto_EncryptAndVerify(t *testing.T) {
	crypto1, err := NewECDHCrypto()
	if err != nil {
		t.Fatalf("创建 ECDHCrypto 实例1失败: %v", err)
	}

	crypto2, err := NewECDHCrypto()
	if err != nil {
		t.Fatalf("创建 ECDHCrypto 实例2失败: %v", err)
	}

	message := "test message"

	// 获取公钥
	encrypted, err := crypto1.Encrypt(message)
	if err != nil {
		t.Fatalf("加密失败: %v", err)
	}

	// 验证共享密钥
	isValid, err := crypto2.Verify(message, encrypted)
	if err != nil {
		t.Fatalf("验证失败: %v", err)
	}

	if !isValid {
		t.Fatal("共享密钥验证未通过")
	}
}
