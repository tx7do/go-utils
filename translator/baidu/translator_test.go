package baidu

import (
	"testing"
)

func TestNewTranslator(t *testing.T) {
	translator := NewTranslator("test_app_id", "test_secret_key")
	if translator == nil {
		t.Fatal("NewTranslator failed")
	}
	if translator.appID != "test_app_id" {
		t.Fatal("appID not set correctly")
	}
	if translator.secretKey != "test_secret_key" {
		t.Fatal("secretKey not set correctly")
	}
}

func TestGenerateSign(t *testing.T) {
	translator := NewTranslator("20230101000001234", "test_secret_key")

	testCases := []struct {
		query string
		salt  int64
	}{
		{"apple", 1234567890},
		{"hello world", 9876543210},
	}

	for _, tc := range testCases {
		sign := translator.generateSign(tc.query, tc.salt)
		if sign == "" {
			t.Fatal("sign should not be empty")
		}
		// Sign should be a valid MD5 hex string (32 characters)
		if len(sign) != 32 {
			t.Fatalf("sign length should be 32, got %d", len(sign))
		}
	}
}

// 实际的翻译测试需要真实的 appID 和 secretKey
// func TestTranslate(t *testing.T) {
// 	translator := NewTranslator("your_app_id", "your_secret_key")
// 	result, err := translator.Translate("Hello", "en", "zh")
// 	if err != nil {
// 		t.Fatal(err)
// 	}
// 	if result == "" {
// 		t.Fatal("result should not be empty")
// 	}
// 	t.Logf("Translation result: %s", result)
// }
