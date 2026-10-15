// Package line LINE Messaging API（push 消息），零 SDK。
package line

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

// DefaultAPIBase LINE 官方 API 域。
const DefaultAPIBase = "https://api.line.me"

const (
	pushTimeout  = 10 * time.Second
	answerLimit  = 8 * 1024
	snippetLimit = 256
)

// Config LINE 渠道配置。Target（userId / groupId / roomId）按消息传。
type Config struct {
	// Ref 渠道配置的展示标识（ID、名字均可）：只进错误文案。
	Ref string

	// ChannelAccessToken 频道的长时访问令牌（开发者控制台签发）。
	ChannelAccessToken string

	// APIBase 服务基地址，留空用官方；测试可覆盖。
	APIBase string
}

// Service LINE 渠道。无状态，可并发使用。
type Service struct {
	client *http.Client
	cfg    Config
}

// New 构造 LINE 渠道。client 传 nil 用默认超时的客户端。
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
func (s *Service) Name() string { return "line" }

// Send 给 Target 推一条文本消息（上限 5000 字符）。
// 正文 = Title + Content（有标题时空行相连）；LINE 文本消息无独立标题位。
func (s *Service) Send(ctx context.Context, msg *notify.Message) (*notify.Receipt, error) {
	receipt := &notify.Receipt{}
	if msg == nil || msg.Target == "" {
		return nil, errors.New("line destination is empty")
	}
	if s.cfg.ChannelAccessToken == "" {
		return receipt, fmt.Errorf("%w: line config incomplete: channel_access_token required", notify.ErrNotConfigured)
	}

	text := msg.Content
	if msg.Title != "" {
		text = msg.Title + "\n" + msg.Content
	}
	payload, err := json.Marshal(map[string]any{
		"to": msg.Target,
		"messages": []map[string]any{
			{"type": "text", "text": text},
		},
	})
	if err != nil {
		return receipt, fmt.Errorf("marshal payload failed: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, s.cfg.APIBase+"/v2/bot/message/push", bytes.NewReader(payload))
	if err != nil {
		return receipt, fmt.Errorf("build request failed: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+s.cfg.ChannelAccessToken)
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
