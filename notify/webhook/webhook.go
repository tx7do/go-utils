// Package webhook HTTP 回调渠道：把通知正文 POST 到对端地址。
//
// 发成什么形状（签名怎么算、放哪、正文长什么样、用什么判成败）由 Config 决定：
// 内置钉钉/飞书/企微三家群机器人的签名风格，外加 HMAC 自有方案（CUSTOM）与不签名（NONE）。
// 出站地址取自 Message.Target——"这条到底发去了哪里"由调用方的投递记录回答，
// 配置里的 URL 只作登记与预检门槛。
package webhook

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"

	"github.com/tx7do/go-utils/notify"
)

const (
	// HeaderSignature / HeaderTimestamp CUSTOM 风格的签名头：对端用 Secret 按
	// `sha256=<hex(hmac("<timestamp>.", body))>` 复算比对。把时间戳折进签名而不是只签 body，
	// 是为了让"同一份 body 重放"必须同时重放时间戳，对端据此拒收旧消息；
	// 只签 body 的话重放是逐字节免费的。其余风格不把签名放头里（见 style.go）。
	HeaderSignature = "X-Gw-Signature"
	HeaderTimestamp = "X-Gw-Timestamp"

	// webhookAnswerBytes 读对端应答的上限：判"200 但其实没收"要看 body，
	// 而各家机器人的结论都是一个几百字节的小 JSON。留 4KB 是给带中文 errmsg 的余量，
	// 再多就当它不是 JSON 放过（checkProviderResponse 不猜）。
	webhookAnswerBytes = 4 * 1024

	// webhookMaxResponseBytes 失败时截入错误文案的对端响应体上限。
	// 只用于排障（"对方 403 说了什么"），所以刻意小：对端可能返回一整页 HTML。
	webhookMaxResponseBytes = 512

	defaultUserAgent = "go-utils-notify/1.0"
)

// Config 一条 WEBHOOK 渠道的配置。
type Config struct {
	// Ref 渠道配置的展示标识（ID、名字均可）：只进错误文案——库里可能配了好几条，
	// 报错要能定位"是哪条"。
	Ref string

	// Secret 签名密钥。各风格对它的用法见 style.go；空串即不签名。
	Secret string

	// SignStyle 签名风格：CUSTOM / NONE / DINGTALK / FEISHU / WECOM。
	// 空串按 CUSTOM（历史行为），认不出的值报错而不是静默兜底。
	SignStyle string

	// PayloadTemplate 载荷模板，占位符 {{name}} 单趟替换；空串表示用默认正文。
	PayloadTemplate string

	// DefaultPayload CUSTOM/NONE 且未配模板时的正文钩子，返回值经 json.Marshal 出站
	// （HTML 转义开启——已按既有字节形状配好验签的对端不该收到字段等价但转义不同的 JSON）。
	// nil 用库内置默认 {"title","content","delivered_at"}。
	DefaultPayload func(msg *notify.Message) (any, error)

	// UserAgent 出站请求头，空串用库默认。
	UserAgent string
}

// Service WEBHOOK 渠道。client 建议传 notify.NewGuardedClient() 的产物；
// 传 nil 用生产默认的 guarded client（不放过任何内网地址）。
type Service struct {
	client *http.Client
	cfg    Config
}

// New 构造 WEBHOOK 渠道。
func New(client *http.Client, cfg Config) *Service {
	if client == nil {
		client = notify.NewGuardedClient(false)
	}
	if cfg.UserAgent == "" {
		cfg.UserAgent = defaultUserAgent
	}
	return &Service{client: client, cfg: cfg}
}

// Name 渠道名。
func (s *Service) Name() string { return "webhook" }

// configErr 配置类失败：发生在联系对端之前，重试也不会变好。
// 双 %w：外层哨兵给台账分类用，内层错误（如 ErrBlockedTarget）保持 errors.Is 可达。
func configErr(err error) error {
	return fmt.Errorf("%w: %w", notify.ErrNotConfigured, err)
}

// Send 投递一次回调。返回 error 一律是"投递失败"；配置类失败包 ErrNotConfigured，
// 调用方据此与"拨过号被拒"区分开。
func (s *Service) Send(ctx context.Context, msg *notify.Message) (*notify.Receipt, error) {
	receipt := &notify.Receipt{}
	if msg == nil || msg.Target == "" {
		return nil, errors.New("webhook target url is empty")
	}

	endpoint, err := NormalizeURL(msg.Target)
	if err != nil {
		return receipt, configErr(err)
	}

	style, err := resolveStyle(s.cfg.SignStyle)
	if err != nil {
		return receipt, configErr(s.prefix(err))
	}

	outbound, err := s.buildOutbound(style, msg)
	if err != nil {
		return receipt, configErr(err)
	}

	if outbound.query != nil {
		// 钉钉把签名拼在 URL query 上：保留对端原本要看的 access_token 等参数，只追加两个键。
		u, parseErr := url.Parse(endpoint)
		if parseErr != nil {
			return receipt, configErr(parseErr)
		}
		q := u.Query()
		for key, values := range outbound.query {
			q.Del(key)
			for _, v := range values {
				q.Add(key, v)
			}
		}
		u.RawQuery = q.Encode()
		endpoint = u.String()
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewReader(outbound.body))
	if err != nil {
		return receipt, fmt.Errorf("build webhook request failed: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json; charset=utf-8")
	httpReq.Header.Set("User-Agent", s.cfg.UserAgent)
	if outbound.signatureHeader != "" {
		httpReq.Header.Set(HeaderTimestamp, outbound.timestampHeader)
		httpReq.Header.Set(HeaderSignature, outbound.signatureHeader)
	}

	resp, err := s.client.Do(httpReq)
	if err != nil {
		if errors.Is(err, notify.ErrBlockedTarget) {
			// 一次都没联系对端 = 地址不合法，属"这条投递没法发"（SKIPPED），不是"发了但被拒"。
			return receipt, configErr(err)
		}
		return receipt, err
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	// 读满一个足够装下"对端那句结论"的量：各家机器人的应答都是几百字节的小 JSON。
	// 只读不排空就 Close，keep-alive 连接会作废——但回调是低频出站，不值得为它多读一整页 HTML。
	snippet, readErr := io.ReadAll(io.LimitReader(resp.Body, webhookAnswerBytes))

	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		return receipt, fmt.Errorf("peer answered %s: %s", resp.Status, truncateForError(snippet))
	}

	// 2xx 之后才轮到 body 说话：钉钉/飞书/企微在拦下消息时给的仍然是 200。
	if verdictFieldOf(style) != "" {
		if readErr != nil && readErr != io.EOF {
			return receipt, fmt.Errorf("read peer answer: %w", readErr)
		}
		if verdictErr := checkProviderResponse(style, snippet); verdictErr != nil {
			return receipt, verdictErr
		}
	}

	return receipt, nil
}

// Precheck 配置自检：风格认得出来、模板渲染得出合法 JSON。只渲染不拨号。
//
// 用一份假正文渲染：模板能引用的变量与内容无关，"能不能渲染成合法 JSON"也与之无关——
// 调用方在入队一次性消息（如验证码）之前先跑它，别等消息躺在队列里失败四次。
// 失败一律包 ErrNotConfigured：这是配置问题。
func (s *Service) Precheck() error {
	style, err := resolveStyle(s.cfg.SignStyle)
	if err != nil {
		return configErr(s.prefix(err))
	}
	if _, err = s.buildOutbound(style, &notify.Message{Title: "precheck", Content: "precheck"}); err != nil {
		return configErr(err)
	}
	return nil
}

// NormalizeURL 只放行绝对 http(s) 地址。
//
// 相对地址（"//10.0.0.5/x"、"/y"）在这一步被挡掉：它们到了 http.Client 那里
// 会报一个不含原址的错，排障时看不出"配在库里的值本来就不是绝对地址"。
func NormalizeURL(raw string) (string, error) {
	u, err := url.Parse(raw)
	if err != nil {
		return "", fmt.Errorf("webhook url %q is invalid: %w", raw, err)
	}
	if u.Scheme != "http" && u.Scheme != "https" {
		return "", fmt.Errorf("webhook url %q must be http(s)", raw)
	}
	if u.Host == "" {
		return "", fmt.Errorf("webhook url %q has no host", raw)
	}
	return u.String(), nil
}
