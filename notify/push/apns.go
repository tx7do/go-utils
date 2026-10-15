package push

import (
	"bytes"
	"context"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/sha256"
	"crypto/x509"
	"encoding/asn1"
	"encoding/base64"
	"encoding/json"
	"encoding/pem"
	"errors"
	"fmt"
	"io"
	"math/big"
	"net/http"
	"sync"
	"time"

	"github.com/tx7do/go-utils/notify"
)

// APNs 端点：生产/沙箱两套域，按 Sandbox 切换；APIBase 可覆盖供测试。
const (
	APNsProductionURL = "https://api.push.apple.com"
	APNsSandboxURL    = "https://api.sandbox.push.apple.com"

	// apnsTokenTTL provider token 的官方有效期 1 小时；提前 10 分钟换新。
	apnsTokenTTL = 50 * time.Minute
)

// APNsConfig Apple APNs 配置（Token 鉴权：.p8 + KeyID + TeamID）。
//
// APNs 要求 provider 复用长连接（HTTP/2 多路复用），频繁建连会被视作 DoS——
// 因此一个配置一个实例、实例内复用连接与 provider token，多配置另起实例。
type APNsConfig struct {
	// Ref 渠道配置的展示标识（ID、名字均可）：只进错误文案。
	Ref string

	// TeamID / KeyID 开发者账号与 .p8 密钥标识。
	TeamID string
	KeyID  string

	// BundleID 推送主题（apns-topic 头）。
	BundleID string

	// PrivateKey .p8 私钥的 PEM 内容（PKCS#8 ECDSA）。
	PrivateKey string

	// Sandbox 走沙箱网关（开发构建的设备令牌只认沙箱）。
	Sandbox bool

	// APIBase 网关基地址，留空按 Sandbox 切默认；测试可覆盖。
	APIBase string
}

// APNs Apple APNs 推送渠道（HTTP/2，标准库实现，无 apns2 依赖）。
type APNs struct {
	client  *http.Client
	cfg     APNsConfig
	apiBase string
	key     *ecdsa.PrivateKey

	mu            sync.Mutex
	token         string
	tokenIssuedAt time.Time
}

// NewAPNs 构造 APNs 渠道。p8 解析在构造期完成——私钥配错当场暴露，不等第一次投递。
// client 传 nil 用默认的 HTTP/2 客户端。
func NewAPNs(client *http.Client, cfg APNsConfig) (*APNs, error) {
	if client == nil {
		client = &http.Client{Timeout: pushTimeout, Transport: &http.Transport{ForceAttemptHTTP2: true}}
	}
	key, err := parseAPNsAuthKey(cfg.PrivateKey)
	if err != nil {
		return nil, err
	}
	apiBase := cfg.APIBase
	if apiBase == "" {
		if cfg.Sandbox {
			apiBase = APNsSandboxURL
		} else {
			apiBase = APNsProductionURL
		}
	}
	return &APNs{client: client, cfg: cfg, apiBase: apiBase, key: key}, nil
}

// Name 渠道名。
func (c *APNs) Name() string { return "apns" }

// Send 给 Target（APNs device token，hex 串）推一条提醒。
// Title/Content 进 aps.alert（title/body），Params 作为自定义键展开进顶层 payload
// （客户端 notiInfo 之类的本地渲染路由按需取）。Receipt.ProviderID 带回 apns-id。
func (c *APNs) Send(ctx context.Context, msg *notify.Message) (*notify.Receipt, error) {
	receipt := &notify.Receipt{}
	if msg == nil || msg.Target == "" {
		return nil, errors.New("apns device token is empty")
	}
	if c.cfg.TeamID == "" || c.cfg.KeyID == "" || c.cfg.BundleID == "" {
		return receipt, notConfigured("apns config incomplete: team_id/key_id/bundle_id required")
	}

	jwt, err := c.providerToken()
	if err != nil {
		return receipt, err
	}

	payload := map[string]any{
		"aps": map[string]any{
			"alert": map[string]string{
				"title": msg.Title,
				"body":  msg.Content,
			},
			"sound": "default",
		},
	}
	for k, v := range msg.Params {
		payload[k] = v
	}
	body, err := json.Marshal(payload)
	if err != nil {
		return receipt, fmt.Errorf("marshal payload failed: %w", err)
	}

	endpoint := c.apiBase + "/3/device/" + msg.Target
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewReader(body))
	if err != nil {
		return receipt, fmt.Errorf("build request failed: %w", err)
	}
	req.Header.Set("Authorization", "bearer "+jwt)
	req.Header.Set("apns-topic", c.cfg.BundleID)
	req.Header.Set("apns-push-type", "alert")
	req.Header.Set("apns-priority", "10")
	req.Header.Set("apns-expiration", "0")

	resp, err := c.client.Do(req)
	if err != nil {
		return receipt, err
	}
	defer func() { _ = resp.Body.Close() }()
	respBody, _ := io.ReadAll(io.LimitReader(resp.Body, answerLimit))

	receipt.ProviderID = resp.Header.Get("apns-id")
	switch {
	case resp.StatusCode == http.StatusOK:
		return receipt, nil
	case resp.StatusCode == http.StatusGone ||
		resp.Header.Get("apns-error-reason") == "Unregistered" ||
		resp.Header.Get("apns-error-reason") == "BadDeviceToken":
		// 令牌失效（应用卸载/系统重置）：调用方删行自清理，不算投递失败重试。
		return receipt, fmt.Errorf("%w: http %s", ErrInvalidToken, resp.Status)
	default:
		return receipt, fmt.Errorf("apns push rejected: status=%d reason=%s: %s",
			resp.StatusCode, resp.Header.Get("apns-error-reason"), errSnippet(respBody))
	}
}

// providerToken 取（缓存的）provider token：ES256 JWT，官方有效期 1 小时。
func (c *APNs) providerToken() (string, error) {
	c.mu.Lock()
	if c.token != "" && time.Since(c.tokenIssuedAt) < apnsTokenTTL {
		tok := c.token
		c.mu.Unlock()
		return tok, nil
	}
	c.mu.Unlock()

	tok, err := MintAPNsToken(c.key, c.cfg.KeyID, c.cfg.TeamID)
	if err != nil {
		return "", err
	}

	c.mu.Lock()
	c.token = tok
	c.tokenIssuedAt = time.Now()
	c.mu.Unlock()
	return tok, nil
}

// MintAPNsToken 签一枚 APNs provider token（ES256 JWT）：
// header {"alg":"ES256","kid":KeyID}，claims {"iss":TeamID,"iat":now}，
// 签名为 P-256 ECDSA 的 ASN.1 DER——与 Apple 官方文档逐字对应。
func MintAPNsToken(key *ecdsa.PrivateKey, keyID, teamID string) (string, error) {
	if key == nil {
		return "", errors.New("apns auth key is nil")
	}
	header := base64.RawURLEncoding.EncodeToString([]byte(
		`{"alg":"ES256","kid":"` + keyID + `"}`))
	claims, err := json.Marshal(map[string]any{
		"iss": teamID,
		"iat": time.Now().Unix(),
	})
	if err != nil {
		return "", err
	}
	signingInput := header + "." + base64.RawURLEncoding.EncodeToString(claims)

	digest := sha256Sum([]byte(signingInput))
	r, s, err := ecdsa.Sign(rand.Reader, key, digest)
	if err != nil {
		return "", err
	}
	sig, err := asn1EcdsaSignature(r, s)
	if err != nil {
		return "", err
	}
	return signingInput + "." + base64.RawURLEncoding.EncodeToString(sig), nil
}

// sha256Sum 摘要助手（Sum 返回定长数组，这里转切片）。
func sha256Sum(b []byte) []byte {
	sum := sha256.Sum256(b)
	return sum[:]
}

// asn1EcdsaSignature ECDSA 签名的 ASN.1 DER 编码（Apple 要的就是这个形态）。
func asn1EcdsaSignature(r, s *big.Int) ([]byte, error) {
	return asn1.Marshal(struct{ R, S *big.Int }{r, s})
}

// parseAPNsAuthKey 解析 .p8 私钥（PKCS#8 ECDSA P-256）。
func parseAPNsAuthKey(pemContent string) (*ecdsa.PrivateKey, error) {
	if pemContent == "" {
		return nil, notConfigured("apns config incomplete: private_key required")
	}
	block, _ := pem.Decode([]byte(pemContent))
	if block == nil {
		return nil, notConfigured("apns private_key is not valid PEM")
	}
	key, err := x509.ParsePKCS8PrivateKey(block.Bytes)
	if err != nil {
		return nil, notConfigured("apns private_key parse failed: %v", err)
	}
	ecKey, ok := key.(*ecdsa.PrivateKey)
	if !ok {
		return nil, notConfigured("apns private_key is not ECDSA")
	}
	if ecKey.Curve != elliptic.P256() {
		return nil, notConfigured("apns private_key curve is not P-256")
	}
	return ecKey, nil
}
