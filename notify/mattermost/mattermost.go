// Package mattermost Mattermost 入站 Webhook：Slack 兼容形态，零 SDK。
package mattermost

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

// Config Mattermost 渠道配置。Webhook URL 即收件身份（集成页创建，可含 channel 覆盖）。
type Config struct {
	// Ref 渠道配置的展示标识（ID、名字均可）：只进错误文案。
	Ref string

	// WebhookURL 完整的入站地址（https://host/hooks/<key>）。
	WebhookURL string

	// Channel 可选：覆盖 webhook 默认频道（@user 私聊或 #channel）。
	Channel string
}

// Service Mattermost 渠道。无状态，可并发使用。
type Service struct {
	client *http.Client
	cfg    Config
}

// New 构造 Mattermost 渠道。client 传 nil 用默认超时的客户端。
func New(client *http.Client, cfg Config) *Service {
	if client == nil {
		client = &http.Client{Timeout: pushTimeout}
	}
	return &Service{client: client, cfg: cfg}
}

// Name 渠道名。
func (s *Service) Name() string { return "mattermost" }

// Send 发一条 markdown 文本。Title 以粗体行起头（Mattermost 的 **text** 即粗体）。
func (s *Service) Send(ctx context.Context, msg *notify.Message) (*notify.Receipt, error) {
	receipt := &notify.Receipt{}
	if msg == nil || msg.Content == "" {
		return nil, errors.New("mattermost message body is empty")
	}
	if s.cfg.WebhookURL == "" {
		return receipt, fmt.Errorf("%w: mattermost config incomplete: webhook_url required", notify.ErrNotConfigured)
	}

	text := msg.Content
	if msg.Title != "" {
		text = "**" + msg.Title + "**\n" + msg.Content
	}
	payload := map[string]any{"text": text}
	if s.cfg.Channel != "" {
		payload["channel"] = s.cfg.Channel
	}
	body, err := json.Marshal(payload)
	if err != nil {
		return receipt, fmt.Errorf("marshal payload failed: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, s.cfg.WebhookURL, bytes.NewReader(body))
	if err != nil {
		return receipt, fmt.Errorf("build request failed: %w", err)
	}
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
	return receipt, nil
}

func snippet(body []byte) string {
	if len(body) > snippetLimit {
		return string(body[:snippetLimit]) + "…"
	}
	return string(body)
}
