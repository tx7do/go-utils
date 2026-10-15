// Package datadog Datadog Events API（事件/告警发布），零 SDK。
package datadog

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

// DefaultAPIBase Datadog US 站点域；欧盟 us3/us5 站点用 APIBase 覆写。
const DefaultAPIBase = "https://api.datadoghq.com"

const (
	pushTimeout  = 10 * time.Second
	answerLimit  = 8 * 1024
	snippetLimit = 256
)

// Config Datadog 配置。个人/站点点对点：Target 无意义（APIKey 即租户身份）。
type Config struct {
	// Ref 渠道配置的展示标识（ID、名字均可）：只进错误文案。
	Ref string

	// APIKey 组织的 API Key。
	APIKey string

	// Priority 事件优先级："normal" / "low"，留空 normal。
	Priority string

	// Tags 事件标签（可选）。
	Tags []string

	// APIBase 站点域，留空用 US；eu/us3/us5/ddog-gov 各站点按需覆盖。
	APIBase string
}

// Service Datadog 渠道。无状态，可并发使用。
type Service struct {
	client *http.Client
	cfg    Config
}

// New 构造 Datadog 渠道。client 传 nil 用默认超时的客户端。
func New(client *http.Client, cfg Config) *Service {
	if client == nil {
		client = &http.Client{Timeout: pushTimeout}
	}
	if cfg.APIBase == "" {
		cfg.APIBase = DefaultAPIBase
	}
	if cfg.Priority == "" {
		cfg.Priority = "normal"
	}
	return &Service{client: client, cfg: cfg}
}

// Name 渠道名。
func (s *Service) Name() string { return "datadog" }

// Send 发布一条事件。Target（可选）作为事件的 host 字段——事件归属到哪台主机。
// Receipt.ProviderID 带回 Datadog 的事件 ID。
func (s *Service) Send(ctx context.Context, msg *notify.Message) (*notify.Receipt, error) {
	receipt := &notify.Receipt{}
	if msg == nil || msg.Content == "" {
		return nil, errors.New("datadog event text is empty")
	}
	if s.cfg.APIKey == "" {
		return receipt, fmt.Errorf("%w: datadog config incomplete: api_key required", notify.ErrNotConfigured)
	}

	payload := map[string]any{
		"title":    msg.Title,
		"text":     msg.Content,
		"priority": s.cfg.Priority,
	}
	if msg.Target != "" {
		payload["host"] = msg.Target
	}
	if len(s.cfg.Tags) > 0 {
		payload["tags"] = s.cfg.Tags
	}
	body, err := json.Marshal(payload)
	if err != nil {
		return receipt, fmt.Errorf("marshal payload failed: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, s.cfg.APIBase+"/api/v1/events", bytes.NewReader(body))
	if err != nil {
		return receipt, fmt.Errorf("build request failed: %w", err)
	}
	req.Header.Set("DD-API-KEY", s.cfg.APIKey)
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
		Status string `json:"status"`
		Event  struct {
			ID int64 `json:"id"`
		} `json:"event"`
		Errors []string `json:"errors"`
	}
	if err = json.Unmarshal(respBody, &answer); err != nil {
		return receipt, fmt.Errorf("peer answer is not JSON: %s", snippet(respBody))
	}
	if answer.Status != "ok" {
		msg := "no message"
		if len(answer.Errors) > 0 {
			msg = answer.Errors[0]
		}
		return receipt, fmt.Errorf("peer answered status=%s: %s", answer.Status, msg)
	}
	if answer.Event.ID != 0 {
		receipt.ProviderID = fmt.Sprintf("%d", answer.Event.ID)
	}
	return receipt, nil
}

func snippet(body []byte) string {
	if len(body) > snippetLimit {
		return string(body[:snippetLimit]) + "…"
	}
	return string(body)
}
