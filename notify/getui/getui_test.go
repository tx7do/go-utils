package getui

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/tx7do/go-utils/notify"
)

func TestSendDeliversNotificationViaCID(t *testing.T) {
	var (
		hitAuth  int
		tokenHdr string
		path     string
		payload  map[string]any
	)
	mux := http.NewServeMux()
	mux.HandleFunc("/v2/app-1/auth", func(w http.ResponseWriter, r *http.Request) {
		hitAuth++
		var req map[string]string
		require.NoError(t, json.NewDecoder(r.Body).Decode(&req))
		require.Equal(t, "key-1", req["appkey"])
		require.NotEmpty(t, req["timestamp"])
		require.Len(t, req["sign"], 64, "sign 是 sha256 hex")
		_, _ = w.Write([]byte(`{"code":0,"msg":"success","data":{"token":"gt-token","expire_time":1760000000000}}`))
	})
	mux.HandleFunc("/v2/app-1/push/single/cid", func(w http.ResponseWriter, r *http.Request) {
		tokenHdr = r.Header.Get("token")
		path = r.URL.Path
		require.NoError(t, json.NewDecoder(r.Body).Decode(&payload))
		_, _ = w.Write([]byte(`{"code":0,"msg":"success"}`))
	})
	srv := httptest.NewServer(mux)
	t.Cleanup(srv.Close)

	svc := New(srv.Client(), Config{Ref: "7", AppID: "app-1", AppKey: "key-1", MasterSecret: "ms-1", APIBase: srv.URL})
	_, err := svc.Send(context.Background(), &notify.Message{
		Target:  "cid-1",
		Title:   "重置验证码",
		Content: "您的验证码是 123456",
	})
	require.NoError(t, err)
	require.Equal(t, 1, hitAuth)
	require.Equal(t, "gt-token", tokenHdr, "鉴权 token 走自定义 token 头")
	require.True(t, path == "/v2/app-1/push/single/cid", path)

	pushMessage := payload["push_message"].(map[string]any)
	notification := pushMessage["notification"].(map[string]any)
	require.Equal(t, "重置验证码", notification["title"])
	require.Equal(t, "您的验证码是 123456", notification["body"])
	require.Equal(t, "notify", notification["push_type"])
	require.NotEmpty(t, payload["request_id"], "request_id 防重必备")
}

func TestSendPeerErrorCode(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(`{"code":10001,"msg":"token expired"}`))
	}))
	t.Cleanup(srv.Close)

	svc := New(srv.Client(), Config{AppID: "a", AppKey: "k", MasterSecret: "s", APIBase: srv.URL})
	_, err := svc.Send(context.Background(), &notify.Message{Target: "cid", Title: "t", Content: "c"})
	require.Error(t, err)
	require.Contains(t, err.Error(), "code=10001")
}

func TestSendConfigIssue(t *testing.T) {
	svc := New(nil, Config{Ref: "7", AppID: "a", AppKey: "k"})
	_, err := svc.Send(context.Background(), &notify.Message{Target: "cid", Title: "t", Content: "c"})
	require.ErrorIs(t, err, notify.ErrNotConfigured)
}
