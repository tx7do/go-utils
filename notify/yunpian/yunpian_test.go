package yunpian

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
		require.Equal(t, "/v2/sms/single_send.json", r.URL.Path)
		require.NoError(t, r.ParseForm())
		form = r.PostForm.Encode()
		_, _ = w.Write([]byte(`{"code":0,"msg":"OK","count":1,"fee":1,"unit":"RMB","mobile":"13800138000","sid":1690001}`))
	}))
	t.Cleanup(srv.Close)

	svc := New(srv.Client(), Config{Ref: "7", ApiKey: "apikey-1", APIBase: srv.URL})
	receipt, err := svc.Send(context.Background(), &notify.Message{
		Target:  "13800138000",
		Content: "【GoWind】您的验证码是 123456",
	})
	require.NoError(t, err)
	require.Equal(t, "1690001", receipt.ProviderID, "云片的 sid 要进回执")

	require.Contains(t, form, "apikey=apikey-1")
	require.Contains(t, form, "mobile=13800138000", "CN 裸号出站（NormalizePhoneNumber 归一）")
	require.Contains(t, form, "text=%E3%80%90GoWind%E3%80%91", "正文整段自由文本（含签名）原样出站")
}

func TestSendE164Normalized(t *testing.T) {
	var form string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		require.NoError(t, r.ParseForm())
		form = r.PostForm.Encode()
		_, _ = w.Write([]byte(`{"code":0,"msg":"OK","sid":1}`))
	}))
	t.Cleanup(srv.Close)

	svc := New(srv.Client(), Config{ApiKey: "k", APIBase: srv.URL})
	_, err := svc.Send(context.Background(), &notify.Message{Target: "+8613800138000", Content: "【G】x"})
	require.NoError(t, err)
	require.Contains(t, form, "mobile=13800138000", "+86 前缀剥掉")
}

func TestSendBadMobileIsConfigClass(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {}))
	t.Cleanup(srv.Close)

	svc := New(srv.Client(), Config{ApiKey: "k", APIBase: srv.URL})
	_, err := svc.Send(context.Background(), &notify.Message{Target: "not-a-phone", Content: "【G】x"})
	require.ErrorIs(t, err, notify.ErrNotConfigured, "号码不合法一次都不碰对端")
}

func TestSendPeerErrorCode(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(`{"code":-4,"msg":"apikey 不存在或被禁用"}`))
	}))
	t.Cleanup(srv.Close)

	svc := New(srv.Client(), Config{ApiKey: "k", APIBase: srv.URL})
	_, err := svc.Send(context.Background(), &notify.Message{Target: "13800138000", Content: "【G】x"})
	require.Error(t, err)
	require.Contains(t, err.Error(), "code=-4")
}

func TestSendConfigIssue(t *testing.T) {
	svc := New(nil, Config{})
	_, err := svc.Send(context.Background(), &notify.Message{Target: "13800138000", Content: "【G】x"})
	require.ErrorIs(t, err, notify.ErrNotConfigured)
}
