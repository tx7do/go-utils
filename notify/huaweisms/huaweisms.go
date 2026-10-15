// Package huaweisms 华为云短信 MSGSMS：WSSE 鉴权 + 模板短信，零 SDK。
//
// 发送端点即"批量发送"路径（官方单发也走它，一人一条只是 result 数组只有一项）。
// from 是签名通道号（csms 开头），不是签名名称。
package huaweisms

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"sort"
	"strings"
	"time"

	"github.com/tx7do/go-utils/notify"
	"github.com/tx7do/go-utils/notify/sms"
)

// DefaultAPIBase APP 接入地址（按申请的短信应用所在区域，控制台可查）。
const DefaultAPIBase = "https://smsapi.cn-north-4.myhuaweicloud.com"

const (
	pushTimeout  = 10 * time.Second
	answerLimit  = 8 * 1024
	snippetLimit = 256
)

// Config 华为云短信配置。Target（手机号）按消息传。
type Config struct {
	// Ref 渠道配置的展示标识（ID、名字均可）：只进错误文案。
	Ref string

	// AppKey / AppSecret APP 凭据（WSSE UsernameToken）。
	AppKey    string
	AppSecret string

	// Sender 签名通道号（csms 开头，即 from 参数）。
	Sender string

	// TemplateID 已报备的模板 ID。
	TemplateID string

	// StatusCallback 可选的状态报告回调地址。
	StatusCallback string

	// APIBase APP 接入地址，留空用默认区域；按应用所在区域覆盖。
	APIBase string
}

// Service 华为云短信渠道。无状态，可并发使用。
type Service struct {
	client *http.Client
	cfg    Config
}

// New 构造华为云短信渠道。client 传 nil 用默认超时的客户端。
func New(client *http.Client, cfg Config) *Service {
	if client == nil {
		client = &http.Client{Timeout: pushTimeout}
	}
	if cfg.APIBase == "" {
		cfg.APIBase = DefaultAPIBase
	}
	return &Service{client: client, cfg: cfg}
}

// Name 渠道名。
func (s *Service) Name() string { return "huaweisms" }

// Send 给 Target（手机号）发一条模板短信。
// Params 为模板变量：按参数名排序取值序列化成 templateParas 数组（单参数无感，
// 多参数时模板 ${1}${2} 的对应关系由排序决定，多参数请核对模板占位顺序）。
// Receipt.ProviderID 带回 smsMsgId。
func (s *Service) Send(ctx context.Context, msg *notify.Message) (*notify.Receipt, error) {
	receipt := &notify.Receipt{}
	if msg == nil || msg.Target == "" {
		return nil, errors.New("huaweisms destination is empty")
	}
	if s.cfg.AppKey == "" || s.cfg.AppSecret == "" || s.cfg.Sender == "" || s.cfg.TemplateID == "" {
		return receipt, fmt.Errorf("%w: huaweisms config incomplete: app_key/app_secret/sender/template_id required", notify.ErrNotConfigured)
	}

	// 目标归一在出站前裁决：E.164 形态（+86 可选，带着最稳）。
	_, e164, err := sms.NormalizePhoneNumber(msg.Target)
	if err != nil {
		return receipt, fmt.Errorf("%w: %v", notify.ErrNotConfigured, err)
	}

	form := url.Values{}
	form.Set("from", s.cfg.Sender)
	form.Set("to", e164)
	form.Set("templateId", s.cfg.TemplateID)
	if len(msg.Params) > 0 {
		form.Set("templateParas", templateParasJSON(msg.Params))
	}
	if s.cfg.StatusCallback != "" {
		form.Set("statusCallback", s.cfg.StatusCallback)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost,
		s.cfg.APIBase+"/sms/batchSendSms/v1", strings.NewReader(form.Encode()))
	if err != nil {
		return receipt, fmt.Errorf("build request failed: %w", err)
	}
	req.Header.Set("Authorization", `WSSE realm="SDP",profile="UsernameToken",type="Appkey"`)
	req.Header.Set("X-WSSE", buildWSSE(s.cfg.AppKey, s.cfg.AppSecret))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := s.client.Do(req)
	if err != nil {
		return receipt, err
	}
	defer func() { _ = resp.Body.Close() }()

	respBody, _ := io.ReadAll(io.LimitReader(resp.Body, answerLimit))
	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		return receipt, fmt.Errorf("peer answered %s: %s", resp.Status, snippet(respBody))
	}

	// 厂商把业务结论放 body 的 code（字符串 "000000" 为成功），HTTP 恒 200。
	var answer struct {
		Code        string `json:"code"`
		Description string `json:"description"`
		Result      []struct {
			SmsMsgID string `json:"smsMsgId"`
			Status   string `json:"status"`
		} `json:"result"`
	}
	if err = json.Unmarshal(respBody, &answer); err != nil {
		return receipt, fmt.Errorf("peer answer is not JSON: %s", snippet(respBody))
	}
	if answer.Code != "000000" {
		if answer.Description == "" {
			answer.Description = "no description"
		}
		return receipt, fmt.Errorf("peer answered code=%s: %s", answer.Code, answer.Description)
	}
	if len(answer.Result) > 0 && answer.Result[0].SmsMsgID != "" {
		receipt.ProviderID = answer.Result[0].SmsMsgID
	}
	return receipt, nil
}

// buildWSSE X-WSSE 头：PasswordDigest = base64(sha256(nonce + created + appSecret))，
// 三段原始串直接相连（不加分隔符）；created 为 UTC 的 RFC3339 秒级。
func buildWSSE(appKey, appSecret string) string {
	nonceBytes := make([]byte, 16)
	_, _ = rand.Read(nonceBytes)
	nonce := hex.EncodeToString(nonceBytes)
	created := time.Now().UTC().Format("2006-01-02T15:04:05Z")

	sum := sha256.Sum256([]byte(nonce + created + appSecret))
	digest := base64.StdEncoding.EncodeToString(sum[:])

	return `UsernameToken Username="` + appKey + `", PasswordDigest="` + digest +
		`", Nonce="` + nonce + `", Created="` + created + `"`
}

// templateParasJSON 模板变量：按参数名排序取值，序列化为 JSON 数组字符串。
// 单参数（验证码 code）无感；多参数时模板 ${1}${2} 的对应关系由这个排序决定。
func templateParasJSON(params map[string]string) string {
	keys := make([]string, 0, len(params))
	for k := range params {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	values := make([]string, 0, len(keys))
	for _, k := range keys {
		values = append(values, params[k])
	}
	encoded, err := json.Marshal(values)
	if err != nil {
		return "[]"
	}
	return string(encoded)
}

func snippet(body []byte) string {
	if len(body) > snippetLimit {
		return string(body[:snippetLimit]) + "…"
	}
	return string(body)
}
