// Package getui 个推（GeTui）v2 REST API：sign 鉴权 + cid 单播，零 SDK。
package getui

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"sync"
	"time"

	crand "crypto/rand"

	"github.com/tx7do/go-utils/notify"
)

// DefaultAPIBase 个推 REST API 域。
const DefaultAPIBase = "https://restapi.getui.com"

const (
	pushTimeout  = 10 * time.Second
	answerLimit  = 8 * 1024
	snippetLimit = 256

	// tokenRefreshMargin 鉴权 token 提前刷新余量（厂商给了毫秒级过期时间）。
	tokenRefreshMargin = 60 * time.Second
)

// Config 个推配置。Target（cid）按消息传。
type Config struct {
	// Ref 渠道配置的展示标识（ID、名字均可）：只进错误文案。
	Ref string

	// AppID / AppKey / MasterSecret 应用三元组凭据。
	AppID        string
	AppKey       string
	MasterSecret string

	// APIBase 服务基地址，留空用官方；测试可覆盖。
	APIBase string
}

// Service 个推渠道。鉴权 token 按实例缓存（并发安全）。
type Service struct {
	client *http.Client
	cfg    Config

	mu             sync.Mutex
	token          string
	tokenExpiresAt time.Time
}

// New 构造个推渠道。client 传 nil 用默认超时的客户端。
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
func (s *Service) Name() string { return "getui" }

// Send 给 Target（个推 cid）推一条通知栏消息。
// Title 为通知栏标题，Content 为通知内容；Receipt 回执无厂商消息 ID（个推 v2
// 单播应答只回受理状态），留空。
func (s *Service) Send(ctx context.Context, msg *notify.Message) (*notify.Receipt, error) {
	receipt := &notify.Receipt{}
	if msg == nil || msg.Target == "" {
		return nil, errors.New("getui cid is empty")
	}
	if s.cfg.AppID == "" || s.cfg.AppKey == "" || s.cfg.MasterSecret == "" {
		return receipt, fmt.Errorf("%w: getui config incomplete: app_id/app_key/master_secret required", notify.ErrNotConfigured)
	}

	token, err := s.authToken(ctx)
	if err != nil {
		return receipt, err
	}

	payload, err := json.Marshal(map[string]any{
		// request_id 要防重：同 id 一小时内重复推送会被厂商去重丢弃。
		"request_id": crand.Text(),
		"push_message": map[string]any{
			"notification": map[string]string{
				"title":     msg.Title,
				"body":      msg.Content,
				"push_type": "notify",
			},
		},
	})
	if err != nil {
		return receipt, fmt.Errorf("marshal payload failed: %w", err)
	}

	endpoint := s.cfg.APIBase + "/v2/" + s.cfg.AppID + "/push/single/cid"
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewReader(payload))
	if err != nil {
		return receipt, fmt.Errorf("build request failed: %w", err)
	}
	req.Header.Set("token", token)
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
		Code int    `json:"code"`
		Msg  string `json:"msg"`
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

// authToken 取（缓存的）鉴权 token：sign = sha256(appkey + timestamp + mastersecret)。
func (s *Service) authToken(ctx context.Context) (string, error) {
	s.mu.Lock()
	if s.token != "" && time.Now().Before(s.tokenExpiresAt) {
		tok := s.token
		s.mu.Unlock()
		return tok, nil
	}
	s.mu.Unlock()

	timestamp := strconv.FormatInt(time.Now().UnixMilli(), 10)
	sum := sha256.Sum256([]byte(s.cfg.AppKey + timestamp + s.cfg.MasterSecret))
	payload, err := json.Marshal(map[string]string{
		"sign":      hex.EncodeToString(sum[:]),
		"timestamp": timestamp,
		"appkey":    s.cfg.AppKey,
	})
	if err != nil {
		return "", fmt.Errorf("marshal auth request failed: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost,
		s.cfg.APIBase+"/v2/"+s.cfg.AppID+"/auth", bytes.NewReader(payload))
	if err != nil {
		return "", fmt.Errorf("build auth request failed: %w", err)
	}
	req.Header.Set("Content-Type", "application/json; charset=utf-8")

	resp, err := s.client.Do(req)
	if err != nil {
		return "", fmt.Errorf("auth failed: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	respBody, _ := io.ReadAll(io.LimitReader(resp.Body, answerLimit))
	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		return "", fmt.Errorf("auth endpoint answered %s: %s", resp.Status, snippet(respBody))
	}

	var answer struct {
		Code int    `json:"code"`
		Msg  string `json:"msg"`
		Data struct {
			Token      string `json:"token"`
			ExpireTime int64  `json:"expire_time"`
		} `json:"data"`
	}
	if err = json.Unmarshal(respBody, &answer); err != nil {
		return "", fmt.Errorf("auth answer is not JSON: %s", snippet(respBody))
	}
	if answer.Code != 0 || answer.Data.Token == "" {
		if answer.Msg == "" {
			answer.Msg = "no message"
		}
		return "", fmt.Errorf("auth endpoint answered code=%d: %s", answer.Code, answer.Msg)
	}

	expires := time.Duration(answer.Data.ExpireTime-time.Now().UnixMilli()) * time.Millisecond
	if expires <= 0 {
		expires = time.Hour
	}

	s.mu.Lock()
	s.token = answer.Data.Token
	s.tokenExpiresAt = time.Now().Add(expires - tokenRefreshMargin)
	s.mu.Unlock()
	return answer.Data.Token, nil
}

func snippet(body []byte) string {
	if len(body) > snippetLimit {
		return string(body[:snippetLimit]) + "…"
	}
	return string(body)
}
