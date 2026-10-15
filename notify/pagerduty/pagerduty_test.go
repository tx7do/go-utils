package pagerduty

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/tx7do/go-utils/notify"
)

func TestSendTriggersEvent(t *testing.T) {
	var payload map[string]any
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, "/v2/enqueue", r.URL.Path)
		require.NoError(t, json.NewDecoder(r.Body).Decode(&payload))
		w.WriteHeader(http.StatusAccepted)
		_, _ = w.Write([]byte(`{"status":"success","message":"Event processed","dedup_key":"disk-alert-1"}`))
	}))
	t.Cleanup(srv.Close)

	svc := New(srv.Client(), Config{Ref: "7", RoutingKey: "rk-1", Severity: "error", Source: "go-wind", APIBase: srv.URL})
	receipt, err := svc.Send(context.Background(), &notify.Message{
		Target:  "disk-alert-1",
		Title:   "磁盘告警",
		Content: "root 分区用量 92%",
	})
	require.NoError(t, err)
	require.Equal(t, "disk-alert-1", receipt.ProviderID, "确认后的 dedup key 要进回执")

	require.Equal(t, "rk-1", payload["routing_key"])
	require.Equal(t, "trigger", payload["event_action"])
	require.Equal(t, "disk-alert-1", payload["dedup_key"], "Target 作为去重键：同 key 关联同一告警")

	pl := payload["payload"].(map[string]any)
	require.Equal(t, "root 分区用量 92%", pl["summary"])
	require.Equal(t, "error", pl["severity"])
	require.Equal(t, "go-wind", pl["source"])
	details := pl["custom_details"].(map[string]any)
	require.Equal(t, "磁盘告警", details["title"], "Title 落 custom_details")
}

func TestSendFailure(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte(`{"status":"invalid_event","message":"Summary is required"}`))
	}))
	t.Cleanup(srv.Close)

	svc := New(srv.Client(), Config{RoutingKey: "rk", APIBase: srv.URL})
	_, err := svc.Send(context.Background(), &notify.Message{Content: "c"})
	require.Error(t, err)
	require.Contains(t, err.Error(), "invalid_event")
}

func TestSendConfigIssue(t *testing.T) {
	svc := New(nil, Config{})
	_, err := svc.Send(context.Background(), &notify.Message{Content: "c"})
	require.ErrorIs(t, err, notify.ErrNotConfigured)
}
