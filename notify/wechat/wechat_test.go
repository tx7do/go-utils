package wechat

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/tx7do/go-utils/notify"
)

// fakeAPI 本地假公众号 API：token 与 template/send 两端点分开计数，
// send 路径用 sendFunc 可编程——token 强刷重试这类"按请求次数给不同答案"的场景全靠它。
type fakeAPI struct {
	t *testing.T

	hitToken int
	hitSend  int
	lastBody map[string]any

	srvURL string
	client *http.Client

	tokenAnswer atomic.Value // string
	sendFunc    func(attempt int, accessToken string, body map[string]any) (int, string)
}

func newFakeAPI(t *testing.T, initialToken string) *fakeAPI {
	f := &fakeAPI{t: t}
	f.tokenAnswer.Store(`{"access_token":"` + initialToken + `","expires_in":7200}`)

	mux := http.NewServeMux()
	mux.HandleFunc("/cgi-bin/token", func(w http.ResponseWriter, r *http.Request) {
		f.hitToken++
		require.Equal(t, "client_credential", r.URL.Query().Get("grant_type"))
		require.Equal(t, "appid-1", r.URL.Query().Get("appid"))
		require.Equal(t, "secret-1", r.URL.Query().Get("secret"))
		_, _ = w.Write([]byte(f.tokenAnswer.Load().(string)))
	})
	mux.HandleFunc("/cgi-bin/message/template/send", func(w http.ResponseWriter, r *http.Request) {
		f.hitSend++
		var body map[string]any
		require.NoError(t, json.NewDecoder(r.Body).Decode(&body))
		f.lastBody = body
		status, answer := http.StatusOK, `{"errcode":0,"errmsg":"ok","msgid":9527}`
		if f.sendFunc != nil {
			status, answer = f.sendFunc(f.hitSend, r.URL.Query().Get("access_token"), body)
		}
		w.WriteHeader(status)
		_, _ = w.Write([]byte(answer))
	})

	srv := httptest.NewServer(mux)
	t.Cleanup(srv.Close)
	f.srvURL = srv.URL
	f.client = srv.Client()
	return f
}

func (f *fakeAPI) start() *Service {
	return New(f.client, Config{Ref: "7", AppID: "appid-1", AppSecret: "secret-1",
		TemplateID: "TMPL-1", APIBase: f.srvURL})
}

func TestSendDeliversTemplateMessage(t *testing.T) {
	api := newFakeAPI(t, "token-1")
	svc := api.start()

	receipt, err := svc.Send(context.Background(), &notify.Message{
		Target: "oX-openid-1",
		Params: map[string]string{"code": "123456", "minute": "10"},
	})
	require.NoError(t, err)
	require.Equal(t, "9527", receipt.ProviderID, "微信的 msgid 要进回执，对账全靠它")

	require.Equal(t, 1, api.hitToken)
	require.Equal(t, 1, api.hitSend)

	data := api.lastBody["data"].(map[string]any)
	require.Equal(t, "oX-openid-1", api.lastBody["touser"])
	require.Equal(t, "TMPL-1", api.lastBody["template_id"])
	require.Equal(t, map[string]any{"value": "123456"}, data["code"], "每个参数键渲染成 {value} 形状")
	require.Equal(t, map[string]any{"value": "10"}, data["minute"])
}

// TestSendContentFallsBackToContentKey 没带参数时整段正文落 "content" 键：
// 对应模板里唯一一个 {{content.data}} 占位的简单用法。
func TestSendContentFallsBackToContentKey(t *testing.T) {
	api := newFakeAPI(t, "token-1")
	svc := api.start()

	_, err := svc.Send(context.Background(), &notify.Message{Target: "oX-1", Content: "整段正文"})
	require.NoError(t, err)

	data := api.lastBody["data"].(map[string]any)
	require.Equal(t, map[string]any{"value": "整段正文"}, data["content"])
	require.Len(t, data, 1)
}

// TestSendTokenCachedAcrossSends access_token 有每日获取配额：第二条消息必须复用缓存，
// 不许每条都去要一枚新 token。
func TestSendTokenCachedAcrossSends(t *testing.T) {
	api := newFakeAPI(t, "token-1")
	svc := api.start()
	ctx := context.Background()

	msg := &notify.Message{Target: "oX-1", Params: map[string]string{"code": "1"}}
	_, err := svc.Send(ctx, msg)
	require.NoError(t, err)
	_, err = svc.Send(ctx, msg)
	require.NoError(t, err)

	require.Equal(t, 1, api.hitToken, "两发只许取一次 token")
	require.Equal(t, 2, api.hitSend)
}

// TestSendInvalidTokenRefreshesAndRetries token 在缓存窗口内被微信侧作废（重置密钥等）：
// 缓存的旧 token 第一次还能用，被作废后回 40001，强刷一枚新 token 重试成功——对调用方无感。
func TestSendInvalidTokenRefreshesAndRetries(t *testing.T) {
	api := newFakeAPI(t, "token-1")
	var firstSendDone atomic.Bool
	api.sendFunc = func(_ int, accessToken string, _ map[string]any) (int, string) {
		// token-1 只在第一次发送时放行（彼时它还新鲜有效）；之后对端已作废它 → 40001。
		if accessToken == "token-1" && !firstSendDone.CompareAndSwap(false, true) {
			return http.StatusOK, `{"errcode":40001,"errmsg":"invalid credential"}`
		}
		return http.StatusOK, `{"errcode":0,"errmsg":"ok","msgid":777}`
	}
	svc := api.start()
	ctx := context.Background()

	msg := &notify.Message{Target: "oX-1", Params: map[string]string{"code": "1"}}
	_, err := svc.Send(ctx, msg) // 第一发：token-1 新鲜有效，成功并缓存
	require.NoError(t, err)
	require.Equal(t, 1, api.hitToken)

	api.tokenAnswer.Store(`{"access_token":"token-2","expires_in":7200}`)
	receipt, err := svc.Send(ctx, msg) // 第二发：旧 token 撞 40001 → 强刷 → 重试成功
	require.NoError(t, err)
	require.Equal(t, "777", receipt.ProviderID)
	require.Equal(t, 2, api.hitToken, "40001 之后必须强刷一枚新 token")
	require.Equal(t, 3, api.hitSend, "一次失败 + 一次重试")
}

// TestSendNonTokenErrorIsNotRetried 业务错误（如频率限制）不走 token 强刷路径：
// 错了就交出去，瞎重试只会把限制撞得更死。
func TestSendNonTokenErrorIsNotRetried(t *testing.T) {
	api := newFakeAPI(t, "token-1")
	api.sendFunc = func(attempt int, _ string, _ map[string]any) (int, string) {
		return http.StatusOK, `{"errcode":45009,"errmsg":"reach max api daily quota limit"}`
	}
	svc := api.start()

	_, err := svc.Send(context.Background(), &notify.Message{Target: "oX-1", Params: map[string]string{"code": "1"}})
	require.Error(t, err)
	require.Contains(t, err.Error(), "45009")
	require.Equal(t, 1, api.hitToken, "非 token 错误不许触发强刷")
	require.Equal(t, 1, api.hitSend, "非 token 错误不许重试")
}

func TestSendConfigIssues(t *testing.T) {
	newSvc := func(cfg Config) *Service { return New(nil, cfg) }

	_, err := newSvc(Config{Ref: "7"}).Send(context.Background(),
		&notify.Message{Target: "oX-1", Params: map[string]string{"code": "1"}})
	require.ErrorIs(t, err, notify.ErrNotConfigured, "没配 app_id/app_secret 在拨号之前就该拦下")

	_, err = newSvc(Config{AppID: "a", AppSecret: "s"}).Send(context.Background(),
		&notify.Message{Target: "oX-1", Params: map[string]string{"code": "1"}})
	require.ErrorIs(t, err, notify.ErrNotConfigured)
	require.Contains(t, err.Error(), "template_id", "模板消息没有 template_id 就是发不出去，错要说清楚")

	_, err = newSvc(Config{AppID: "a", AppSecret: "s", TemplateID: "t"}).Send(context.Background(),
		&notify.Message{Params: map[string]string{"code": "1"}})
	require.Error(t, err)
	require.False(t, errors.Is(err, notify.ErrNotConfigured), "目标为空是调用方的问题")
}

func TestSendTokenEndpointError(t *testing.T) {
	api := newFakeAPI(t, "token-1")
	api.tokenAnswer.Store(`{"errcode":40125,"errmsg":"invalid appsecret"}`)
	svc := api.start()

	_, err := svc.Send(context.Background(), &notify.Message{Target: "oX-1", Params: map[string]string{"code": "1"}})
	require.Error(t, err)
	require.Contains(t, err.Error(), "40125")
	require.Equal(t, 1, api.hitToken)
	require.Equal(t, 0, api.hitSend, "token 拿不到就不该去碰发送端点")
}
