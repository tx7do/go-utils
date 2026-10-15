package webex

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/tx7do/go-utils/notify"
)

func TestSendDeliversToRoom(t *testing.T) {
	var (
		auth    string
		payload map[string]any
	)
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, "/v1/messages", r.URL.Path)
		auth = r.Header.Get("Authorization")
		require.NoError(t, json.NewDecoder(r.Body).Decode(&payload))
		_, _ = w.Write([]byte(`{"id":"msg-1","roomId":"ROOM1","text":"x"}`))
	}))
	t.Cleanup(srv.Close)

	svc := New(srv.Client(), Config{Ref: "7", BotToken: "bt-1", APIBase: srv.URL})
	receipt, err := svc.Send(context.Background(), &notify.Message{
		Target:  "ROOM1",
		Title:   "部署完成",
		Content: "v1.2.0 已上线",
	})
	require.NoError(t, err)
	require.Equal(t, "msg-1", receipt.ProviderID)
	require.Equal(t, "Bearer bt-1", auth)
	require.Equal(t, "ROOM1", payload["roomId"])
	_, hasEmail := payload["toPersonEmail"]
	require.False(t, hasEmail, "无 @ 的 Target 按空间投递")
	require.Equal(t, "部署完成\nv1.2.0 已上线", payload["text"])
}

func TestSendToPersonEmail(t *testing.T) {
	var payload map[string]any
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		require.NoError(t, json.NewDecoder(r.Body).Decode(&payload))
		_, _ = w.Write([]byte(`{"id":"msg-2"}`))
	}))
	t.Cleanup(srv.Close)

	svc := New(srv.Client(), Config{BotToken: "bt", APIBase: srv.URL})
	_, err := svc.Send(context.Background(), &notify.Message{Target: "alice@example.com", Content: "c"})
	require.NoError(t, err)
	require.Equal(t, "alice@example.com", payload["toPersonEmail"], "含 @ 按 1v1 私聊投递")
	_, hasRoom := payload["roomId"]
	require.False(t, hasRoom)
}

func TestSendConfigIssue(t *testing.T) {
	svc := New(nil, Config{})
	_, err := svc.Send(context.Background(), &notify.Message{Target: "ROOM1", Content: "c"})
	require.ErrorIs(t, err, notify.ErrNotConfigured)
}
