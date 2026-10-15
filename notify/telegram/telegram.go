// Package telegram Telegram Bot API 渠道：把通知正文用 sendMessage 发到指定 chat。
//
// 走裸 HTTP，零 SDK。实现时踩过的一个坑钉在这里：Telegram 会用 HTTP 200 返回
// {"ok":false,...}（代理/网关场景尤其常见），只看状态码会把"对方没收"记成发送成功——
// 所以 2xx 之后还要看 body 的 ok 字段。
package telegram

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"time"

	"github.com/tx7do/go-utils/notify"
)

const (
	defaultAPIBase = "https://api.telegram.org"

	// telegramTimeout 单次出站请求的总超时。
	telegramTimeout = 10 * time.Second

	// telegramAnswerBytes 读对端应答的上限：结论是一个几百字节的小 JSON。
	telegramAnswerBytes = 4 * 1024
)

// Config Telegram 渠道配置。
type Config struct {
	// Ref 渠道配置的展示标识（ID、名字均可）：只进错误文案。
	Ref string

	// BotToken 机器人令牌（BotFather 签发）。
	BotToken string

	// APIBase 服务端基地址，留空用官方 API；测试与自建网关可覆盖。
	APIBase string

	// ParseMode 可选解析模式：MarkdownV2 / HTML / 空串（纯文本）。
	// 配了之后 Message.Title/Content 里的标记按该模式解析，转义是调用方的事。
	ParseMode string
}

// Service Telegram 渠道。无状态：配置在构造时注入，可并发使用。
type Service struct {
	client *http.Client
	cfg    Config
}

// New 构造 Telegram 渠道。client 传 nil 用默认超时的客户端。
func New(client *http.Client, cfg Config) *Service {
	if client == nil {
		client = &http.Client{Timeout: telegramTimeout}
	}
	if cfg.APIBase == "" {
		cfg.APIBase = defaultAPIBase
	}
	return &Service{client: client, cfg: cfg}
}

// Name 渠道名。
func (s *Service) Name() string { return "telegram" }

// Send 投递一条消息：Target 为 chat_id（数字或 @频道名），Title 与 Content 以空行相连
// 作为消息文本。Receipt.ProviderID 带回 Telegram 的 message_id。
func (s *Service) Send(ctx context.Context, msg *notify.Message) (*notify.Receipt, error) {
	receipt := &notify.Receipt{}
	if msg == nil || msg.Target == "" {
		return nil, errors.New("telegram chat_id is empty")
	}
	if s.cfg.BotToken == "" {
		return nil, fmt.Errorf("%w: telegram config missing bot_token", notify.ErrNotConfigured)
	}

	text := msg.Content
	if msg.Title != "" {
		text = msg.Title + "\n\n" + msg.Content
	}

	payload := map[string]string{
		"chat_id": msg.Target,
		"text":    text,
	}
	if s.cfg.ParseMode != "" {
		payload["parse_mode"] = s.cfg.ParseMode
	}
	body, err := json.Marshal(payload)
	if err != nil {
		return receipt, fmt.Errorf("marshal payload failed: %w", err)
	}

	endpoint := s.cfg.APIBase + "/bot" + s.cfg.BotToken + "/sendMessage"
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewReader(body))
	if err != nil {
		return receipt, fmt.Errorf("build request failed: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := s.client.Do(req)
	if err != nil {
		return receipt, err
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	respBody, readErr := io.ReadAll(io.LimitReader(resp.Body, telegramAnswerBytes))
	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		return receipt, fmt.Errorf("peer answered %s: %s", resp.Status, snippet(respBody, readErr))
	}
	if readErr != nil && readErr != io.EOF {
		return receipt, fmt.Errorf("read peer answer: %w", readErr)
	}

	// HTTP 200 不代表发送成功：Telegram 用 body 的 ok=false 报业务失败。
	var answer struct {
		Ok          bool   `json:"ok"`
		ErrorCode   int    `json:"error_code"`
		Description string `json:"description"`
		Result      struct {
			MessageID int64 `json:"message_id"`
		} `json:"result"`
	}
	if err = json.Unmarshal(respBody, &answer); err != nil {
		return receipt, fmt.Errorf("peer answer is not JSON: %s", snippet(respBody, readErr))
	}
	if !answer.Ok {
		if answer.Description == "" {
			answer.Description = "no description"
		}
		return receipt, fmt.Errorf("peer answered ok=false code=%d: %s", answer.ErrorCode, answer.Description)
	}

	if answer.Result.MessageID != 0 {
		receipt.ProviderID = strconv.FormatInt(answer.Result.MessageID, 10)
	}
	return receipt, nil
}

// snippet 进错误文案的响应片段截断：对端可能回一整页 HTML。
func snippet(body []byte, readErr error) string {
	const limit = 512
	if readErr != nil && readErr != io.EOF {
		return "<unreadable>"
	}
	if len(body) <= limit {
		return string(body)
	}
	return string(body[:limit]) + "…"
}
