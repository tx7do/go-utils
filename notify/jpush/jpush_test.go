package jpush

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/tx7do/go-utils/notify"
)

func TestSendDeliversBasicAuthAndAudience(t *testing.T) {
	var (
		auth    string
		path    string
		payload map[string]any
	)
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		path = r.URL.Path
		auth = r.Header.Get("Authorization")
		require.NoError(t, json.NewDecoder(r.Body).Decode(&payload))
		_, _ = w.Write([]byte(`{"sendno":"0","msg_id":1828256768}`))
	}))
	t.Cleanup(srv.Close)

	svc := New(srv.Client(), Config{Ref: "7", AppKey: "key-1", MasterSecret: "secret-1", APIBase: srv.URL})
	receipt, err := svc.Send(context.Background(), &notify.Message{
		Target:  "reg-token-1",
		Title:   "重置验证码",
		Content: "您的验证码是 123456",
	})
	require.NoError(t, err)
	require.Equal(t, "1828256768", receipt.ProviderID, "JPush 的 msg_id 要进回执")

	require.Equal(t, "/v3/push", path)
	require.Equal(t, "Basic "+base64.StdEncoding.EncodeToString([]byte("key-1:secret-1")), auth)

	audience := payload["audience"].(map[string]any)
	require.Equal(t, []any{"reg-token-1"}, audience["registration_id"], "单播 audience")

	notification := payload["notification"].(map[string]any)
	require.Equal(t, "您的验证码是 123456", notification["alert"])
	android := notification["android"].(map[string]any)
	require.Equal(t, "重置验证码", android["title"], "安卓有独立通知栏标题位")
	ios := notification["ios"].(map[string]any)
	require.Equal(t, map[string]any{"title": "重置验证码", "body": "您的验证码是 123456"}, ios["alert"], "iOS 10+ 的 alert 支持标题子段")
}

func TestSendInvalidRegistrationID(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte(`{"error":{"code":1011,"message":"cannot find user by this audience","request_id":"req-1"}}`))
	}))
	t.Cleanup(srv.Close)

	svc := New(srv.Client(), Config{AppKey: "k", MasterSecret: "s", APIBase: srv.URL})
	_, err := svc.Send(context.Background(), &notify.Message{Target: "gone", Title: "t", Content: "c"})
	require.ErrorIs(t, err, notify.ErrInvalidToken, "1011 = registration_id 失效，调用方要删行自清理")
}

func TestSendPeerError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		_, _ = w.Write([]byte(`{"error":{"code":1003,"message":"appkey not exist"}}`))
	}))
	t.Cleanup(srv.Close)

	svc := New(srv.Client(), Config{AppKey: "k", MasterSecret: "s", APIBase: srv.URL})
	_, err := svc.Send(context.Background(), &notify.Message{Target: "r", Title: "t", Content: "c"})
	require.Error(t, err)
	require.NotErrorIs(t, err, notify.ErrInvalidToken)
	require.Contains(t, err.Error(), "1003")
}

func TestSendConfigIssue(t *testing.T) {
	svc := New(nil, Config{Ref: "7", AppKey: "k"})
	_, err := svc.Send(context.Background(), &notify.Message{Target: "r", Title: "t", Content: "c"})
	require.ErrorIs(t, err, notify.ErrNotConfigured)
}
