package ntfy

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

func TestSendJSONPublishShape(t *testing.T) {
	var (
		payload map[string]any
		auth    string
	)
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, "/", r.URL.Path, "JSON publish 走根路径：主题在 body 里，绕开头注入")
		auth = r.Header.Get("Authorization")
		require.NoError(t, json.NewDecoder(r.Body).Decode(&payload))
		_, _ = w.Write([]byte(`{"id":"sP8vG0","time":1760000000,"topic":"alerts"}`))
	}))
	t.Cleanup(srv.Close)

	svc := New(srv.Client(), Config{Ref: "7", ServerURL: srv.URL, User: "u", Password: "p", Tags: []string{"warning"}})
	receipt, err := svc.Send(context.Background(), &notify.Message{
		Target:  "alerts",
		Title:   "磁盘告警",
		Content: "用量 92%",
		Params:  map[string]string{"priority": "4"},
	})
	require.NoError(t, err)
	require.Equal(t, "sP8vG0", receipt.ProviderID, "ntfy 的消息 ID 要进回执")

	require.Equal(t, "Basic dTpw", auth, "基础认证 user:pass")
	require.Equal(t, "alerts", payload["topic"])
	require.Equal(t, "磁盘告警", payload["title"])
	require.Equal(t, float64(4), payload["priority"], "Params[priority] 按条覆盖默认档位")
	require.Equal(t, []any{"warning"}, payload["tags"])
}

func TestSendTokenAuthOverridesBasic(t *testing.T) {
	var auth string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		auth = r.Header.Get("Authorization")
		_, _ = w.Write([]byte(`{"id":"x"}`))
	}))
	t.Cleanup(srv.Close)

	svc := New(srv.Client(), Config{ServerURL: srv.URL, User: "u", Password: "p", AccessToken: "tk_x"})
	_, err := svc.Send(context.Background(), &notify.Message{Target: "t", Content: "c"})
	require.NoError(t, err)
	require.Equal(t, "Bearer tk_x", auth, "令牌认证优先于基础认证")
}

func TestSendBadPriorityParam(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {}))
	t.Cleanup(srv.Close)

	svc := New(srv.Client(), Config{ServerURL: srv.URL})
	_, err := svc.Send(context.Background(), &notify.Message{
		Target: "t", Content: "c", Params: map[string]string{"priority": "high"},
	})
	require.Error(t, err, "优先级要数字档位，垃圾值在拨号之前拦下")
	require.False(t, errors.Is(err, notify.ErrNotConfigured))
}
