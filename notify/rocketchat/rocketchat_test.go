package rocketchat

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/tx7do/go-utils/notify"
)

func TestSendDeliversUserHeaders(t *testing.T) {
	var (
		authID  string
		authTok string
		payload map[string]any
	)
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, "/api/v1/chat.postMessage", r.URL.Path)
		authID = r.Header.Get("X-User-Id")
		authTok = r.Header.Get("X-Auth-Token")
		require.NoError(t, json.NewDecoder(r.Body).Decode(&payload))
		_, _ = w.Write([]byte(`{"success":true,"message":{"_id":"msg-9"}}`))
	}))
	t.Cleanup(srv.Close)

	svc := New(srv.Client(), Config{Ref: "7", ServerURL: srv.URL, UserID: "uid-1", AuthToken: "tok-1"})
	receipt, err := svc.Send(context.Background(), &notify.Message{
		Target:  "#ops",
		Title:   "部署完成",
		Content: "v1.2.0 已上线",
	})
	require.NoError(t, err)
	require.Equal(t, "msg-9", receipt.ProviderID, "消息 _id 要进回执")
	require.Equal(t, "uid-1", authID)
	require.Equal(t, "tok-1", authTok)
	require.Equal(t, "#ops", payload["channel"])
	require.Equal(t, "部署完成\nv1.2.0 已上线", payload["text"])
}

func TestSendSuccessFalse(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(`{"success":false,"error":"invalid-channel"}`))
	}))
	t.Cleanup(srv.Close)

	svc := New(srv.Client(), Config{ServerURL: srv.URL, UserID: "u", AuthToken: "t"})
	_, err := svc.Send(context.Background(), &notify.Message{Target: "#gone", Content: "c"})
	require.Error(t, err)
	require.Contains(t, err.Error(), "invalid-channel")
}

func TestSendConfigIssue(t *testing.T) {
	svc := New(nil, Config{ServerURL: "https://example.com"})
	_, err := svc.Send(context.Background(), &notify.Message{Target: "#ops", Content: "c"})
	require.ErrorIs(t, err, notify.ErrNotConfigured)
}
