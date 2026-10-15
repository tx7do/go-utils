package serverchan

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/tx7do/go-utils/notify"
)

func TestSendDeliversFormToKeyedEndpoint(t *testing.T) {
	var (
		path string
		form string
	)
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		path = r.URL.Path
		require.NoError(t, r.ParseForm())
		form = r.PostForm.Encode()
		_, _ = w.Write([]byte(`{"code":0,"message":"","data":{"pushed":"1","total":"1"}}`))
	}))
	t.Cleanup(srv.Close)

	svc := New(srv.Client(), Config{Ref: "7", SendKey: "SCT123456", APIBase: srv.URL})
	_, err := svc.Send(context.Background(), &notify.Message{
		Title:   "服务器告警",
		Content: "CPU 95%",
	})
	require.NoError(t, err)
	require.Equal(t, "/SCT123456.send", path, "SendKey 进路径")
	require.Contains(t, form, "title=%E6%9C%8D%E5%8A%A1%E5%99%A8%E5%91%8A%E8%AD%A6")
	require.Contains(t, form, "desp=CPU+95%25", "desp 是 Markdown 正文")
}

func TestSendPeerErrorCode(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(`{"code":40001,"message":"用户不存在"}`))
	}))
	t.Cleanup(srv.Close)

	svc := New(srv.Client(), Config{SendKey: "bad", APIBase: srv.URL})
	_, err := svc.Send(context.Background(), &notify.Message{Content: "c"})
	require.Error(t, err)
	require.Contains(t, err.Error(), "code=40001")
}

func TestSendConfigIssue(t *testing.T) {
	svc := New(nil, Config{})
	_, err := svc.Send(context.Background(), &notify.Message{Content: "c"})
	require.ErrorIs(t, err, notify.ErrNotConfigured)
}
