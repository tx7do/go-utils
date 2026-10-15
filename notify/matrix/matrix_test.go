package matrix

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/tx7do/go-utils/notify"
)

func TestSendDeliversEventPut(t *testing.T) {
	var (
		rawURI  string
		query   string
		payload map[string]any
	)
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		rawURI = r.RequestURI
		query = r.URL.RawQuery
		require.Equal(t, "PUT", r.Method, "事务发送是 PUT（txnId 幂等）")
		require.NoError(t, json.NewDecoder(r.Body).Decode(&payload))
		_, _ = w.Write([]byte(`{"event_id":"$abc123:example.org"}`))
	}))
	t.Cleanup(srv.Close)

	svc := New(srv.Client(), Config{Ref: "7", ServerURL: srv.URL, AccessToken: "syt-1"})
	receipt, err := svc.Send(context.Background(), &notify.Message{
		Target:  "!room:example.org",
		Title:   "部署完成",
		Content: "v1.2.0 已上线",
	})
	require.NoError(t, err)
	require.Equal(t, "$abc123:example.org", receipt.ProviderID, "event_id 要进回执")

	// RequestURI 是原始请求行：r.URL.Path 是解码视图，看不出转义是否发生。
	require.True(t, strings.HasPrefix(rawURI, "/_matrix/client/v3/rooms/%21room%3Aexample.org/send/m.text/"),
		"房间 ID 要 URL 转义进路径（! 和 : 是保留字）: %s", rawURI)
	require.Contains(t, query, "access_token=syt-1")
	require.Equal(t, "m.text", payload["msgtype"])
	require.Equal(t, "部署完成\nv1.2.0 已上线", payload["body"])
}

func TestSendMatrixError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusForbidden)
		_, _ = w.Write([]byte(`{"errcode":"M_FORBIDDEN","error":"You are not in the room"}`))
	}))
	t.Cleanup(srv.Close)

	svc := New(srv.Client(), Config{ServerURL: srv.URL, AccessToken: "syt-1"})
	_, err := svc.Send(context.Background(), &notify.Message{Target: "!r:s.org", Content: "c"})
	require.Error(t, err)
	require.Contains(t, err.Error(), "M_FORBIDDEN")
}

func TestSendConfigIssue(t *testing.T) {
	svc := New(nil, Config{ServerURL: "https://example.org"})
	_, err := svc.Send(context.Background(), &notify.Message{Target: "!r:s.org", Content: "c"})
	require.ErrorIs(t, err, notify.ErrNotConfigured)
}
