package sms

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"time"
)

// 腾讯云短信（sms SendSms，TC3-HMAC-SHA256 签名）。与阿里云最大的协议差异：
// 签名不只盖参数，还盖整个 JSON 请求体；手机号只认 E.164（+86 前缀由 NormalizePhoneNumber 补齐）。
const (
	tencentSmsHost          = "sms.tencentcloudapi.com"
	tencentSmsService       = "sms"
	tencentSmsVersion       = "2021-01-11"
	tencentSmsDefaultRegion = "ap-guangzhou"

	// tencentAnswerBytes 读对端应答的上限：结论是一个几百字节的小 JSON。
	tencentAnswerBytes = 4 * 1024
)

// sendTencent 给单个手机号下发一条短信。
// phone 收 E.164（带 +86 前缀，由 NormalizePhoneNumber 归一）。
// templateParams 为位置参数（按参数名排序取值，见 templateParamsOrdered）。
func (s *Service) sendTencent(ctx context.Context, phone string, templateParams []string) error {
	region := s.cfg.Region
	if region == "" {
		region = tencentSmsDefaultRegion
	}

	payload, err := json.Marshal(struct {
		PhoneNumberSet   []string `json:"PhoneNumberSet"`
		SmsSdkAppId      string   `json:"SmsSdkAppId"`
		SignName         string   `json:"SignName"`
		TemplateId       string   `json:"TemplateId"`
		TemplateParamSet []string `json:"TemplateParamSet"`
	}{
		PhoneNumberSet:   []string{phone},
		SmsSdkAppId:      s.cfg.AppID,
		SignName:         s.cfg.SignName,
		TemplateId:       s.cfg.TemplateCode,
		TemplateParamSet: templateParams,
	})
	if err != nil {
		return fmt.Errorf("encode request payload failed: %w", err)
	}

	now := time.Now()
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, "https://"+tencentSmsHost+"/", bytes.NewReader(payload))
	if err != nil {
		return fmt.Errorf("build request failed: %w", err)
	}
	req.Header.Set("Content-Type", "application/json; charset=utf-8")
	req.Header.Set("X-TC-Action", "SendSms")
	req.Header.Set("X-TC-Version", tencentSmsVersion)
	req.Header.Set("X-TC-Timestamp", strconv.FormatInt(now.Unix(), 10))
	req.Header.Set("X-TC-Region", region)
	req.Header.Set("Authorization", SignTencentTC3(s.cfg.AccessKeyID, s.cfg.AccessSecret, now, payload))

	resp, err := s.client.Do(req)
	if err != nil {
		return err
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	body, readErr := io.ReadAll(io.LimitReader(resp.Body, tencentAnswerBytes))
	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		return fmt.Errorf("peer answered %s: %s", resp.Status, truncateAnswer(body))
	}
	if readErr != nil && readErr != io.EOF {
		return fmt.Errorf("read peer answer: %w", readErr)
	}

	// 腾讯云两级结论都要看：API 级错误在 Response.Error；API 成功后每条号码各有一个 Code，
	// 单发只有第一项，Code=="Ok"（这个 Ok 大小写与阿里云不同）才是发出去了。
	var answer struct {
		Response struct {
			Error *struct {
				Code    string `json:"Code"`
				Message string `json:"Message"`
			} `json:"Error"`
			SendStatusSet []struct {
				Code     string `json:"Code"`
				Message  string `json:"Message"`
				SerialNo string `json:"SerialNo"`
			} `json:"SendStatusSet"`
		} `json:"Response"`
	}
	if err = json.Unmarshal(body, &answer); err != nil {
		return fmt.Errorf("peer answer is not JSON: %s", truncateAnswer(body))
	}
	if e := answer.Response.Error; e != nil {
		return fmt.Errorf("provider answered Error=%s: %s", e.Code, e.Message)
	}
	if len(answer.Response.SendStatusSet) == 0 {
		return fmt.Errorf("peer answered without SendStatusSet: %s", truncateAnswer(body))
	}
	if status := answer.Response.SendStatusSet[0]; status.Code != "Ok" {
		if status.Message == "" {
			status.Message = "no message"
		}
		return fmt.Errorf("provider answered Code=%s: %s", status.Code, status.Message)
	}
	return nil
}

// SignTencentTC3 TC3-HMAC-SHA256 请求签名，可独立测试的纯函数。
//
// 签名链四段派生（date → service → tc3_request → stringToSign），每段都是 HMAC-SHA256；
// Credential 里必须带 AccessKeyId，它参与"这份签名是谁的"的判定。
func SignTencentTC3(accessKeyID, accessSecret string, now time.Time, payload []byte) string {
	date := now.UTC().Format("2006-01-02")
	timestamp := strconv.FormatInt(now.Unix(), 10)

	hashedPayload := sha256Hex(payload)
	canonicalRequest := "POST\n/\n\n" +
		"content-type:application/json; charset=utf-8\n" +
		"host:" + tencentSmsHost + "\n" +
		"\n" +
		"content-type;host\n" +
		hashedPayload
	stringToSign := "TC3-HMAC-SHA256\n" + timestamp + "\n" +
		date + "/" + tencentSmsService + "/tc3_request\n" +
		sha256Hex([]byte(canonicalRequest))

	kDate := hmacSha256([]byte("TC3"+accessSecret), []byte(date))
	kService := hmacSha256(kDate, []byte(tencentSmsService))
	kSigning := hmacSha256(kService, []byte("tc3_request"))
	signature := hex.EncodeToString(hmacSha256(kSigning, []byte(stringToSign)))

	return "TC3-HMAC-SHA256 Credential=" + accessKeyID + "/" + date + "/" + tencentSmsService + "/tc3_request" +
		", SignedHeaders=content-type;host, Signature=" + signature
}

// truncateAnswer 进错误文案的响应片段截断：对端可能回一整页 HTML。
func truncateAnswer(body []byte) string {
	const limit = 512
	if len(body) <= limit {
		return string(body)
	}
	return string(body[:limit]) + "…"
}

func sha256Hex(data []byte) string {
	sum := sha256.Sum256(data)
	return hex.EncodeToString(sum[:])
}

func hmacSha256(key, data []byte) []byte {
	mac := hmac.New(sha256.New, key)
	mac.Write(data)
	return mac.Sum(nil)
}
