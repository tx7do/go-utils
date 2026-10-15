// Package dingtalk 钉钉工作通知（企业内部应用异步消息），零 SDK。
//
// 与群机器人（notify/webhook 的 DINGTALK 风格）分工：机器人推群、凭签名 webhook；
// 本包推**工作通知到成员个人**（userid_list），凭 appKey/appSecret/agentId。
package dingtalk

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"sync"
	"time"

	"github.com/tx7do/go-utils/notify"
)

// DefaultAPIBase 钉钉官方 API 域。
const DefaultAPIBase = "https://oapi.dingtalk.com"

const (
	pushTimeout  = 10 * time.Second
	answerLimit  = 8 * 1024
	snippetLimit = 256

	// tokenRefreshMargin access_token 提前刷新余量。
	tokenRefreshMargin = 60 * time.Second
)

// Config 钉钉工作通知配置。
type Config struct {
	// Ref 渠道配置的展示标识（ID、名字均可）：只进错误文案。
	Ref string

	// AppKey / AppSecret 企业内部应用凭据。
	AppKey    string
	AppSecret string
	AgentID   int64
	APIBase   string
}

// Service 钉钉工作通知渠道。access_token 按实例缓存（并发安全）。
type Service struct {
	client *http.Client
	cfg    Config

	mu             sync.Mutex
	token          string
	tokenExpiresAt time.Time
}

// New 构造钉钉工作通知渠道。client 传 nil 用默认超时的客户端。
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
func (s *Service) Name() string { return "dingtalk" }

// Send 给 Target（userid_list，成员 userid 以 | 相连）发工作通知。
// 正文 = Title + Content（text 类型无独立标题位，有标题时空行相连）。
// 异步接口：errcode==0 表示已受理，task_id 进回执；到 达结果靠回调查询，不在本包。
func (s *Service) Send(ctx context.Context, msg *notify.Message) (*notify.Receipt, error) {
	receipt := &notify.Receipt{}
	if msg == nil || msg.Target == "" {
		return nil, errors.New("dingtalk userid_list is empty")
	}
	if s.cfg.AppKey == "" || s.cfg.AppSecret == "" || s.cfg.AgentID == 0 {
		return receipt, fmt.Errorf("%w: dingtalk config incomplete: app_key/app_secret/agent_id required", notify.ErrNotConfigured)
	}

	content := msg.Content
	if msg.Title != "" {
		content = msg.Title + "\n" + msg.Content
	}
	payload, err := json.Marshal(map[string]any{
		"agent_id":    s.cfg.AgentID,
		"userid_list": msg.Target,
		"msg": map[string]any{
			"msgtype": "text",
			"text":    map[string]string{"content": content},
		},
	})
	if err != nil {
		return receipt, fmt.Errorf("marshal payload failed: %w", err)
	}

	token, err := s.accessToken(ctx)
	if err != nil {
		return receipt, err
	}

	endpoint := s.cfg.APIBase + "/topapi/message/corpconversation/asyncsend_v2?access_token=" + url.QueryEscape(token)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewReader(payload))
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

	var answer struct {
		ErrCode int    `json:"errcode"`
		ErrMsg  string `json:"errmsg"`
		TaskID  int64  `json:"task_id"`
	}
	if err = json.Unmarshal(respBody, &answer); err != nil {
		return receipt, fmt.Errorf("peer answer is not JSON: %s", snippet(respBody))
	}
	if answer.ErrCode != 0 {
		if answer.ErrMsg == "" {
			answer.ErrMsg = "no message"
		}
		return receipt, fmt.Errorf("peer answered errcode=%d: %s", answer.ErrCode, answer.ErrMsg)
	}
	if answer.TaskID != 0 {
		receipt.ProviderID = fmt.Sprintf("%d", answer.TaskID)
	}
	return receipt, nil
}

// accessToken 取（缓存的）access_token。钉钉对失效 token 的错误码各家文档口径不一，
// 本包不做失效强刷重试——token 按到期余量缓存，轮换凭据另起实例。
func (s *Service) accessToken(ctx context.Context) (string, error) {
	s.mu.Lock()
	if s.token != "" && time.Now().Before(s.tokenExpiresAt) {
		tok := s.token
		s.mu.Unlock()
		return tok, nil
	}
	s.mu.Unlock()

	endpoint := fmt.Sprintf("%s/gettoken?appkey=%s&appsecret=%s",
		s.cfg.APIBase, url.QueryEscape(s.cfg.AppKey), url.QueryEscape(s.cfg.AppSecret))
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return "", fmt.Errorf("build token request failed: %w", err)
	}

	resp, err := s.client.Do(req)
	if err != nil {
		return "", fmt.Errorf("get access_token failed: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	respBody, _ := io.ReadAll(io.LimitReader(resp.Body, answerLimit))
	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		return "", fmt.Errorf("token endpoint answered %s: %s", resp.Status, snippet(respBody))
	}

	var answer struct {
		ErrCode     int    `json:"errcode"`
		ErrMsg      string `json:"errmsg"`
		AccessToken string `json:"access_token"`
		ExpiresIn   int    `json:"expires_in"`
	}
	if err = json.Unmarshal(respBody, &answer); err != nil {
		return "", fmt.Errorf("token answer is not JSON: %s", snippet(respBody))
	}
	if answer.ErrCode != 0 || answer.AccessToken == "" {
		if answer.ErrMsg == "" {
			answer.ErrMsg = "no message"
		}
		return "", fmt.Errorf("token endpoint answered errcode=%d: %s", answer.ErrCode, answer.ErrMsg)
	}

	expires := time.Duration(answer.ExpiresIn) * time.Second
	if expires <= 0 {
		expires = 7200 * time.Second
	}

	s.mu.Lock()
	s.token = answer.AccessToken
	s.tokenExpiresAt = time.Now().Add(expires - tokenRefreshMargin)
	s.mu.Unlock()
	return answer.AccessToken, nil
}

func snippet(body []byte) string {
	if len(body) > snippetLimit {
		return string(body[:snippetLimit]) + "…"
	}
	return string(body)
}
