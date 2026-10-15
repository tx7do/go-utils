package push

import (
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

// OPPO Push 端点（默认生产域；AuthURL/SendURL 可覆盖供测试）。
const (
	OppoAuthURL = "https://push.oppomobile.com/server/v1/message/auth/aliyun"
	OppoSendURL = "https://push.oppomobile.com/server/v1/message/notification/unicast"
)

// oppoInvalidCode OPPO 响应码中"registration_id 无效"类（厂商码漂移只改这里）。
const oppoInvalidCode = 2008

// oppoTitleMax OPPO 通知标题上限 32 字符，按 28 runes 截断防 400。
const oppoTitleMax = 28

// OppoConfig OPPO 推送配置（key/secret 两段式鉴权）。
type OppoConfig struct {
	// Ref 渠道配置的展示标识（ID、名字均可）：只进错误文案。
	Ref string

	// AppKey / AppSecret OPPO 推送凭据。
	AppKey    string
	AppSecret string

	// AuthURL / SendURL 端点，留空用官方；测试可覆盖。
	AuthURL string
	SendURL string
}

// Oppo OPPO 推送渠道。auth token 按实例缓存，并发安全。
type Oppo struct {
	client *http.Client
	cfg    OppoConfig

	authURL string
	sendURL string

	mu         sync.Mutex
	authToken  string
	tokenExiry time.Time
}

// NewOppo 构造 OPPO 推送渠道。client 传 nil 用默认超时的客户端。
func NewOppo(client *http.Client, cfg OppoConfig) *Oppo {
	if client == nil {
		client = &http.Client{Timeout: pushTimeout}
	}
	return &Oppo{
		client:  client,
		cfg:     cfg,
		authURL: defaultStr(cfg.AuthURL, OppoAuthURL),
		sendURL: defaultStr(cfg.SendURL, OppoSendURL),
	}
}

// Name 渠道名。
func (c *Oppo) Name() string { return "oppo" }

// Send 给 Target（OPPO registration_id）推一条通知栏消息。
// Title 按 28 runes 截断（厂商上限 32 字符，超了 400）；
// OPPO 通知栏形态没有数据面——Message.Params 被忽略，业务数据请自行并进 Content。
func (c *Oppo) Send(ctx context.Context, msg *notify.Message) (*notify.Receipt, error) {
	receipt := &notify.Receipt{}
	if msg == nil || msg.Target == "" {
		return nil, errors.New("oppo registration_id is empty")
	}
	if c.cfg.AppKey == "" || c.cfg.AppSecret == "" {
		return receipt, notConfigured("oppo config incomplete: app_key/app_secret required")
	}

	authToken, err := c.getAuthToken(ctx)
	if err != nil {
		return receipt, err
	}

	form := url.Values{}
	form.Set("auth_token", authToken)
	form.Set("registration_id", msg.Target)
	form.Set("title", truncateRunes(msg.Title, oppoTitleMax))
	form.Set("content", msg.Content)
	form.Set("click_action_type", "0")

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.sendURL, strings.NewReader(form.Encode()))
	if err != nil {
		return receipt, fmt.Errorf("build request failed: %w", err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := c.client.Do(req)
	if err != nil {
		return receipt, err
	}
	defer func() { _ = resp.Body.Close() }()
	respBody, _ := io.ReadAll(io.LimitReader(resp.Body, answerLimit))

	var sr struct {
		Code    int    `json:"code"`
		Message string `json:"message"`
	}
	_ = json.Unmarshal(respBody, &sr)
	if resp.StatusCode >= 200 && resp.StatusCode < 300 && sr.Code == 0 {
		return receipt, nil
	}
	if sr.Code == oppoInvalidCode {
		return receipt, fmt.Errorf("%w: invalid registration_id (code %d)", ErrInvalidToken, sr.Code)
	}
	return receipt, fmt.Errorf("oppo push failed: http %d code %d: %s", resp.StatusCode, sr.Code, errSnippet(respBody))
}

// getAuthToken 取（缓存至到期的）auth token：两段式鉴权第一阶段。
func (c *Oppo) getAuthToken(ctx context.Context) (string, error) {
	c.mu.Lock()
	if c.authToken != "" && time.Now().Before(c.tokenExiry) {
		tok := c.authToken
		c.mu.Unlock()
		return tok, nil
	}
	c.mu.Unlock()

	form := url.Values{}
	form.Set("key", c.cfg.AppKey)
	form.Set("secret", c.cfg.AppSecret)

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.authURL, strings.NewReader(form.Encode()))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := c.client.Do(req)
	if err != nil {
		return "", err
	}
	defer func() { _ = resp.Body.Close() }()
	body, _ := io.ReadAll(io.LimitReader(resp.Body, answerLimit))
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("oppo auth failed: http %d: %s", resp.StatusCode, errSnippet(body))
	}

	var ar struct {
		Code int `json:"code"`
		Data struct {
			AuthToken string `json:"authToken"`
			ExpiresIn int    `json:"expiresIn"`
		} `json:"data"`
	}
	if err := json.Unmarshal(body, &ar); err != nil || ar.Code != 0 || ar.Data.AuthToken == "" {
		return "", errors.New("oppo auth response malformed")
	}

	c.mu.Lock()
	c.authToken = ar.Data.AuthToken
	c.tokenExiry = time.Now().Add(time.Duration(ar.Data.ExpiresIn) * time.Second)
	c.mu.Unlock()
	return ar.Data.AuthToken, nil
}
