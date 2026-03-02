package volc

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

const (
	volcService = "translate"
	host        = "translate.volcengineapi.com"
)

// Translator 火山翻译器
type Translator struct {
	accessKey string
	secretKey string // 预处理后的secretKey（去除末尾多余换行）
	region    string
	client    *http.Client
}

// NewTranslator 创建火山翻译器
func NewTranslator(accessKey, secretKey string, opts ...Option) (*Translator, error) {
	if accessKey == "" || secretKey == "" {
		return nil, fmt.Errorf("accessKey和secretKey不能为空")
	}

	// 核心修复：预处理SecretKey，去除末尾所有换行符（避免重复）
	secretKey = strings.TrimSuffix(secretKey, "\n")
	secretKey = strings.TrimSuffix(secretKey, "\r\n")

	translator := &Translator{
		accessKey: accessKey,
		secretKey: secretKey,
		region:    "cn-north-1",
		client:    &http.Client{Timeout: 30 * time.Second},
	}

	for _, opt := range opts {
		opt(translator)
	}

	return translator, nil
}

// Translate 单文本翻译
func (t *Translator) Translate(source, sourceLang, targetLang string) (string, error) {
	if source == "" {
		return "", fmt.Errorf("待翻译文本不能为空")
	}

	reqBody := translateTextRequest{
		SourceLanguage: sourceLang,
		TargetLanguage: targetLang,
		TextList:       []string{source},
		Scene:          "general",
	}

	resp, err := t.callAPI(reqBody)
	if err != nil {
		return "", fmt.Errorf("翻译失败: %w", err)
	}

	if len(resp.TranslationList) == 0 {
		return "", fmt.Errorf("翻译结果为空")
	}

	return resp.TranslationList[0].Text, nil
}

// TranslateBatch 批量翻译
func (t *Translator) TranslateBatch(sources []string, sourceLang, targetLang string) ([]string, error) {
	if len(sources) == 0 {
		return []string{}, nil
	}

	reqBody := translateTextRequest{
		SourceLanguage: sourceLang,
		TargetLanguage: targetLang,
		TextList:       sources,
		Scene:          "general",
	}

	resp, err := t.callAPI(reqBody)
	if err != nil {
		return nil, fmt.Errorf("批量翻译失败: %w", err)
	}

	if len(resp.TranslationList) != len(sources) {
		return nil, fmt.Errorf("结果数量不匹配：输入%d条，返回%d条", len(sources), len(resp.TranslationList))
	}

	results := make([]string, len(sources))
	for i, item := range resp.TranslationList {
		results[i] = item.Text
	}

	return results, nil
}

// callAPI 调用API
func (t *Translator) callAPI(body interface{}) (translateTextResponse, error) {
	bodyBytes, err := json.Marshal(body)
	if err != nil {
		return translateTextResponse{}, fmt.Errorf("序列化请求体失败: %w", err)
	}

	req, err := t.buildAndSignRequest(bodyBytes)
	if err != nil {
		return translateTextResponse{}, err
	}

	// 调试日志
	fmt.Printf("Request URL: %s\n", req.URL.String())
	fmt.Printf("Authorization: %s\n", req.Header.Get("Authorization"))
	fmt.Printf("X-Date: %s\n", req.Header.Get("X-Date"))
	fmt.Printf("Request Body: %s\n", string(bodyBytes))

	resp, err := t.client.Do(req)
	if err != nil {
		return translateTextResponse{}, fmt.Errorf("发送请求失败: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return translateTextResponse{}, fmt.Errorf("读取响应失败: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return translateTextResponse{}, fmt.Errorf("HTTP错误: %d, 响应: %s", resp.StatusCode, string(respBody))
	}

	var volcResp volcResponse
	if err := json.Unmarshal(respBody, &volcResp); err != nil {
		return translateTextResponse{}, fmt.Errorf("解析响应失败: %w, 响应: %s", err, string(respBody))
	}

	if volcResp.ResponseMetadata.Error != nil {
		return translateTextResponse{}, fmt.Errorf("API错误: %s - %s",
			volcResp.ResponseMetadata.Error.Code,
			volcResp.ResponseMetadata.Error.Message)
	}

	return volcResp.TranslationResponse, nil
}

// buildAndSignRequest 构建并签名请求
func (t *Translator) buildAndSignRequest(body []byte) (*http.Request, error) {
	now := time.Now().UTC()
	timestamp := now.Format("20060102T150405Z")
	dateStamp := now.Format("20060102")

	scheme := "https"
	path := "/"
	rawQuery := "Action=TranslateText&Version=2020-06-01"

	fullURL := fmt.Sprintf("%s://%s%s?%s", scheme, host, path, rawQuery)
	req, err := http.NewRequest("POST", fullURL, bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("创建请求失败: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Host", host)
	req.Header.Set("X-Date", timestamp)

	// 计算请求体哈希
	bodyHash := sha256.Sum256(body)
	bodyHashBase64 := base64.StdEncoding.EncodeToString(bodyHash[:])
	req.Header.Set("X-Content-Sha256", bodyHashBase64)

	// 生成签名
	signatureHex := t.generateSignature(body, timestamp, dateStamp)

	// 构建Authorization头
	signedHeaders := "content-type;host;x-content-sha256;x-date"
	credential := fmt.Sprintf("%s/%s/%s/%s/request", t.accessKey, dateStamp, t.region, volcService)
	authHeader := fmt.Sprintf(
		"HMAC-SHA256 Credential=%s, SignedHeaders=%s, Signature=%s",
		credential,
		signedHeaders,
		signatureHex,
	)
	req.Header.Set("Authorization", authHeader)

	return req, nil
}

// generateSignature 生成火山引擎签名（最终版）
func (t *Translator) generateSignature(body []byte, timestamp, date string) string {
	// 1. 计算请求体哈希
	bodyHash := sha256.Sum256(body)
	bodyHashBase64 := base64.StdEncoding.EncodeToString(bodyHash[:])

	// 2. 构建规范请求（完全匹配火山官方）
	canonicalHeaders := fmt.Sprintf(
		"content-type:application/json\nhost:%s\nx-content-sha256:%s\nx-date:%s\n",
		host, bodyHashBase64, timestamp,
	)
	signedHeaders := "content-type;host;x-content-sha256;x-date"
	canonicalRequest := fmt.Sprintf(
		"POST\n/\nAction=TranslateText&Version=2020-06-01\n%s%s\n%s",
		canonicalHeaders, signedHeaders, bodyHashBase64,
	)

	// 3. 构建待签名字符串
	credentialScope := fmt.Sprintf("%s/%s/%s/request", date, t.region, volcService)
	canonicalRequestHash := sha256.Sum256([]byte(canonicalRequest))
	canonicalRequestHashHex := hex.EncodeToString(canonicalRequestHash[:])
	stringToSign := fmt.Sprintf(
		"HMAC-SHA256\n%s\n%s\n%s",
		timestamp, credentialScope, canonicalRequestHashHex,
	)

	// 4. 派生签名密钥（仅拼接一次\n）
	kDate := t.hmacSHA256([]byte(t.secretKey+"\n"), []byte(date))
	kRegion := t.hmacSHA256(kDate, []byte(t.region))
	kService := t.hmacSHA256(kRegion, []byte(volcService))
	kSigning := t.hmacSHA256(kService, []byte("request"))

	// 5. 计算最终签名
	signature := t.hmacSHA256(kSigning, []byte(stringToSign))
	return hex.EncodeToString(signature)
}

// hmacSHA256 标准HMAC-SHA256实现
func (t *Translator) hmacSHA256(key, data []byte) []byte {
	h := hmac.New(sha256.New, key)
	h.Write(data)
	return h.Sum(nil)
}

// String 调试用字符串
func (t *Translator) String() string {
	return fmt.Sprintf("Translator{region: %s, accessKey: %s****%s}",
		t.region,
		t.accessKey[:6],
		t.accessKey[len(t.accessKey)-4:],
	)
}

// VerifySignature 手动校验签名（传入测试日志中的参数）
func (t *Translator) VerifySignature(testTimestamp, testDate, testBody string, expectedSignature string) {
	bodyBytes := []byte(testBody)
	calculatedSignature := t.generateSignature(bodyBytes, testTimestamp, testDate)
	fmt.Printf("\n=== 签名校验 ===\n")
	fmt.Printf("预期签名: %s\n", expectedSignature)
	fmt.Printf("计算签名: %s\n", calculatedSignature)
	fmt.Printf("签名是否一致: %t\n", calculatedSignature == expectedSignature)
}
