// Package wechat 微信公众号模板消息渠道：access_token 流程 + template/send。
//
// 走裸 HTTP，零 SDK。两条协议事实钉在实现里：
//   - 模板消息必须有已备案的 template_id，正文形状是 data 字段里的键值对
//     （每个键渲染成 {value}），没有自由文本通道；
//   - access_token 有每日获取配额与 2 小时有效期，所以按过期前余量缓存、
//     失效时强刷重试一次，而不是每条消息都去要一枚新 token。
package wechat

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/tx7do/go-utils/notify"
)

const (
	defaultAPIBase = "https://api.weixin.qq.com"

	// wechatTimeout 单次出站请求的总超时。
	wechatTimeout = 10 * time.Second

	// wechatAnswerBytes 读对端应答的上限。
	wechatAnswerBytes = 4 * 1024

	// tokenRefreshMargin access_token 提前这么多时间量刷新，免得边界上用过期 token。
	tokenRefreshMargin = 60 * time.Second

	// errCodeTokenInvalid 微信侧 token 失效的两个错误码：强刷一次再试。
	errCodeTokenInvalid = 40001 // access_token invalid
	errCodeTokenExpired = 42001 // access_token expired
)

// Config 微信公众号渠道配置。
type Config struct {
	// Ref 渠道配置的展示标识（ID、名字均可）：只进错误文案。
	Ref string

	// AppID / AppSecret 公众号凭据。
	AppID     string
	AppSecret string

	// TemplateID 已备案的模板消息 ID。模板消息没有它发不出去。
	TemplateID string

	// APIBase 服务端基地址，留空用官方 API；测试与自建网关可覆盖。
	APIBase string
}

// Service 微信公众号渠道。access_token 按 Service 实例缓存（并发安全）；
// 一个实例对应一份凭据，凭据轮换时另起新实例。
type Service struct {
	client *http.Client
	cfg    Config

	mu             sync.Mutex
	token          string
	tokenExpiresAt time.Time
}

// New 构造微信公众号渠道。client 传 nil 用默认超时的客户端。
func New(client *http.Client, cfg Config) *Service {
	if client == nil {
		client = &http.Client{Timeout: wechatTimeout}
	}
	if cfg.APIBase == "" {
		cfg.APIBase = defaultAPIBase
	}
	return &Service{client: client, cfg: cfg}
}

// Name 渠道名。
func (s *Service) Name() string { return "wechat" }

// Send 给 Target（openid）发一条模板消息。
//
// 正文形状：Params 的每个键渲染成模板的 {KEY} 占位值；Params 为空且 Content 非空时，
// 整段正文落在 "content" 键里（对应模板里唯一一个 {{content.data}}）。Title 不进模板消息——
// 公众号模板的形状全在 data 里，没有标题位。
func (s *Service) Send(ctx context.Context, msg *notify.Message) (*notify.Receipt, error) {
	receipt := &notify.Receipt{}
	if msg == nil || msg.Target == "" {
		return nil, errors.New("wechat openid is empty")
	}
	if s.cfg.AppID == "" || s.cfg.AppSecret == "" {
		return nil, fmt.Errorf("%w: wechat config missing app_id/app_secret", notify.ErrNotConfigured)
	}
	if s.cfg.TemplateID == "" {
		return nil, fmt.Errorf("%w: wechat config missing template_id: template messages require a provider-side approved template", notify.ErrNotConfigured)
	}

	data := make(map[string]map[string]string, len(msg.Params))
	for k, v := range msg.Params {
		data[k] = map[string]string{"value": v}
	}
	if len(data) == 0 && msg.Content != "" {
		data["content"] = map[string]string{"value": msg.Content}
	}

	token, err := s.accessToken(ctx)
	if err != nil {
		return receipt, err
	}

	receipt, err = s.sendTemplate(ctx, token, msg.Target, data)
	if err == nil {
		return receipt, nil
	}

	// token 在缓存窗口内被微信侧作废（重置密钥/后台刷新）是常态外的事：
	// 强刷一枚再试一次，仍失败才把错误交出去。
	if isTokenErr(err) {
		s.invalidateToken()
		token, tokenErr := s.accessToken(ctx)
		if tokenErr != nil {
			return receipt, tokenErr
		}
		return s.sendTemplate(ctx, token, msg.Target, data)
	}
	return receipt, err
}

// accessToken 取一枚可用的 access_token：缓存未到期直接用，否则向微信换取。
func (s *Service) accessToken(ctx context.Context) (string, error) {
	s.mu.Lock()
	if s.token != "" && time.Now().Before(s.tokenExpiresAt) {
		token := s.token
		s.mu.Unlock()
		return token, nil
	}
	s.mu.Unlock()

	endpoint := fmt.Sprintf("%s/cgi-bin/token?grant_type=client_credential&appid=%s&secret=%s",
		s.cfg.APIBase, url.QueryEscape(s.cfg.AppID), url.QueryEscape(s.cfg.AppSecret))

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return "", fmt.Errorf("build token request failed: %w", err)
	}

	resp, err := s.client.Do(req)
	if err != nil {
		return "", fmt.Errorf("get access_token failed: %w", err)
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	respBody, readErr := io.ReadAll(io.LimitReader(resp.Body, wechatAnswerBytes))
	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		return "", fmt.Errorf("token endpoint answered %s: %s", resp.Status, snippet(respBody, readErr))
	}
	if readErr != nil && readErr != io.EOF {
		return "", fmt.Errorf("read token answer: %w", readErr)
	}

	var answer struct {
		AccessToken string `json:"access_token"`
		ExpiresIn   int    `json:"expires_in"`
		ErrCode     int    `json:"errcode"`
		ErrMsg      string `json:"errmsg"`
	}
	if err = json.Unmarshal(respBody, &answer); err != nil {
		return "", fmt.Errorf("token answer is not JSON: %s", snippet(respBody, readErr))
	}
	if answer.ErrCode != 0 || answer.AccessToken == "" {
		if answer.ErrMsg == "" {
			answer.ErrMsg = "no message"
		}
		return "", fmt.Errorf("token endpoint answered errcode=%d: %s", answer.ErrCode, answer.ErrMsg)
	}

	expires := time.Duration(answer.ExpiresIn) * time.Second
	if expires <= 0 {
		expires = 7200 * time.Second // 微信文档默认 7200s；对端没给时按默认兜
	}

	s.mu.Lock()
	s.token = answer.AccessToken
	s.tokenExpiresAt = time.Now().Add(expires - tokenRefreshMargin)
	s.mu.Unlock()

	return answer.AccessToken, nil
}

// invalidateToken 丢弃缓存的 access_token：对端报 token 失效时强刷用。
func (s *Service) invalidateToken() {
	s.mu.Lock()
	s.token = ""
	s.tokenExpiresAt = time.Time{}
	s.mu.Unlock()
}

// sendTemplate 发一次模板消息。Receipt.ProviderID 带回微信的 msgid。
func (s *Service) sendTemplate(ctx context.Context, token, openid string, data map[string]map[string]string) (*notify.Receipt, error) {
	receipt := &notify.Receipt{}

	payload, err := json.Marshal(map[string]any{
		"touser":      openid,
		"template_id": s.cfg.TemplateID,
		"data":        data,
	})
	if err != nil {
		return receipt, fmt.Errorf("marshal payload failed: %w", err)
	}

	endpoint := fmt.Sprintf("%s/cgi-bin/message/template/send?access_token=%s",
		s.cfg.APIBase, url.QueryEscape(token))
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewReader(payload))
	if err != nil {
		return receipt, fmt.Errorf("build request failed: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := s.client.Do(req)
	if err != nil {
		return receipt, err
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	respBody, readErr := io.ReadAll(io.LimitReader(resp.Body, wechatAnswerBytes))
	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		return receipt, fmt.Errorf("peer answered %s: %s", resp.Status, snippet(respBody, readErr))
	}
	if readErr != nil && readErr != io.EOF {
		return receipt, fmt.Errorf("read peer answer: %w", readErr)
	}

	var answer struct {
		ErrCode int    `json:"errcode"`
		ErrMsg  string `json:"errmsg"`
		MsgID   int64  `json:"msgid"`
	}
	if err = json.Unmarshal(respBody, &answer); err != nil {
		return receipt, fmt.Errorf("peer answer is not JSON: %s", snippet(respBody, readErr))
	}
	if answer.ErrCode != 0 {
		if answer.ErrMsg == "" {
			answer.ErrMsg = "no message"
		}
		return receipt, fmt.Errorf("peer answered errcode=%d: %s", answer.ErrCode, answer.ErrMsg)
	}

	if answer.MsgID != 0 {
		receipt.ProviderID = strconv.FormatInt(answer.MsgID, 10)
	}
	return receipt, nil
}

func isTokenErr(err error) bool {
	if err == nil {
		return false
	}
	msg := err.Error()
	return strings.Contains(msg, fmt.Sprintf("errcode=%d", errCodeTokenInvalid)) ||
		strings.Contains(msg, fmt.Sprintf("errcode=%d", errCodeTokenExpired))
}

// snippet 进错误文案的响应片段截断：对端可能回一整页 HTML。
func snippet(body []byte, readErr error) string {
	const limit = 512
	if readErr != nil && readErr != io.EOF {
		return "<unreadable>"
	}
	if len(body) <= limit {
		return string(body)
	}
	return string(body[:limit]) + "…"
}
