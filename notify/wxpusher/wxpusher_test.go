package wxpusher

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/tx7do/go-utils/notify"
)

func TestSendDeliversUIDMessage(t *testing.T) {
	var payload map[string]any
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, "/api/send/message", r.URL.Path)
		require.NoError(t, json.NewDecoder(r.Body).Decode(&payload))
		_, _ = w.Write([]byte(`{"code":1000,"msg":"处理成功","data":[{"uid":"UID_x","messageId":731}]}`))
	}))
	t.Cleanup(srv.Close)

	svc := New(srv.Client(), Config{Ref: "7", AppToken: "AT_x", APIBase: srv.URL})
	receipt, err := svc.Send(context.Background(), &notify.Message{
		Target:  "UID_x",
		Title:   "重置验证码",
		Content: "您的验证码是 123456",
	})
	require.NoError(t, err)
	require.Equal(t, "731", receipt.ProviderID, "WxPusher 的 messageId 要进回执")

	require.Equal(t, "AT_x", payload["appToken"])
	require.Equal(t, []any{"UID_x"}, payload["uids"])
	require.Equal(t, "您的验证码是 123456", payload["content"])
	require.Equal(t, float64(1), payload["contentType"], "默认文本形态")
	require.Equal(t, "重置验证码", payload["summary"], "summary 是通知栏预览（Title 兜底）")
}

// TestSendSuccessCodeIs1000 WxPusher 的成功码是 1000 不是 0——钉住，别被顺手"统一"成 0。
func TestSendSuccessCodeIs1000(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(`{"code":1001,"msg":"appToken无效"}`))
	}))
	t.Cleanup(srv.Close)

	svc := New(srv.Client(), Config{AppToken: "bad", APIBase: srv.URL})
	_, err := svc.Send(context.Background(), &notify.Message{Target: "UID_x", Content: "c"})
	require.Error(t, err)
	require.Contains(t, err.Error(), "code=1001")
}

func TestSendConfigIssue(t *testing.T) {
	svc := New(nil, Config{})
	_, err := svc.Send(context.Background(), &notify.Message{Target: "UID_x", Content: "c"})
	require.ErrorIs(t, err, notify.ErrNotConfigured)
}
