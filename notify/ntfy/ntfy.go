// Package ntfy 主题推送（ntfy.sh 及自建服务）：JSON publish，零依赖。
package ntfy

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

// DefaultServerURL ntfy 官方服务域。
const DefaultServerURL = "https://ntfy.sh"

// Priority 优先级档位（1 最低 ~ 5 最高；0 默认）。
const (
	PriorityMin     = 1
	PriorityDefault = 0
	PriorityHigh    = 4
	PriorityMax     = 5
)

// Config ntfy 渠道配置。
type Config struct {
	// Ref 渠道配置的展示标识（ID、名字均可）：只进错误文案。
	Ref string

	// ServerURL 服务基地址，留空用官方 ntfy.sh；自建服务可覆盖。
	ServerURL string

	// Priority 默认优先级（1~5，0 默认）；Message.Params["priority"] 可按条覆盖。
	Priority int

	// User / Password 基础认证（自建服务常见）；与 AccessToken 二选一。
	User     string
	Password string

	// AccessToken 访问令牌（ntfy access tokens，tk_ 开头）；优先于 User/Password。
	AccessToken string

	// Tags 默认标签（emoji shortcode 或文本）。
	Tags []string
}

// Service ntfy 渠道。无状态，可并发使用。
type Service struct {
	client *http.Client
	cfg    Config
}

// New 构造 ntfy 渠道。client 传 nil 用默认超时的客户端。
func New(client *http.Client, cfg Config) *Service {
	if client == nil {
		client = &http.Client{Timeout: pushTimeout}
	}
	if cfg.ServerURL == "" {
		cfg.ServerURL = DefaultServerURL
	}
	return &Service{client: client, cfg: cfg}
}

// Name 渠道名。
func (s *Service) Name() string { return "ntfy" }

// Send 给 Target（ntfy 主题名）发一条消息。JSON publish 走根路径——
// 主题名/标题/正文都在 body 里，绕开表单转义与头注入两类坑。
// Receipt.ProviderID 带回 ntfy 的消息 ID。
func (s *Service) Send(ctx context.Context, msg *notify.Message) (*notify.Receipt, error) {
	receipt := &notify.Receipt{}
	if msg == nil || msg.Target == "" {
		return nil, errors.New("ntfy topic is empty")
	}

	payload := map[string]any{
		"topic":   msg.Target,
		"title":   msg.Title,
		"message": msg.Content,
	}
	if len(s.cfg.Tags) > 0 {
		payload["tags"] = s.cfg.Tags
	}
	priority := s.cfg.Priority
	if p, ok := msg.Params["priority"]; ok {
		n, err := strconv.Atoi(p)
		if err != nil {
			return receipt, fmt.Errorf("params[priority] %q is not a number 0-5: %w", p, err)
		}
		priority = n
	}
	if priority != 0 {
		payload["priority"] = priority
	}
	body, err := json.Marshal(payload)
	if err != nil {
		return receipt, fmt.Errorf("marshal payload failed: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, s.cfg.ServerURL, bytes.NewReader(body))
	if err != nil {
		return receipt, fmt.Errorf("build request failed: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	if s.cfg.AccessToken != "" {
		req.Header.Set("Authorization", "Bearer "+s.cfg.AccessToken)
	} else if s.cfg.User != "" {
		req.SetBasicAuth(s.cfg.User, s.cfg.Password)
	}

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
		ID string `json:"id"`
	}
	_ = json.Unmarshal(respBody, &answer)
	receipt.ProviderID = answer.ID
	return receipt, nil
}

const (
	pushTimeout  = 10 * time.Second
	answerLimit  = 4 * 1024
	snippetLimit = 256
)

func snippet(body []byte) string {
	if len(body) > snippetLimit {
		return string(body[:snippetLimit]) + "…"
	}
	return string(body)
}
