// Send 全链路的 httptest 测试：真发一次 HTTP 才暴露得出的几件事。
//
// 本机 httptest 的地址是 127.0.0.1，正好落在 SSRF 拦网表里，因此"发得出去"的用例
// 用 NewGuardedClient(true) 放行；"拦得住"的用例用生产默认的 guarded client。
package webhook

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/tx7do/go-utils/notify"
)

// recorder 起一个记录请求的本地服务，响应状态与正文由参数定。
type recorder struct {
	server *httptest.Server

	hit    int
	url    *url.URL
	header http.Header
	body   string
}

func receiver(t *testing.T, status int, responseBody string) *recorder {
	t.Helper()

	rec := &recorder{}
	rec.server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		raw, _ := io.ReadAll(r.Body)
		rec.hit++
		rec.url = r.URL
		rec.header = r.Header.Clone()
		rec.body = string(raw)
		w.WriteHeader(status)
		_, _ = io.WriteString(w, responseBody)
	}))
	t.Cleanup(rec.server.Close)

	return rec
}

func TestSendDeliversSignedDefaultBody(t *testing.T) {
	const secret = "s3cr3t-key"

	rec := receiver(t, http.StatusNoContent, "")
	svc := New(notify.NewGuardedClient(true), Config{
		Ref:       "7",
		Secret:    secret,
		SignStyle: "CUSTOM",
	})

	receipt, err := svc.Send(context.Background(), &notify.Message{
		Target:  rec.server.URL,
		Title:   "重置验证码",
		Content: "您的验证码是 123456",
	})
	require.NoError(t, err)
	require.NotNil(t, receipt)

	require.Equal(t, 1, rec.hit)
	require.Equal(t, "application/json; charset=utf-8", rec.header.Get("Content-Type"))

	var got map[string]any
	require.NoError(t, json.Unmarshal([]byte(rec.body), &got))
	require.Equal(t, "重置验证码", got["title"])
	require.Equal(t, "您的验证码是 123456", got["content"])
	require.Contains(t, got, "delivered_at")

	ts := rec.header.Get(HeaderTimestamp)
	require.NotEmpty(t, ts)
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte(ts))
	mac.Write([]byte("."))
	mac.Write([]byte(rec.body))
	require.Equal(t, "sha256="+hex.EncodeToString(mac.Sum(nil)), rec.header.Get(HeaderSignature),
		"签名必须覆盖时间戳：只签 body 的话重放是逐字节免费的")
}

func TestSendNoSecretHeaderWhenUnconfigured(t *testing.T) {
	rec := receiver(t, http.StatusOK, `{"ok":true}`)
	svc := New(notify.NewGuardedClient(true), Config{SignStyle: "CUSTOM"})

	_, err := svc.Send(context.Background(), &notify.Message{Target: rec.server.URL, Title: "t", Content: "c"})
	require.NoError(t, err)
	require.Empty(t, rec.header.Get(HeaderSignature))
	require.Empty(t, rec.header.Get(HeaderTimestamp))
}

// TestSendBlockedTargetIsConfigClass SSRF 防线：生产默认（guarded client）下 127.0.0.1
// 一次都拨不出去，并且归"没尝试过投递"（ErrNotConfigured），同时 ErrBlockedTarget 可沿链 Is 到。
func TestSendBlockedTargetIsConfigClass(t *testing.T) {
	rec := receiver(t, http.StatusOK, "")
	svc := New(notify.NewGuardedClient(false), Config{Ref: "7"})

	_, err := svc.Send(context.Background(), &notify.Message{Target: rec.server.URL, Title: "t", Content: "c"})
	require.Error(t, err)
	require.Equal(t, 0, rec.hit, "对端压根没收到请求才算拦住了")
	require.ErrorIs(t, err, notify.ErrNotConfigured)
	require.ErrorIs(t, err, notify.ErrBlockedTarget)
	require.Contains(t, err.Error(), "127.0.0.1")
}

func TestSendPeerErrorIsTransportClass(t *testing.T) {
	rec := receiver(t, http.StatusInternalServerError, "downstream queue is full")
	svc := New(notify.NewGuardedClient(true), Config{Ref: "7"})

	_, err := svc.Send(context.Background(), &notify.Message{Target: rec.server.URL, Title: "t", Content: "c"})
	require.Error(t, err)
	require.False(t, errors.Is(err, notify.ErrNotConfigured), "发出去了但被拒：属投递失败")
	require.Contains(t, err.Error(), "500")
	require.Contains(t, err.Error(), "downstream queue is full")
}

// TestSendRedirectIsNotFollowed 3xx 一律当失败：跟随 Location 等于让对端决定这一发打到哪。
func TestSendRedirectIsNotFollowed(t *testing.T) {
	hit := receiver(t, http.StatusOK, "")

	redirect := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, hit.server.URL, http.StatusFound)
	}))
	t.Cleanup(redirect.Close)

	svc := New(notify.NewGuardedClient(true), Config{})
	_, err := svc.Send(context.Background(), &notify.Message{Target: redirect.URL, Title: "t", Content: "c"})
	require.Error(t, err)
	require.Equal(t, 0, hit.hit, "重定向的目标一次都不该被命中")
	require.Contains(t, err.Error(), "302")
}

// TestSendPeerRejectsWithin200 三家机器人都会在 HTTP 200 里报失败
// （钉钉被安全策略拦下时给的就是 200 + errcode 310000）。只看状态码会把"对方没收"记成成功；
// 同时它仍是"拨过号被拒"，不是 ErrNotConfigured。
func TestSendPeerRejectsWithin200(t *testing.T) {
	cases := []struct {
		name      string
		style     string
		answer    string
		wantCause string
	}{
		{"dingtalk rejected", "DINGTALK", `{"errcode":310000,"errmsg":"keywords not in content"}`, "keywords not in content"},
		{"dingtalk ok", "DINGTALK", `{"errcode":0,"errmsg":"ok"}`, ""},
		{"wecom rejected", "WECOM", `{"errcode":93000,"errmsg":"invalid webhook k"}`, "invalid webhook k"},
		{"feishu rejected", "FEISHU", `{"code":19021,"msg":"sign match fail"}`, "sign match fail"},
		{"feishu ok", "FEISHU", `{"code":0,"msg":"success"}`, ""},
		// 判据字段各认各的：钉钉看到 code 不动作。
		{"field mismatch is not a verdict", "DINGTALK", `{"code":19021}`, ""},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			rec := receiver(t, http.StatusOK, c.answer)
			svc := New(notify.NewGuardedClient(true), Config{Ref: "7", Secret: "k", SignStyle: c.style})

			_, err := svc.Send(context.Background(), &notify.Message{Target: rec.server.URL, Title: "t", Content: "c"})
			require.Equal(t, 1, rec.hit)

			if c.wantCause == "" {
				require.NoError(t, err)
				return
			}
			require.Error(t, err)
			require.Contains(t, err.Error(), c.wantCause)
			require.Contains(t, err.Error(), "peer answered 200")
			require.False(t, errors.Is(err, notify.ErrNotConfigured), "拨过号且被对端拒 = 投递失败")
		})
	}
}

// TestSendBrokenTemplateNeverDials 模板渲染不出合法 JSON：错在出站之前，一次都不拨号。
func TestSendBrokenTemplateNeverDials(t *testing.T) {
	rec := receiver(t, http.StatusOK, "")
	svc := New(notify.NewGuardedClient(true), Config{Ref: "7", SignStyle: "CUSTOM", PayloadTemplate: `通知：{{title}}`})

	_, err := svc.Send(context.Background(), &notify.Message{Target: rec.server.URL, Title: "t", Content: "c"})
	require.Error(t, err)
	require.ErrorIs(t, err, notify.ErrNotConfigured, "配置错重试也不会变好")
	require.Contains(t, err.Error(), "channel [7]")
	require.Equal(t, 0, rec.hit, "渲染失败发生在拨号之前")

	require.ErrorIs(t, svc.Precheck(), notify.ErrNotConfigured, "预检要提前撞出同一个错")
}

// TestSendDingtalkSignsUrlQuery 钉钉签名在真发一次出站里的落点：
// 进 URL query（13 位毫秒、与时间戳成对）、且不能吃掉地址原有的 access_token；
// 正文是钉钉的 text 形状，自定义签名头一个都不发（钉钉只认 query）。
func TestSendDingtalkSignsUrlQuery(t *testing.T) {
	const secret = "ding-secret"

	rec := receiver(t, http.StatusOK, `{"errcode":0,"errmsg":"ok"}`)
	svc := New(notify.NewGuardedClient(true), Config{Secret: secret, SignStyle: "DINGTALK"})

	_, err := svc.Send(context.Background(), &notify.Message{
		Target: rec.server.URL + "/?access_token=abc123", Title: "t", Content: "c"})
	require.NoError(t, err)
	require.Equal(t, 1, rec.hit)

	q := rec.url.Query()
	require.Equal(t, "abc123", q.Get("access_token"), "拼签名必须保留地址原有的参数，否则钉钉先报 token 无效")

	ts := q.Get("timestamp")
	require.Len(t, ts, 13, "钉钉要 13 位毫秒，10 位秒会被判成 timestamp outside validity period")

	// 用 stdlib 独立复算，不调生产里的 signDingtalk：否则生产函数算错了这条断言也照样绿。
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte(ts + "\n" + secret))
	require.Equal(t, base64.StdEncoding.EncodeToString(mac.Sum(nil)), q.Get("sign"))

	require.Empty(t, rec.header.Get(HeaderSignature))
	require.Empty(t, rec.header.Get(HeaderTimestamp))

	require.Contains(t, rec.body, `"msgtype":"text"`)
	require.NotContains(t, rec.body, "sign", "钉钉的签名不进正文")
}

// TestSendEmptyTarget 目标为空是调用方的问题，不是配置问题。
func TestSendEmptyTarget(t *testing.T) {
	svc := New(nil, Config{})
	_, err := svc.Send(context.Background(), &notify.Message{Title: "t"})
	require.Error(t, err)
	require.False(t, errors.Is(err, notify.ErrNotConfigured))
}

// TestNormalizeURL 只有绝对 http(s) 地址放行。
func TestNormalizeURL(t *testing.T) {
	for _, bad := range []string{"", "/hooks/1", "//127.0.0.1/x", "ftp://example.com/x", "not a url"} {
		_, err := NormalizeURL(bad)
		require.Error(t, err, "%q 不该放行", bad)
	}

	for _, ok := range []string{"https://example.com/hook", "http://example.com:8080/h?a=1"} {
		got, err := NormalizeURL(ok)
		require.NoError(t, err, "%q 应放行", ok)
		require.Equal(t, ok, got)
	}
}
