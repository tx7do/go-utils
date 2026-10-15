// Package gchat Google Chat 入站 Webhook：裸 HTTP POST，零依赖。
package gchat

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

const (
	pushTimeout  = 10 * time.Second
	answerLimit  = 8 * 1024
	snippetLimit = 256
)

// Config Google Chat 渠道配置。Webhook URL 即收件身份（空间级），Target 无意义。
// URL 从空间的「Webhook」应用创建：https://chat.googleapis.com/v1/spaces/XXX/messages?key=...&token=...
type Config struct {
	// Ref 渠道配置的展示标识（ID、名字均可）：只进错误文案。
	Ref string

	// WebhookURL 完整的入站 webhook 地址。
	WebhookURL string
}

// Service Google Chat 渠道。无状态，可并发使用。
type Service struct {
	client *http.Client
	cfg    Config
}

// New 构造 Google Chat 渠道。client 传 nil 用默认超时的客户端。
func New(client *http.Client, cfg Config) *Service {
	if client == nil {
		client = &http.Client{Timeout: pushTimeout}
	}
	return &Service{client: client, cfg: cfg}
}

// Name 渠道名。
func (s *Service) Name() string { return "gchat" }

// Send 发一条文本消息。Title 以粗体行起头（Google Chat 的 *text* 即粗体）。
// Receipt.ProviderID 带回消息资源名（spaces/XXX/messages/YYY）。
func (s *Service) Send(ctx context.Context, msg *notify.Message) (*notify.Receipt, error) {
	receipt := &notify.Receipt{}
	if msg == nil || msg.Content == "" {
		return nil, errors.New("gchat message body is empty")
	}
	if s.cfg.WebhookURL == "" {
		return receipt, fmt.Errorf("%w: gchat config incomplete: webhook_url required", notify.ErrNotConfigured)
	}

	text := msg.Content
	if msg.Title != "" {
		text = "*" + msg.Title + "*\n" + msg.Content
	}
	body, err := json.Marshal(map[string]any{"text": text})
	if err != nil {
		return receipt, fmt.Errorf("marshal payload failed: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, s.cfg.WebhookURL, bytes.NewReader(body))
	if err != nil {
		return receipt, fmt.Errorf("build request failed: %w", err)
	}
	req.Header.Set("Content-Type", "application/json; charset=utf-8")

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
		Name string `json:"name"`
	}
	_ = json.Unmarshal(respBody, &answer)
	receipt.ProviderID = answer.Name
	return receipt, nil
}

func snippet(body []byte) string {
	if len(body) > snippetLimit {
		return string(body[:snippetLimit]) + "…"
	}
	return string(body)
}
