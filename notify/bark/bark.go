// Package Bark iOS 推送（Bark App）：裸 HTTP 调 /push 接口，零依赖。
//
// Bark 服务可自建（ServerURL 覆写）；官方域 api.day.app。
package bark

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

// DefaultServerURL Bark 官方服务域。
const DefaultServerURL = "https://api.day.app"

const (
	pushTimeout  = 10 * time.Second
	answerLimit  = 4 * 1024
	snippetLimit = 256
)

// Config Bark 渠道配置。Target（设备密钥）按消息传，不进配置——
// 一个 Service 可以服务多台设备。
type Config struct {
	// Ref 渠道配置的展示标识（ID、名字均可）：只进错误文案。
	Ref string

	// ServerURL 服务基地址，留空用官方 api.day.app；自建服务可覆盖。
	ServerURL string
}

// Service Bark 渠道。无状态，可并发使用。
type Service struct {
	client    *http.Client
	serverURL string
}

// New 构造 Bark 渠道。client 传 nil 用默认超时的客户端。
func New(client *http.Client, cfg Config) *Service {
	if client == nil {
		client = &http.Client{Timeout: pushTimeout}
	}
	if cfg.ServerURL == "" {
		cfg.ServerURL = DefaultServerURL
	}
	return &Service{client: client, serverURL: cfg.ServerURL}
}

// Name 渠道名。
func (s *Service) Name() string { return "bark" }

// Send 给 Target（Bark App 的 device_key）推一条通知。
// Params 的键原样并进出站 JSON（level/badge/sound/group/url/icon 等 Bark 原生字段，
// 字符串值即可），拼错的字段名由对端忽略——Bark 对未知字段宽容。
func (s *Service) Send(ctx context.Context, msg *notify.Message) (*notify.Receipt, error) {
	receipt := &notify.Receipt{}
	if msg == nil || msg.Target == "" {
		return nil, errors.New("bark device_key is empty")
	}

	payload := map[string]any{
		"device_key": msg.Target,
		"title":      msg.Title,
		"body":       msg.Content,
	}
	for k, v := range msg.Params {
		payload[k] = v
	}
	body, err := json.Marshal(payload)
	if err != nil {
		return receipt, fmt.Errorf("marshal payload failed: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, s.serverURL+"/push", bytes.NewReader(body))
	if err != nil {
		return receipt, fmt.Errorf("build request failed: %w", err)
	}
	req.Header.Set("Content-Type", "application/json; charset=utf-8")

	resp, err := s.client.Do(req)
	if err != nil {
		return receipt, err
	}
	defer func() { _ = resp.Body.Close() }()

	respBody, _ := io.ReadAll(io.LimitReader(resp.Body, answerLimit))
	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		return receipt, fmt.Errorf("peer answered %s: %s", resp.Status, snippet(respBody))
	}

	// Bark 把业务结论放 body 的 code 里（HTTP 恒 200 时代的行为，仍旧兼容判断）。
	var answer struct {
		Code    int    `json:"code"`
		Message string `json:"message"`
	}
	if err = json.Unmarshal(respBody, &answer); err != nil {
		return receipt, fmt.Errorf("peer answer is not JSON: %s", snippet(respBody))
	}
	if answer.Code != 200 {
		if answer.Message == "" {
			answer.Message = "no message"
		}
		return receipt, fmt.Errorf("peer answered code=%d: %s", answer.Code, answer.Message)
	}
	return receipt, nil
}

func snippet(body []byte) string {
	if len(body) > snippetLimit {
		return string(body[:snippetLimit]) + "…"
	}
	return string(body)
}
