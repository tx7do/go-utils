package bark

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/tx7do/go-utils/notify"
)

func TestSendDeliversPayloadShape(t *testing.T) {
	var payload map[string]any
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, "/push", r.URL.Path)
		require.NoError(t, json.NewDecoder(r.Body).Decode(&payload))
		_, _ = w.Write([]byte(`{"code":200,"message":"success","timestamp":1760000000}`))
	}))
	t.Cleanup(srv.Close)

	svc := New(srv.Client(), Config{Ref: "7", ServerURL: srv.URL})
	_, err := svc.Send(context.Background(), &notify.Message{
		Target:  "device-key-1",
		Title:   "重置验证码",
		Content: "您的验证码是 123456",
		Params:  map[string]string{"group": "auth", "sound": "minuet"},
	})
	require.NoError(t, err)
	require.Equal(t, "device-key-1", payload["device_key"])
	require.Equal(t, "重置验证码", payload["title"])
	require.Equal(t, "您的验证码是 123456", payload["body"])
	require.Equal(t, "auth", payload["group"], "Params 原生字段直传 JSON")
}

func TestSendNon200BusinessCode(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = io.WriteString(w, `{"code":400,"message":"key not found"}`)
	}))
	t.Cleanup(srv.Close)

	svc := New(srv.Client(), Config{ServerURL: srv.URL})
	_, err := svc.Send(context.Background(), &notify.Message{Target: "k", Title: "t", Content: "c"})
	require.Error(t, err)
	require.Contains(t, err.Error(), "code=400")
}

func TestSendEmptyTarget(t *testing.T) {
	svc := New(nil, Config{})
	_, err := svc.Send(context.Background(), &notify.Message{Title: "t", Content: "c"})
	require.Error(t, err)
}
