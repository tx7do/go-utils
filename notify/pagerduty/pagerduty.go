// Package pagerduty PagerDuty Events API v2（trigger 告警），零 SDK。
package pagerduty

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

// DefaultAPIBase Events API 官方域。
const DefaultAPIBase = "https://events.pagerduty.com"

const (
	pushTimeout  = 10 * time.Second
	answerLimit  = 8 * 1024
	snippetLimit = 256
)

// Config PagerDuty 渠道配置。Target（dedup key）按消息传，供事件去重/关联。
type Config struct {
	// Ref 渠道配置的展示标识（ID、名字均可）：只进错误文案。
	Ref string

	// RoutingKey 集成的 Routing Key（Events API v2）。
	RoutingKey string

	// Severity 告警级别：critical / error / warning / info，留空 info。
	Severity string

	// Source 事件来源描述，留空 "notify"。
	Source string

	// APIBase 服务基地址，留空用官方；测试可覆盖。
	APIBase string
}

// Service PagerDuty 渠道。无状态，可并发使用。
type Service struct {
	client  *http.Client
	apiBase string
	cfg     Config
}

// New 构造 PagerDuty 渠道。client 传 nil 用默认超时的客户端。
func New(client *http.Client, cfg Config) *Service {
	if client == nil {
		client = &http.Client{Timeout: pushTimeout}
	}
	if cfg.APIBase == "" {
		cfg.APIBase = DefaultAPIBase
	}
	if cfg.Severity == "" {
		cfg.Severity = "info"
	}
	if cfg.Source == "" {
		cfg.Source = "notify"
	}
	return &Service{client: client, apiBase: cfg.APIBase, cfg: cfg}
}

// Name 渠道名。
func (s *Service) Name() string { return "pagerduty" }

// Send 触发一条告警。Content 为 summary（必填，一行内说清是什么）；
// Target 为 dedup key——同 key 的事件被 PagerDuty 关联为同一告警的重现，而非新告警。
// Receipt.ProviderID 带回确认后的 dedup key。
func (s *Service) Send(ctx context.Context, msg *notify.Message) (*notify.Receipt, error) {
	receipt := &notify.Receipt{}
	if msg == nil || msg.Content == "" {
		return nil, errors.New("pagerduty summary is empty")
	}
	if s.cfg.RoutingKey == "" {
		return receipt, fmt.Errorf("%w: pagerduty config incomplete: routing_key required", notify.ErrNotConfigured)
	}

	payload := map[string]any{
		"routing_key":  s.cfg.RoutingKey,
		"event_action": "trigger",
		"payload": map[string]any{
			"summary":   msg.Content,
			"source":    s.cfg.Source,
			"severity":  s.cfg.Severity,
			"timestamp": time.Now().UTC().Format(time.RFC3339),
		},
	}
	if msg.Title != "" {
		payload["payload"].(map[string]any)["custom_details"] = map[string]string{"title": msg.Title}
	}
	if msg.Target != "" {
		payload["dedup_key"] = msg.Target
	}
	body, err := json.Marshal(payload)
	if err != nil {
		return receipt, fmt.Errorf("marshal payload failed: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, s.apiBase+"/v2/enqueue", bytes.NewReader(body))
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
		Status   string `json:"status"`
		Message  string `json:"message"`
		DedupKey string `json:"dedup_key"`
	}
	if err = json.Unmarshal(respBody, &answer); err != nil {
		return receipt, fmt.Errorf("peer answer is not JSON: %s", snippet(respBody))
	}
	if answer.Status != "success" {
		if answer.Message == "" {
			answer.Message = "no message"
		}
		return receipt, fmt.Errorf("peer answered status=%s: %s", answer.Status, answer.Message)
	}
	receipt.ProviderID = answer.DedupKey
	return receipt, nil
}

func snippet(body []byte) string {
	if len(body) > snippetLimit {
		return string(body[:snippetLimit]) + "…"
	}
	return string(body)
}
