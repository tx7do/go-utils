package telegram

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/tx7do/go-utils/notify"
)

// fakeAPI 本地假 Bot API：记录请求，按参数回 ok/错误。
type fakeAPI struct {
	t *testing.T

	hit     int
	payload map[string]string

	status int // 默认 200
	answer string
}

func newFakeAPI(t *testing.T) *fakeAPI {
	return &fakeAPI{t: t, status: http.StatusOK}
}

func (f *fakeAPI) start() *Service {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		f.hit++
		require.Equal(f.t, "/botTOKEN/sendMessage", r.URL.Path, "token 必须在路径上，query 里传等于把它泄进访问日志")
		require.Equal(f.t, http.MethodPost, r.Method)
		_ = json.NewDecoder(r.Body).Decode(&f.payload)
		w.WriteHeader(f.status)
		_, _ = w.Write([]byte(f.answer))
	}))
	f.t.Cleanup(srv.Close)

	return New(srv.Client(), Config{Ref: "7", BotToken: "TOKEN", APIBase: srv.URL})
}

func TestSendDeliversAndParsesMessageID(t *testing.T) {
	api := newFakeAPI(t)
	api.answer = `{"ok":true,"result":{"message_id":42,"chat_id":123}}`
	svc := api.start()

	receipt, err := svc.Send(context.Background(), &notify.Message{
		Target:  "123456",
		Title:   "重置验证码",
		Content: "您的验证码是 123456",
	})
	require.NoError(t, err)
	require.Equal(t, "42", receipt.ProviderID, "Telegram 的 message_id 要进回执，对账全靠它")

	require.Equal(t, "123456", api.payload["chat_id"])
	require.Equal(t, "重置验证码\n\n您的验证码是 123456", api.payload["text"], "标题与正文以空行相连")
	require.NotContains(t, api.payload, "parse_mode", "没配解析模式就不传：对端按纯文本处理")
}

func TestSendParseModePassedThrough(t *testing.T) {
	api := newFakeAPI(t)
	api.answer = `{"ok":true,"result":{"message_id":1}}`
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewDecoder(r.Body).Decode(&api.payload)
		_, _ = w.Write([]byte(api.answer))
	}))
	t.Cleanup(srv.Close)

	svc := New(srv.Client(), Config{BotToken: "t", APIBase: srv.URL, ParseMode: "HTML"})
	_, err := svc.Send(context.Background(), &notify.Message{Target: "1", Content: "<b>hi</b>"})
	require.NoError(t, err)
	require.Equal(t, "HTML", api.payload["parse_mode"])
}

// TestSendOKFalseWithin200 Telegram 会用 HTTP 200 返回 ok=false（代理/网关场景尤其常见）：
// 只看状态码会把"对方没收"记成发送成功。
func TestSendOKFalseWithin200(t *testing.T) {
	api := newFakeAPI(t)
	api.answer = `{"ok":false,"error_code":403,"description":"bot was blocked by the user"}`
	svc := api.start()

	_, err := svc.Send(context.Background(), &notify.Message{Target: "1", Content: "c"})
	require.Error(t, err)
	require.Contains(t, err.Error(), "ok=false")
	require.Contains(t, err.Error(), "bot was blocked by the user")
	require.NotErrorIs(t, err, notify.ErrNotConfigured, "对端拒收是投递失败，不是配置问题")
}

func TestSendPeerErrorStatus(t *testing.T) {
	api := newFakeAPI(t)
	api.status = http.StatusTooManyRequests
	api.answer = `{"ok":false,"error_code":429,"description":"Too Many Requests: retry after 5"}`
	svc := api.start()

	_, err := svc.Send(context.Background(), &notify.Message{Target: "1", Content: "c"})
	require.Error(t, err)
	require.Contains(t, err.Error(), "429")
}

func TestSendConfigIssues(t *testing.T) {
	svc := New(nil, Config{Ref: "7"})
	_, err := svc.Send(context.Background(), &notify.Message{Target: "1", Content: "c"})
	require.Error(t, err)
	require.ErrorIs(t, err, notify.ErrNotConfigured, "没配 bot_token 在拨号之前就该拦下")

	svc2 := New(nil, Config{BotToken: "t"})
	_, err = svc2.Send(context.Background(), &notify.Message{Content: "c"})
	require.Error(t, err)
	require.False(t, errors.Is(err, notify.ErrNotConfigured), "目标为空是调用方的问题")
}
