package datadog

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/tx7do/go-utils/notify"
)

func TestSendPublishesEvent(t *testing.T) {
	var (
		ddKey   string
		payload map[string]any
	)
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, "/api/v1/events", r.URL.Path)
		ddKey = r.Header.Get("DD-API-KEY")
		require.NoError(t, json.NewDecoder(r.Body).Decode(&payload))
		_, _ = w.Write([]byte(`{"status":"ok","event":{"id":1728312,"title":"磁盘告警","text":"用量 92%","tags":[]}}`))
	}))
	t.Cleanup(srv.Close)

	svc := New(srv.Client(), Config{Ref: "7", APIKey: "dd-key", Priority: "low",
		Tags: []string{"env:prod", "team:infra"}, APIBase: srv.URL})
	receipt, err := svc.Send(context.Background(), &notify.Message{
		Target:  "worker-7",
		Title:   "磁盘告警",
		Content: "用量 92%",
	})
	require.NoError(t, err)
	require.Equal(t, "1728312", receipt.ProviderID, "事件 ID 要进回执")
	require.Equal(t, "dd-key", ddKey, "Datadog 的鉴权头是 DD-API-KEY")

	require.Equal(t, "磁盘告警", payload["title"])
	require.Equal(t, "用量 92%", payload["text"])
	require.Equal(t, "low", payload["priority"])
	require.Equal(t, []any{"env:prod", "team:infra"}, payload["tags"])
	require.Equal(t, "worker-7", payload["host"], "Target 落 host 字段")
}

func TestSendErrorAnswer(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusForbidden)
		_, _ = w.Write([]byte(`{"status":"error","errors":["API key invalid"]}`))
	}))
	t.Cleanup(srv.Close)

	svc := New(srv.Client(), Config{APIKey: "bad", APIBase: srv.URL})
	_, err := svc.Send(context.Background(), &notify.Message{Content: "c"})
	require.Error(t, err)
	require.Contains(t, err.Error(), "API key invalid")
}

func TestSendConfigIssue(t *testing.T) {
	svc := New(nil, Config{})
	_, err := svc.Send(context.Background(), &notify.Message{Content: "c"})
	require.ErrorIs(t, err, notify.ErrNotConfigured)
}
