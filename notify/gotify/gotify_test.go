package gotify

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/tx7do/go-utils/notify"
)

func TestSendDeliversTokenAndShape(t *testing.T) {
	var (
		query   string
		payload map[string]any
	)
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		query = r.URL.RawQuery
		require.Equal(t, "/message", r.URL.Path)
		require.NoError(t, json.NewDecoder(r.Body).Decode(&payload))
		_, _ = w.Write([]byte(`{"id":42,"appid":1,"message":"m","title":"t","priority":4}`))
	}))
	t.Cleanup(srv.Close)

	svc := New(srv.Client(), Config{Ref: "7", ServerURL: srv.URL, AppToken: "Atoken"})
	receipt, err := svc.Send(context.Background(), &notify.Message{
		Title:   "重置验证码",
		Content: "您的验证码是 123456",
		Params:  map[string]string{"priority": "4"},
	})
	require.NoError(t, err)
	require.Equal(t, "token=Atoken", query, "application token 走 query（Gotify 的惯例）")
	require.Equal(t, "42", receipt.ProviderID, "Gotify 的消息 ID 要进回执")
	require.Equal(t, float64(4), payload["priority"])
	require.Equal(t, "重置验证码", payload["title"])
}

func TestSendConfigIssues(t *testing.T) {
	svc := New(nil, Config{ServerURL: "https://example.com"})
	_, err := svc.Send(context.Background(), &notify.Message{Content: "c"})
	require.ErrorIs(t, err, notify.ErrNotConfigured, "缺 app_token 在拨号之前就该拦下")
}
