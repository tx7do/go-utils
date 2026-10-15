// Package homeassistant Home Assistant 通知服务（REST API），零 SDK。
package homeassistant

import (
	"bytes"
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

const (
	pushTimeout  = 10 * time.Second
	answerLimit  = 8 * 1024
	snippetLimit = 256
)

// ServiceDefault 默认 notify 服务（手机伴侣 App 的通知平台）。
const ServiceDefault = "persistent_notification"

// Config Home Assistant 配置。个人/家庭推送：Target 无意义（Service 决定通知面）。
type Config struct {
	// Ref 渠道配置的展示标识（ID、名字均可）：只进错误文案。
	Ref string

	// ServerURL HA 实例基地址（必填，如 http://homeassistant.local:8123）。
	ServerURL string

	// Token 长期访问令牌（Long-Lived Access Token）。
	Token string

	// Service notify 服务名（notify.persistent_notification / notify.mobile_app_xxx）。
	Service string

	// APIBase 兼容字段：与 ServerURL 等价（二选一配置均可）。
	APIBase string
}

// Service HA 通知渠道。无状态，可并发使用。
type HASService struct {
	client *http.Client
	cfg    Config
}

// New 构造 Home Assistant 通知渠道。client 传 nil 用默认超时的客户端。
func New(client *http.Client, cfg Config) *HASService {
	if client == nil {
		client = &http.Client{Timeout: pushTimeout}
	}
	if cfg.ServerURL == "" {
		cfg.ServerURL = cfg.APIBase
	}
	if cfg.Service == "" {
		cfg.Service = ServiceDefault
	}
	return &HASService{client: client, cfg: cfg}
}

// Name 渠道名。
func (s *HASService) Name() string { return "homeassistant" }

// Send 调一次 notify 服务。Target 可选——填了会放进 data.target（按平台解释）。
// Params 原样进 data（各家 notify 平台的自定义字段，如 markdown 的 title/或推送渠道）。
func (s *HASService) Send(ctx context.Context, msg *notify.Message) (*notify.Receipt, error) {
	receipt := &notify.Receipt{}
	if msg == nil || msg.Content == "" {
		return nil, errors.New("homeassistant message body is empty")
	}
	if s.cfg.ServerURL == "" || s.cfg.Token == "" {
		return receipt, fmt.Errorf("%w: homeassistant config incomplete: server_url/token required", notify.ErrNotConfigured)
	}

	data := map[string]any{}
	for k, v := range msg.Params {
		data[k] = v
	}
	if msg.Target != "" {
		data["target"] = msg.Target
	}

	payload := map[string]any{"message": msg.Content}
	if msg.Title != "" {
		payload["title"] = msg.Title
	}
	if len(data) > 0 {
		payload["data"] = data
	}
	body, err := json.Marshal(payload)
	if err != nil {
		return receipt, fmt.Errorf("marshal payload failed: %w", err)
	}

	endpoint := fmt.Sprintf("%s/api/services/notify/%s",
		strings.TrimRight(s.cfg.ServerURL, "/"), url.PathEscape(s.cfg.Service))
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewReader(body))
	if err != nil {
		return receipt, fmt.Errorf("build request failed: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+s.cfg.Token)
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
