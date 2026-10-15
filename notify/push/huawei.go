package push

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
	"net/url"
	"strings"
	"sync"
	"time"

	"github.com/tx7do/go-utils/notify"
)

// 华为 / 荣耀 Push Kit 端点（默认生产域；TokenURL/SendURL 可覆盖供测试）。
const (
	HuaweiTokenURL = "https://oauth-login.cloud.huawei.com/oauth2/v3/token"
	HuaweiSendURL  = "https://push-api.cloud.huawei.com/v1/%s/messages:send"
	HonorTokenURL  = "https://oauth-login.hihonorcloud.com/oauth2/v3/token"
	HonorSendURL   = "https://push-drcn.hihonorcloud.com/v1/%s/messages:send"
)

// 华为系响应码：0x80000000 段。0x80200xxx 为令牌/目标无效类（厂商码漂移只改这里的表）。
const (
	huaweiCodeOK       = 80000000
	huaweiErrInvalidLo = 80200000
	huaweiErrInvalidHi = 80209999
)

// HuaweiConfig 华为/荣耀推送（HMS / Honor Push Kit）配置。两者协议同形，只差域名。
type HuaweiConfig struct {
	// Ref 渠道配置的展示标识（ID、名字均可）：只进错误文案。
	Ref string

	// ClientID / ClientSecret 推送服务凭据。
	ClientID     string
	ClientSecret string

	// TokenURL / SendURL 端点，留空按 Honor 开关切华为/荣耀的默认生产域；测试可覆盖。
	TokenURL string
	SendURL  string // 含 %s app_id 占位
}

// huaweiClient 华为/荣耀共用客户端：OAuth2 client_credentials 换访问令牌
// （按凭证指纹缓存——多租户各拿各的凭据轮用同一实例时不串令牌），
// 逐台设备 POST messages:send。
type huaweiClient struct {
	httpClient *http.Client

	tokenURL string
	sendURL  string // 含 %s app_id 占位

	mu          sync.Mutex
	fingerprint string
	accessToken string
	tokenExpiry time.Time
}

func newHuaweiClient(client *http.Client, tokenURL, sendURL string) *huaweiClient {
	if client == nil {
		client = &http.Client{Timeout: pushTimeout}
	}
	return &huaweiClient{httpClient: client, tokenURL: tokenURL, sendURL: sendURL}
}

// Huawei 华为推送渠道（HMS Push Kit）。
type Huawei struct {
	hc  *huaweiClient
	cfg HuaweiConfig
}

// Honor 荣耀推送渠道。协议与华为同形，只差域名。
type Honor struct {
	hc  *huaweiClient
	cfg HuaweiConfig
}

// NewHuawei 构造华为推送渠道。client 传 nil 用默认超时的客户端。
func NewHuawei(client *http.Client, cfg HuaweiConfig) *Huawei {
	return &Huawei{hc: newHuaweiClient(client, defaultStr(cfg.TokenURL, HuaweiTokenURL), defaultStr(cfg.SendURL, HuaweiSendURL)), cfg: cfg}
}

// NewHonor 构造荣耀推送渠道。
func NewHonor(client *http.Client, cfg HuaweiConfig) *Honor {
	return &Honor{hc: newHuaweiClient(client, defaultStr(cfg.TokenURL, HonorTokenURL), defaultStr(cfg.SendURL, HonorSendURL)), cfg: cfg}
}

// Name 渠道名。
func (c *Huawei) Name() string { return "huawei" }

// Name 渠道名。
func (c *Honor) Name() string { return "honor" }

// Send 给 Target（华为/荣耀 regId）推一条通知。
// Title/Content 进通知面板，Params 进 data 面。Receipt.ProviderID 带回 requestId。
func sendHuaweiStyle(ctx context.Context, hc *huaweiClient, cfg HuaweiConfig, msg *notify.Message) (*notify.Receipt, error) {
	receipt := &notify.Receipt{}
	if msg == nil || msg.Target == "" {
		return nil, errors.New("device token is empty")
	}
	if cfg.ClientID == "" || cfg.ClientSecret == "" {
		return receipt, notConfigured("config incomplete: client_id/client_secret required")
	}

	accessToken, err := hc.getAccessToken(ctx, cfg.ClientID, cfg.ClientSecret)
	if err != nil {
		return receipt, err
	}

	payload, err := json.Marshal(map[string]any{
		"message": map[string]any{
			"token": msg.Target,
			"notification": map[string]string{
				"title": msg.Title,
				"body":  msg.Content,
			},
			"data": msg.Params,
		},
	})
	if err != nil {
		return receipt, fmt.Errorf("marshal payload failed: %w", err)
	}

	endpoint := fmt.Sprintf(hc.sendURL, url.PathEscape(cfg.ClientID))
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewReader(payload))
	if err != nil {
		return receipt, fmt.Errorf("build request failed: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+accessToken)
	req.Header.Set("Content-Type", "application/json; charset=UTF-8")

	resp, err := hc.httpClient.Do(req)
	if err != nil {
		return receipt, err
	}
	defer func() { _ = resp.Body.Close() }()
	respBody, _ := io.ReadAll(io.LimitReader(resp.Body, answerLimit))

	var sr struct {
		Code      int    `json:"code"`
		Msg       string `json:"msg"`
		RequestID string `json:"requestId"`
	}
	_ = json.Unmarshal(respBody, &sr)
	if resp.StatusCode >= 200 && resp.StatusCode < 300 && sr.Code == huaweiCodeOK {
		receipt.ProviderID = sr.RequestID
		return receipt, nil
	}
	if sr.Code >= huaweiErrInvalidLo && sr.Code <= huaweiErrInvalidHi {
		return receipt, fmt.Errorf("%w: invalid push token (code %d)", ErrInvalidToken, sr.Code)
	}
	return receipt, fmt.Errorf("push send failed: http %d code %d: %s", resp.StatusCode, sr.Code, errSnippet(respBody))
}

// Send 实现华为渠道投递。
func (c *Huawei) Send(ctx context.Context, msg *notify.Message) (*notify.Receipt, error) {
	return sendHuaweiStyle(ctx, c.hc, c.cfg, msg)
}

// Send 实现荣耀渠道投递。
func (c *Honor) Send(ctx context.Context, msg *notify.Message) (*notify.Receipt, error) {
	return sendHuaweiStyle(ctx, c.hc, c.cfg, msg)
}

// getAccessToken 取（按凭证指纹缓存的）OAuth2 访问令牌。
func (hc *huaweiClient) getAccessToken(ctx context.Context, clientID, clientSecret string) (string, error) {
	fp := credentialFingerprint(clientID, clientSecret)

	hc.mu.Lock()
	if hc.accessToken != "" && hc.fingerprint == fp && time.Now().Before(hc.tokenExpiry) {
		tok := hc.accessToken
		hc.mu.Unlock()
		return tok, nil
	}
	hc.mu.Unlock()

	form := url.Values{}
	form.Set("grant_type", "client_credentials")
	form.Set("client_id", clientID)
	form.Set("client_secret", clientSecret)

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, hc.tokenURL, strings.NewReader(form.Encode()))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := hc.httpClient.Do(req)
	if err != nil {
		return "", err
	}
	defer func() { _ = resp.Body.Close() }()
	body, _ := io.ReadAll(io.LimitReader(resp.Body, answerLimit))
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("oauth token failed: http %d: %s", resp.StatusCode, errSnippet(body))
	}

	var tr struct {
		AccessToken string `json:"access_token"`
		ExpiresIn   int    `json:"expires_in"`
	}
	if err := json.Unmarshal(body, &tr); err != nil || tr.AccessToken == "" {
		return "", errors.New("oauth token response malformed")
	}

	hc.mu.Lock()
	hc.fingerprint = fp
	hc.accessToken = tr.AccessToken
	hc.tokenExpiry = time.Now().Add(time.Duration(tr.ExpiresIn) * time.Second)
	hc.mu.Unlock()
	return tr.AccessToken, nil
}

// credentialFingerprint 凭证指纹：sha256 hex。
func credentialFingerprint(clientID, clientSecret string) string {
	sum := sha256.Sum256([]byte(clientID + "\x00" + clientSecret))
	return hex.EncodeToString(sum[:])
}

func defaultStr(v, fallback string) string {
	if v != "" {
		return v
	}
	return fallback
}
