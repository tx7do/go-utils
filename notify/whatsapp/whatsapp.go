// Package whatsapp WhatsApp Business Cloud API（Meta Graph）：单条文本，零 SDK。
//
// 收件人须先与业务号码开启会话（24h 窗口或模板消息）；本包只管 text 形态直发。
package whatsapp

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
	// DefaultAPIBase Meta Graph API 域。
	DefaultAPIBase = "https://graph.facebook.com"

	// DefaultAPIVersion Graph API 版本（路径段）。
	DefaultAPIVersion = "v21.0"

	pushTimeout  = 10 * time.Second
	answerLimit  = 8 * 1024
	snippetLimit = 256
)

// Config WhatsApp 渠道配置。Target（收件号 E.164 无 + 号，如 8613800138000）按消息传。
type Config struct {
	// Ref 渠道配置的展示标识（ID、名字均可）：只进错误文案。
	Ref string

	// PhoneNumberID 业务电话号码 ID（Meta 后台的 Phone number ID）。
	PhoneNumberID string

	// AccessToken 永久系统用户令牌。
	AccessToken string

	// APIVersion Graph API 版本段，留空用默认。
	APIVersion string

	// APIBase 服务基地址，留空用官方；测试可覆盖。
	APIBase string
}

// Service WhatsApp 渠道。无状态，可并发使用。
type Service struct {
	client *http.Client
	cfg    Config
}

// New 构造 WhatsApp 渠道。client 传 nil 用默认超时的客户端。
func New(client *http.Client, cfg Config) *Service {
	if client == nil {
		client = &http.Client{Timeout: pushTimeout}
	}
	if cfg.APIBase == "" {
		cfg.APIBase = DefaultAPIBase
	}
	if cfg.APIVersion == "" {
		cfg.APIVersion = DefaultAPIVersion
	}
	return &Service{client: client, cfg: cfg}
}

// Name 渠道名。
func (s *Service) Name() string { return "whatsapp" }

// Send 给 Target 发一条文本消息。
// 正文 = Title + Content（有标题时空行相连）；Receipt.ProviderID 带回 wamid 消息 ID。
func (s *Service) Send(ctx context.Context, msg *notify.Message) (*notify.Receipt, error) {
	receipt := &notify.Receipt{}
	if msg == nil || msg.Target == "" {
		return nil, errors.New("whatsapp destination is empty")
	}
	if s.cfg.PhoneNumberID == "" || s.cfg.AccessToken == "" {
		return receipt, fmt.Errorf("%w: whatsapp config incomplete: phone_number_id/access_token required", notify.ErrNotConfigured)
	}

	body := msg.Content
	if msg.Title != "" {
		body = msg.Title + "\n" + msg.Content
	}
	payload, err := json.Marshal(map[string]any{
		"messaging_product": "whatsapp",
		"recipient_type":    "individual",
		"to":                msg.Target,
		"type":              "text",
		"text":              map[string]any{"preview_url": false, "body": body},
	})
	if err != nil {
		return receipt, fmt.Errorf("marshal payload failed: %w", err)
	}

	endpoint := fmt.Sprintf("%s/%s/%s/messages", s.cfg.APIBase, s.cfg.APIVersion, s.cfg.PhoneNumberID)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewReader(payload))
	if err != nil {
		return receipt, fmt.Errorf("build request failed: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+s.cfg.AccessToken)
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
		Messages []struct {
			ID string `json:"id"`
		} `json:"messages"`
	}
	_ = json.Unmarshal(respBody, &answer)
	if len(answer.Messages) > 0 {
		receipt.ProviderID = answer.Messages[0].ID
	}
	return receipt, nil
}

func snippet(body []byte) string {
	if len(body) > snippetLimit {
		return string(body[:snippetLimit]) + "…"
	}
	return string(body)
}
