// Package rocketchat Rocket.Chat chat.postMessage：用户头鉴权，零 SDK。
package rocketchat

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

const (
	pushTimeout  = 10 * time.Second
	answerLimit  = 8 * 1024
	snippetLimit = 256
)

// Config Rocket.Chat 渠道配置。Target（频道名 @user 或 #channel）按消息传。
type Config struct {
	// Ref 渠道配置的展示标识（ID、名字均可）：只进错误文案。
	Ref string

	// ServerURL 服务基地址（必填——Rocket.Chat 都是自建/私有化部署）。
	ServerURL string

	// UserID / AuthToken 个人访问令牌（账户页创建）。
	UserID    string
	AuthToken string
}

// Service Rocket.Chat 渠道。无状态，可并发使用。
type Service struct {
	client *http.Client
	cfg    Config
}

// New 构造 Rocket.Chat 渠道。client 传 nil 用默认超时的客户端。
func New(client *http.Client, cfg Config) *Service {
	if client == nil {
		client = &http.Client{Timeout: pushTimeout}
	}
	return &Service{client: client, cfg: cfg}
}

// Name 渠道名。
func (s *Service) Name() string { return "rocketchat" }

// Send 给 Target 发一条消息。正文 = Title + Content（有标题时空行相连）。
// Receipt.ProviderID 带回消息的 _id。
func (s *Service) Send(ctx context.Context, msg *notify.Message) (*notify.Receipt, error) {
	receipt := &notify.Receipt{}
	if msg == nil || msg.Target == "" {
		return nil, errors.New("rocketchat channel is empty")
	}
	if s.cfg.ServerURL == "" || s.cfg.UserID == "" || s.cfg.AuthToken == "" {
		return receipt, fmt.Errorf("%w: rocketchat config incomplete: server_url/user_id/auth_token required", notify.ErrNotConfigured)
	}

	text := msg.Content
	if msg.Title != "" {
		text = msg.Title + "\n" + msg.Content
	}
	payload, err := json.Marshal(map[string]any{
		"channel": msg.Target,
		"text":    text,
	})
	if err != nil {
		return receipt, fmt.Errorf("marshal payload failed: %w", err)
	}

	endpoint := strings.TrimRight(s.cfg.ServerURL, "/") + "/api/v1/chat.postMessage"
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewReader(payload))
	if err != nil {
		return receipt, fmt.Errorf("build request failed: %w", err)
	}
	req.Header.Set("X-Auth-Token", s.cfg.AuthToken)
	req.Header.Set("X-User-Id", s.cfg.UserID)
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
		Success bool `json:"success"`
		Message struct {
			ID string `json:"_id"`
		} `json:"message"`
		Error string `json:"error"`
	}
	if err = json.Unmarshal(respBody, &answer); err != nil {
		return receipt, fmt.Errorf("peer answer is not JSON: %s", snippet(respBody))
	}
	if !answer.Success {
		if answer.Error == "" {
			answer.Error = "no message"
		}
		return receipt, fmt.Errorf("peer answered success=false: %s", answer.Error)
	}
	receipt.ProviderID = answer.Message.ID
	return receipt, nil
}

func snippet(body []byte) string {
	if len(body) > snippetLimit {
		return string(body[:snippetLimit]) + "…"
	}
	return string(body)
}
