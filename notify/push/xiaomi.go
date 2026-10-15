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

	"github.com/tx7do/go-utils/notify"
)

// XiaomiSendURL MiPush 通知栏消息端点（通知栏形态，非透传）。
const XiaomiSendURL = "https://api.xmpush.xiaomi.com/v2/message/regid"

// xiaomiInvalidCodes MiPush 响应码中"目标 regId 无效"类（厂商码漂移只改这里的表）。
var xiaomiInvalidCodes = map[int]bool{
	20201: true,
	20202: true,
}

// XiaomiConfig 小米推送（MiPush）配置。
type XiaomiConfig struct {
	// Ref 渠道配置的展示标识（ID、名字均可）：只进错误文案。
	Ref string

	// AppSecret MiPush 的应用 Secret（请求头 `Authorization: Key=<appSecret>`）。
	AppSecret string

	// SendURL 端点，留空用官方；测试可覆盖。
	SendURL string
}

// Xiaomi 小米推送渠道。无状态，可并发使用。
type Xiaomi struct {
	client  *http.Client
	sendURL string
	cfg     XiaomiConfig
}

// NewXiaomi 构造小米推送渠道。client 传 nil 用默认超时的客户端。
func NewXiaomi(client *http.Client, cfg XiaomiConfig) *Xiaomi {
	if client == nil {
		client = &http.Client{Timeout: pushTimeout}
	}
	return &Xiaomi{client: client, sendURL: defaultStr(cfg.SendURL, XiaomiSendURL), cfg: cfg}
}

// Name 渠道名。
func (c *Xiaomi) Name() string { return "xiaomi" }

// Send 给 Target（MiPush regId）推一条通知栏消息。
// Title/Content 进通知面板（title/description），Params 以 extra.{key} 表单键透传——
// 通知栏消息点击经 extra.uri 深链回端内。
func (c *Xiaomi) Send(ctx context.Context, msg *notify.Message) (*notify.Receipt, error) {
	receipt := &notify.Receipt{}
	if msg == nil || msg.Target == "" {
		return nil, errors.New("xiaomi regId is empty")
	}
	if c.cfg.AppSecret == "" {
		return receipt, notConfigured("xiaomi config incomplete: app_secret required")
	}

	form := url.Values{}
	form.Set("registration_id", msg.Target)
	form.Set("title", msg.Title)
	form.Set("description", msg.Content)
	form.Set("pass_through", "0")
	for k, v := range msg.Params {
		form.Set("extra."+k, v)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.sendURL, strings.NewReader(form.Encode()))
	if err != nil {
		return receipt, fmt.Errorf("build request failed: %w", err)
	}
	req.Header.Set("Authorization", "Key="+c.cfg.AppSecret)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := c.client.Do(req)
	if err != nil {
		return receipt, err
	}
	defer func() { _ = resp.Body.Close() }()
	respBody, _ := io.ReadAll(io.LimitReader(resp.Body, answerLimit))

	var xr struct {
		Result bool `json:"result"`
		Code   int  `json:"code"`
	}
	_ = json.Unmarshal(respBody, &xr)
	if resp.StatusCode >= 200 && resp.StatusCode < 300 && xr.Code == 0 {
		return receipt, nil
	}
	if xiaomiInvalidCodes[xr.Code] {
		return receipt, fmt.Errorf("%w: invalid regid (code %d)", ErrInvalidToken, xr.Code)
	}
	return receipt, fmt.Errorf("xiaomi push failed: http %d code %d: %s", resp.StatusCode, xr.Code, errSnippet(respBody))
}
