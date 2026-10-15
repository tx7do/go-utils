package push

import (
	"bytes"
	"context"
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"encoding/pem"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/tx7do/go-utils/notify"
)

// FCM HTTP v1 API。service account 的 RS256 JWT 换 OAuth2 访问令牌（进程内缓存），
// 逐台设备 POST messages:send。
const (
	fcmSendEndpoint    = "https://fcm.googleapis.com/v1/projects/%s/messages:send"
	fcmDefaultTokenURI = "https://oauth2.googleapis.com/token"
	fcmScope           = "https://www.googleapis.com/auth/firebase.messaging"

	// fcmTokenRefresh 访问令牌提前刷新窗口：到期前 60 秒视为过期。
	fcmTokenRefresh = 60 * time.Second
)

// FCMConfig Firebase Cloud Messaging（HTTP v1 API）配置。
type FCMConfig struct {
	// Ref 渠道配置的展示标识（ID、名字均可）：只进错误文案。
	Ref string

	// ProjectID Firebase 项目 ID。
	ProjectID string

	// ClientEmail / PrivateKey 服务账号凭据（Firebase 控制台下载的 JSON 里的
	// client_email / private_key 两个字段；PEM 支持 PKCS#8 与 PKCS#1）。
	ClientEmail string
	PrivateKey  string

	// TokenURI OAuth2 令牌端点，留空用 Google 官方；测试可覆盖。
	TokenURI string

	// APIBase 发送端点基地址，留空用官方；测试可覆盖。
	APIBase string
}

// FCM FCM 推送渠道。单实例可复用（连接池 + 令牌按服务账号缓存），并发安全。
type FCM struct {
	client *http.Client
	cfg    FCMConfig

	mu          sync.Mutex
	tokenEmail  string
	tokenExpiry time.Time
	accessToken string
}

// NewFCM 构造 FCM 渠道。client 传 nil 用默认超时的客户端。
func NewFCM(client *http.Client, cfg FCMConfig) *FCM {
	if client == nil {
		client = &http.Client{Timeout: pushTimeout}
	}
	if cfg.APIBase == "" {
		cfg.APIBase = "https://fcm.googleapis.com"
	}
	return &FCM{client: client, cfg: cfg}
}

// Name 渠道名。
func (c *FCM) Name() string { return "fcm" }

// Send 给 Target（FCM registration token）推一条通知。
// Title/Content 进 notification 面板，Params 进 data 面（安卓端 onMessageReceived 可读）。
func (c *FCM) Send(ctx context.Context, msg *notify.Message) (*notify.Receipt, error) {
	receipt := &notify.Receipt{}
	if msg == nil || msg.Target == "" {
		return nil, errors.New("fcm device token is empty")
	}
	if err := c.checkConfig(); err != nil {
		return receipt, err
	}

	accessToken, err := c.getAccessToken(ctx)
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
			"data":    msg.Params,
			"android": map[string]any{"priority": "HIGH"},
		},
	})
	if err != nil {
		return receipt, fmt.Errorf("marshal payload failed: %w", err)
	}

	endpoint := c.cfg.APIBase + "/v1/projects/" + c.cfg.ProjectID + "/messages:send"
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewReader(payload))
	if err != nil {
		return receipt, fmt.Errorf("build request failed: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+accessToken)
	req.Header.Set("Content-Type", "application/json; charset=UTF-8")

	resp, err := c.client.Do(req)
	if err != nil {
		return receipt, err
	}
	defer func() { _ = resp.Body.Close() }()

	respBody, _ := io.ReadAll(io.LimitReader(resp.Body, answerLimit))
	switch {
	case resp.StatusCode >= 200 && resp.StatusCode < 300:
		// 应答 {"name":"projects/x/messages/<id>"}：尾部即厂商侧消息 ID。
		var answer struct {
			Name string `json:"name"`
		}
		_ = json.Unmarshal(respBody, &answer)
		if i := strings.LastIndex(answer.Name, "/"); i >= 0 {
			receipt.ProviderID = answer.Name[i+1:]
		}
		return receipt, nil
	case resp.StatusCode == http.StatusNotFound || resp.StatusCode == http.StatusGone:
		// 令牌失效（应用卸载/实例删除）：调用方删行自清理，不算投递失败重试。
		return receipt, fmt.Errorf("%w: http %s", ErrInvalidToken, resp.Status)
	default:
		return receipt, fmt.Errorf("fcm send failed: http %d: %s", resp.StatusCode, errSnippet(respBody))
	}
}

func (c *FCM) checkConfig() error {
	if c.cfg.ProjectID == "" {
		return notConfigured("fcm config incomplete: project_id required")
	}
	if c.cfg.ClientEmail == "" || c.cfg.PrivateKey == "" {
		return notConfigured("fcm config incomplete: client_email/private_key required")
	}
	return nil
}

// getAccessToken 取（缓存的）OAuth2 访问令牌：RS256 JWT assertion 换取。
func (c *FCM) getAccessToken(ctx context.Context) (string, error) {
	c.mu.Lock()
	if c.accessToken != "" && c.tokenEmail == c.cfg.ClientEmail && time.Now().Before(c.tokenExpiry) {
		tok := c.accessToken
		c.mu.Unlock()
		return tok, nil
	}
	c.mu.Unlock()

	assertion, expiry, err := MintFcmJWT(c.cfg.ClientEmail, c.cfg.PrivateKey, c.tokenURI())
	if err != nil {
		return "", err
	}

	form := strings.NewReader("grant_type=urn:ietf:params:oauth:grant-type:jwt-bearer&assertion=" + assertion)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.tokenURI(), form)
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
		return "", fmt.Errorf("fcm oauth token failed: http %d: %s", resp.StatusCode, errSnippet(body))
	}
	var tok struct {
		AccessToken string `json:"access_token"`
		ExpiresIn   int    `json:"expires_in"`
	}
	if err = json.Unmarshal(body, &tok); err != nil {
		return "", err
	}
	if tok.AccessToken == "" {
		return "", errors.New("fcm oauth token empty access_token")
	}

	exp := expiry
	if d := time.Duration(tok.ExpiresIn) * time.Second; d > 0 && time.Now().Add(d).Before(exp) {
		exp = time.Now().Add(d)
	}

	c.mu.Lock()
	c.accessToken = tok.AccessToken
	c.tokenEmail = c.cfg.ClientEmail
	c.tokenExpiry = exp
	c.mu.Unlock()
	return tok.AccessToken, nil
}

func (c *FCM) tokenURI() string {
	if c.cfg.TokenURI != "" {
		return c.cfg.TokenURI
	}
	return fcmDefaultTokenURI
}

// MintFcmJWT 用服务账号私钥签 RS256 JWT（scope=FCM），返回 (assertion, 过期时间)。
//
// 签名链：header+claims 各 base64url，RSA PKCS#1 v15 / SHA-256——与 golang.org/x/oauth2
// 的 JWT 拨号同一形状，纯标准库实现避免为此引入 oauth2 依赖。
func MintFcmJWT(clientEmail, privateKeyPEM, tokenURI string) (string, time.Time, error) {
	if clientEmail == "" || privateKeyPEM == "" {
		return "", time.Time{}, errors.New("fcm service account incomplete (client_email/private_key required)")
	}

	privateKey, err := parseRSAPrivateKeyFromPEM(privateKeyPEM)
	if err != nil {
		return "", time.Time{}, err
	}

	now := time.Now()
	expiry := now.Add(time.Hour)
	header := base64.RawURLEncoding.EncodeToString([]byte(`{"alg":"RS256","typ":"JWT"}`))
	claims, err := json.Marshal(map[string]any{
		"iss":   clientEmail,
		"scope": fcmScope,
		"aud":   tokenURI,
		"iat":   now.Unix(),
		"exp":   expiry.Unix(),
	})
	if err != nil {
		return "", time.Time{}, err
	}
	claimsB64 := base64.RawURLEncoding.EncodeToString(claims)
	signingInput := header + "." + claimsB64

	digest := sha256.Sum256([]byte(signingInput))
	sig, err := rsa.SignPKCS1v15(rand.Reader, privateKey, crypto.SHA256, digest[:])
	if err != nil {
		return "", time.Time{}, err
	}

	return signingInput + "." + base64.RawURLEncoding.EncodeToString(sig), expiry, nil
}

// parseRSAPrivateKeyFromPEM 解析服务账号 PEM 私钥（PKCS#8 或 PKCS#1）。
func parseRSAPrivateKeyFromPEM(pemContent string) (*rsa.PrivateKey, error) {
	block, _ := pem.Decode([]byte(pemContent))
	if block == nil {
		return nil, errors.New("private_key is not valid PEM")
	}
	if key, err := x509.ParsePKCS8PrivateKey(block.Bytes); err == nil {
		if rsaKey, ok := key.(*rsa.PrivateKey); ok {
			return rsaKey, nil
		}
		return nil, errors.New("private_key is not RSA")
	}
	return x509.ParsePKCS1PrivateKey(block.Bytes)
}
