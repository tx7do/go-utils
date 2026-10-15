// Package mailgun Mailgun Messages API（Basic Auth 表单邮件），零 SDK。
package mailgun

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

// DefaultAPIBase Mailgun 官方 API 域（欧盟租户覆写 https://api.eu.mailgun.net）。
const DefaultAPIBase = "https://api.mailgun.net"

const (
	pushTimeout  = 10 * time.Second
	answerLimit  = 8 * 1024
	snippetLimit = 256
)

// Config Mailgun 渠道配置。Target（收件邮箱）按消息传。
type Config struct {
	// Ref 渠道配置的展示标识（ID、名字均可）：只进错误文案。
	Ref string

	// Domain 已验证的发信域名（v3 路径段）。
	Domain string

	// From 发件人地址。
	From string

	// APIKey 域 API 密钥（Basic Auth 的密码，用户名固定 "api"）。
	APIKey string

	// APIBase 服务基地址，留空用官方；欧盟租户与测试可覆盖。
	APIBase string
}

// Service Mailgun 渠道。无状态，可并发使用。
type Service struct {
	client *http.Client
	cfg    Config
}

// New 构造 Mailgun 渠道。client 传 nil 用默认超时的客户端。
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
func (s *Service) Name() string { return "mailgun" }

// Send 给 Target（收件邮箱）发一封纯文本邮件。
// Receipt.ProviderID 带回 Mailgun 的消息 id（形如 <2026...@domain>）。
func (s *Service) Send(ctx context.Context, msg *notify.Message) (*notify.Receipt, error) {
	receipt := &notify.Receipt{}
	if msg == nil || msg.Target == "" {
		return nil, errors.New("mailgun recipient is empty")
	}
	if s.cfg.Domain == "" || s.cfg.APIKey == "" {
		return receipt, fmt.Errorf("%w: mailgun config incomplete: domain/api_key required", notify.ErrNotConfigured)
	}
	if msg.Content == "" {
		return nil, errors.New("mailgun message body is empty")
	}

	form := url.Values{}
	form.Set("from", s.cfg.From)
	form.Set("to", msg.Target)
	form.Set("subject", msg.Title)
	form.Set("text", msg.Content)

	endpoint := fmt.Sprintf("%s/v3/%s/messages", strings.TrimRight(s.cfg.APIBase, "/"), url.PathEscape(s.cfg.Domain))
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, strings.NewReader(form.Encode()))
	if err != nil {
		return receipt, fmt.Errorf("build request failed: %w", err)
	}
	req.SetBasicAuth("api", s.cfg.APIKey)
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
		ID      string `json:"id"`
		Message string `json:"message"`
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
