package alibaba

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewTranslator(t *testing.T) {
	translator, err := NewTranslator(
		"test_access_key_id",
		"test_access_key_secret",
	)
	if err != nil {
		t.Fatalf("NewTranslator failed: %v", err)
	}
	if translator == nil {
		t.Fatal("NewTranslator returned nil translator")
	}
	if translator.regionID != "cn-hangzhou" {
		t.Fatal("regionID should default to cn-hangzhou")
	}

	const text string = `Hello, World!`
	result, err := translator.Translate(text, "auto", "zh-TW")
	assert.Nil(t, err)
	t.Log(result)
}

func TestNewTranslatorWithOptions(t *testing.T) {
	translator, err := NewTranslator(
		"test_access_key_id",
		"test_access_key_secret",
		WithRegionID("cn-hangzhou"),
	)
	if err != nil {
		t.Fatalf("NewTranslator with options failed: %v", err)
	}
	if translator == nil {
		t.Fatal("NewTranslator returned nil translator")
	}
	if translator.regionID != "cn-hangzhou" {
		t.Fatal("regionID not set correctly via WithRegionID option")
	}

	const text string = `Hello, World!`
	result, err := translator.Translate(text, "auto", "zh-TW")
	assert.Nil(t, err)
	t.Log(result)
}

// 实际的翻译测试需要真实的 AccessKeyID 和 AccessKeySecret
// func TestTranslate(t *testing.T) {
// 	translator, err := NewTranslator("your_access_key_id", "your_access_key_secret")
// 	if err != nil {
// 		t.Fatal(err)
// 	}
// 	result, err := translator.Translate("Hello, World!", "en", "zh")
// 	if err != nil {
// 		t.Fatal(err)
// 	}
// 	if result == "" {
// 		t.Fatal("result should not be empty")
// 	}
// 	t.Logf("Translation result: %s", result)
// }

// func TestTranslate(t *testing.T) {
// 	translator := NewTranslator("your_access_key_id", "your_access_key_secret")
// 	result, err := translator.Translate("Hello", "en", "zh")
// 	if err != nil {
// 		t.Fatal(err)
// 	}
// 	if result == "" {
// 		t.Fatal("result should not be empty")
// 	}
// 	t.Logf("Translation result: %s", result)
// }
