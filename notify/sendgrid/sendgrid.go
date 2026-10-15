// Package sendgrid SendGrid Mail Send v3（API 邮件），零 SDK。
package sendgrid

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/tx7do/go-utils/notify"
)

// DefaultAPIBase SendGrid 官方 API 域。
const DefaultAPIBase = "https://api.sendgrid.com"

const (
	pushTimeout  = 10 * time.Second
	answerLimit  = 8 * 1024
	snippetLimit = 256
)

// Config SendGrid 渠道配置。Target（收件邮箱）按消息传。
type Config struct {
	// Ref 渠道配置的展示标识（ID、名字均可）：只进错误文案。
	Ref string

	// APIKey 发信 API Key。
	APIKey string

	// From 发件人地址。
	From string

	// APIBase 服务基地址，留空用官方；测试可覆盖。
	APIBase string
}

// Service SendGrid 渠道。无状态，可并发使用。
type Service struct {
	client *http.Client
	cfg    Config
}

// New 构造 SendGrid 渠道。client 传 nil 用默认超时的客户端。
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
func (s *Service) Name() string { return "sendgrid" }

// Send 给 Target（收件邮箱）发一封纯文本邮件。受理（202）即成功，
// SendGrid 的消息 ID 在响应头 X-Message-Id，进回执供对账。
func (s *Service) Send(ctx context.Context, msg *notify.Message) (*notify.Receipt, error) {
	receipt := &notify.Receipt{}
	if msg == nil || msg.Target == "" {
		return nil, errors.New("sendgrid recipient is empty")
	}
	if s.cfg.APIKey == "" || s.cfg.From == "" {
		return receipt, fmt.Errorf("%w: sendgrid config incomplete: api_key/from required", notify.ErrNotConfigured)
	}
	if msg.Content == "" {
		return nil, errors.New("sendgrid message body is empty")
	}

	payload, err := json.Marshal(map[string]any{
		"personalizations": []map[string]any{
			{"to": []map[string]string{{"email": msg.Target}}},
		},
		"from":    map[string]string{"email": s.cfg.From},
		"subject": msg.Title,
		"content": []map[string]string{
			{"type": "text/plain", "value": msg.Content},
		},
	})
	if err != nil {
		return receipt, fmt.Errorf("marshal payload failed: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, s.cfg.APIBase+"/v3/mail/send", bytes.NewReader(payload))
	if err != nil {
		return receipt, fmt.Errorf("build request failed: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+s.cfg.APIKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := s.client.Do(req)
	if err != nil {
		return receipt, err
	}
	defer func() { _ = resp.Body.Close() }()

	respBody, _ := io.ReadAll(io.LimitReader(resp.Body, answerLimit))
	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		return receipt, fmt.Errorf("peer answered %s: %s", resp.Status, snippet(respBody))
	}
	receipt.ProviderID = resp.Header.Get("X-Message-Id")
	return receipt, nil
}

func snippet(body []byte) string {
	if len(body) > snippetLimit {
		return string(body[:snippetLimit]) + "…"
	}
	return string(body)
}
