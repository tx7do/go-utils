// Package opsgenie Opsgenie Alerts v2（创建告警），零 SDK。
package opsgenie

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

// DefaultAPIBase Opsgenie 官方 API 域（欧盟租户覆写 api.eu.opsgenie.com）。
const DefaultAPIBase = "https://api.opsgenie.com"

const (
	pushTimeout  = 10 * time.Second
	answerLimit  = 8 * 1024
	snippetLimit = 256
)

// Config Opsgenie 渠道配置。Target（alias）按消息传，供告警去重/关联。
type Config struct {
	// Ref 渠道配置的展示标识（ID、名字均可）：只进错误文案。
	Ref string

	// APIKey 集成的 API Key（请求头 GenieKey）。
	APIKey string

	// APIBase 服务基地址，留空用官方；欧盟租户与测试可覆盖。
	APIBase string
}

// Service Opsgenie 渠道。无状态，可并发使用。
type Service struct {
	client *http.Client
	cfg    Config
}

// New 构造 Opsgenie 渠道。client 传 nil 用默认超时的客户端。
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
func (s *Service) Name() string { return "opsgenie" }

// Send 创建一条告警。message 必填（一行说清是什么）——Title 优先、Content 兜底；
// Content 全文进 description。Target 为 alias：同 alias 的告警被 Opsgenie 去重关联。
// Receipt.ProviderID 带回 requestId。
func (s *Service) Send(ctx context.Context, msg *notify.Message) (*notify.Receipt, error) {
	receipt := &notify.Receipt{}
	if msg == nil || msg.Content == "" {
		return nil, errors.New("opsgenie message is empty")
	}
	if s.cfg.APIKey == "" {
		return receipt, fmt.Errorf("%w: opsgenie config incomplete: api_key required", notify.ErrNotConfigured)
	}

	message := msg.Title
	if message == "" {
		message = msg.Content
	}
	payload := map[string]any{
		"message":     message,
		"description": msg.Content,
	}
	if msg.Target != "" {
		payload["alias"] = msg.Target
	}
	body, err := json.Marshal(payload)
	if err != nil {
		return receipt, fmt.Errorf("marshal payload failed: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost,
		strings.TrimRight(s.cfg.APIBase, "/")+"/v2/alerts", bytes.NewReader(body))
	if err != nil {
		return receipt, fmt.Errorf("build request failed: %w", err)
	}
	req.Header.Set("Authorization", "GenieKey "+s.cfg.APIKey)
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
		Result    string `json:"result"`
		RequestID string `json:"requestId"`
	}
	_ = json.Unmarshal(respBody, &answer)
	receipt.ProviderID = answer.RequestID
	return receipt, nil
}

func snippet(body []byte) string {
	if len(body) > snippetLimit {
		return string(body[:snippetLimit]) + "…"
	}
	return string(body)
}
