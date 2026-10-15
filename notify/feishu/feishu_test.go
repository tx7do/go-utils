package feishu

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/tx7do/go-utils/notify"
)

func TestSendDeliversTextWithJSONContent(t *testing.T) {
	var (
		hitToken int
		auth     string
		query    string
		payload  map[string]any
	)
	mux := http.NewServeMux()
	mux.HandleFunc("/open-apis/auth/v3/tenant_access_token/internal", func(w http.ResponseWriter, r *http.Request) {
		hitToken++
		var req map[string]string
		require.NoError(t, json.NewDecoder(r.Body).Decode(&req))
		require.Equal(t, "app-1", req["app_id"])
		require.Equal(t, "secret-1", req["app_secret"])
		_, _ = w.Write([]byte(`{"code":0,"msg":"ok","tenant_access_token":"t-1","expire":7200}`))
	})
	mux.HandleFunc("/open-apis/im/v1/messages", func(w http.ResponseWriter, r *http.Request) {
		auth = r.Header.Get("Authorization")
		query = r.URL.RawQuery
		require.NoError(t, json.NewDecoder(r.Body).Decode(&payload))
		_, _ = w.Write([]byte(`{"code":0,"msg":"success","data":{"message_id":"om_abc123"}}`))
	})
	srv := httptest.NewServer(mux)
	t.Cleanup(srv.Close)

	svc := New(srv.Client(), Config{Ref: "7", AppID: "app-1", AppSecret: "secret-1", APIBase: srv.URL})
	receipt, err := svc.Send(context.Background(), &notify.Message{
		Target:  "ou_user1",
		Title:   "审批提醒",
		Content: "您有一张单据待审批",
	})
	require.NoError(t, err)
	require.Equal(t, "om_abc123", receipt.ProviderID, "飞书的 message_id 要进回执")
	require.Equal(t, "Bearer t-1", auth)
	require.Equal(t, "receive_id_type=open_id", query, "接收人类型默认 open_id")

	require.Equal(t, "ou_user1", payload["receive_id"])
	require.Equal(t, "text", payload["msg_type"])

	// content 是"JSON 字符串套 JSON"：内层先序列化成串。
	var inner map[string]string
	require.NoError(t, json.Unmarshal([]byte(payload["content"].(string)), &inner))
	require.Equal(t, "审批提醒\n您有一张单据待审批", inner["text"])
}

func TestSendPeerErrorCode(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(`{"code":230001,"msg":"invalid receive_id"}`))
	}))
	t.Cleanup(srv.Close)

	svc := New(srv.Client(), Config{AppID: "a", AppSecret: "s", APIBase: srv.URL})
	_, err := svc.Send(context.Background(), &notify.Message{Target: "ou_x", Content: "c"})
	require.Error(t, err)
	require.Contains(t, err.Error(), "code=230001")
}

func TestSendConfigIssue(t *testing.T) {
	svc := New(nil, Config{Ref: "7", AppID: "a"})
	_, err := svc.Send(context.Background(), &notify.Message{Target: "ou_x", Content: "c"})
	require.ErrorIs(t, err, notify.ErrNotConfigured)
}
