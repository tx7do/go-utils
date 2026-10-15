package push

import (
	"context"
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/tx7do/go-utils/notify"
)

// VivoSendURL vivo Push 单播通知端点。
const VivoSendURL = "https://push-api.vivo.com.cn/message/notify/unicast"

// vivoInvalidCode vivo 响应码中"regId 无效"类（厂商码漂移只改这里）。
const vivoInvalidCode = 10080

// VivoConfig vivo 推送配置（appId/appKey/appSecret 鉴权三元组）。
type VivoConfig struct {
	// Ref 渠道配置的展示标识（ID、名字均可）：只进错误文案。
	Ref string

	AppID     string
	AppKey    string
	AppSecret string

	// SendURL 端点，留空用官方；测试可覆盖。
	SendURL string
}

// Vivo vivo 推送渠道。无状态，可并发使用。
// vivo 有每小时推送量窗等厂商硬限制：失败如实上报，不做静默整形。
type Vivo struct {
	client  *http.Client
	sendURL string
	cfg     VivoConfig
}

// NewVivo 构造 vivo 推送渠道。client 传 nil 用默认超时的客户端。
func NewVivo(client *http.Client, cfg VivoConfig) *Vivo {
	if client == nil {
		client = &http.Client{Timeout: pushTimeout}
	}
	return &Vivo{client: client, sendURL: defaultStr(cfg.SendURL, VivoSendURL), cfg: cfg}
}

// Name 渠道名。
func (c *Vivo) Name() string { return "vivo" }

// Send 给 Target（vivo regId）推一条通知栏消息。
// vivo 通知栏消息不支持数据面——Message.Params 被忽略，业务数据请自行并进 Content。
func (c *Vivo) Send(ctx context.Context, msg *notify.Message) (*notify.Receipt, error) {
	receipt := &notify.Receipt{}
	if msg == nil || msg.Target == "" {
		return nil, errors.New("vivo regId is empty")
	}
	if c.cfg.AppID == "" || c.cfg.AppKey == "" || c.cfg.AppSecret == "" {
		return receipt, notConfigured("vivo config incomplete: app_id/app_key/app_secret required")
	}

	timestamp := time.Now().Unix()
	ts := fmt.Sprintf("%d", timestamp)
	form := url.Values{}
	form.Set("appId", c.cfg.AppID)
	form.Set("appKey", c.cfg.AppKey)
	form.Set("timestamp", ts)
	form.Set("sign", VivoSign(c.cfg.AppKey, ts, c.cfg.AppSecret))
	form.Set("regId", msg.Target)
	form.Set("title", msg.Title)
	form.Set("content", msg.Content)

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

	var vr struct {
		Code int    `json:"code"`
		Msg  string `json:"msg"`
	}
	_ = json.Unmarshal(respBody, &vr)
	if resp.StatusCode >= 200 && resp.StatusCode < 300 && vr.Code == 0 {
		return receipt, nil
	}
	if vr.Code == vivoInvalidCode {
		return receipt, fmt.Errorf("%w: invalid regId (code %d)", ErrInvalidToken, vr.Code)
	}
	return receipt, fmt.Errorf("vivo push failed: http %d code %d: %s", resp.StatusCode, vr.Code, errSnippet(respBody))
}

// VivoSign vivo 请求签名：MD5(appKey + timestamp + appSecret) 小写 hex。
// 签名细节漂移只改这里。
func VivoSign(appKey, timestamp, appSecret string) string {
	sum := md5.Sum([]byte(appKey + timestamp + appSecret))
	return hex.EncodeToString(sum[:])
}
