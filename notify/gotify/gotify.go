// Package gotify 自建推送服务（Gotify Server）：application token + /message，零依赖。
package gotify

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"github.com/tx7do/go-utils/notify"
)

const (
	pushTimeout  = 10 * time.Second
	answerLimit  = 4 * 1024
	snippetLimit = 256
)

// Config Gotify 渠道配置。Gotify 的 application token 即收件身份，
// 所以 Target 在本渠道无意义——一个 Service 对应一个应用。
type Config struct {
	// Ref 渠道配置的展示标识（ID、名字均可）：只进错误文案。
	Ref string

	// ServerURL 自建 Gotify 服务基地址（必填——Gotify 没有官方公有域）。
	ServerURL string

	// AppToken Gotify 应用的 application token。
	AppToken string

	// Priority 默认优先级（0~10，各家客户端对高优先级的呈现不同）；
	// Message.Params["priority"] 可按条覆盖。
	Priority int
}

// Service Gotify 渠道。无状态，可并发使用。
type Service struct {
	client *http.Client
	cfg    Config
}

// New 构造 Gotify 渠道。client 传 nil 用默认超时的客户端。
func New(client *http.Client, cfg Config) *Service {
	if client == nil {
		client = &http.Client{Timeout: pushTimeout}
	}
	return &Service{client: client, cfg: cfg}
}

// Name 渠道名。
func (s *Service) Name() string { return "gotify" }

// Send 给 Target 所在的 Gotify 应用发一条消息（Target 忽略——收件身份是 AppToken）。
// Receipt.ProviderID 带回 Gotify 的消息 ID。
func (s *Service) Send(ctx context.Context, msg *notify.Message) (*notify.Receipt, error) {
	receipt := &notify.Receipt{}
	if msg == nil || msg.Content == "" {
		return nil, errors.New("gotify message body is empty")
	}
	if s.cfg.ServerURL == "" || s.cfg.AppToken == "" {
		return receipt, fmt.Errorf("%w: gotify config incomplete: server_url/app_token required", notify.ErrNotConfigured)
	}

	payload := map[string]any{
		"title":   msg.Title,
		"message": msg.Content,
	}
	priority := s.cfg.Priority
	if p, ok := msg.Params["priority"]; ok {
		n, err := strconv.Atoi(p)
		if err != nil {
			return receipt, fmt.Errorf("params[priority] %q is not a number: %w", p, err)
		}
		priority = n
	}
	if priority != 0 {
		payload["priority"] = priority
	}
	body, err := json.Marshal(payload)
	if err != nil {
		return receipt, fmt.Errorf("marshal payload failed: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, s.cfg.ServerURL+"/message?token="+url.QueryEscape(s.cfg.AppToken), bytes.NewReader(body))
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

	var answer struct {
		ID int64 `json:"id"`
	}
	_ = json.Unmarshal(respBody, &answer)
	if answer.ID != 0 {
		receipt.ProviderID = strconv.FormatInt(answer.ID, 10)
	}
	return receipt, nil
}

func snippet(body []byte) string {
	if len(body) > snippetLimit {
		return string(body[:snippetLimit]) + "…"
	}
	return string(body)
}
