package mailgun

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/tx7do/go-utils/notify"
)

func TestSendDeliversBasicAuthForm(t *testing.T) {
	var (
		auth string
		form string
		path string
	)
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		path = r.URL.Path
		auth = r.Header.Get("Authorization")
		require.NoError(t, r.ParseForm())
		form = r.PostForm.Encode()
		_, _ = w.Write([]byte(`{"id":"<20261015.0@mg.example.com>","message":"Queued. Thank you."}`))
	}))
	t.Cleanup(srv.Close)

	svc := New(srv.Client(), Config{Ref: "7", Domain: "mg.example.com", APIKey: "key-1",
		From: "noreply@mg.example.com", APIBase: srv.URL})
	receipt, err := svc.Send(context.Background(), &notify.Message{
		Target:  "alice@example.com",
		Title:   "重置验证码",
		Content: "您的验证码是 123456",
	})
	require.NoError(t, err)
	require.Equal(t, "<20261015.0@mg.example.com>", receipt.ProviderID, "Mailgun 的消息 id 要进回执")

	require.Equal(t, "Basic YXBpOmtleS0x", auth, "Basic auth(api:key)，用户名固定 api")
	require.Equal(t, "/v3/mg.example.com/messages", path, "域名进路径")
	require.Contains(t, form, "from=noreply%40mg.example.com")
	require.Contains(t, form, "to=alice%40example.com")
	require.Contains(t, form, "subject=%E9%87%8D%E7%BD%AE%E9%AA%8C%E8%AF%81%E7%A0%81")
	require.Contains(t, form, "text=%E6%82%A8%E7%9A%84%E9%AA%8C%E8%AF%81%E7%A0%81%E6%98%AF+123456")
}

func TestSendConfigIssue(t *testing.T) {
	svc := New(nil, Config{})
	_, err := svc.Send(context.Background(), &notify.Message{Target: "a@b.c", Content: "c"})
	require.ErrorIs(t, err, notify.ErrNotConfigured)
}
