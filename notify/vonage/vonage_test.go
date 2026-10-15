package vonage

import (
	"context"
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
		require.Equal(t, "/sms/json", r.URL.Path)
		require.NoError(t, r.ParseForm())
		form = r.PostForm.Encode()
		_, _ = w.Write([]byte(`{"message-count":"1","messages":[{"status":"0","message-id":"0A00000012AB","to":"8613800138000"}]}`))
	}))
	t.Cleanup(srv.Close)

	svc := New(srv.Client(), Config{Ref: "7", APIKey: "k1", APISecret: "s1",
		From: "Acme", APIBase: srv.URL})
	receipt, err := svc.Send(context.Background(), &notify.Message{
		Target:  "+8613800138000",
		Title:   "重置验证码",
		Content: "您的验证码是 123456",
	})
	require.NoError(t, err)
	require.Equal(t, "0A00000012AB", receipt.ProviderID, "message-id 要进回执")

	require.Contains(t, form, "api_key=k1")
	require.Contains(t, form, "api_secret=s1")
	require.Contains(t, form, "to=8613800138000", "Vonage 收件号无 + 前缀")
	require.Contains(t, form, "from=Acme")
	require.Contains(t, form, "text=%E9%87%8D%E7%BD%AE%E9%AA%8C%E8%AF%81%E7%A0%81")
}

func TestSendNonZeroStatusIsError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(`{"message-count":"1","messages":[{"status":"9","error-text":"Invalid api key"}]}`))
	}))
	t.Cleanup(srv.Close)

	svc := New(srv.Client(), Config{APIKey: "k", APISecret: "s", From: "F", APIBase: srv.URL})
	_, err := svc.Send(context.Background(), &notify.Message{Target: "8613800138000", Content: "c"})
	require.Error(t, err)
	require.Contains(t, err.Error(), "status=9")
	require.Contains(t, err.Error(), "Invalid api key")
}

func TestSendConfigIssue(t *testing.T) {
	svc := New(nil, Config{APIKey: "k"})
	_, err := svc.Send(context.Background(), &notify.Message{Target: "8613800138000", Content: "c"})
	require.ErrorIs(t, err, notify.ErrNotConfigured)
}
