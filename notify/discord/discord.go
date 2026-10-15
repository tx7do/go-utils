// Package discord Discord Webhook：裸 HTTP POST，零依赖。
//
// ?wait=true 让 Discord 回完整消息对象，message id 进回执供对账。
package discord

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

// Config Discord 渠道配置。Webhook URL 即收件身份（创建时绑定频道），Target 无意义。
type Config struct {
	// Ref 渠道配置的展示标识（ID、名字均可）：只进错误文案。
	Ref string

	// WebhookURL 完整的 webhook 地址（https://discord.com/api/webhooks/...）。
	WebhookURL string
}

// Service Discord 渠道。无状态，可并发使用。
type Service struct {
	client *http.Client
	cfg    Config
}

// New 构造 Discord 渠道。client 传 nil 用默认超时的客户端。
func New(client *http.Client, cfg Config) *Service {
	if client == nil {
		client = &http.Client{Timeout: pushTimeout}
	}
	return &Service{client: client, cfg: cfg}
}

// Name 渠道名。
func (s *Service) Name() string { return "discord" }

// Send 发一条消息。带 Title 时走 embed（标题 + 描述的卡片形态），否则纯文本。
// Receipt.ProviderID 带回 Discord 的 message id。
func (s *Service) Send(ctx context.Context, msg *notify.Message) (*notify.Receipt, error) {
	receipt := &notify.Receipt{}
	if msg == nil || msg.Content == "" {
		return nil, errors.New("discord message body is empty")
	}
	if s.cfg.WebhookURL == "" {
		return receipt, fmt.Errorf("%w: discord config incomplete: webhook_url required", notify.ErrNotConfigured)
	}

	var payload map[string]any
	if msg.Title != "" {
		payload = map[string]any{
			"embeds": []map[string]any{{
				"title":       msg.Title,
				"description": msg.Content,
			}},
		}
	} else {
		payload = map[string]any{"content": msg.Content}
	}
	body, err := json.Marshal(payload)
	if err != nil {
		return receipt, fmt.Errorf("marshal payload failed: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, s.cfg.WebhookURL+"?wait=true", bytes.NewReader(body))
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
		ID string `json:"id"`
	}
	_ = json.Unmarshal(respBody, &answer)
	receipt.ProviderID = answer.ID
	return receipt, nil
}

func snippet(body []byte) string {
	if len(body) > snippetLimit {
		return string(body[:snippetLimit]) + "…"
	}
	return string(body)
}
