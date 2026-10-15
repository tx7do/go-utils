// Package webex Cisco Webex 消息（Bot），零 SDK。
package webex

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/tx7do/go-utils/notify"
)

// DefaultAPIBase Webex 官方 API 域。
const DefaultAPIBase = "https://webexapis.com"

const (
	pushTimeout  = 10 * time.Second
	answerLimit  = 8 * 1024
	snippetLimit = 256
)

// Config Webex 配置。Target（roomId 或人员邮箱）按消息传。
type Config struct {
	// Ref 渠道配置的展示标识（ID、名字均可）：只进错误文案。
	Ref string

	// BotToken 机器人的 Bearer 令牌。
	BotToken string

	// APIBase 服务基地址，留空用官方；测试可覆盖。
	APIBase string
}

// Service Webex 渠道。无状态，可并发使用。
type Service struct {
	client *http.Client
	cfg    Config
}

// New 构造 Webex 渠道。client 传 nil 用默认超时的客户端。
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
func (s *Service) Name() string { return "webex" }

// Send 给 Target 发一条消息：含 "@" 按 toPersonEmail（私聊人员），
// 否则按 roomId（空间）。正文 = Title + Content（有标题时空行相连）。
// Receipt.ProviderID 带回 Webex 的消息 id。
func (s *Service) Send(ctx context.Context, msg *notify.Message) (*notify.Receipt, error) {
	receipt := &notify.Receipt{}
	if msg == nil || msg.Target == "" {
		return nil, errors.New("webex destination is empty")
	}
	if s.cfg.BotToken == "" {
		return receipt, fmt.Errorf("%w: webex config incomplete: bot_token required", notify.ErrNotConfigured)
	}

	text := msg.Content
	if msg.Title != "" {
		text = msg.Title + "\n" + msg.Content
	}
	payload := map[string]any{"text": text}
	if strings.Contains(msg.Target, "@") {
		payload["toPersonEmail"] = msg.Target
	} else {
		payload["roomId"] = msg.Target
	}
	body, err := json.Marshal(payload)
	if err != nil {
		return receipt, fmt.Errorf("marshal payload failed: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost,
		strings.TrimRight(s.cfg.APIBase, "/")+"/v1/messages", bytes.NewReader(body))
	if err != nil {
		return receipt, fmt.Errorf("build request failed: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+s.cfg.BotToken)
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
