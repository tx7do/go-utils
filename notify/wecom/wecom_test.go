package wecom

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

type wecomFixture struct {
	t *testing.T

	srvURL string
	client *http.Client

	hitToken int
	hitSend  int
	sendBody map[string]any

	tokenAnswer atomic.Value
	sendFunc    func(accessToken string, body map[string]any) (int, string)
}

func newWecomFixture(t *testing.T, initialToken string) *wecomFixture {
	f := &wecomFixture{t: t}
	f.tokenAnswer.Store(`{"errcode":0,"errmsg":"ok","access_token":"` + initialToken + `","expires_in":7200}`)

	mux := http.NewServeMux()
	mux.HandleFunc("/cgi-bin/gettoken", func(w http.ResponseWriter, r *http.Request) {
		f.hitToken++
		require.Equal(t, "corp-1", r.URL.Query().Get("corpid"))
		require.Equal(t, "secret-1", r.URL.Query().Get("corpsecret"))
		_, _ = w.Write([]byte(f.tokenAnswer.Load().(string)))
	})
	mux.HandleFunc("/cgi-bin/message/send", func(w http.ResponseWriter, r *http.Request) {
		f.hitSend++
		var body map[string]any
		require.NoError(t, json.NewDecoder(r.Body).Decode(&body))
		f.sendBody = body
		status, answer := http.StatusOK, `{"errcode":0,"errmsg":"ok","invaliduser":""}`
		if f.sendFunc != nil {
			status, answer = f.sendFunc(r.URL.Query().Get("access_token"), body)
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

func (f *wecomFixture) service() *Service {
	return New(f.client, Config{Ref: "7", CorpID: "corp-1", CorpSecret: "secret-1", AgentID: 1000002, APIBase: f.srvURL})
}

func TestSendDeliversTextToAgent(t *testing.T) {
	f := newWecomFixture(t, "token-1")
	svc := f.service()

	_, err := svc.Send(context.Background(), &notify.Message{
		Target:  "zhangsan",
		Title:   "工单提醒",
		Content: "您的工单已受理",
	})
	require.NoError(t, err)

	require.Equal(t, 1, f.hitToken)
	require.Equal(t, "zhangsan", f.sendBody["touser"])
	require.Equal(t, "text", f.sendBody["msgtype"])
	require.Equal(t, float64(1000002), f.sendBody["agentid"], "agentid 要按数字出站（企微要求）")
	text := f.sendBody["text"].(map[string]any)
	require.Equal(t, "工单提醒\n您的工单已受理", text["content"], "text 无独立标题位，标题空行相连")
}

func TestSendTokenCachedAcrossSends(t *testing.T) {
	f := newWecomFixture(t, "token-1")
	svc := f.service()
	ctx := context.Background()

	_, err := svc.Send(ctx, &notify.Message{Target: "@all", Content: "全员公告"})
	require.NoError(t, err)
	_, err = svc.Send(ctx, &notify.Message{Target: "@all", Content: "全员公告"})
	require.NoError(t, err)

	require.Equal(t, 1, f.hitToken, "两发只许换一次 access_token")
	require.Equal(t, 2, f.hitSend)
}

// TestSendInvalidTokenRefreshesAndRetries access_token 在缓存窗口内被企微侧作废：
// 用旧 token 的第一次发送回 40014，强刷一枚新 token 重试成功——对调用方无感。
func TestSendInvalidTokenRefreshesAndRetries(t *testing.T) {
	f := newWecomFixture(t, "token-1")
	var firstSendDone atomic.Bool
	f.sendFunc = func(accessToken string, _ map[string]any) (int, string) {
		// token-1 只在第一次发送时放行（彼时它还新鲜有效）；之后企微已作废它 → 40014。
		if accessToken == "token-1" && !firstSendDone.CompareAndSwap(false, true) {
			return http.StatusOK, `{"errcode":40014,"errmsg":"invalid access_token"}`
		}
		return http.StatusOK, `{"errcode":0,"errmsg":"ok"}`
	}
	svc := f.service()
	ctx := context.Background()

	_, err := svc.Send(ctx, &notify.Message{Target: "a", Content: "c"}) // 第一发：token-1 有效
	require.NoError(t, err)
	require.Equal(t, 1, f.hitToken)

	f.tokenAnswer.Store(`{"errcode":0,"access_token":"token-2","expires_in":7200}`)
	_, err = svc.Send(ctx, &notify.Message{Target: "a", Content: "c"}) // 第二发：40014 → 强刷 → 重试
	require.NoError(t, err)
	require.Equal(t, 2, f.hitToken, "40014 之后必须强刷一枚新 token")
	require.Equal(t, 3, f.hitSend, "一次失败 + 一次重试")
}

func TestSendConfigIssues(t *testing.T) {
	svc := New(nil, Config{Ref: "7", CorpID: "c", CorpSecret: "s"})
	_, err := svc.Send(context.Background(), &notify.Message{Target: "a", Content: "c"})
	require.ErrorIs(t, err, notify.ErrNotConfigured, "缺 agent_id 在拨号之前就该拦下")

	_, err = svc.Send(context.Background(), &notify.Message{Content: "c"})
	require.Error(t, err)
	require.False(t, errors.Is(err, notify.ErrNotConfigured), "touser 为空是调用方的问题")
}
