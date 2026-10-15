package gchat

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/tx7do/go-utils/notify"
)

func TestSendDeliversTextAndParsesResourceName(t *testing.T) {
	var payload map[string]any
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		require.NoError(t, json.NewDecoder(r.Body).Decode(&payload))
		_, _ = w.Write([]byte(`{"sender":{},"text":"*t*\nc","thread":{},"name":"spaces/AAA/messages/BBB"}`))
	}))
	t.Cleanup(srv.Close)

	svc := New(srv.Client(), Config{Ref: "7", WebhookURL: srv.URL})
	receipt, err := svc.Send(context.Background(), &notify.Message{
		Title:   "重置验证码",
		Content: "您的验证码是 123456",
	})
	require.NoError(t, err)
	require.Equal(t, "spaces/AAA/messages/BBB", receipt.ProviderID, "消息资源名要进回执")
	require.Equal(t, "*重置验证码*\n您的验证码是 123456", payload["text"])
}

func TestSendConfigIssue(t *testing.T) {
	svc := New(nil, Config{})
	_, err := svc.Send(context.Background(), &notify.Message{Content: "c"})
	require.ErrorIs(t, err, notify.ErrNotConfigured)
}
