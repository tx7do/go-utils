package push

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/tx7do/go-utils/notify"
)

// huaweiFixture 本地假华为/荣耀 Push Kit：token + send 两端点分开计数。
type huaweiFixture struct {
	t *testing.T

	srvURL string

	tokenHits int
	sendHits  int
	sendAuth  string
	sendPath  string
	sendBody  map[string]any
	sendCode  int // send 端点 body 里的厂商码
}

func newHuaweiFixture(t *testing.T, sendCode int) *huaweiFixture {
	f := &huaweiFixture{t: t, sendCode: sendCode}
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case strings.HasSuffix(r.URL.Path, "/token"):
			f.tokenHits++
			require.Equal(t, "client_credentials", r.FormValue("grant_type"))
			require.Equal(t, "client-1", r.FormValue("client_id"))
			require.Equal(t, "secret-1", r.FormValue("client_secret"))
			_, _ = w.Write([]byte(`{"access_token":"hw-token","expires_in":3600}`))

		case strings.HasSuffix(r.URL.Path, "messages:send"):
			f.sendHits++
			f.sendAuth = r.Header.Get("Authorization")
			f.sendPath = r.URL.Path
			require.NoError(t, json.NewDecoder(r.Body).Decode(&f.sendBody))
			_, _ = w.Write([]byte(`{"code":` + strconv.Itoa(sendCode) + `,"msg":"answer","requestId":"req-42"}`))

		default:
			t.Fatalf("unexpected path %q", r.URL.Path)
		}
	}))
	t.Cleanup(srv.Close)
	f.srvURL = srv.URL
	return f
}

func TestHuaweiSendDeliversAndParsesRequestID(t *testing.T) {
	f := newHuaweiFixture(t, huaweiCodeOK)

	// 荣耀/华为协议同形，只差默认域名；端点被 Config 覆写后两者等价——
	// 这里顺带走一遍 Honor 构造器，验证默认域切换与覆盖逻辑共用。
	honor := NewHonor(nil, HuaweiConfig{Ref: "7", ClientID: "client-1", ClientSecret: "secret-1"})
	require.Contains(t, "honor", honor.Name())

	hw := NewHuawei(nil, HuaweiConfig{Ref: "7", ClientID: "client-1", ClientSecret: "secret-1",
		TokenURL: f.srvURL + "/token", SendURL: f.srvURL + "/v1/app-%s/messages:send"})

	receipt, err := hw.Send(context.Background(), &notify.Message{
		Target:  "hw-reg-1",
		Title:   "重置验证码",
		Content: "您的验证码是 123456",
		Params:  map[string]string{"msg_id": "7"},
	})
	require.NoError(t, err)
	require.Equal(t, "req-42", receipt.ProviderID, "华为的 requestId 要进回执")

	require.Equal(t, 1, f.tokenHits)
	require.Equal(t, "Bearer hw-token", f.sendAuth)
	require.True(t, strings.HasSuffix(f.sendPath, "/v1/app-client-1/messages:send"), "app_id 要进路径: %s", f.sendPath)

	message := f.sendBody["message"].(map[string]any)
	require.Equal(t, "hw-reg-1", message["token"])
	require.Equal(t, map[string]any{"title": "重置验证码", "body": "您的验证码是 123456"}, message["notification"])
	require.Equal(t, map[string]any{"msg_id": "7"}, message["data"])
}

func TestHuaweiInvalidTokenCode(t *testing.T) {
	// 0x80200xxx 段是令牌无效类；取段中一个代表值钉住。
	f := newHuaweiFixture(t, 80200007)
	hw := NewHuawei(nil, HuaweiConfig{ClientID: "client-1", ClientSecret: "secret-1",
		TokenURL: f.srvURL + "/token", SendURL: f.srvURL + "/v1/app-%s/messages:send"})

	_, err := hw.Send(context.Background(), &notify.Message{Target: "gone", Title: "t", Content: "c"})
	require.ErrorIs(t, err, ErrInvalidToken, "80200xxx 段 = 令牌失效，调用方要删行自清理")
}

func TestHuaweiNonOKCodeIsTransportError(t *testing.T) {
	f := newHuaweiFixture(t, 80000001)
	hw := NewHuawei(nil, HuaweiConfig{ClientID: "client-1", ClientSecret: "secret-1",
		TokenURL: f.srvURL + "/token", SendURL: f.srvURL + "/v1/app-%s/messages:send"})

	_, err := hw.Send(context.Background(), &notify.Message{Target: "reg", Title: "t", Content: "c"})
	require.Error(t, err)
	require.False(t, errors.Is(err, ErrInvalidToken), "非 80200xxx 段的厂商码是投递失败，不是令牌失效")
	require.Contains(t, err.Error(), "80000001")
}

func TestHuaweiConfigIssues(t *testing.T) {
	hw := NewHuawei(nil, HuaweiConfig{Ref: "7"})
	_, err := hw.Send(context.Background(), &notify.Message{Target: "reg", Title: "t", Content: "c"})
	require.ErrorIs(t, err, notify.ErrNotConfigured, "缺凭据在拨号之前就该拦下")
}
