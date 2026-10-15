// Package wecom 企业微信应用消息：access_token 流 + message/send，零依赖。
//
// 与群机器人 webhook 的分工：机器人（notify/webhook 的 WECOM 风格）推群、只凭 webhook 地址；
// 本包推**应用消息到成员个人**（touser），凭 corpid/corpsecret/agentid。
package wecom

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

// DefaultAPIBase 企业微信官方 API 域。
const DefaultAPIBase = "https://qyapi.weixin.qq.com"

const (
	pushTimeout  = 10 * time.Second
	answerLimit  = 8 * 1024
	snippetLimit = 256

	// tokenRefreshMargin access_token 提前刷新余量。
	tokenRefreshMargin = 60 * time.Second
)

// token 失效码：强刷一枚重试一次（密钥重置/后台刷新是常态外的事，重试一次就够）。
const (
	errCodeTokenInvalid = 40014
	errCodeTokenExpired = 42001
)

// Config 企业微信应用消息配置。
type Config struct {
	// Ref 渠道配置的展示标识（ID、名字均可）：只进错误文案。
	Ref string

	// CorpID / CorpSecret 企业 ID 与自建应用的 Secret。
	CorpID     string
	CorpSecret string

	// AgentID 自建应用的 AgentId（应用消息必填）。
	AgentID int64

	// APIBase 服务基地址，留空用官方；测试可覆盖。
	APIBase string
}

// Service 企业微信应用消息渠道。access_token 按实例缓存（并发安全）；
// 一个实例对应一份凭据，凭据轮换时另起新实例。
type Service struct {
	client *http.Client
	cfg    Config

	mu             sync.Mutex
	token          string
	tokenExpiresAt time.Time
}

// New 构造企业微信应用消息渠道。client 传 nil 用默认超时的客户端。
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
func (s *Service) Name() string { return "wecom" }

// Send 给 Target 发应用消息。touser 语法：成员账号以 | 相连（"zhangsan|lisi"），
// "@all" 表示全员。正文 = Title + Content（text 类型无独立标题位，有标题时空行相连）。
// 企业微信对应用消息不回消息 ID，Receipt 为空；无效成员列在 errcode 0 应答的
// invaliduser 字段里，由调用方按需取（本包不因部分无效而判失败）。
func (s *Service) Send(ctx context.Context, msg *notify.Message) (*notify.Receipt, error) {
	receipt := &notify.Receipt{}
	if msg == nil || msg.Target == "" {
		return nil, errors.New("wecom touser is empty")
	}
	if s.cfg.CorpID == "" || s.cfg.CorpSecret == "" || s.cfg.AgentID == 0 {
		return receipt, fmt.Errorf("%w: wecom config incomplete: corp_id/corp_secret/agent_id required", notify.ErrNotConfigured)
	}

	content := msg.Content
	if msg.Title != "" {
		content = msg.Title + "\n" + msg.Content
	}
	payload, err := json.Marshal(map[string]any{
		"touser":  msg.Target,
		"msgtype": "text",
		"agentid": s.cfg.AgentID,
		"text":    map[string]string{"content": content},
	})
	if err != nil {
		return receipt, fmt.Errorf("marshal payload failed: %w", err)
	}

	token, err := s.accessToken(ctx)
	if err != nil {
		return receipt, err
	}

	err = s.sendText(ctx, token, payload)
	if isTokenErr(err) {
		s.invalidateToken()
		token, tokenErr := s.accessToken(ctx)
		if tokenErr != nil {
			return receipt, tokenErr
		}
		err = s.sendText(ctx, token, payload)
	}
	return receipt, err
}

// sendText 发一次应用消息。errcode==0 即受理（无效成员在 invaliduser 字段，不影响受理）。
func (s *Service) sendText(ctx context.Context, token string, payload []byte) error {
	endpoint := s.cfg.APIBase + "/cgi-bin/message/send?access_token=" + url.QueryEscape(token)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewReader(payload))
	if err != nil {
		return fmt.Errorf("build request failed: %w", err)
	}
	req.Header.Set("Content-Type", "application/json; charset=utf-8")

	resp, err := s.client.Do(req)
	if err != nil {
		return err
	}
	defer func() { _ = resp.Body.Close() }()

	respBody, _ := io.ReadAll(io.LimitReader(resp.Body, answerLimit))
	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		return fmt.Errorf("peer answered %s: %s", resp.Status, snippet(respBody))
	}

	var answer struct {
		ErrCode     int    `json:"errcode"`
		ErrMsg      string `json:"errmsg"`
		InvalidUser string `json:"invaliduser"`
	}
	if err = json.Unmarshal(respBody, &answer); err != nil {
		return fmt.Errorf("peer answer is not JSON: %s", snippet(respBody))
	}
	if answer.ErrCode != 0 {
		if answer.ErrMsg == "" {
			answer.ErrMsg = "no message"
		}
		return fmt.Errorf("peer answered errcode=%d: %s", answer.ErrCode, answer.ErrMsg)
	}
	return nil
}

// accessToken 取（缓存的）access_token：corpid+secret 换取。
func (s *Service) accessToken(ctx context.Context) (string, error) {
	s.mu.Lock()
	if s.token != "" && time.Now().Before(s.tokenExpiresAt) {
		tok := s.token
		s.mu.Unlock()
		return tok, nil
	}
	s.mu.Unlock()

	endpoint := fmt.Sprintf("%s/cgi-bin/gettoken?corpid=%s&corpsecret=%s",
		s.cfg.APIBase, url.QueryEscape(s.cfg.CorpID), url.QueryEscape(s.cfg.CorpSecret))
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
	return strings.Contains(msg, fmt.Sprintf("errcode=%d", errCodeTokenInvalid)) ||
		strings.Contains(msg, fmt.Sprintf("errcode=%d", errCodeTokenExpired))
}

func snippet(body []byte) string {
	if len(body) > snippetLimit {
		return string(body[:snippetLimit]) + "…"
	}
	return string(body)
}
