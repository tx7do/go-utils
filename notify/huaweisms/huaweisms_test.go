package huaweisms

import (
	"context"
	"crypto/sha256"
	"encoding/base64"
	"net/http"
	"net/http/httptest"
	"regexp"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/tx7do/go-utils/notify"
)

func TestSendDeliversWSSEAndTemplateForm(t *testing.T) {
	var (
		wsse          string
		authorization string
		form          string
	)
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, "/sms/batchSendSms/v1", r.URL.Path, "官方单发也走批量路径")
		wsse = r.Header.Get("X-WSSE")
		authorization = r.Header.Get("Authorization")
		require.NoError(t, r.ParseForm())
		form = r.PostForm.Encode()
		_, _ = w.Write([]byte(`{"code":"000000","description":"Success","result":[{"smsMsgId":"d6e3cdd0","status":"000000","originTo":"+8613800138000"}]}`))
	}))
	t.Cleanup(srv.Close)

	svc := New(srv.Client(), Config{Ref: "7", AppKey: "app-key-1", AppSecret: "secret-1",
		Sender: "csms100000001", TemplateID: "tpl-1", APIBase: srv.URL})
	receipt, err := svc.Send(context.Background(), &notify.Message{
		Target: "13800138000",
		Params: map[string]string{"code": "123456"},
	})
	require.NoError(t, err)
	require.Equal(t, "d6e3cdd0", receipt.ProviderID, "smsMsgId 要进回执")

	require.Equal(t, `WSSE realm="SDP",profile="UsernameToken",type="Appkey"`, authorization)

	// X-WSSE 结构钉住，并用 PasswordDigest 的算法独立复算一遍。
	wsseRe := regexp.MustCompile(`UsernameToken Username="([^"]+)", PasswordDigest="([^"]+)", Nonce="([^"]+)", Created="([^"]+)"`)
	m := wsseRe.FindStringSubmatch(wsse)
	require.NotNil(t, m, "X-WSSE 形态不符: %s", wsse)
	require.Equal(t, "app-key-1", m[1])
	sum := sha256.Sum256([]byte(m[3] + m[4] + "secret-1"))
	require.Equal(t, base64.StdEncoding.EncodeToString(sum[:]), m[2],
		"PasswordDigest 必须等于 base64(sha256(nonce+created+secret))，三段原始串直连")
	require.Regexp(t, `^\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2}Z$`, m[4], "created 是 UTC 秒级")

	require.Contains(t, form, "from=csms100000001", "from 是签名通道号不是签名名称")
	require.Contains(t, form, "to=%2B8613800138000", "手机号 E.164 出站")
	require.Contains(t, form, "templateId=tpl-1")
	require.Contains(t, form, "templateParas=%5B%22123456%22%5D", "模板变量是 JSON 数组串")
}

func TestSendPeerErrorCode(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(`{"code":"E200028","description":"Template variable count is inconsistent"}`))
	}))
	t.Cleanup(srv.Close)

	svc := New(srv.Client(), Config{AppKey: "k", AppSecret: "s", Sender: "csms1", TemplateID: "t", APIBase: srv.URL})
	_, err := svc.Send(context.Background(), &notify.Message{Target: "13800138000", Params: map[string]string{"code": "1"}})
	require.Error(t, err)
	require.Contains(t, err.Error(), "code=E200028")
}

func TestSendConfigAndTargetIssues(t *testing.T) {
	svc := New(nil, Config{Ref: "7", AppKey: "k", AppSecret: "s", Sender: "csms1"})
	_, err := svc.Send(context.Background(), &notify.Message{Target: "13800138000", Params: map[string]string{"code": "1"}})
	require.ErrorIs(t, err, notify.ErrNotConfigured, "缺 template_id 在拨号之前就该拦下")

	_, err = svc.Send(context.Background(), &notify.Message{Target: "not-a-phone", Params: map[string]string{"code": "1"}})
	require.ErrorIs(t, err, notify.ErrNotConfigured, "号码不合法一次都不碰对端")
}
