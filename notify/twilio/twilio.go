// Package twilio Twilio SMS：裸 HTTP 调 Messages API（Basic Auth），零依赖。
package twilio

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

// DefaultAPIBase Twilio 官方 API 域。
const DefaultAPIBase = "https://api.twilio.com"

const (
	pushTimeout  = 10 * time.Second
	answerLimit  = 8 * 1024
	snippetLimit = 256
)

// Config Twilio 渠道配置。Target（收件号 E.164）按消息传。
type Config struct {
	// Ref 渠道配置的展示标识（ID、名字均可）：只进错误文案。
	Ref string

	// AccountSID / AuthToken 账号凭据。
	AccountSID string
	AuthToken  string

	// From 发送方：Twilio 号码（E.164）或已验证的 Sender ID / Messaging Service SID（MG 开头）。
	From string

	// APIBase 服务基地址，留空用官方；测试可覆盖。
	APIBase string
}

// Service Twilio 渠道。无状态，可并发使用。
type Service struct {
	client  *http.Client
	apiBase string
	cfg     Config
}

// New 构造 Twilio 渠道。client 传 nil 用默认超时的客户端。
func New(client *http.Client, cfg Config) *Service {
	if client == nil {
		client = &http.Client{Timeout: pushTimeout}
	}
	if cfg.APIBase == "" {
		cfg.APIBase = DefaultAPIBase
	}
	return &Service{client: client, apiBase: cfg.APIBase, cfg: cfg}
}

// Name 渠道名。
func (s *Service) Name() string { return "twilio" }

// Send 给 Target（收件手机号，E.164，如 +8613800138000）发一条短信。
// 正文 = Title + Content（有标题时空行相连）；Receipt.ProviderID 带回 Twilio 的消息 SID。
func (s *Service) Send(ctx context.Context, msg *notify.Message) (*notify.Receipt, error) {
	receipt := &notify.Receipt{}
	if msg == nil || msg.Target == "" {
		return nil, errors.New("twilio destination is empty")
	}
	if s.cfg.AccountSID == "" || s.cfg.AuthToken == "" || s.cfg.From == "" {
		return receipt, fmt.Errorf("%w: twilio config incomplete: account_sid/auth_token/from required", notify.ErrNotConfigured)
	}

	body := msg.Content
	if msg.Title != "" {
		body = msg.Title + "\n" + msg.Content
	}

	form := url.Values{}
	form.Set("To", msg.Target)
	form.Set("From", s.cfg.From)
	form.Set("Body", body)

	endpoint := fmt.Sprintf("%s/2010-04-01/Accounts/%s/Messages.json", s.apiBase, s.cfg.AccountSID)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, strings.NewReader(form.Encode()))
	if err != nil {
		return receipt, fmt.Errorf("build request failed: %w", err)
	}
	req.SetBasicAuth(s.cfg.AccountSID, s.cfg.AuthToken)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := s.client.Do(req)
	if err != nil {
		return receipt, err
	}
	defer func() { _ = resp.Body.Close() }()

	respBody, _ := io.ReadAll(io.LimitReader(resp.Body, answerLimit))
	// Twilio 出站消息创建成功回 201；把 2xx 都算可解析，201 之外按对端语义报。
	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		return receipt, fmt.Errorf("peer answered %s: %s", resp.Status, snippet(respBody))
	}

	var answer struct {
		SID string `json:"sid"`
	}
	_ = json.Unmarshal(respBody, &answer)
	receipt.ProviderID = answer.SID
	return receipt, nil
}

func snippet(body []byte) string {
	if len(body) > snippetLimit {
		return string(body[:snippetLimit]) + "…"
	}
	return string(body)
}
