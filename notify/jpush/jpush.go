// Package jpush 极光推送（JPush）v3/push：Basic Auth + registration_id 单播，零 SDK。
package jpush

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"time"

	"github.com/tx7do/go-utils/notify"
)

// DefaultAPIBase JPush 官方 API 域。
const DefaultAPIBase = "https://api.jpush.cn"

// jpushErrInvalidRegistrationID JPush 错误码：registration_id 无效（应用卸载/实例删除）。
const jpushErrInvalidRegistrationID = 1011

const (
	pushTimeout  = 10 * time.Second
	answerLimit  = 8 * 1024
	snippetLimit = 256
)

// Config 极光推送配置。Target（registration_id）按消息传。
type Config struct {
	// Ref 渠道配置的展示标识（ID、名字均可）：只进错误文案。
	Ref string

	// AppKey / MasterSecret 应用凭据（Basic Auth 的用户名/密码）。
	AppKey       string
	MasterSecret string

	// APIBase 服务基地址，留空用官方；测试可覆盖。
	APIBase string
}

// Service 极光推送渠道。无状态，可并发使用。
type Service struct {
	client  *http.Client
	apiBase string
	cfg     Config
}

// New 构造极光推送渠道。client 传 nil 用默认超时的客户端。
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
func (s *Service) Name() string { return "jpush" }

// Send 给 Target（registration_id）推一条通知。
// Title 进安卓通知栏标题；iOS 无独立标题（进 alert 的 title 子段，iOS 10+ 支持）。
// Receipt.ProviderID 带回 JPush 的 msg_id。
func (s *Service) Send(ctx context.Context, msg *notify.Message) (*notify.Receipt, error) {
	receipt := &notify.Receipt{}
	if msg == nil || msg.Target == "" {
		return nil, errors.New("jpush registration_id is empty")
	}
	if s.cfg.AppKey == "" || s.cfg.MasterSecret == "" {
		return receipt, fmt.Errorf("%w: jpush config incomplete: app_key/master_secret required", notify.ErrNotConfigured)
	}

	notification := map[string]any{
		"alert": msg.Content,
	}
	android := map[string]any{"alert": msg.Content}
	if msg.Title != "" {
		android["title"] = msg.Title
		notification["ios"] = map[string]any{
			"alert": map[string]any{"title": msg.Title, "body": msg.Content},
			"sound": "default",
		}
	} else {
		notification["ios"] = map[string]any{
			"alert": msg.Content,
			"sound": "default",
		}
	}
	notification["android"] = android

	payload, err := json.Marshal(map[string]any{
		"platform":     "all",
		"audience":     map[string]any{"registration_id": []string{msg.Target}},
		"notification": notification,
	})
	if err != nil {
		return receipt, fmt.Errorf("marshal payload failed: %w", err)
	}

	endpoint := s.apiBase + "/v3/push"
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewReader(payload))
	if err != nil {
		return receipt, fmt.Errorf("build request failed: %w", err)
	}
	req.Header.Set("Authorization", "Basic "+base64.StdEncoding.EncodeToString(
		[]byte(s.cfg.AppKey+":"+s.cfg.MasterSecret)))
	req.Header.Set("Content-Type", "application/json; charset=utf-8")

	resp, err := s.client.Do(req)
	if err != nil {
		return receipt, err
	}
	defer func() { _ = resp.Body.Close() }()

	respBody, _ := io.ReadAll(io.LimitReader(resp.Body, answerLimit))
	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		return receipt, jpushError(resp.StatusCode, respBody)
	}

	var answer struct {
		MsgID int64 `json:"msg_id"`
	}
	_ = json.Unmarshal(respBody, &answer)
	if answer.MsgID != 0 {
		receipt.ProviderID = strconv.FormatInt(answer.MsgID, 10)
	}
	return receipt, nil
}

// jpushError JPush 的错误在 {"error":{"code","message"}} 里：1011 是令牌失效，
// 其余按投递失败带出状态码与原文。
func jpushError(status int, body []byte) error {
	var answer struct {
		Error struct {
			Code    int    `json:"code"`
			Message string `json:"message"`
		} `json:"error"`
	}
	if err := json.Unmarshal(body, &answer); err != nil || answer.Error.Code == 0 {
		return fmt.Errorf("peer answered %s: %s", http.StatusText(status), snippet(body))
	}
	if answer.Error.Code == jpushErrInvalidRegistrationID {
		return fmt.Errorf("%w: code %d: %s", notify.ErrInvalidToken, answer.Error.Code, answer.Error.Message)
	}
	if answer.Error.Message == "" {
		answer.Error.Message = "no message"
	}
	return fmt.Errorf("peer answered error code=%d: %s", answer.Error.Code, answer.Error.Message)
}

func snippet(body []byte) string {
	if len(body) > snippetLimit {
		return string(body[:snippetLimit]) + "…"
	}
	return string(body)
}
