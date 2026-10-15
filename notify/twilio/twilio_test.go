package twilio

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/tx7do/go-utils/notify"
)

func TestSendDeliversBasicAuthAndForm(t *testing.T) {
	var (
		auth string
		form map[string][]string
		path string
	)
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		path = r.URL.Path
		auth = r.Header.Get("Authorization")
		require.NoError(t, r.ParseForm())
		form = r.PostForm
		w.WriteHeader(http.StatusCreated)
		_, _ = w.Write([]byte(`{"sid":"SM8f","status":"queued","to":"+8613800138000"}`))
	}))
	t.Cleanup(srv.Close)

	svc := New(srv.Client(), Config{Ref: "7", AccountSID: "AC123", AuthToken: "tok",
		From: "+15550001111", APIBase: srv.URL})
	receipt, err := svc.Send(context.Background(), &notify.Message{
		Target:  "+8613800138000",
		Title:   "重置验证码",
		Content: "您的验证码是 123456",
	})
	require.NoError(t, err)
	require.Equal(t, "SM8f", receipt.ProviderID, "Twilio 的消息 SID 要进回执")

	require.Equal(t, "Basic QUMxMjM6dG9r", auth, "Basic auth(sid:token)")
	require.True(t, path == "/2010-04-01/Accounts/AC123/Messages.json", "SID 要进路径: %s", path)
	require.Equal(t, "+8613800138000", form["To"][0])
	require.Equal(t, "+15550001111", form["From"][0])
	require.Equal(t, "重置验证码\n您的验证码是 123456", form["Body"][0])
}

func TestSendPeerError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte(`{"code":21211,"message":"The 'To' number is not a valid phone number"}`))
	}))
	t.Cleanup(srv.Close)

	svc := New(srv.Client(), Config{AccountSID: "AC", AuthToken: "t", From: "F", APIBase: srv.URL})
	_, err := svc.Send(context.Background(), &notify.Message{Target: "bad", Content: "c"})
	require.Error(t, err)
	require.Contains(t, err.Error(), "21211")
}

func TestSendConfigIssues(t *testing.T) {
	svc := New(nil, Config{AccountSID: "AC"})
	_, err := svc.Send(context.Background(), &notify.Message{Target: "+8613800138000", Content: "c"})
	require.ErrorIs(t, err, notify.ErrNotConfigured)

	_, err = svc.Send(context.Background(), &notify.Message{Content: "c"})
	require.Error(t, err)
}
