package password

import (
	"bytes"
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

// TestECDHCrypto_SharedSecretSymmetry 验证双方各自用对方公钥推导出的共享密钥
// 完全一致（含前导零填充：大整数序列化丢前导零会导致跨端结果不稳定）。
func TestECDHCrypto_SharedSecretSymmetry(t *testing.T) {
	crypto1, err := NewECDHCrypto()
	if err != nil {
		t.Fatalf("创建 ECDHCrypto 实例1失败: %v", err)
	}
	crypto2, err := NewECDHCrypto()
	if err != nil {
		t.Fatalf("创建 ECDHCrypto 实例2失败: %v", err)
	}

	blob1, err := crypto1.Encrypt("pwd")
	if err != nil {
		t.Fatalf("导出公钥1失败: %v", err)
	}
	blob2, err := crypto2.Encrypt("pwd")
	if err != nil {
		t.Fatalf("导出公钥2失败: %v", err)
	}

	secret1, err := crypto2.DeriveSharedSecret(blob1)
	if err != nil {
		t.Fatalf("实例2推导共享密钥失败: %v", err)
	}
	secret2, err := crypto1.DeriveSharedSecret(blob2)
	if err != nil {
		t.Fatalf("实例1推导共享密钥失败: %v", err)
	}

	if len(secret1) != 32 || len(secret2) != 32 {
		t.Fatalf("共享密钥长度 = %d/%d, want 32/32 (P-256 定长)", len(secret1), len(secret2))
	}
	if !bytes.Equal(secret1, secret2) {
		t.Fatal("双方推导的共享密钥不一致")
	}
}
