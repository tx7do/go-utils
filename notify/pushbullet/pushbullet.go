// Package pushbullet Pushbullet 推送（note），零 SDK。
package pushbullet

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

// DefaultAPIBase Pushbullet 官方 API 域。
const DefaultAPIBase = "https://api.pushbullet.com"

const (
	pushTimeout  = 10 * time.Second
	answerLimit  = 8 * 1024
	snippetLimit = 256
)

// Config Pushbullet 配置。Target（设备标识 device_iden，可选）按消息传；
// 留空推到账号的所有设备。
type Config struct {
	// Ref 渠道配置的展示标识（ID、名字均可）：只进错误文案。
	Ref string

	// APIKey 账户 Access Token。
	APIKey string

	// APIBase 服务基地址，留空用官方；测试可覆盖。
	APIBase string
}

// Service Pushbullet 渠道。无状态，可并发使用。
type Service struct {
	client *http.Client
	cfg    Config
}

// New 构造 Pushbullet 渠道。client 传 nil 用默认超时的客户端。
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
func (s *Service) Name() string { return "pushbullet" }

// Send 推一条 note。Target 为目标设备 device_iden（可选，留空全部设备）。
// Receipt.ProviderID 带回 Pushbullet 的 iden。
func (s *Service) Send(ctx context.Context, msg *notify.Message) (*notify.Receipt, error) {
	receipt := &notify.Receipt{}
	if msg == nil || msg.Content == "" {
		return nil, errors.New("pushbullet message body is empty")
	}
	if s.cfg.APIKey == "" {
		return receipt, fmt.Errorf("%w: pushbullet config incomplete: api_key required", notify.ErrNotConfigured)
	}

	payload := map[string]any{
		"type":  "note",
		"title": msg.Title,
		"body":  msg.Content,
	}
	if msg.Target != "" {
		payload["device_iden"] = msg.Target
	}
	body, err := json.Marshal(payload)
	if err != nil {
		return receipt, fmt.Errorf("marshal payload failed: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, s.cfg.APIBase+"/v2/pushes", bytes.NewReader(body))
	if err != nil {
		return receipt, fmt.Errorf("build request failed: %w", err)
	}
	req.Header.Set("Access-Token", s.cfg.APIKey)
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
		IDEN string `json:"iden"`
	}
	_ = json.Unmarshal(respBody, &answer)
	receipt.ProviderID = answer.IDEN
	return receipt, nil
}

func snippet(body []byte) string {
	if len(body) > snippetLimit {
		return string(body[:snippetLimit]) + "…"
	}
	return string(body)
}
