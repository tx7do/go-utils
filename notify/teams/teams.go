// Package teams Microsoft Teams 工作流 Webhook：Adaptive Card 形态，零依赖。
//
// O365 组连接器已于 2024 退役，现行的入站地址是 Power Automate「发布到 Web 时」
// 工作流生成的 URL，收 Adaptive Card JSON。
package teams

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

// Config Teams 渠道配置。工作流 Webhook URL 即收件身份，Target 无意义。
type Config struct {
	// Ref 渠道配置的展示标识（ID、名字均可）：只进错误文案。
	Ref string

	// WebhookURL Power Automate 工作流的入站地址（https://*.webhook.office.com/... 或 logic.azure.com）。
	WebhookURL string
}

// Service Teams 渠道。无状态，可并发使用。
type Service struct {
	client *http.Client
	cfg    Config
}

// New 构造 Teams 渠道。client 传 nil 用默认超时的客户端。
func New(client *http.Client, cfg Config) *Service {
	if client == nil {
		client = &http.Client{Timeout: pushTimeout}
	}
	return &Service{client: client, cfg: cfg}
}

// Name 渠道名。
func (s *Service) Name() string { return "teams" }

// Send 发一张 Adaptive Card：Title 为加粗大字块，Content 为正文块。
// 工作流 Webhook 只回 200（无消息 ID），Receipt 为空。
func (s *Service) Send(ctx context.Context, msg *notify.Message) (*notify.Receipt, error) {
	receipt := &notify.Receipt{}
	if msg == nil || msg.Content == "" {
		return nil, errors.New("teams message body is empty")
	}
	if s.cfg.WebhookURL == "" {
		return receipt, fmt.Errorf("%w: teams config incomplete: webhook_url required", notify.ErrNotConfigured)
	}

	cardBody := []map[string]any{{
		"type": "TextBlock",
		"text": msg.Content,
		"wrap": true,
	}}
	if msg.Title != "" {
		cardBody = append([]map[string]any{{
			"type": "TextBlock", "text": msg.Title, "weight": "Bolder", "size": "Medium", "wrap": true,
		}}, cardBody...)
	}

	payload := map[string]any{
		"type": "message",
		"attachments": []map[string]any{{
			"contentType": "application/vnd.microsoft.card.adaptive",
			"content": map[string]any{
				"type":    "AdaptiveCard",
				"version": "1.4",
				"body":    cardBody,
			},
		}},
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
