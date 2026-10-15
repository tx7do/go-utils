// Package vonage Vonage（原 Nexmo）SMS：API Key/Secret 表单，零 SDK。
package vonage

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/tx7do/go-utils/notify"
)

// DefaultAPIBase Vonage 官方 REST 域。
const DefaultAPIBase = "https://rest.nexmo.com"

const (
	pushTimeout  = 10 * time.Second
	answerLimit  = 8 * 1024
	snippetLimit = 256
)

// Config Vonage 短信配置。Target（收件号 E.164，无 + 前缀）按消息传。
type Config struct {
	// Ref 渠道配置的展示标识（ID、名字均可）：只进错误文案。
	Ref string

	// APIKey / APISecret 账户凭据。
	APIKey    string
	APISecret string

	// From 发送方：已注册的发信 ID 或 Vonage 号码（E.164）。
	From string

	// APIBase 服务基地址，留空用官方；测试可覆盖。
	APIBase string
}

// Service Vonage 短信渠道。无状态，可并发使用。
type Service struct {
	client *http.Client
	cfg    Config
}

// New 构造 Vonage 短信渠道。client 传 nil 用默认超时的客户端。
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
func (s *Service) Name() string { return "vonage" }

// Send 给 Target 发一条短信。正文 = Title + Content（有标题时空行相连）。
// Receipt.ProviderID 带回 Vonage 的 message-id。
func (s *Service) Send(ctx context.Context, msg *notify.Message) (*notify.Receipt, error) {
	receipt := &notify.Receipt{}
	if msg == nil || msg.Target == "" {
		return nil, errors.New("vonage destination is empty")
	}
	if s.cfg.APIKey == "" || s.cfg.APISecret == "" || s.cfg.From == "" {
		return receipt, fmt.Errorf("%w: vonage config incomplete: api_key/api_secret/from required", notify.ErrNotConfigured)
	}

	body := msg.Content
	if msg.Title != "" {
		body = msg.Title + "\n" + msg.Content
	}

	form := url.Values{}
	form.Set("api_key", s.cfg.APIKey)
	form.Set("api_secret", s.cfg.APISecret)
	form.Set("to", strings.TrimPrefix(msg.Target, "+"))
	form.Set("from", s.cfg.From)
	form.Set("text", body)

	req, err := http.NewRequestWithContext(ctx, http.MethodPost,
		s.cfg.APIBase+"/sms/json", strings.NewReader(form.Encode()))
	if err != nil {
		return receipt, fmt.Errorf("build request failed: %w", err)
	}
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

	var answer struct {
		Messages []struct {
			Status    string `json:"status"`
			MessageID string `json:"message-id"`
			ErrorText string `json:"error-text"`
		} `json:"messages"`
	}
	if err = json.Unmarshal(respBody, &answer); err != nil {
		return receipt, fmt.Errorf("peer answer is not JSON: %s", snippet(respBody))
	}
	if len(answer.Messages) == 0 {
		return receipt, fmt.Errorf("peer answered without messages: %s", snippet(respBody))
	}
	m := answer.Messages[0]
	if m.Status != "0" {
		if m.ErrorText == "" {
			m.ErrorText = "no message"
		}
		return receipt, fmt.Errorf("peer answered status=%s: %s", m.Status, m.ErrorText)
	}
	receipt.ProviderID = m.MessageID
	return receipt, nil
}

func snippet(body []byte) string {
	if len(body) > snippetLimit {
		return string(body[:snippetLimit]) + "…"
	}
	return string(body)
}
