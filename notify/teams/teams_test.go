package teams

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/tx7do/go-utils/notify"
)

func TestSendDeliversAdaptiveCard(t *testing.T) {
	var payload map[string]any
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		require.NoError(t, json.NewDecoder(r.Body).Decode(&payload))
		_, _ = w.Write([]byte("1"))
	}))
	t.Cleanup(srv.Close)

	svc := New(srv.Client(), Config{Ref: "7", WebhookURL: srv.URL})
	_, err := svc.Send(context.Background(), &notify.Message{
		Title:   "重置验证码",
		Content: "您的验证码是 123456",
	})
	require.NoError(t, err)

	require.Equal(t, "message", payload["type"])
	attachments := payload["attachments"].([]any)
	att := attachments[0].(map[string]any)
	require.Equal(t, "application/vnd.microsoft.card.adaptive", att["contentType"])

	content := att["content"].(map[string]any)
	require.Equal(t, "AdaptiveCard", content["type"])
	blocks := content["body"].([]any)
	require.Len(t, blocks, 2, "Title 加粗大字块 + Content 正文块")
	title := blocks[0].(map[string]any)
	body := blocks[1].(map[string]any)
	require.Equal(t, "TextBlock", title["type"])
	require.Equal(t, "Bolder", title["weight"])
	require.Equal(t, "重置验证码", title["text"])
	require.Equal(t, "您的验证码是 123456", body["text"])
}

func TestSendConfigIssue(t *testing.T) {
	svc := New(nil, Config{})
	_, err := svc.Send(context.Background(), &notify.Message{Content: "c"})
	require.ErrorIs(t, err, notify.ErrNotConfigured)
}
