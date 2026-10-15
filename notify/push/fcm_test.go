package push

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/tx7do/go-utils/notify"
)

// fcmFixture 本地假 Google 侧：token 端点断言 JWT-bearer 拨号并回 access_token，
// send 端点记录请求并可编程应答。
type fcmFixture struct {
	t *testing.T

	srvURL string
	client *http.Client

	tokenHits int
	assertion string // token 端点收到的 JWT assertion

	sendHits int
	sendPath string
	sendAuth string
	sendBody map[string]any
	sendFunc func() (int, string)
}

func newFCMFixture(t *testing.T) *fcmFixture {
	f := &fcmFixture{t: t}
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case strings.HasSuffix(r.URL.Path, "/token"):
			f.tokenHits++
			require.Equal(t, http.MethodPost, r.Method)
			require.Equal(t, "application/x-www-form-urlencoded", r.Header.Get("Content-Type"))
			require.Contains(t, r.FormValue("grant_type"), "jwt-bearer")
			f.assertion = r.FormValue("assertion")
			_, _ = w.Write([]byte(`{"access_token":"fake-oauth-token","expires_in":3600}`))

		case strings.HasSuffix(r.URL.Path, "messages:send"):
			f.sendHits++
			f.sendPath = r.URL.Path
			f.sendAuth = r.Header.Get("Authorization")
			require.NoError(t, json.NewDecoder(r.Body).Decode(&f.sendBody))
			status, answer := http.StatusOK, `{"name":"projects/proj-x/messages/2026101508"}`
			if f.sendFunc != nil {
				status, answer = f.sendFunc()
			}
			w.WriteHeader(status)
			_, _ = w.Write([]byte(answer))

		default:
			t.Fatalf("unexpected path %q", r.URL.Path)
		}
	}))
	t.Cleanup(srv.Close)
	f.srvURL = srv.URL
	f.client = srv.Client()
	return f
}

func (f *fcmFixture) service() *FCM {
	return NewFCM(f.client, FCMConfig{
		Ref:         "7",
		ProjectID:   "proj-x",
		ClientEmail: "sa@proj-x.iam.gserviceaccount.com",
		PrivateKey:  requireTestKeyPEM(f.t),
		TokenURI:    f.srvURL + "/token",
		APIBase:     f.srvURL,
	})
}

func TestFCMSendDeliversAndParsesMessageName(t *testing.T) {
	f := newFCMFixture(t)
	svc := f.service()

	receipt, err := svc.Send(context.Background(), &notify.Message{
		Target:  "reg-token-1",
		Title:   "重置验证码",
		Content: "您的验证码是 123456",
		Params:  map[string]string{"conversation_id": "42"},
	})
	require.NoError(t, err)
	require.Equal(t, "2026101508", receipt.ProviderID, "FCM 的 message name 尾段要进回执")

	// token 拨号：assertion 是三段式 JWT，claims 携带服务账号与 FCM scope。
	require.Equal(t, 1, f.tokenHits)
	parts := strings.Split(f.assertion, ".")
	require.Len(t, parts, 3)
	claimsJSON, decErr := base64DecodeSegment(parts[1])
	require.NoError(t, decErr)
	var claims map[string]any
	require.NoError(t, json.Unmarshal(claimsJSON, &claims))
	require.Equal(t, "sa@proj-x.iam.gserviceaccount.com", claims["iss"])
	require.Contains(t, claims["scope"], "firebase.messaging")

	// 发送：Bearer + 端点带项目段 + 消息形状。
	require.Equal(t, "Bearer fake-oauth-token", f.sendAuth)
	require.True(t, strings.HasSuffix(f.sendPath, "/v1/projects/proj-x/messages:send"), "端点要带项目 ID: %s", f.sendPath)
	msg := f.sendBody["message"].(map[string]any)
	require.Equal(t, "reg-token-1", msg["token"])
	require.Equal(t, map[string]any{"title": "重置验证码", "body": "您的验证码是 123456"}, msg["notification"])
	require.Equal(t, map[string]any{"conversation_id": "42"}, msg["data"])
}

// TestFCMTokenCachedAcrossSends 访问令牌按服务账号缓存：第二发不再拨 token 端点。
func TestFCMTokenCachedAcrossSends(t *testing.T) {
	f := newFCMFixture(t)
	svc := f.service()
	ctx := context.Background()

	msg := &notify.Message{Target: "reg-1", Title: "t", Content: "c"}
	_, err := svc.Send(ctx, msg)
	require.NoError(t, err)
	_, err = svc.Send(ctx, msg)
	require.NoError(t, err)

	require.Equal(t, 1, f.tokenHits, "两发只许换一次 OAuth 令牌")
	require.Equal(t, 2, f.sendHits)
}

func TestFCMSendInvalidToken(t *testing.T) {
	f := newFCMFixture(t)
	f.sendFunc = func() (int, string) { return http.StatusNotFound, `{"error":{"status":"NOT_FOUND"}}` }
	svc := f.service()

	_, err := svc.Send(context.Background(), &notify.Message{Target: "gone", Title: "t", Content: "c"})
	require.ErrorIs(t, err, ErrInvalidToken, "404 = 令牌已从注册表除名，调用方要删行自清理")
}

func TestFCMSendPeerError(t *testing.T) {
	f := newFCMFixture(t)
	f.sendFunc = func() (int, string) { return http.StatusTooManyRequests, `{"error":"quota exceeded"}` }
	svc := f.service()

	_, err := svc.Send(context.Background(), &notify.Message{Target: "reg-1", Title: "t", Content: "c"})
	require.Error(t, err)
	require.False(t, errors.Is(err, ErrInvalidToken))
	require.Contains(t, err.Error(), "429")
}

func TestFCMConfigIssues(t *testing.T) {
	svc := NewFCM(nil, FCMConfig{Ref: "7", ProjectID: "p"})
	_, err := svc.Send(context.Background(), &notify.Message{Target: "t", Title: "t", Content: "c"})
	require.ErrorIs(t, err, notify.ErrNotConfigured, "缺服务账号凭据在拨号之前就该拦下")

	_, err = svc.Send(context.Background(), &notify.Message{Title: "t", Content: "c"})
	require.Error(t, err)
	require.False(t, errors.Is(err, notify.ErrNotConfigured), "目标为空是调用方的问题")
}
