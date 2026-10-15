// Package pushdeer PushDeer（自建微信推送），零 SDK。
package pushdeer

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

// DefaultServerURL PushDeer 官方服务域（自建部署常见，ServerURL 覆写）。
const DefaultServerURL = "https://api2.pushdeer.com"

const (
	pushTimeout  = 10 * time.Second
	answerLimit  = 8 * 1024
	snippetLimit = 256
)

// Config PushDeer 配置。个人推送：Target 无意义（PushKey 即收件绑定）。
type Config struct {
	// Ref 渠道配置的展示标识（ID、名字均可）：只进错误文案。
	Ref string

	// PushKey 设备的 PushKey。
	PushKey string

	// ServerURL 服务基地址，留空用官方；自建部署覆盖。
	ServerURL string
}

// Service PushDeer 渠道。无状态，可并发使用。
type Service struct {
	client *http.Client
	cfg    Config
}

// New 构造 PushDeer 渠道。client 传 nil 用默认超时的客户端。
func New(client *http.Client, cfg Config) *Service {
	if client == nil {
		client = &http.Client{Timeout: pushTimeout}
	}
	if cfg.ServerURL == "" {
		cfg.ServerURL = DefaultServerURL
	}
	return &Service{client: client, cfg: cfg}
}

// Name 渠道名。
func (s *Service) Name() string { return "pushdeer" }

// Send 推一条消息。text 进通知预览（Title），desp 为 Markdown 正文（Content）。
func (s *Service) Send(ctx context.Context, msg *notify.Message) (*notify.Receipt, error) {
	receipt := &notify.Receipt{}
	if msg == nil || msg.Content == "" {
		return nil, errors.New("pushdeer message body is empty")
	}
	if s.cfg.PushKey == "" {
		return receipt, fmt.Errorf("%w: pushdeer config incomplete: push_key required", notify.ErrNotConfigured)
	}

	form := url.Values{}
	form.Set("pushkey", s.cfg.PushKey)
	form.Set("text", msg.Title)
	form.Set("desp", msg.Content)

	req, err := http.NewRequestWithContext(ctx, http.MethodPost,
		strings.TrimRight(s.cfg.ServerURL, "/")+"/message/push", strings.NewReader(form.Encode()))
	if err != nil {
		return receipt, fmt.Errorf("build request failed: %w", err)
	}
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
		Code  int    `json:"code"`
		Error string `json:"error"`
	}
	if err = json.Unmarshal(respBody, &answer); err != nil {
		return receipt, fmt.Errorf("peer answer is not JSON: %s", snippet(respBody))
	}
	if answer.Code != 0 {
		if answer.Error == "" {
			answer.Error = "no message"
		}
		return receipt, fmt.Errorf("peer answered code=%d: %s", answer.Code, answer.Error)
	}
	return receipt, nil
}

func snippet(body []byte) string {
	if len(body) > snippetLimit {
		return string(body[:snippetLimit]) + "…"
	}
	return string(body)
}
