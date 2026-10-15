package push

import (
	"context"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/tx7do/go-utils/notify"
)

// oppoFixture 本地假 OPPO Push：auth + send 两端点，记录表单。
type oppoFixture struct {
	authHits int
	sendHits int
	sendForm string

	authURL string
	sendURL string
	svc     *Oppo
}

func newOppoFixture(t *testing.T) *oppoFixture {
	f := &oppoFixture{}
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		raw, _ := io.ReadAll(r.Body)
		switch {
		case strings.Contains(r.URL.Path, "/auth/"):
			f.authHits++
			_, _ = w.Write([]byte(`{"code":0,"data":{"authToken":"oppo-token","expiresIn":3600}}`))
		default:
			f.sendHits++
			f.sendForm = string(raw)
			_, _ = w.Write([]byte(`{"code":0,"message":"ok"}`))
		}
	}))
	t.Cleanup(srv.Close)
	f.authURL = srv.URL + "/auth/aliyun"
	f.sendURL = srv.URL + "/notification/unicast"
	f.svc = NewOppo(srv.Client(), OppoConfig{Ref: "7", AppKey: "key-1", AppSecret: "secret-1",
		AuthURL: f.authURL, SendURL: f.sendURL})
	return f
}

func TestOppoSendDeliversTwoStageAuth(t *testing.T) {
	f := newOppoFixture(t)
	ctx := context.Background()

	msg := &notify.Message{Target: "reg-1", Title: "重置验证码", Content: "您的验证码是 123456"}
	_, err := f.svc.Send(ctx, msg)
	require.NoError(t, err)
	_, err = f.svc.Send(ctx, msg)
	require.NoError(t, err)

	require.Equal(t, 1, f.authHits, "auth token 按实例缓存：两发只许换一次 token")
	require.Equal(t, 2, f.sendHits)

	form, parseErr := parseForm(f.sendForm)
	require.NoError(t, parseErr)
	require.Equal(t, "oppo-token", form["auth_token"])
	require.Equal(t, "reg-1", form["registration_id"])
	require.Equal(t, "0", form["click_action_type"], "点击打开应用")
}

// TestOppoTitleTruncatedTo28Runes OPPO 通知标题上限 32 字符：超限直接 400，
// 所以在 28 runes 处截断——按 rune 不按字节，中文标题才不会腰斩出半个字。
func TestOppoTitleTruncatedTo28Runes(t *testing.T) {
	f := newOppoFixture(t)

	long := ""
	for i := 0; i < 40; i++ {
		long += "重"
	}
	_, err := f.svc.Send(context.Background(), &notify.Message{Target: "reg-1", Title: long, Content: "c"})
	require.NoError(t, err)

	form, parseErr := parseForm(f.sendForm)
	require.NoError(t, parseErr)
	require.Len(t, []rune(form["title"]), oppoTitleMax, "标题按 rune 截到 28 个")
	require.Equal(t, long[:1], form["title"][:1], "截断是前缀截断")
}

func TestOppoInvalidRegistrationID(t *testing.T) {
	f := newOppoFixture(t)
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(`{"code":2008,"message":"registration_id invalid"}`))
	}))
	t.Cleanup(srv.Close)
	f.svc = NewOppo(srv.Client(), OppoConfig{AppKey: "k", AppSecret: "s",
		AuthURL: f.authURL, SendURL: srv.URL})

	_, err := f.svc.Send(context.Background(), &notify.Message{Target: "gone", Title: "t", Content: "c"})
	require.ErrorIs(t, err, ErrInvalidToken, "2008 = registration_id 失效，调用方要删行自清理")
}

func TestOppoConfigIssue(t *testing.T) {
	svc := NewOppo(nil, OppoConfig{Ref: "7"})
	_, err := svc.Send(context.Background(), &notify.Message{Target: "reg", Title: "t", Content: "c"})
	require.ErrorIs(t, err, notify.ErrNotConfigured)
	require.False(t, errors.Is(err, ErrInvalidToken))
}
