// Package join Join by joaoapps（设备推送），零 SDK。
package join

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

// DefaultAPIBase Join 官方 API 域。
const DefaultAPIBase = "https://joinjoaomgcd.appspot.com"

const (
	pushTimeout  = 10 * time.Second
	answerLimit  = 8 * 1024
	snippetLimit = 256
)

// Config Join 配置。Target（设备 ID，多台以逗号相连）按消息传；留空推全部设备。
type Config struct {
	// Ref 渠道配置的展示标识（ID、名字均可）：只进错误文案。
	Ref string

	// APIKey 账户 API Key（Join 官网个人页获取）。
	APIKey string

	// APIBase 服务基地址，留空用官方；测试可覆盖。
	APIBase string
}

// Service Join 渠道。无状态，可并发使用。
type Service struct {
	client *http.Client
	cfg    Config
}

// New 构造 Join 渠道。client 传 nil 用默认超时的客户端。
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
func (s *Service) Name() string { return "join" }

// Send 推一条通知到 Target 设备。Join 的 text 参数 = 标题行 + \n + 正文。
func (s *Service) Send(ctx context.Context, msg *notify.Message) (*notify.Receipt, error) {
	receipt := &notify.Receipt{}
	if msg == nil || msg.Content == "" {
		return nil, errors.New("join message body is empty")
	}
	if s.cfg.APIKey == "" {
		return receipt, fmt.Errorf("%w: join config incomplete: api_key required", notify.ErrNotConfigured)
	}

	text := msg.Content
	if msg.Title != "" {
		text = msg.Title + "\n" + msg.Content
	}

	form := url.Values{}
	form.Set("apikey", s.cfg.APIKey)
	form.Set("title", msg.Title)
	form.Set("text", text)
	if msg.Target != "" {
		form.Set("deviceIds", msg.Target)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost,
		s.cfg.APIBase+"/_ah/api/messaging/v1/sendPush", strings.NewReader(form.Encode()))
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
		Success      bool   `json:"success"`
		ErrorMessage string `json:"errorMessage"`
	}
	if err = json.Unmarshal(respBody, &answer); err != nil {
		return receipt, fmt.Errorf("peer answer is not JSON: %s", snippet(respBody))
	}
	if !answer.Success {
		if answer.ErrorMessage == "" {
			answer.ErrorMessage = "no message"
		}
		return receipt, fmt.Errorf("peer answered success=false: %s", answer.ErrorMessage)
	}
	return receipt, nil
}

func snippet(body []byte) string {
	if len(body) > snippetLimit {
		return string(body[:snippetLimit]) + "…"
	}
	return string(body)
}
