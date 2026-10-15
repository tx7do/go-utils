// Package matrix Matrix 客户端-服务端 API（房间文本消息），零 SDK。
package matrix

import (
	"bytes"
	"context"
	crand "crypto/rand"
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

const (
	pushTimeout  = 10 * time.Second
	answerLimit  = 8 * 1024
	snippetLimit = 256
)

// Config Matrix 渠道配置。Target（房间 ID，如 !abc:example.org）按消息传。
type Config struct {
	// Ref 渠道配置的展示标识（ID、名字均可）：只进错误文案。
	Ref string

	// ServerURL homeserver 基地址（必填——Matrix 都是自建/托管域）。
	ServerURL string

	// AccessToken 用户 access token（登录或注册应用服务获得）。
	AccessToken string
}

// Service Matrix 渠道。无状态，可并发使用。
type Service struct {
	client *http.Client
	cfg    Config
}

// New 构造 Matrix 渠道。client 传 nil 用默认超时的客户端。
func New(client *http.Client, cfg Config) *Service {
	if client == nil {
		client = &http.Client{Timeout: pushTimeout}
	}
	return &Service{client: client, cfg: cfg}
}

// Name 渠道名。
func (s *Service) Name() string { return "matrix" }

// Send 给 Target（房间 ID）发一条 m.text 消息。
// 正文 = Title + Content（有标题时空行相连）；Receipt.ProviderID 带回 event_id。
func (s *Service) Send(ctx context.Context, msg *notify.Message) (*notify.Receipt, error) {
	receipt := &notify.Receipt{}
	if msg == nil || msg.Target == "" {
		return nil, errors.New("matrix room id is empty")
	}
	if s.cfg.ServerURL == "" || s.cfg.AccessToken == "" {
		return receipt, fmt.Errorf("%w: matrix config incomplete: server_url/access_token required", notify.ErrNotConfigured)
	}

	body := msg.Content
	if msg.Title != "" {
		body = msg.Title + "\n" + msg.Content
	}
	payload, err := json.Marshal(map[string]any{
		"msgtype": "m.text",
		"body":    body,
	})
	if err != nil {
		return receipt, fmt.Errorf("marshal payload failed: %w", err)
	}

	// 事务 ID 只需会话内唯一：客户端重复发送时对端据此去重。
	endpoint := fmt.Sprintf("%s/_matrix/client/v3/rooms/%s/send/m.text/%s?access_token=%s",
		strings.TrimRight(s.cfg.ServerURL, "/"), matrixRoomIDPath(msg.Target), crand.Text(),
		url.QueryEscape(s.cfg.AccessToken))
	req, err := http.NewRequestWithContext(ctx, http.MethodPut, endpoint, bytes.NewReader(payload))
	if err != nil {
		return receipt, fmt.Errorf("build request failed: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

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
		EventID string `json:"event_id"`
		ErrCode string `json:"errcode"`
		Error   string `json:"error"`
	}
	if err = json.Unmarshal(respBody, &answer); err != nil {
		return receipt, fmt.Errorf("peer answer is not JSON: %s", snippet(respBody))
	}
	if answer.EventID == "" {
		if answer.Error == "" {
			answer.Error = "no event_id"
		}
		return receipt, fmt.Errorf("peer answered errcode=%s: %s", answer.ErrCode, answer.Error)
	}
	receipt.ProviderID = answer.EventID
	return receipt, nil
}

// matrixRoomIDPath 房间 ID 进 URL 路径段：! 与 : 是路径合法字符，
// url.PathEscape 不会动它们，而 Matrix 规范要求显式转义（%21 / %3A）。
func matrixRoomIDPath(roomID string) string {
	return strings.NewReplacer("%", "%25", "!", "%21", ":", "%3A").Replace(roomID)
}

func snippet(body []byte) string {
	if len(body) > snippetLimit {
		return string(body[:snippetLimit]) + "…"
	}
	return string(body)
}
