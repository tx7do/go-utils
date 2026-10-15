package pushdeer

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/tx7do/go-utils/notify"
)

func TestSendDeliversForm(t *testing.T) {
	var (
		form string
		path string
	)
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		path = r.URL.Path
		require.NoError(t, r.ParseForm())
		form = r.PostForm.Encode()
		_, _ = w.Write([]byte(`{"code":0,"content":{"result":[{"token":"ok"}]}}`))
	}))
	t.Cleanup(srv.Close)

	svc := New(srv.Client(), Config{Ref: "7", PushKey: "PK123", ServerURL: srv.URL})
	_, err := svc.Send(context.Background(), &notify.Message{
		Title:   "下载完成",
		Content: "**报告已生成**",
	})
	require.NoError(t, err)
	require.Equal(t, "/message/push", path)
	require.Contains(t, form, "pushkey=PK123")
	require.Contains(t, form, "text=%E4%B8%8B%E8%BD%BD%E5%AE%8C%E6%88%90", "text 是通知预览")
	require.Contains(t, form, "desp=%2A%2A%E6%8A%A5%E5%91%8A%E5%B7%B2%E7%94%9F%E6%88%90%2A%2A", "desp 是 Markdown 正文")
}

func TestSendPeerError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(`{"code":1000,"error":"wrong key"}`))
	}))
	t.Cleanup(srv.Close)

	svc := New(srv.Client(), Config{PushKey: "bad", ServerURL: srv.URL})
	_, err := svc.Send(context.Background(), &notify.Message{Content: "c"})
	require.Error(t, err)
	require.Contains(t, err.Error(), "wrong key")
}

func TestSendConfigIssue(t *testing.T) {
	svc := New(nil, Config{})
	_, err := svc.Send(context.Background(), &notify.Message{Content: "c"})
	require.ErrorIs(t, err, notify.ErrNotConfigured)
}
