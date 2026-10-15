// Package wxpusher 微信个人推送（WxPusher），零 SDK。
package wxpusher

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

// DefaultAPIBase WxPusher 官方域。
const DefaultAPIBase = "https://wxpusher.zjiecode.com"

// ContentText / ContentHTML / ContentMarkdown contentType 取值。
const (
	ContentText     = 1
	ContentHTML     = 2
	ContentMarkdown = 3
)

const (
	pushTimeout  = 10 * time.Second
	answerLimit  = 8 * 1024
	snippetLimit = 256
)

// Config WxPusher 配置。Target（UID，形如 UID_xxx）按消息传。
type Config struct {
	// Ref 渠道配置的展示标识（ID、名字均可）：只进错误文案。
	Ref string

	// AppToken WxPusher 应用的 APP_TOKEN。
	AppToken string

	// ContentType 正文形态：1 文本 / 2 HTML / 3 Markdown，留空按文本。
	ContentType int

	// APIBase 服务基地址，留空用官方。
	APIBase string
}

// Service WxPusher 渠道。无状态，可并发使用。
type Service struct {
	client *http.Client
	cfg    Config
}

// New 构造 WxPusher 渠道。client 传 nil 用默认超时的客户端。
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
func (s *Service) Name() string { return "wxpusher" }

// Send 给 Target（UID）推一条消息。summary 是微信通知栏预览（Title 兜 Content 前缀）。
// Receipt.ProviderID 带回 WxPusher 的 messageId。
func (s *Service) Send(ctx context.Context, msg *notify.Message) (*notify.Receipt, error) {
	receipt := &notify.Receipt{}
	if msg == nil || msg.Target == "" {
		return nil, errors.New("wxpusher uid is empty")
	}
	if s.cfg.AppToken == "" {
		return receipt, fmt.Errorf("%w: wxpusher config incomplete: app_token required", notify.ErrNotConfigured)
	}

	summary := msg.Title
	if summary == "" && len(msg.Content) > 20 {
		summary = msg.Content[:20]
	} else if summary == "" {
		summary = msg.Content
	}

	contentType := s.cfg.ContentType
	if contentType == 0 {
		contentType = ContentText
	}

	payload, err := json.Marshal(map[string]any{
		"appToken":    s.cfg.AppToken,
		"content":     msg.Content,
		"summary":     summary,
		"contentType": contentType,
		"uids":        []string{msg.Target},
	})
	if err != nil {
		return receipt, fmt.Errorf("marshal payload failed: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, s.cfg.APIBase+"/api/send/message", bytes.NewReader(payload))
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

	// WxPusher 的成功码是 1000，不是 0。
	var answer struct {
		Code int    `json:"code"`
		Msg  string `json:"msg"`
		Data []struct {
			MessageID int64  `json:"messageId"`
			UID       string `json:"uid"`
		} `json:"data"`
	}
	if err = json.Unmarshal(respBody, &answer); err != nil {
		return receipt, fmt.Errorf("peer answer is not JSON: %s", snippet(respBody))
	}
	if answer.Code != 1000 {
		if answer.Msg == "" {
			answer.Msg = "no message"
		}
		return receipt, fmt.Errorf("peer answered code=%d: %s", answer.Code, answer.Msg)
	}
	if len(answer.Data) > 0 && answer.Data[0].MessageID != 0 {
		receipt.ProviderID = fmt.Sprintf("%d", answer.Data[0].MessageID)
	}
	return receipt, nil
}

func snippet(body []byte) string {
	if len(body) > snippetLimit {
		return string(body[:snippetLimit]) + "…"
	}
	return string(body)
}
