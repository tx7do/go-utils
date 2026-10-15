package ifttt

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
	var (
		path    string
		query   string
		payload map[string]any
	)
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		path = r.URL.Path
		query = r.URL.RawQuery
		require.NoError(t, json.NewDecoder(r.Body).Decode(&payload))
		_, _ = w.Write([]byte("Congratulations! You've fired the build_done event"))
	}))
	t.Cleanup(srv.Close)

	svc := New(srv.Client(), Config{Ref: "7", APIKey: "k-1", APIBase: srv.URL})
	_, err := svc.Send(context.Background(), &notify.Message{
		Target:  "build_done",
		Title:   "构建完成",
		Content: "产物已就绪",
		Params:  map[string]string{"value3": "v1.2.0"},
	})
	require.NoError(t, err)
	require.Equal(t, "/trigger/build_done/json", path, "事件名进路径")
	require.Equal(t, "key=k-1", query)

	require.Equal(t, "构建完成", payload["value1"])
	require.Equal(t, "产物已就绪", payload["value2"])
	require.Equal(t, "v1.2.0", payload["value3"], "value3 从 Params 取")
}

func TestSendErrorAnswer(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		_, _ = w.Write([]byte(`{"errors":[{"message":"Authentication failed"}]}`))
	}))
	t.Cleanup(srv.Close)

	svc := New(srv.Client(), Config{APIKey: "bad", APIBase: srv.URL})
	_, err := svc.Send(context.Background(), &notify.Message{Target: "evt", Content: "c"})
	require.Error(t, err)
	require.Contains(t, err.Error(), "401")
}

func TestSendConfigIssue(t *testing.T) {
	svc := New(nil, Config{})
	_, err := svc.Send(context.Background(), &notify.Message{Target: "evt", Content: "c"})
	require.ErrorIs(t, err, notify.ErrNotConfigured)
}
