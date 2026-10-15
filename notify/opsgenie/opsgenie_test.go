package opsgenie

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/tx7do/go-utils/notify"
)

func TestSendCreatesAlert(t *testing.T) {
	var (
		auth    string
		payload map[string]any
	)
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, "/v2/alerts", r.URL.Path)
		auth = r.Header.Get("Authorization")
		require.NoError(t, json.NewDecoder(r.Body).Decode(&payload))
		_, _ = w.Write([]byte(`{"result":"Request will be processed","took":1,"requestId":"rq-1"}`))
	}))
	t.Cleanup(srv.Close)

	svc := New(srv.Client(), Config{Ref: "7", APIKey: "gk-1", APIBase: srv.URL})
	receipt, err := svc.Send(context.Background(), &notify.Message{
		Target:  "disk-alert-1",
		Title:   "磁盘告警",
		Content: "root 分区用量 92%",
	})
	require.NoError(t, err)
	require.Equal(t, "rq-1", receipt.ProviderID, "requestId 要进回执")
	require.Equal(t, "GenieKey gk-1", auth)

	require.Equal(t, "磁盘告警", payload["message"], "message 必填位由 Title 顶上")
	require.Equal(t, "root 分区用量 92%", payload["description"], "Content 全文进 description")
	require.Equal(t, "disk-alert-1", payload["alias"], "alias 即去重关联键")
}

func TestSendTitleFallbackToContent(t *testing.T) {
	var payload map[string]any
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		require.NoError(t, json.NewDecoder(r.Body).Decode(&payload))
		_, _ = w.Write([]byte(`{"result":"Request will be processed","requestId":"rq"}`))
	}))
	t.Cleanup(srv.Close)

	svc := New(srv.Client(), Config{APIKey: "gk", APIBase: srv.URL})
	_, err := svc.Send(context.Background(), &notify.Message{Content: "没有标题的告警"})
	require.NoError(t, err)
	require.Equal(t, "没有标题的告警", payload["message"], "无 Title 时 message 由 Content 兜底")
}

func TestSendConfigIssue(t *testing.T) {
	svc := New(nil, Config{})
	_, err := svc.Send(context.Background(), &notify.Message{Content: "c"})
	require.ErrorIs(t, err, notify.ErrNotConfigured)
}
