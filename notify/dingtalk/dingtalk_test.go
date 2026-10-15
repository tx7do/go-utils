package dingtalk

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/tx7do/go-utils/notify"
)

func TestSendDeliversWorkNotice(t *testing.T) {
	var (
		hitToken int
		sendAuth string
		payload  map[string]any
	)
	mux := http.NewServeMux()
	mux.HandleFunc("/gettoken", func(w http.ResponseWriter, r *http.Request) {
		hitToken++
		require.Equal(t, "key-1", r.URL.Query().Get("appkey"))
		require.Equal(t, "secret-1", r.URL.Query().Get("appsecret"))
		_, _ = w.Write([]byte(`{"errcode":0,"errmsg":"ok","access_token":"tk-1","expires_in":7200}`))
	})
	mux.HandleFunc("/topapi/message/corpconversation/asyncsend_v2", func(w http.ResponseWriter, r *http.Request) {
		sendAuth = r.URL.Query().Get("access_token")
		require.NoError(t, json.NewDecoder(r.Body).Decode(&payload))
		_, _ = w.Write([]byte(`{"errcode":0,"errmsg":"ok","task_id":66779}`))
	})
	srv := httptest.NewServer(mux)
	t.Cleanup(srv.Close)

	svc := New(srv.Client(), Config{Ref: "7", AppKey: "key-1", AppSecret: "secret-1", AgentID: 123, APIBase: srv.URL})
	receipt, err := svc.Send(context.Background(), &notify.Message{
		Target:  "zhangsan|lisi",
		Title:   "审批提醒",
		Content: "您有一张单据待审批",
	})
	require.NoError(t, err)
	require.Equal(t, "66779", receipt.ProviderID, "异步受理的 task_id 要进回执")
	require.Equal(t, "tk-1", sendAuth)

	require.Equal(t, float64(123), payload["agent_id"], "agent_id 按数字出站")
	require.Equal(t, "zhangsan|lisi", payload["userid_list"])
	msg := payload["msg"].(map[string]any)
	require.Equal(t, "text", msg["msgtype"])
	require.Equal(t, map[string]any{"content": "审批提醒\n您有一张单据待审批"}, msg["text"])
}

func TestSendPeerErrorCode(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(`{"errcode":40007,"errmsg":"invalid agent_id"}`))
	}))
	t.Cleanup(srv.Close)

	svc := New(srv.Client(), Config{AppKey: "k", AppSecret: "s", AgentID: 1, APIBase: srv.URL})
	_, err := svc.Send(context.Background(), &notify.Message{Target: "a", Content: "c"})
	require.Error(t, err)
	require.Contains(t, err.Error(), "errcode=40007")
}

func TestSendConfigIssues(t *testing.T) {
	svc := New(nil, Config{Ref: "7", AppKey: "k", AppSecret: "s"})
	_, err := svc.Send(context.Background(), &notify.Message{Target: "a", Content: "c"})
	require.ErrorIs(t, err, notify.ErrNotConfigured, "缺 agent_id 在拨号之前就该拦下")
}
