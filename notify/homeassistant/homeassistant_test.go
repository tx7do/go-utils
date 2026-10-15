package homeassistant

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/tx7do/go-utils/notify"
)

func TestSendCallsNotifyService(t *testing.T) {
	var (
		auth    string
		path    string
		payload map[string]any
	)
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		path = r.URL.Path
		auth = r.Header.Get("Authorization")
		require.NoError(t, json.NewDecoder(r.Body).Decode(&payload))
		_, _ = w.Write([]byte(`[{"message_id":1}]`))
	}))
	t.Cleanup(srv.Close)

	svc := New(srv.Client(), Config{Ref: "7", ServerURL: srv.URL, Token: "llat-1",
		Service: "mobile_app_phone"})
	_, err := svc.Send(context.Background(), &notify.Message{
		Title:   "门铃",
		Content: "有人在门口",
		Target:  "entity.kitchen_display",
		Params:  map[string]string{"channel": "critical"},
	})
	require.NoError(t, err)
	require.Equal(t, "/api/services/notify/mobile_app_phone", path, "服务名进路径")
	require.Equal(t, "Bearer llat-1", auth)

	require.Equal(t, "有人在门口", payload["message"])
	require.Equal(t, "门铃", payload["title"])
	data := payload["data"].(map[string]any)
	require.Equal(t, "entity.kitchen_display", data["target"], "Target 落 data.target 由平台解释")
	require.Equal(t, "critical", data["channel"], "Params 原样进 data")
}

func TestSendDefaultService(t *testing.T) {
	var path string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		path = r.URL.Path
		_, _ = w.Write([]byte(`[]`))
	}))
	t.Cleanup(srv.Close)

	svc := New(srv.Client(), Config{ServerURL: srv.URL, Token: "t"})
	_, err := svc.Send(context.Background(), &notify.Message{Content: "c"})
	require.NoError(t, err)
	require.Equal(t, "/api/services/notify/persistent_notification", path, "缺省走持久通知服务")
}

func TestSendConfigIssue(t *testing.T) {
	svc := New(nil, Config{ServerURL: "http://ha:8123"})
	_, err := svc.Send(context.Background(), &notify.Message{Content: "c"})
	require.ErrorIs(t, err, notify.ErrNotConfigured)
}
