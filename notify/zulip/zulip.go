// Package zulip Zulip 消息（stream 形态）：Basic Auth 机器人，零 SDK。
package zulip

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

const (
	pushTimeout  = 10 * time.Second
	answerLimit  = 8 * 1024
	snippetLimit = 256
)

// Config Zulip 渠道配置。Target（stream 频道名）按消息传。
type Config struct {
	// Ref 渠道配置的展示标识（ID、名字均可）：只进错误文案。
	Ref string

	// ServerURL 服务基地址（必填——Zulip 都是自建/云租户域）。
	ServerURL string

	// BotEmail / BotAPIKey 机器人凭据（Basic Auth 的用户名/密码）。
	BotEmail  string
	BotAPIKey string

	// Topic stream 消息必须有 topic；Params["topic"] 可按条覆盖。
	Topic string
}

// Service Zulip 渠道。无状态，可并发使用。
type Service struct {
	client *http.Client
	cfg    Config
}

// New 构造 Zulip 渠道。client 传 nil 用默认超时的客户端。
func New(client *http.Client, cfg Config) *Service {
	if client == nil {
		client = &http.Client{Timeout: pushTimeout}
	}
	return &Service{client: client, cfg: cfg}
}

// Name 渠道名。
func (s *Service) Name() string { return "zulip" }

// Send 给 Target（stream 频道名）发一条消息。正文 = Title + Content（有标题时空行相连）。
// Receipt.ProviderID 带回 Zulip 的消息 ID。
func (s *Service) Send(ctx context.Context, msg *notify.Message) (*notify.Receipt, error) {
	receipt := &notify.Receipt{}
	if msg == nil || msg.Target == "" {
		return nil, errors.New("zulip stream is empty")
	}
	if s.cfg.ServerURL == "" || s.cfg.BotEmail == "" || s.cfg.BotAPIKey == "" {
		return receipt, fmt.Errorf("%w: zulip config incomplete: server_url/bot_email/bot_api_key required", notify.ErrNotConfigured)
	}

	topic := s.cfg.Topic
	if t, ok := msg.Params["topic"]; ok {
		topic = t
	}
	if topic == "" {
		return receipt, fmt.Errorf("%w: zulip requires a topic: set Config.Topic or params[\"topic\"]", notify.ErrNotConfigured)
	}

	content := msg.Content
	if msg.Title != "" {
		content = msg.Title + "\n" + msg.Content
	}

	form := url.Values{}
	form.Set("type", "stream")
	form.Set("to", msg.Target)
	form.Set("topic", topic)
	form.Set("content", content)

	req, err := http.NewRequestWithContext(ctx, http.MethodPost,
		strings.TrimRight(s.cfg.ServerURL, "/")+"/api/v1/messages", strings.NewReader(form.Encode()))
	if err != nil {
		return receipt, fmt.Errorf("build request failed: %w", err)
	}
	req.SetBasicAuth(s.cfg.BotEmail, s.cfg.BotAPIKey)
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
		Result string `json:"result"`
		Msg    string `json:"msg"`
		ID     int64  `json:"id"`
	}
	if err = json.Unmarshal(respBody, &answer); err != nil {
		return receipt, fmt.Errorf("peer answer is not JSON: %s", snippet(respBody))
	}
	if answer.Result != "success" {
		if answer.Msg == "" {
			answer.Msg = "no message"
		}
		return receipt, fmt.Errorf("peer answered result=%s: %s", answer.Result, answer.Msg)
	}
	if answer.ID != 0 {
		receipt.ProviderID = fmt.Sprintf("%d", answer.ID)
	}
	return receipt, nil
}

func snippet(body []byte) string {
	if len(body) > snippetLimit {
		return string(body[:snippetLimit]) + "…"
	}
	return string(body)
}
