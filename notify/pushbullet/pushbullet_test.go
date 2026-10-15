package pushbullet

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/tx7do/go-utils/notify"
)

func TestSendDeliversNote(t *testing.T) {
	var (
		auth    string
		payload map[string]any
	)
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, "/v2/pushes", r.URL.Path)
		auth = r.Header.Get("Access-Token")
		require.NoError(t, json.NewDecoder(r.Body).Decode(&payload))
		_, _ = w.Write([]byte(`{"type":"note","iden":"ujzX123","title":"t","body":"c"}`))
	}))
	t.Cleanup(srv.Close)

	svc := New(srv.Client(), Config{Ref: "7", APIKey: "o.token-1", APIBase: srv.URL})
	receipt, err := svc.Send(context.Background(), &notify.Message{
		Target:  "device-iden-9",
		Title:   "下载完成",
		Content: "报告已生成",
	})
	require.NoError(t, err)
	require.Equal(t, "ujzX123", receipt.ProviderID, "Pushbullet 的 iden 要进回执")
	require.Equal(t, "o.token-1", auth, "Pushbullet 的鉴权头是 Access-Token 不是 Bearer")
	require.Equal(t, "note", payload["type"])
	require.Equal(t, "device-iden-9", payload["device_iden"], "Target 作为目标设备")
}

func TestSendWithoutTargetPushesAllDevices(t *testing.T) {
	var payload map[string]any
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		require.NoError(t, json.NewDecoder(r.Body).Decode(&payload))
		_, _ = w.Write([]byte(`{"iden":"x"}`))
	}))
	t.Cleanup(srv.Close)

	svc := New(srv.Client(), Config{APIKey: "k", APIBase: srv.URL})
	_, err := svc.Send(context.Background(), &notify.Message{Title: "t", Content: "c"})
	require.NoError(t, err)
	_, hasDevice := payload["device_iden"]
	require.False(t, hasDevice, "无 Target 推全部设备")
}

func TestSendConfigIssue(t *testing.T) {
	svc := New(nil, Config{})
	_, err := svc.Send(context.Background(), &notify.Message{Content: "c"})
	require.ErrorIs(t, err, notify.ErrNotConfigured)
}
