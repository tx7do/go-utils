package push

import (
	"context"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/tx7do/go-utils/notify"
)

// xiaomiFixture 本地假 MiPush：记录请求并可编程应答。
type xiaomiFixture struct {
	svc *Xiaomi

	hit    int
	form   string
	auth   string
	answer string
}

func newXiaomiFixture(t *testing.T, answer string) *xiaomiFixture {
	f := &xiaomiFixture{answer: answer}
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		f.hit++
		f.auth = r.Header.Get("Authorization")
		raw, _ := io.ReadAll(r.Body)
		f.form = string(raw)
		_, _ = w.Write([]byte(f.answer))
	}))
	t.Cleanup(srv.Close)
	f.svc = NewXiaomi(srv.Client(), XiaomiConfig{Ref: "7", AppSecret: "secret-1", SendURL: srv.URL})
	return f
}

func TestXiaomiSendDeliversNotificationForm(t *testing.T) {
	f := newXiaomiFixture(t, `{"result":"ok","code":0,"data":{"id":"msg-1"}}`)

	_, err := f.svc.Send(context.Background(), &notify.Message{
		Target:  "reg-1",
		Title:   "重置验证码",
		Content: "您的验证码是 123456",
		Params:  map[string]string{"uri": "app://chat/42"},
	})
	require.NoError(t, err)
	require.Equal(t, `Key=secret-1`, f.auth, "MiPush 的凭据是静态 Key 头")

	form, err := parseForm(f.form)
	require.NoError(t, err)
	require.Equal(t, "reg-1", form["registration_id"])
	require.Equal(t, "重置验证码", form["title"])
	require.Equal(t, "您的验证码是 123456", form["description"])
	require.Equal(t, "0", form["pass_through"], "通知栏形态，非透传")
	require.Equal(t, "app://chat/42", form["extra.uri"], "Params 以 extra.{key} 表单键透传，点击深链靠它")
}

func TestXiaomiInvalidRegId(t *testing.T) {
	f := newXiaomiFixture(t, `{"result":"fail","code":20201}`)
	_, err := f.svc.Send(context.Background(), &notify.Message{Target: "gone", Title: "t", Content: "c"})
	require.ErrorIs(t, err, ErrInvalidToken, "20201/20202 = regId 失效，调用方要删行自清理")
}

func TestXiaomiPeerError(t *testing.T) {
	f := newXiaomiFixture(t, `{"result":"fail","code":1001,"reason":"quota"}`)
	_, err := f.svc.Send(context.Background(), &notify.Message{Target: "reg", Title: "t", Content: "c"})
	require.Error(t, err)
	require.False(t, errors.Is(err, ErrInvalidToken))
	require.Contains(t, err.Error(), "1001")
}

func TestXiaomiConfigIssue(t *testing.T) {
	svc := NewXiaomi(nil, XiaomiConfig{Ref: "7"})
	_, err := svc.Send(context.Background(), &notify.Message{Target: "reg", Title: "t", Content: "c"})
	require.ErrorIs(t, err, notify.ErrNotConfigured)
}
