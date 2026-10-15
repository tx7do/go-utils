package pushover

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/tx7do/go-utils/notify"
)

func TestSendDeliversFormShape(t *testing.T) {
	var (
		form string
	)
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		require.NoError(t, r.ParseForm())
		form = r.Form.Encode()
		_, _ = w.Write([]byte(`{"status":1,"request":"e4e3a4"}`))
	}))
	t.Cleanup(srv.Close)

	svc := New(srv.Client(), Config{Ref: "7", AppToken: "app-token", APIBase: srv.URL})
	receipt, err := svc.Send(context.Background(), &notify.Message{
		Target:  "user-key-30chars",
		Title:   "重置验证码",
		Content: "您的验证码是 123456",
		Params:  map[string]string{"priority": "2", "sound": "siren"},
	})
	require.NoError(t, err)
	require.Equal(t, "e4e3a4", receipt.ProviderID, "Pushover 的 request ID 要进回执")

	require.Contains(t, form, "token=app-token")
	require.Contains(t, form, "user=user-key-30chars")
	require.Contains(t, form, "title=%E9%87%8D%E7%BD%AE%E9%AA%8C%E8%AF%81%E7%A0%81")
	require.Contains(t, form, "priority=2", "Params 原生字段直传表单")
	require.Contains(t, form, "sound=siren")
}

func TestSendStatusZeroMeansRejected(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(`{"status":0,"request":"x","errors":["user identifier is not a valid user, group, or subscribed user key"]}`))
	}))
	t.Cleanup(srv.Close)

	svc := New(srv.Client(), Config{AppToken: "t", APIBase: srv.URL})
	_, err := svc.Send(context.Background(), &notify.Message{Target: "bad", Content: "c"})
	require.Error(t, err)
	require.Contains(t, err.Error(), "user identifier is not a valid user")
}

func TestSendConfigIssues(t *testing.T) {
	svc := New(nil, Config{})
	_, err := svc.Send(context.Background(), &notify.Message{Target: "user-key", Content: "c"})
	require.ErrorIs(t, err, notify.ErrNotConfigured, "缺 app_token 在拨号之前就该拦下")

	_, err = svc.Send(context.Background(), &notify.Message{Content: "c"})
	require.Error(t, err)
	require.False(t, errors.Is(err, notify.ErrNotConfigured), "user key 为空是调用方的问题")
}
