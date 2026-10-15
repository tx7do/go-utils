// Package serverchan Server酱（sct.ftqq.com）：微信个人推送，零 SDK。
package serverchan

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

// DefaultAPIBase Server酱 Turbo 官方域。
const DefaultAPIBase = "https://sctapi.ftqq.com"

const (
	pushTimeout  = 10 * time.Second
	answerLimit  = 8 * 1024
	snippetLimit = 256
)

// Config Server酱配置。个人推送：Target 无意义（SendKey 即收件人绑定）。
type Config struct {
	// Ref 渠道配置的展示标识（ID、名字均可）：只进错误文案。
	Ref string

	// SendKey SendKey（sctpNNNNNx 形态，官网个人页获取）。
	SendKey string

	// APIBase 服务基地址，留空用官方。
	APIBase string
}

// Service Server酱渠道。无状态，可并发使用。
type Service struct {
	client *http.Client
	cfg    Config
}

// New 构造 Server酱渠道。client 传 nil 用默认超时的客户端。
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
func (s *Service) Name() string { return "serverchan" }

// Send 推一条消息到绑定的微信。title 进通知预览，desp 为 Markdown 正文。
func (s *Service) Send(ctx context.Context, msg *notify.Message) (*notify.Receipt, error) {
	receipt := &notify.Receipt{}
	if msg == nil || msg.Content == "" {
		return nil, errors.New("serverchan message body is empty")
	}
	if s.cfg.SendKey == "" {
		return receipt, fmt.Errorf("%w: serverchan config incomplete: send_key required", notify.ErrNotConfigured)
	}

	form := url.Values{}
	form.Set("title", msg.Title)
	form.Set("desp", msg.Content)

	req, err := http.NewRequestWithContext(ctx, http.MethodPost,
		fmt.Sprintf("%s/%s.send", s.cfg.APIBase, url.PathEscape(s.cfg.SendKey)), strings.NewReader(form.Encode()))
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
		Code int    `json:"code"`
		Msg  string `json:"message"`
	}
	if err = json.Unmarshal(respBody, &answer); err != nil {
		return receipt, fmt.Errorf("peer answer is not JSON: %s", snippet(respBody))
	}
	if answer.Code != 0 {
		if answer.Msg == "" {
			answer.Msg = "no message"
		}
		return receipt, fmt.Errorf("peer answered code=%d: %s", answer.Code, answer.Msg)
	}
	return receipt, nil
}

func snippet(body []byte) string {
	if len(body) > snippetLimit {
		return string(body[:snippetLimit]) + "…"
	}
	return string(body)
}
