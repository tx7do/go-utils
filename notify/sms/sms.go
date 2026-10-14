// Package sms 短信渠道：阿里云 Dysmsapi 与腾讯云 SMS 的裸 HTTP 实现（零 SDK）。
//
// 协议差异一览：
//   - 阿里云是 RPC 风格：全部参数排序拼接做 HMAC-SHA1 签名，请求是一张普通表单；
//   - 腾讯云是 TC3-HMAC-SHA256：签名不只盖参数，还盖整个 JSON 请求体，手机号只认 E.164。
//
// 两家都把业务结论放在 HTTP 200 的 body 里（阿里云 Code=="OK"、腾讯云无 Error 且
// SendStatusSet[0].Code=="Ok"），只看状态码会把"对方没收"记成发送成功。
package sms

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"sort"
	"time"

	"github.com/tx7do/go-utils/notify"
)

// smsTimeout 单次出站请求的总超时（拨号 + 写请求 + 读响应）。
const smsTimeout = 10 * time.Second

// Provider 服务商标识。存储里这一列的值域与这里的常量逐字相同（按名字配对）。
const (
	ProviderAliyun  = "ALIYUN"
	ProviderTencent = "TENCENT"
)

// Config 短信渠道配置。Provider 决定哪些字段必填（见 Precheck）。
type Config struct {
	// Provider 服务商：ALIYUN / TENCENT。
	Provider string

	// AccessKeyID / AccessSecret 服务商访问密钥对。
	AccessKeyID  string
	AccessSecret string

	// SignName 服务商侧已备案的短信签名。
	SignName string

	// TemplateCode 短信模板 ID（阿里云正文含 ${code}，腾讯云正文含 {1}）。
	TemplateCode string

	// AppID 腾讯云 SmsSdkAppId（必填）；阿里云留空。
	AppID string
	// Region 地域：腾讯云必填（空兜默认 ap-guangzhou）；阿里云留空兜 cn-hangzhou。
	Region string
}

// Service 短信渠道。无状态：配置在构造时注入，可并发使用。
type Service struct {
	client *http.Client
	cfg    Config
}

// New 构造短信渠道。Provider 认不出当场报错——拼错的服务商名不该等到第一次投递才暴露。
// client 传 nil 用默认超时的客户端。
func New(client *http.Client, cfg Config) (*Service, error) {
	switch cfg.Provider {
	case ProviderAliyun, ProviderTencent:
	default:
		return nil, fmt.Errorf("unknown sms provider %q (expected ALIYUN / TENCENT)", cfg.Provider)
	}
	if client == nil {
		client = &http.Client{Timeout: smsTimeout}
	}
	return &Service{client: client, cfg: cfg}, nil
}

// Name 渠道名。
func (s *Service) Name() string { return "sms" }

// Send 投递一条短信：Target 为手机号（见 NormalizePhoneNumber），
// Params 为模板参数——阿里云按参数名整表传 JSON（模板写 ${name}），
// 腾讯云按参数名排序取值传位置参数（模板写 {1}{2}…，单参数无感，多参数请核对模板占位顺序）。
// Params 为空直接拒绝：短信没有自由文本通道，正文出自服务商侧审核过的模板。
func (s *Service) Send(ctx context.Context, msg *notify.Message) (*notify.Receipt, error) {
	receipt := &notify.Receipt{}
	if msg == nil || msg.Target == "" {
		return nil, errors.New("sms target is empty")
	}
	if len(msg.Params) == 0 {
		return nil, errors.New("sms template params are empty: SMS has no free-text channel, the body comes from a provider-side approved template")
	}

	national, e164, err := NormalizePhoneNumber(msg.Target)
	if err != nil {
		// 目标错了没有"发到别家去"的软失败，也没碰过对端：归配置/目标类。
		return receipt, fmt.Errorf("%w: %v", notify.ErrNotConfigured, err)
	}

	switch s.cfg.Provider {
	case ProviderAliyun:
		if err = s.sendAliyun(ctx, national, msg.Params); err != nil {
			return receipt, err
		}
	case ProviderTencent:
		if err = s.sendTencent(ctx, e164, templateParamsOrdered(msg.Params)); err != nil {
			return receipt, err
		}
	}
	return receipt, nil
}

// Precheck 配置自检：按服务商核对必填项，不拨号。
// 错误文案直接点名缺哪个字段；用切片保序，缺多项时报第一个。
func (s *Service) Precheck() error {
	var required []struct{ name, value string }
	switch s.cfg.Provider {
	case ProviderAliyun:
		required = []struct{ name, value string }{
			{"sms_access_key_id", s.cfg.AccessKeyID},
			{"sms_access_secret", s.cfg.AccessSecret},
			{"sms_sign_name", s.cfg.SignName},
			{"sms_template_code", s.cfg.TemplateCode},
		}
	case ProviderTencent:
		required = []struct{ name, value string }{
			{"sms_access_key_id", s.cfg.AccessKeyID},
			{"sms_access_secret", s.cfg.AccessSecret},
			{"sms_sign_name", s.cfg.SignName},
			{"sms_template_code", s.cfg.TemplateCode},
			{"sms_app_id", s.cfg.AppID},
			{"sms_region", s.cfg.Region},
		}
	default:
		return fmt.Errorf("%w: unknown sms provider %q (expected ALIYUN / TENCENT)", notify.ErrNotConfigured, s.cfg.Provider)
	}
	for _, item := range required {
		if item.value == "" {
			return fmt.Errorf("%w: %s is required for provider %s", notify.ErrNotConfigured, item.name, s.cfg.Provider)
		}
	}
	return nil
}

// templateParamsOrdered 腾讯云的位置参数：按参数名排序取值。
// 单参数（验证码 code）无感；多参数时模板 {1}{2} 的对应关系由这个排序决定。
func templateParamsOrdered(params map[string]string) []string {
	keys := make([]string, 0, len(params))
	for k := range params {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	values := make([]string, 0, len(keys))
	for _, k := range keys {
		values = append(values, params[k])
	}
	return values
}
