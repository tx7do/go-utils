// Package pushover Pushover 推送：裸 HTTP 表单 API，零依赖。
package pushover

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

// DefaultAPIBase Pushover 官方 API 域。
const DefaultAPIBase = "https://api.pushover.net"

const (
	pushTimeout  = 10 * time.Second
	answerLimit  = 4 * 1024
	snippetLimit = 256
)

// Config Pushover 渠道配置。Target（用户 key）按消息传——一个应用可以推给多个用户。
type Config struct {
	// Ref 渠道配置的展示标识（ID、名字均可）：只进错误文案。
	Ref string

	// AppToken 应用的 API Token。
	AppToken string

	// APIBase 服务基地址，留空用官方；测试可覆盖。
	APIBase string
}

// Service Pushover 渠道。无状态，可并发使用。
type Service struct {
	client  *http.Client
	apiBase string
	cfg     Config
}

// New 构造 Pushover 渠道。client 传 nil 用默认超时的客户端。
func New(client *http.Client, cfg Config) *Service {
	if client == nil {
		client = &http.Client{Timeout: pushTimeout}
	}
	if cfg.APIBase == "" {
		cfg.APIBase = DefaultAPIBase
	}
	return &Service{client: client, apiBase: cfg.APIBase, cfg: cfg}
}

// Name 渠道名。
func (s *Service) Name() string { return "pushover" }

// Send 给 Target（Pushover 用户 key，30 字母）推一条通知。
// Params 原生字段直传表单（priority/sound/device/url/url_title/retry/expire 等，
// 字符串值即可）——priority=-1~2 的重试语义是应用的事，本包不越权校验。
// Receipt.ProviderID 带回 Pushover 的 request ID。
func (s *Service) Send(ctx context.Context, msg *notify.Message) (*notify.Receipt, error) {
	receipt := &notify.Receipt{}
	if msg == nil || msg.Target == "" {
		return nil, errors.New("pushover user key is empty")
	}
	if s.cfg.AppToken == "" {
		return receipt, fmt.Errorf("%w: pushover config incomplete: app_token required", notify.ErrNotConfigured)
	}

	form := url.Values{}
	form.Set("token", s.cfg.AppToken)
	form.Set("user", msg.Target)
	form.Set("message", msg.Content)
	if msg.Title != "" {
		form.Set("title", msg.Title)
	}
	for k, v := range msg.Params {
		form.Set(k, v)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, s.apiBase+"/1/messages.json", strings.NewReader(form.Encode()))
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

	// status==1 才是发出去了；失败时 errors 数组点名原因。
	var answer struct {
		Status  int      `json:"status"`
		Request string   `json:"request"`
		Errors  []string `json:"errors"`
	}
	if err = json.Unmarshal(respBody, &answer); err != nil {
		return receipt, fmt.Errorf("peer answer is not JSON: %s", snippet(respBody))
	}
	if answer.Status != 1 {
		return receipt, fmt.Errorf("peer answered status=0: %s", strings.Join(answer.Errors, "; "))
	}
	receipt.ProviderID = answer.Request
	return receipt, nil
}

func snippet(body []byte) string {
	if len(body) > snippetLimit {
		return string(body[:snippetLimit]) + "…"
	}
	return string(body)
}
