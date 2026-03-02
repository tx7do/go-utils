package volc

import (
	"fmt"
	"testing"
)

func TestNewTranslator(t *testing.T) {
	translator, err := NewTranslator(
		"test_access_key",
		"test_secret_key",
	)
	if err != nil {
		t.Fatalf("NewTranslator failed: %v", err)
	}
	if translator == nil {
		t.Fatal("NewTranslator returned nil")
	}
	if translator.region != "cn-north-1" {
		t.Fatal("region should default to cn-north-1")
	}

	// 2. 传入测试日志中的参数
	testTimestamp := "20260302T124626Z"
	testDate := "20260302"
	testBody := `{"SourceLanguage":"zh","TargetLanguage":"en","TextList":["你好，世界！"],"Scene":"general"}`
	expectedSignature := "dd55fbfd9fb799339968d5562f0a93eb7b7437e22f8e69cb22473132cf8c06c3"

	// 3. 校验签名
	translator.VerifySignature(testTimestamp, testDate, testBody, expectedSignature)

	// 2. 单文本翻译
	result, err := translator.Translate("你好，世界！", "zh", "en")
	if err != nil {
		fmt.Printf("单文本翻译失败：%v\n", err)
		return
	}
	fmt.Printf("单文本翻译结果：%s\n", result) // 输出：Hello, world!

	// 3. 批量翻译
	batchResult, err := translator.TranslateBatch([]string{"火山翻译", "Go客户端"}, "zh", "en")
	if err != nil {
		fmt.Printf("批量翻译失败：%v\n", err)
		return
	}
	fmt.Printf("批量翻译结果：%v\n", batchResult) // 输出：[Volcengine Translate Go Client]
}

func TestNewTranslatorWithOptions(t *testing.T) {
	translator, err := NewTranslator(
		"test_access_key",
		"test_secret_key",
		WithRegion("cn-shanghai"),
	)
	if err != nil {
		t.Fatalf("NewTranslator with options failed: %v", err)
	}
	if translator == nil {
		t.Fatal("NewTranslator returned nil")
	}
	if translator.region != "cn-shanghai" {
		t.Fatal("region not set correctly via WithRegion option")
	}
}

// 实际的翻译测试需要真实的 AccessKey 和 SecretKey
// 取消注释以下代码并填入真实凭证进行测试
//
// func TestTranslate(t *testing.T) {
// 	translator, err := NewTranslator("your_access_key", "your_secret_key")
// 	if err != nil {
// 		t.Fatal(err)
// 	}
// 	result, err := translator.Translate("Hello", "en", "zh")
// 	if err != nil {
// 		t.Fatal(err)
// 	}
// 	if result == "" {
// 		t.Fatal("result should not be empty")
// 	}
// 	t.Logf("Translation result: %s", result)
// }
//
// func TestTranslateBatch(t *testing.T) {
// 	translator, err := NewTranslator("your_access_key", "your_secret_key")
// 	if err != nil {
// 		t.Fatal(err)
// 	}
// 	sources := []string{"Hello", "World", "Good morning"}
// 	results, err := translator.TranslateBatch(sources, "en", "zh")
// 	if err != nil {
// 		t.Fatal(err)
// 	}
// 	if len(results) != len(sources) {
// 		t.Fatalf("expected %d results, got %d", len(sources), len(results))
// 	}
// 	for i, result := range results {
// 		t.Logf("Translation %d: %s -> %s", i, sources[i], result)
// 	}
// }
