package sendgrid

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/tx7do/go-utils/notify"
)

func TestSendDeliversV3Shape(t *testing.T) {
	var (
		auth      string
		messageID = "sg-msg-1"
		payload   map[string]any
	)
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, "/v3/mail/send", r.URL.Path)
		auth = r.Header.Get("Authorization")
		require.NoError(t, json.NewDecoder(r.Body).Decode(&payload))
		w.Header().Set("X-Message-Id", messageID)
		w.WriteHeader(http.StatusAccepted)
	}))
	t.Cleanup(srv.Close)

	svc := New(srv.Client(), Config{Ref: "7", APIKey: "SG.key", From: "noreply@example.com", APIBase: srv.URL})
	receipt, err := svc.Send(context.Background(), &notify.Message{
		Target:  "alice@example.com",
		Title:   "重置验证码",
		Content: "您的验证码是 123456",
	})
	require.NoError(t, err)
	require.Equal(t, "sg-msg-1", receipt.ProviderID, "X-Message-Id 响应头要进回执")
	require.Equal(t, "Bearer SG.key", auth)

	personalizations := payload["personalizations"].([]any)
	to := personalizations[0].(map[string]any)["to"].([]any)[0].(map[string]any)
	require.Equal(t, "alice@example.com", to["email"])
	require.Equal(t, "noreply@example.com", payload["from"].(map[string]any)["email"])
	require.Equal(t, "重置验证码", payload["subject"])
	content := payload["content"].([]any)[0].(map[string]any)
	require.Equal(t, "text/plain", content["type"])
	require.Equal(t, "您的验证码是 123456", content["value"])
}

func TestSendConfigIssues(t *testing.T) {
	svc := New(nil, Config{})
	_, err := svc.Send(context.Background(), &notify.Message{Target: "a@b.c", Content: "c"})
	require.ErrorIs(t, err, notify.ErrNotConfigured)
}
