// Package yunpian 云片短信（single_send 自由文本形态），零 SDK。
//
// 与 notify/sms 的分工：阿里云/腾讯云是模板短信（正文出自审核过的模板，
// 只传占位值）；云片的 single_send 收**整段自由文本**（含签名，如
// "【GoWind】您的验证码是 123456"），签名合规是调用方的事。
package yunpian

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/tx7do/go-utils/notify"
	"github.com/tx7do/go-utils/notify/sms"
)

// DefaultAPIBase 云片官方 API 域。
const DefaultAPIBase = "https://sms.yunpian.com"

const (
	pushTimeout  = 10 * time.Second
	answerLimit  = 4 * 1024
	snippetLimit = 256
)

// Config 云片短信配置。Target（手机号）按消息传。
type Config struct {
	// Ref 渠道配置的展示标识（ID、名字均可）：只进错误文案。
	Ref string

	// ApiKey 云片账户的 API Key。
	ApiKey string

	// APIBase 服务基地址，留空用官方；测试可覆盖。
	APIBase string
}

// Service 云片短信渠道。无状态，可并发使用。
type Service struct {
	client  *http.Client
	apiBase string
	cfg     Config
}

// New 构造云片短信渠道。client 传 nil 用默认超时的客户端。
func New(client *http.Client, cfg Config) *Service {
	if client == nil {
		client = &http.Client{Timeout: pushTimeout}
	}
	if cfg.APIBase == "" {
		cfg.APIBase = DefaultAPIBase
	}
	return &Service{client: client, apiBase: cfg.APIBase, cfg: cfg}
}

// Name 渠道名。
func (s *Service) Name() string { return "yunpian" }

// Send 给 Target（手机号，CN 裸号或 E.164）发一条自由文本短信。
// 正文 = Content 原样出站（须含已报备的签名【XX】，否则对端拒）；
// Title 忽略——按 67 字符计费的通道不替调用方注水。
// Receipt.ProviderID 带回云片的 sid。
func (s *Service) Send(ctx context.Context, msg *notify.Message) (*notify.Receipt, error) {
	receipt := &notify.Receipt{}
	if msg == nil || msg.Content == "" {
		return nil, errors.New("yunpian sms text is empty")
	}
	if s.cfg.ApiKey == "" {
		return receipt, fmt.Errorf("%w: yunpian config incomplete: api_key required", notify.ErrNotConfigured)
	}

	// 目标归一在出站前裁决：号码不合法一次都不碰对端，归配置/目标类。
	national, _, err := sms.NormalizePhoneNumber(msg.Target)
	if err != nil {
		return receipt, fmt.Errorf("%w: %v", notify.ErrNotConfigured, err)
	}

	form := url.Values{}
	form.Set("apikey", s.cfg.ApiKey)
	form.Set("mobile", national)
	form.Set("text", msg.Content)

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, s.apiBase+"/v2/sms/single_send.json", strings.NewReader(form.Encode()))
	if err != nil {
		return receipt, fmt.Errorf("build request failed: %w", err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

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
		Sid  int64  `json:"sid"`
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
	if answer.Sid != 0 {
		receipt.ProviderID = strconv.FormatInt(answer.Sid, 10)
	}
	return receipt, nil
}

func snippet(body []byte) string {
	if len(body) > snippetLimit {
		return string(body[:snippetLimit]) + "…"
	}
	return string(body)
}
