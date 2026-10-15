package zulip

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/tx7do/go-utils/notify"
)

func TestSendDeliversStreamMessage(t *testing.T) {
	var (
		form   string
		method string
	)
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		method = r.Method
		require.NoError(t, r.ParseForm())
		form = r.PostForm.Encode()
		_, _ = w.Write([]byte(`{"result":"success","msg":"","id":3159981}`))
	}))
	t.Cleanup(srv.Close)

	svc := New(srv.Client(), Config{Ref: "7", ServerURL: srv.URL, BotEmail: "bot@example.com",
		BotAPIKey: "key-1", Topic: "告警"})
	receipt, err := svc.Send(context.Background(), &notify.Message{
		Target:  "ops",
		Title:   "磁盘告警",
		Content: "用量 92%",
	})
	require.NoError(t, err)
	require.Equal(t, "3159981", receipt.ProviderID, "Zulip 的消息 ID 要进回执")
	require.Equal(t, "POST", method)

	require.Contains(t, form, "type=stream")
	require.Contains(t, form, "to=ops")
	require.Contains(t, form, "topic=%E5%91%8A%E8%AD%A6", "stream 消息必须有 topic（Config 默认）")
	require.Contains(t, form, "content=%E7%A3%81%E7%9B%98%E5%91%8A%E8%AD%A6")
}

func TestSendTopicParamOverrides(t *testing.T) {
	var form string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		require.NoError(t, r.ParseForm())
		form = r.PostForm.Encode()
		_, _ = w.Write([]byte(`{"result":"success","id":2}`))
	}))
	t.Cleanup(srv.Close)

	svc := New(srv.Client(), Config{ServerURL: srv.URL, BotEmail: "b@e", BotAPIKey: "k", Topic: "default"})
	_, err := svc.Send(context.Background(), &notify.Message{
		Target: "ops", Content: "c", Params: map[string]string{"topic": "deploy"},
	})
	require.NoError(t, err)
	require.Contains(t, form, "topic=deploy", "Params[topic] 按条覆盖")
}

func TestSendRequiresTopic(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {}))
	t.Cleanup(srv.Close)

	svc := New(srv.Client(), Config{ServerURL: srv.URL, BotEmail: "b@e", BotAPIKey: "k"})
	_, err := svc.Send(context.Background(), &notify.Message{Target: "ops", Content: "c"})
	require.ErrorIs(t, err, notify.ErrNotConfigured, "stream 消息没 topic 发不出去，拨号之前拦下")
}

func TestSendConfigIssue(t *testing.T) {
	svc := New(nil, Config{Topic: "t"})
	_, err := svc.Send(context.Background(), &notify.Message{Target: "ops", Content: "c"})
	require.ErrorIs(t, err, notify.ErrNotConfigured)
}
