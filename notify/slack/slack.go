// Package slack Slack Incoming Webhook：裸 HTTP POST，零依赖。
package slack

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
	answerLimit  = 4 * 1024
	snippetLimit = 256
)

// Config Slack 渠道配置。Incoming Webhook 的 URL 即收件身份（创建时绑定到具体频道），
// Target 无意义。Webhook URL 在 Slack 的 App 管理页创建。
type Config struct {
	// Ref 渠道配置的展示标识（ID、名字均可）：只进错误文案。
	Ref string

	// WebhookURL 完整的 incoming webhook 地址（https://hooks.slack.com/services/...）。
	WebhookURL string
}

// Service Slack 渠道。无状态，可并发使用。
type Service struct {
	client *http.Client
	cfg    Config
}

// New 构造 Slack 渠道。client 传 nil 用默认超时的客户端。
func New(client *http.Client, cfg Config) *Service {
	if client == nil {
		client = &http.Client{Timeout: pushTimeout}
	}
	return &Service{client: client, cfg: cfg}
}

// Name 渠道名。
func (s *Service) Name() string { return "slack" }

// Send 发一条 mrkdwn 文本消息。Title 以粗体行起头（Slack 无独立标题位）。
// 对端回 200 "ok" 即成功；Incoming Webhook 不回消息 ID，Receipt 为空。
func (s *Service) Send(ctx context.Context, msg *notify.Message) (*notify.Receipt, error) {
	receipt := &notify.Receipt{}
	if msg == nil || msg.Content == "" {
		return nil, errors.New("slack message body is empty")
	}
	if s.cfg.WebhookURL == "" {
		return receipt, fmt.Errorf("%w: slack config incomplete: webhook_url required", notify.ErrNotConfigured)
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
	return receipt, nil
}

func snippet(body []byte) string {
	if len(body) > snippetLimit {
		return string(body[:snippetLimit]) + "…"
	}
	return string(body)
}
