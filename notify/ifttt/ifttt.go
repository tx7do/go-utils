// Package ifttt IFTTT Maker Webhooks（自动化触发），零 SDK。
package ifttt

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

// DefaultAPIBase IFTTT Maker Webhooks 官方域。
const DefaultAPIBase = "https://maker.ifttt.com"

const (
	pushTimeout  = 10 * time.Second
	answerLimit  = 4 * 1024
	snippetLimit = 256
)

// Config IFTTT 配置。Target（event 事件名）按消息传。
type Config struct {
	// Ref 渠道配置的展示标识（ID、名字均可）：只进错误文案。
	Ref string

	// APIKey 账户的 Webhooks Key。
	APIKey string

	// APIBase 服务基地址，留空用官方；测试可覆盖。
	APIBase string
}

// Service IFTTT 渠道。无状态，可并发使用。
type Service struct {
	client *http.Client
	cfg    Config
}

// New 构造 IFTTT 渠道。client 传 nil 用默认超时的客户端。
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
func (s *Service) Name() string { return "ifttt" }

// Send 触发一次 Target 事件。值槽映射：value1=Title、value2=Content、value3=Params["value3"]。
// IFTTT 只回 200 "Congratulations!"，无消息 ID，Receipt 为空。
func (s *Service) Send(ctx context.Context, msg *notify.Message) (*notify.Receipt, error) {
	receipt := &notify.Receipt{}
	if msg == nil || msg.Target == "" {
		return nil, errors.New("ifttt event is empty")
	}
	if s.cfg.APIKey == "" {
		return receipt, fmt.Errorf("%w: ifttt config incomplete: api_key required", notify.ErrNotConfigured)
	}

	payload := map[string]string{
		"value1": msg.Title,
		"value2": msg.Content,
	}
	if v, ok := msg.Params["value3"]; ok {
		payload["value3"] = v
	}
	body, err := json.Marshal(payload)
	if err != nil {
		return receipt, fmt.Errorf("marshal payload failed: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost,
		fmt.Sprintf("%s/trigger/%s/json?key=%s", s.cfg.APIBase, msg.Target, s.cfg.APIKey), bytes.NewReader(body))
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
	return receipt, nil
}

func snippet(body []byte) string {
	if len(body) > snippetLimit {
		return string(body[:snippetLimit]) + "…"
	}
	return string(body)
}
