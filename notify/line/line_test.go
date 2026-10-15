package line

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/tx7do/go-utils/notify"
)

func TestSendDeliversPushMessage(t *testing.T) {
	var (
		auth    string
		payload map[string]any
	)
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, "/v2/bot/message/push", r.URL.Path)
		auth = r.Header.Get("Authorization")
		require.NoError(t, json.NewDecoder(r.Body).Decode(&payload))
		_, _ = w.Write([]byte(`{}`))
	}))
	t.Cleanup(srv.Close)

	svc := New(srv.Client(), Config{Ref: "7", ChannelAccessToken: "cat-1", APIBase: srv.URL})
	_, err := svc.Send(context.Background(), &notify.Message{
		Target:  "U1234",
		Title:   "重置验证码",
		Content: "您的验证码是 123456",
	})
	require.NoError(t, err)
	require.Equal(t, "Bearer cat-1", auth)
	require.Equal(t, "U1234", payload["to"])
	messages := payload["messages"].([]any)
	require.Len(t, messages, 1)
	m := messages[0].(map[string]any)
	require.Equal(t, "text", m["type"])
	require.Equal(t, "重置验证码\n您的验证码是 123456", m["text"])
}

func TestSendConfigIssue(t *testing.T) {
	svc := New(nil, Config{})
	_, err := svc.Send(context.Background(), &notify.Message{Target: "U1", Content: "c"})
	require.ErrorIs(t, err, notify.ErrNotConfigured)
}
