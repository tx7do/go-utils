package join

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/tx7do/go-utils/notify"
)

func TestSendDeliversToDeviceIDs(t *testing.T) {
	var (
		form string
		path string
	)
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		path = r.URL.Path
		require.NoError(t, r.ParseForm())
		form = r.PostForm.Encode()
		_, _ = w.Write([]byte(`{"success":true,"userAccountId":"acc"}`))
	}))
	t.Cleanup(srv.Close)

	svc := New(srv.Client(), Config{Ref: "7", APIKey: "key-1", APIBase: srv.URL})
	_, err := svc.Send(context.Background(), &notify.Message{
		Target:  "dev1,dev2",
		Title:   "构建完成",
		Content: "产物已就绪",
	})
	require.NoError(t, err)
	require.Equal(t, "/_ah/api/messaging/v1/sendPush", path)
	require.Contains(t, form, "apikey=key-1")
	require.Contains(t, form, "deviceIds=dev1%2Cdev2", "多设备逗号相连走 deviceIds")
	require.Contains(t, form, "title=%E6%9E%84%E5%BB%BA%E5%AE%8C%E6%88%90")
	require.Contains(t, form, "text=%E6%9E%84%E5%BB%BA%E5%AE%8C%E6%88%90%0A%E4%BA%A7%E7%89%A9%E5%B7%B2%E5%B0%B1%E7%BB%AA")
}

func TestSendFailureAnswer(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(`{"success":false,"errorMessage":"API key not valid"}`))
	}))
	t.Cleanup(srv.Close)

	svc := New(srv.Client(), Config{APIKey: "bad", APIBase: srv.URL})
	_, err := svc.Send(context.Background(), &notify.Message{Content: "c"})
	require.Error(t, err)
	require.Contains(t, err.Error(), "API key not valid")
}

func TestSendConfigIssue(t *testing.T) {
	svc := New(nil, Config{})
	_, err := svc.Send(context.Background(), &notify.Message{Content: "c"})
	require.ErrorIs(t, err, notify.ErrNotConfigured)
}
