// Package postmark Postmark 邮件 API，零 SDK。
package postmark

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

// DefaultAPIBase Postmark 官方 API 域。
const DefaultAPIBase = "https://api.postmarkapp.com"

const (
	pushTimeout  = 10 * time.Second
	answerLimit  = 8 * 1024
	snippetLimit = 256
)

// Config Postmark 配置。Target（收件邮箱）按消息传。
type Config struct {
	// Ref 渠道配置的展示标识（ID、名字均可）：只进错误文案。
	Ref string

	// ServerToken 服务级令牌（请求头 X-Postmark-Server-Token）。
	ServerToken string

	// From 发件人地址（或 "Name <addr>" 形态）。
	From string

	// MessageStream 可选的消息流 Tag（默认 outbound）。
	MessageStream string

	// APIBase 服务基地址，留空用官方；测试可覆盖。
	APIBase string
}

// Service Postmark 渠道。无状态，可并发使用。
type Service struct {
	client *http.Client
	cfg    Config
}

// New 构造 Postmark 渠道。client 传 nil 用默认超时的客户端。
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
func (s *Service) Name() string { return "postmark" }

// Send 给 Target（收件邮箱）发一封纯文本邮件。
// Receipt.ProviderID 带回 Postmark 的数字 MessageID。
func (s *Service) Send(ctx context.Context, msg *notify.Message) (*notify.Receipt, error) {
	receipt := &notify.Receipt{}
	if msg == nil || msg.Target == "" {
		return nil, errors.New("postmark recipient is empty")
	}
	if s.cfg.ServerToken == "" || s.cfg.From == "" {
		return receipt, fmt.Errorf("%w: postmark config incomplete: server_token/from required", notify.ErrNotConfigured)
	}
	if msg.Content == "" {
		return nil, errors.New("postmark message body is empty")
	}

	payload := map[string]any{
		"From":     s.cfg.From,
		"To":       msg.Target,
		"Subject":  msg.Title,
		"TextBody": msg.Content,
	}
	if s.cfg.MessageStream != "" {
		payload["MessageStream"] = s.cfg.MessageStream
	}
	body, err := json.Marshal(payload)
	if err != nil {
		return receipt, fmt.Errorf("marshal payload failed: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, s.cfg.APIBase+"/email", bytes.NewReader(body))
	if err != nil {
		return receipt, fmt.Errorf("build request failed: %w", err)
	}
	req.Header.Set("X-Postmark-Server-Token", s.cfg.ServerToken)
	req.Header.Set("Accept", "application/json")
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

	var answer struct {
		MessageID string `json:"MessageID"`
		ErrorCode int    `json:"ErrorCode"`
		Message   string `json:"Message"`
	}
	if err = json.Unmarshal(respBody, &answer); err != nil {
		return receipt, fmt.Errorf("peer answer is not JSON: %s", snippet(respBody))
	}
	if answer.ErrorCode != 0 {
		if answer.Message == "" {
			answer.Message = "no message"
		}
		return receipt, fmt.Errorf("peer answered error_code=%d: %s", answer.ErrorCode, answer.Message)
	}
	receipt.ProviderID = answer.MessageID
	return receipt, nil
}

func snippet(body []byte) string {
	if len(body) > snippetLimit {
		return string(body[:snippetLimit]) + "…"
	}
	return string(body)
}
