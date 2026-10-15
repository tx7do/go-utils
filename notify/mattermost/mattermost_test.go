package mattermost

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/tx7do/go-utils/notify"
)

func TestSendDeliversMarkdownText(t *testing.T) {
	var payload map[string]any
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, "/hooks/key-1", r.URL.Path)
		require.NoError(t, json.NewDecoder(r.Body).Decode(&payload))
		_, _ = w.Write([]byte("ok"))
	}))
	t.Cleanup(srv.Close)

	svc := New(srv.Client(), Config{Ref: "7", WebhookURL: srv.URL + "/hooks/key-1", Channel: "#ops"})
	_, err := svc.Send(context.Background(), &notify.Message{
		Title:   "磁盘告警",
		Content: "用量 92%",
	})
	require.NoError(t, err)
	require.Equal(t, "**磁盘告警**\n用量 92%", payload["text"], "Mattermost 的粗体是 **text**")
	require.Equal(t, "#ops", payload["channel"], "Channel 覆盖 webhook 默认频道")
}

func TestSendConfigIssue(t *testing.T) {
	svc := New(nil, Config{})
	_, err := svc.Send(context.Background(), &notify.Message{Content: "c"})
	require.ErrorIs(t, err, notify.ErrNotConfigured)
}
