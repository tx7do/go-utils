package sms

import (
	"context"
	"crypto/hmac"
	crand "crypto/rand"
	"crypto/sha1"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"sort"
	"strings"
	"time"
)

// 阿里云短信（Dysmsapi SendSms）。协议是阿里云 RPC 风格：全部参数排序拼接做 HMAC-SHA1 签名，
// 请求本身是一张普通表单——签名只管"请求没被篡改"，与 GET/POST 形态无关。
const (
	aliyunSmsEndpoint      = "https://dysmsapi.aliyuncs.com/"
	aliyunSmsVersion       = "2017-05-26"
	aliyunSmsDefaultRegion = "cn-hangzhou"

	// aliyunAnswerBytes 读对端应答的上限：结论是一个几百字节的小 JSON。
	aliyunAnswerBytes = 4 * 1024
)

// sendAliyun 给单个手机号下发一条短信。
// phone 收国内裸号（无 +86 前缀，由 NormalizePhoneNumber 归一）。
// templateParams 按参数名整表序列化为 TemplateParam JSON（模板里写 ${name}）。
func (s *Service) sendAliyun(ctx context.Context, phone string, templateParams map[string]string) error {
	region := s.cfg.Region
	if region == "" {
		region = aliyunSmsDefaultRegion
	}

	templateParam, err := json.Marshal(templateParams)
	if err != nil {
		return fmt.Errorf("encode template param failed: %w", err)
	}

	params := url.Values{}
	params.Set("AccessKeyId", s.cfg.AccessKeyID)
	params.Set("Action", "SendSms")
	params.Set("Format", "JSON")
	params.Set("PhoneNumbers", phone)
	params.Set("RegionId", region)
	params.Set("SignName", s.cfg.SignName)
	params.Set("SignatureMethod", "HMAC-SHA1")
	params.Set("SignatureNonce", crand.Text())
	params.Set("SignatureVersion", "1.0")
	params.Set("TemplateCode", s.cfg.TemplateCode)
	params.Set("TemplateParam", string(templateParam))
	params.Set("Timestamp", time.Now().UTC().Format("2006-01-02T15:04:05Z"))
	params.Set("Version", aliyunSmsVersion)

	params.Set("Signature", SignAliyunRPC(params, s.cfg.AccessSecret))

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, aliyunSmsEndpoint, strings.NewReader(params.Encode()))
	if err != nil {
		return fmt.Errorf("build request failed: %w", err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := s.client.Do(req)
	if err != nil {
		return err
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	body, readErr := io.ReadAll(io.LimitReader(resp.Body, aliyunAnswerBytes))
	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		return fmt.Errorf("peer answered %s: %s", resp.Status, truncateAnswer(body))
	}
	if readErr != nil && readErr != io.EOF {
		return fmt.Errorf("read peer answer: %w", readErr)
	}

	// 阿里云把业务结论放在 body 的 Code 里（HTTP 恒 200）：Code=="OK" 才是发出去了。
	var answer struct {
		Code      string `json:"Code"`
		Message   string `json:"Message"`
		RequestId string `json:"RequestId"`
		BizId     string `json:"BizId"`
	}
	if err = json.Unmarshal(body, &answer); err != nil {
		return fmt.Errorf("peer answer is not JSON: %s", truncateAnswer(body))
	}
	if answer.Code != "OK" {
		if answer.Message == "" {
			answer.Message = "no message"
		}
		return fmt.Errorf("provider answered Code=%s: %s", answer.Code, answer.Message)
	}
	return nil
}

// SignAliyunRPC 阿里云 RPC 风格签名（HMAC-SHA1），可独立测试的纯函数。
//
// stringToSign = "POST&%2F&" + percentEncode(按参数名排序的 canonical query)，
// 签名密钥是 AccessKeySecret + "&"——那个悬空的 & 是协议的一部分，不是笔误。
func SignAliyunRPC(params url.Values, accessSecret string) string {
	keys := make([]string, 0, len(params))
	for k := range params {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	var canonical strings.Builder
	for i, k := range keys {
		if i > 0 {
			canonical.WriteByte('&')
		}
		canonical.WriteString(aliyunPercentEncode(k))
		canonical.WriteByte('=')
		canonical.WriteString(aliyunPercentEncode(params.Get(k)))
	}

	stringToSign := "POST&" + aliyunPercentEncode("/") + "&" + aliyunPercentEncode(canonical.String())
	mac := hmac.New(sha1.New, []byte(accessSecret+"&"))
	mac.Write([]byte(stringToSign))
	return base64.StdEncoding.EncodeToString(mac.Sum(nil))
}

// aliyunPercentEncode 阿里云规范的百分号编码（RFC3986）：空格用 %20（不是 +）、
// 星号要转义、波浪号不转义——QueryEscape 三处都与规范不同，逐条修回来。
func aliyunPercentEncode(s string) string {
	encoded := url.QueryEscape(s)
	encoded = strings.ReplaceAll(encoded, "+", "%20")
	encoded = strings.ReplaceAll(encoded, "*", "%2A")
	encoded = strings.ReplaceAll(encoded, "%7E", "~")
	return encoded
}
