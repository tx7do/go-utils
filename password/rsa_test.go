package password

import (
	"testing"
)

func TestRSACrypto_EncryptAndDecrypt(t *testing.T) {
	// 创建 RSACrypto 实例
	crypto, err := NewRSACrypto(2048)
	if err != nil {
		t.Fatalf("创建 RSACrypto 实例失败: %v", err)
	}

	// 测试加密
	originalData := "securedata"
	encryptedData, err := crypto.Encrypt(originalData)
	if err != nil {
		t.Fatalf("加密失败: %v", err)
	}

	if encryptedData == "" {
		t.Fatal("加密结果为空")
	}
	t.Log(encryptedData)

	// 测试解密
	decryptedData, err := crypto.Decrypt(encryptedData)
	if err != nil {
		t.Fatalf("解密失败: %v", err)
	}

	if decryptedData != originalData {
		t.Fatalf("解密结果不匹配，期望: %s，实际: %s", originalData, decryptedData)
	}
}

func TestRSACrypto_ExportKeys(t *testing.T) {
	// 创建 RSACrypto 实例
	crypto, err := NewRSACrypto(2048)
	if err != nil {
		t.Fatalf("创建 RSACrypto 实例失败: %v", err)
	}

	// 测试导出私钥
	privateKey, err := crypto.ExportPrivateKey()
	if err != nil {
		t.Fatalf("导出私钥失败: %v", err)
	}

	if privateKey == "" {
		t.Fatal("导出的私钥为空")
	}

	// 测试导出公钥
	publicKey, err := crypto.ExportPublicKey()
	if err != nil {
		t.Fatalf("导出公钥失败: %v", err)
	}

	if publicKey == "" {
		t.Fatal("导出的公钥为空")
	}
}
