package push

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/tx7do/go-utils/notify"
)

func TestVivoSign(t *testing.T) {
	// 签名 = MD5(appKey + timestamp + appSecret) 小写 hex——钉一条独立复算值，
	// 厂商签名公式漂移时这里是第一道闸。
	require.Equal(t, VivoSign("key-1", "1760000000", "secret-1"), VivoSign("key-1", "1760000000", "secret-1"))
	require.Equal(t, md5hex("key-11760000000secret-1"), VivoSign("key-1", "1760000000", "secret-1"))
	require.NotEqual(t, VivoSign("key-1", "1760000000", "secret-1"), VivoSign("key-2", "1760000000", "secret-1"),
		"不同 appKey 的签名必须不同：签名错对端只会回 sign fail")
}

func TestVivoSendDeliversAuthQuadAndPayload(t *testing.T) {
	var (
		hit     int
		form    string
		answer  = `{"code":0,"msg":"ok","taskId":"task-1"}`
		capture = func(raw string) { form = raw }
	)
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		hit++
		raw, _ := io.ReadAll(r.Body)
		capture(string(raw))
		_, _ = w.Write([]byte(answer))
	}))
	t.Cleanup(srv.Close)

	svc := NewVivo(srv.Client(), VivoConfig{Ref: "7", AppID: "100", AppKey: "key-1",
		AppSecret: "secret-1", SendURL: srv.URL})
	_, err := svc.Send(context.Background(), &notify.Message{
		Target:  "reg-1",
		Title:   "重置验证码",
		Content: "您的验证码是 123456",
	})
	require.NoError(t, err)
	require.Equal(t, 1, hit)

	got, parseErr := parseForm(form)
	require.NoError(t, parseErr)
	require.Equal(t, "100", got["appId"])
	require.Equal(t, "key-1", got["appKey"])
	require.Equal(t, VivoSign("key-1", got["timestamp"], "secret-1"), got["sign"],
		"表单里的 sign 必须能用表单里的 timestamp 复算出来：四元组不成对对端直接拒")
	require.Equal(t, "reg-1", got["regId"])
	require.Equal(t, "重置验证码", got["title"])
}

func TestVivoInvalidRegID(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(`{"code":10080,"msg":"regId invalid"}`))
	}))
	t.Cleanup(srv.Close)

	svc := NewVivo(srv.Client(), VivoConfig{AppID: "100", AppKey: "k", AppSecret: "s", SendURL: srv.URL})
	_, err := svc.Send(context.Background(), &notify.Message{Target: "gone", Title: "t", Content: "c"})
	require.ErrorIs(t, err, ErrInvalidToken, "10080 = regId 失效，调用方要删行自清理")
}

func TestVivoConfigIssue(t *testing.T) {
	svc := NewVivo(nil, VivoConfig{Ref: "7", AppID: "100", AppKey: "k"})
	_, err := svc.Send(context.Background(), &notify.Message{Target: "reg", Title: "t", Content: "c"})
	require.ErrorIs(t, err, notify.ErrNotConfigured, "缺 app_secret 在拨号之前就该拦下")
}
