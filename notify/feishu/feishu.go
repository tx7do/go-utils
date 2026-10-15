// Package feishu 飞书应用消息（im/v1/messages），零 SDK。
//
// 与群机器人（notify/webhook 的 FEISHU 风格）分工：机器人推群、凭签名 webhook；
// 本包凭 app_id/app_secret 发**单聊/指定接收人**的应用消息。
package feishu

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
	"sync"
	"time"

	"github.com/tx7do/go-utils/notify"
)

// DefaultAPIBase 飞书开放平台域（国际版 Lark 覆写为 https://open.larksuite.com）。
const DefaultAPIBase = "https://open.feishu.cn"

// ReceiveIDTypeDefault 默认接收人类型（open_id）；支持 open_id/user_id/union_id/email/chat_id。
const ReceiveIDTypeDefault = "open_id"

const (
	pushTimeout  = 10 * time.Second
	answerLimit  = 8 * 1024
	snippetLimit = 256

	tokenRefreshMargin = 60 * time.Second

	// token 失效码：强刷一枚重试一次。
	errCodeTokenInvalid = 99991663
	errCodeTokenExpired = 99991661
)

// Config 飞书应用消息配置。
type Config struct {
	// Ref 渠道配置的展示标识（ID、名字均可）：只进错误文案。
	Ref string

	// AppID / AppSecret 应用凭据。
	AppID     string
	AppSecret string

	// ReceiveIDType 接收人 ID 类型，留空 open_id。
	ReceiveIDType string

	// APIBase 域，留空用飞书；Lark 国际版可覆盖。
	APIBase string
}

// Service 飞书应用消息渠道。tenant_access_token 按实例缓存（并发安全）。
type Service struct {
	client *http.Client
	cfg    Config

	mu             sync.Mutex
	token          string
	tokenExpiresAt time.Time
}

// New 构造飞书应用消息渠道。client 传 nil 用默认超时的客户端。
func New(client *http.Client, cfg Config) *Service {
	if client == nil {
		client = &http.Client{Timeout: pushTimeout}
	}
	if cfg.APIBase == "" {
		cfg.APIBase = DefaultAPIBase
	}
	if cfg.ReceiveIDType == "" {
		cfg.ReceiveIDType = ReceiveIDTypeDefault
	}
	return &Service{client: client, cfg: cfg}
}

// Name 渠道名。
func (s *Service) Name() string { return "feishu" }

// Send 给 Target（按 ReceiveIDType 解释：open_id/user_id/union_id/email/chat_id）
// 发一条文本消息。正文 = Title + Content（text 类型无独立标题位，有标题时空行相连）。
// Receipt.ProviderID 带回飞书的 message_id。
func (s *Service) Send(ctx context.Context, msg *notify.Message) (*notify.Receipt, error) {
	receipt := &notify.Receipt{}
	if msg == nil || msg.Target == "" {
		return nil, errors.New("feishu receive_id is empty")
	}
	if s.cfg.AppID == "" || s.cfg.AppSecret == "" {
		return receipt, fmt.Errorf("%w: feishu config incomplete: app_id/app_secret required", notify.ErrNotConfigured)
	}

	content := msg.Content
	if msg.Title != "" {
		content = msg.Title + "\n" + msg.Content
	}
	// 飞书的 content 字段是"JSON 字符串套 JSON"：内层对象先序列化成串再嵌进去。
	inner, err := json.Marshal(map[string]string{"text": content})
	if err != nil {
		return receipt, fmt.Errorf("marshal inner content failed: %w", err)
	}
	payload, err := json.Marshal(map[string]any{
		"receive_id": msg.Target,
		"msg_type":   "text",
		"content":    string(inner),
	})
	if err != nil {
		return receipt, fmt.Errorf("marshal payload failed: %w", err)
	}

	token, err := s.tenantAccessToken(ctx)
	if err != nil {
		return receipt, err
	}

	messageID, err := s.sendText(ctx, token, payload)
	if isTokenErr(err) {
		s.invalidateToken()
		token, tokenErr := s.tenantAccessToken(ctx)
		if tokenErr != nil {
			return receipt, tokenErr
		}
		messageID, err = s.sendText(ctx, token, payload)
	}
	if err == nil {
		receipt.ProviderID = messageID
	}
	return receipt, err
}

func (s *Service) sendText(ctx context.Context, token string, payload []byte) (string, error) {
	endpoint := s.cfg.APIBase + "/open-apis/im/v1/messages?receive_id_type=" + url.QueryEscape(s.cfg.ReceiveIDType)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewReader(payload))
	if err != nil {
		return "", fmt.Errorf("build request failed: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json; charset=utf-8")

	resp, err := s.client.Do(req)
	if err != nil {
		return "", err
	}
	defer func() { _ = resp.Body.Close() }()

	respBody, _ := io.ReadAll(io.LimitReader(resp.Body, answerLimit))
	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		return "", fmt.Errorf("peer answered %s: %s", resp.Status, snippet(respBody))
	}

	var answer struct {
		Code int    `json:"code"`
		Msg  string `json:"msg"`
		Data struct {
			MessageID string `json:"message_id"`
		} `json:"data"`
	}
	if err = json.Unmarshal(respBody, &answer); err != nil {
		return "", fmt.Errorf("peer answer is not JSON: %s", snippet(respBody))
	}
	if answer.Code != 0 {
		if answer.Msg == "" {
			answer.Msg = "no message"
		}
		return "", fmt.Errorf("peer answered code=%d: %s", answer.Code, answer.Msg)
	}
	return answer.Data.MessageID, nil
}

// tenantAccessToken 取（缓存的）tenant_access_token。
func (s *Service) tenantAccessToken(ctx context.Context) (string, error) {
	s.mu.Lock()
	if s.token != "" && time.Now().Before(s.tokenExpiresAt) {
		tok := s.token
		s.mu.Unlock()
		return tok, nil
	}
	s.mu.Unlock()

	payload, err := json.Marshal(map[string]string{
		"app_id":     s.cfg.AppID,
		"app_secret": s.cfg.AppSecret,
	})
	if err != nil {
		return "", fmt.Errorf("marshal token request failed: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost,
		s.cfg.APIBase+"/open-apis/auth/v3/tenant_access_token/internal", bytes.NewReader(payload))
	if err != nil {
		return "", fmt.Errorf("build token request failed: %w", err)
	}
	req.Header.Set("Content-Type", "application/json; charset=utf-8")

	resp, err := s.client.Do(req)
	if err != nil {
		return "", fmt.Errorf("get tenant_access_token failed: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	respBody, _ := io.ReadAll(io.LimitReader(resp.Body, answerLimit))
	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		return "", fmt.Errorf("token endpoint answered %s: %s", resp.Status, snippet(respBody))
	}

	var answer struct {
		Code              int    `json:"code"`
		Msg               string `json:"msg"`
		TenantAccessToken string `json:"tenant_access_token"`
		Expire            int    `json:"expire"`
	}
	if err = json.Unmarshal(respBody, &answer); err != nil {
		return "", fmt.Errorf("token answer is not JSON: %s", snippet(respBody))
	}
	if answer.Code != 0 || answer.TenantAccessToken == "" {
		if answer.Msg == "" {
			answer.Msg = "no message"
		}
		return "", fmt.Errorf("token endpoint answered code=%d: %s", answer.Code, answer.Msg)
	}

	expires := time.Duration(answer.Expire) * time.Second
	if expires <= 0 {
		expires = 7200 * time.Second
	}

	s.mu.Lock()
	s.token = answer.TenantAccessToken
	s.tokenExpiresAt = time.Now().Add(expires - tokenRefreshMargin)
	s.mu.Unlock()
	return answer.TenantAccessToken, nil
}

func (s *Service) invalidateToken() {
	s.mu.Lock()
	s.token = ""
	s.tokenExpiresAt = time.Time{}
	s.mu.Unlock()
}

func isTokenErr(err error) bool {
	if err == nil {
		return false
	}
	msg := err.Error()
	return strings.Contains(msg, fmt.Sprintf("code=%d", errCodeTokenInvalid)) ||
		strings.Contains(msg, fmt.Sprintf("code=%d", errCodeTokenExpired))
}

func snippet(body []byte) string {
	if len(body) > snippetLimit {
		return string(body[:snippetLimit]) + "…"
	}
	return string(body)
}
