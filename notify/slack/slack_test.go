package slack

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/tx7do/go-utils/notify"
)

func TestSendDeliversMrkdwnText(t *testing.T) {
	var payload map[string]any
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		require.NoError(t, json.NewDecoder(r.Body).Decode(&payload))
		_, _ = w.Write([]byte("ok"))
	}))
	t.Cleanup(srv.Close)

	svc := New(srv.Client(), Config{Ref: "7", WebhookURL: srv.URL})
	_, err := svc.Send(context.Background(), &notify.Message{
		Title:   "重置验证码",
		Content: "您的验证码是 123456",
	})
	require.NoError(t, err)
	require.Equal(t, "*重置验证码*\n您的验证码是 123456", payload["text"], "Title 以 mrkdwn 粗体行起头")
}

func TestSendPlainWhenNoTitle(t *testing.T) {
	var payload map[string]any
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		require.NoError(t, json.NewDecoder(r.Body).Decode(&payload))
		_, _ = w.Write([]byte("ok"))
	}))
	t.Cleanup(srv.Close)

	svc := New(srv.Client(), Config{WebhookURL: srv.URL})
	_, err := svc.Send(context.Background(), &notify.Message{Content: "只有正文"})
	require.NoError(t, err)
	require.Equal(t, "只有正文", payload["text"])
}

func TestSendPeerError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte("invalid_payload"))
	}))
	t.Cleanup(srv.Close)

	svc := New(srv.Client(), Config{WebhookURL: srv.URL})
	_, err := svc.Send(context.Background(), &notify.Message{Content: "c"})
	require.Error(t, err)
	require.Contains(t, err.Error(), "invalid_payload")
}

func TestSendConfigIssue(t *testing.T) {
	svc := New(nil, Config{})
	_, err := svc.Send(context.Background(), &notify.Message{Content: "c"})
	require.ErrorIs(t, err, notify.ErrNotConfigured)
}
