package discord

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/tx7do/go-utils/notify"
)

func TestSendEmbedWithTitle(t *testing.T) {
	var (
		query   string
		payload map[string]any
	)
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		query = r.URL.RawQuery
		require.NoError(t, json.NewDecoder(r.Body).Decode(&payload))
		_, _ = w.Write([]byte(`{"id":"119900211"}`))
	}))
	t.Cleanup(srv.Close)

	svc := New(srv.Client(), Config{Ref: "7", WebhookURL: srv.URL})
	receipt, err := svc.Send(context.Background(), &notify.Message{
		Title:   "重置验证码",
		Content: "您的验证码是 123456",
	})
	require.NoError(t, err)
	require.Equal(t, "wait=true", query, "wait=true 才回消息对象，message id 全靠它")
	require.Equal(t, "119900211", receipt.ProviderID)

	embeds := payload["embeds"].([]any)
	embed := embeds[0].(map[string]any)
	require.Equal(t, "重置验证码", embed["title"])
	require.Equal(t, "您的验证码是 123456", embed["description"])
	_, hasContent := payload["content"]
	require.False(t, hasContent, "embed 形态下不该再发重复的 content")
}

func TestSendPlainWhenNoTitle(t *testing.T) {
	var payload map[string]any
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		require.NoError(t, json.NewDecoder(r.Body).Decode(&payload))
		_, _ = w.Write([]byte(`{"id":"1"}`))
	}))
	t.Cleanup(srv.Close)

	svc := New(srv.Client(), Config{WebhookURL: srv.URL})
	_, err := svc.Send(context.Background(), &notify.Message{Content: "只有正文"})
	require.NoError(t, err)
	require.Equal(t, "只有正文", payload["content"])
}

func TestSendRateLimited(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusTooManyRequests)
		_, _ = w.Write([]byte(`{"message":"You are being rate limited.","code": 200029}`))
	}))
	t.Cleanup(srv.Close)

	svc := New(srv.Client(), Config{WebhookURL: srv.URL})
	_, err := svc.Send(context.Background(), &notify.Message{Content: "c"})
	require.Error(t, err)
	require.Contains(t, err.Error(), "429")
}

func TestSendConfigIssue(t *testing.T) {
	svc := New(nil, Config{})
	_, err := svc.Send(context.Background(), &notify.Message{Content: "c"})
	require.ErrorIs(t, err, notify.ErrNotConfigured)
}
