package push

import (
	"context"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/x509"
	"encoding/json"
	"encoding/pem"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/tx7do/go-utils/notify"
)

// newH2TLSServer httptest 的 TLS 服务端默认不开 HTTP/2（测试基架限制），
// 显式挂上 x/net/http2 让测试真正走 APNs 的协议。
func newH2TLSServer(t *testing.T, h http.HandlerFunc) *httptest.Server {
	t.Helper()
	srv := httptest.NewUnstartedServer(h)
	srv.EnableHTTP2 = true
	srv.StartTLS()
	t.Cleanup(srv.Close)
	return srv
}

// newTestAPNsKeyPEM 进程内生成一把 P-256 测试私钥的 PKCS#8 PEM（.p8 形态）。
func newTestAPNsKeyPEM(t *testing.T) (*ecdsa.PrivateKey, string) {
	t.Helper()
	key, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	require.NoError(t, err)
	der, err := x509.MarshalPKCS8PrivateKey(key)
	require.NoError(t, err)
	pemContent := string(pem.EncodeToMemory(&pem.Block{Type: "PRIVATE KEY", Bytes: der}))
	return key, pemContent
}

func TestNewAPNsParsesP8Key(t *testing.T) {
	_, pemContent := newTestAPNsKeyPEM(t)

	svc, err := NewAPNs(nil, APNsConfig{Ref: "7", TeamID: "TEAM", KeyID: "KEY-1",
		BundleID: "com.example.app", PrivateKey: pemContent})
	require.NoError(t, err)
	require.NotNil(t, svc.key, "p8 在构造期解析完成：私钥配错当场暴露，不等第一次投递")
	require.Equal(t, elliptic.P256(), svc.key.Curve)
	require.Equal(t, APNsProductionURL, svc.apiBase, "默认生产网关")

	sandbox, err := NewAPNs(nil, APNsConfig{TeamID: "T", KeyID: "K", BundleID: "b",
		PrivateKey: pemContent, Sandbox: true})
	require.NoError(t, err)
	require.Equal(t, APNsSandboxURL, sandbox.apiBase, "沙箱开关切沙箱网关：开发构建的令牌只认沙箱")

	_, err = NewAPNs(nil, APNsConfig{TeamID: "T", KeyID: "K", BundleID: "b", PrivateKey: "not a pem"})
	require.Error(t, err)
	require.ErrorIs(t, err, notify.ErrNotConfigured, "坏 PEM 是配置问题，归配置类哨兵")
}

// TestMintAPNsTokenStructure provider token 的结构：三段式 JWT，header 带 alg/kid，
// claims 是 iss(TeamID)+iat。签名正确性由 Send 的真投递用例（对端接受与否）兜底。
func TestMintAPNsTokenStructure(t *testing.T) {
	key, _ := newTestAPNsKeyPEM(t)

	tok, err := MintAPNsToken(key, "KEY-1", "TEAM-9")
	require.NoError(t, err)

	parts := strings.Split(tok, ".")
	require.Len(t, parts, 3)

	headerJSON, err := base64DecodeSegment(parts[0])
	require.NoError(t, err)
	var header map[string]any
	require.NoError(t, json.Unmarshal(headerJSON, &header))
	require.Equal(t, "ES256", header["alg"])
	require.Equal(t, "KEY-1", header["kid"])

	claimsJSON, err := base64DecodeSegment(parts[1])
	require.NoError(t, err)
	var claims map[string]any
	require.NoError(t, json.Unmarshal(claimsJSON, &claims))
	require.Equal(t, "TEAM-9", claims["iss"], "iss 是 TeamID：APNs 按它找账号")
	require.Contains(t, claims, "iat")
}

func TestAPNsSendDeliversOverH2(t *testing.T) {
	_, pemContent := newTestAPNsKeyPEM(t)

	var (
		hit      int
		proto    string
		auth     string
		topic    string
		pushType string
		priority string
		apnsID   = "apns-msg-1"
		payload  map[string]any
	)

	srv := newH2TLSServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		hit++
		proto = r.Proto
		auth = r.Header.Get("Authorization")
		topic = r.Header.Get("apns-topic")
		pushType = r.Header.Get("apns-push-type")
		priority = r.Header.Get("apns-priority")
		raw, _ := io.ReadAll(r.Body)
		require.NoError(t, json.Unmarshal(raw, &payload))
		w.Header().Set("apns-id", apnsID)
		w.WriteHeader(http.StatusOK)
	}))
	t.Cleanup(srv.Close)

	transport := srv.Client().Transport.(*http.Transport)
	transport.ForceAttemptHTTP2 = true // APNs 只讲 HTTP/2；httptest TLS 支持 ALPN h2

	svc, err := NewAPNs(&http.Client{Transport: transport},
		APNsConfig{TeamID: "TEAM", KeyID: "KEY-1", BundleID: "com.example.app",
			PrivateKey: pemContent, APIBase: srv.URL})
	require.NoError(t, err)

	receipt, err := svc.Send(context.Background(), &notify.Message{
		Target:  strings.Repeat("a", 64),
		Title:   "重置验证码",
		Content: "您的验证码是 123456",
		Params:  map[string]string{"msg_id": "7"},
	})
	require.NoError(t, err)
	require.Equal(t, 1, hit)
	require.Equal(t, "HTTP/2.0", proto, "APNs 只讲 HTTP/2：降级到 h1 说明 transport 配置错了")
	require.Equal(t, "com.example.app", topic)
	require.Equal(t, "alert", pushType)
	require.Equal(t, "10", priority)
	require.True(t, strings.HasPrefix(auth, "bearer ey"), "provider token 走 bearer 头")

	aps := payload["aps"].(map[string]any)
	require.Equal(t, map[string]any{"title": "重置验证码", "body": "您的验证码是 123456"}, aps["alert"])
	require.Equal(t, "default", aps["sound"])
	require.Equal(t, "7", payload["msg_id"], "Params 展开为顶层自定义键，客户端本地路由按需取")

	require.Equal(t, apnsID, receipt.ProviderID, "apns-id 是厂商侧消息 ID，进回执供对账")
}

// TestAPNsInvalidToken 令牌失效（应用卸载/系统重置）：410 或 reason=Unregistered/BadDeviceToken
// 都要归 ErrInvalidToken——调用方删行自清理，不算投递失败重试。
func TestAPNsInvalidToken(t *testing.T) {
	_, pemContent := newTestAPNsKeyPEM(t)

	cases := []struct {
		name   string
		status int
		reason string
	}{
		{"gone", http.StatusGone, "Unregistered"},
		{"bad device token on 400", http.StatusBadRequest, "BadDeviceToken"},
		{"unregistered on 400", http.StatusBadRequest, "Unregistered"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			srv := newH2TLSServer(t, http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
				w.Header().Set("apns-error-reason", tc.reason)
				w.WriteHeader(tc.status)
			}))
			t.Cleanup(srv.Close)

			transport := srv.Client().Transport.(*http.Transport)
			transport.ForceAttemptHTTP2 = true
			svc, err := NewAPNs(&http.Client{Transport: transport},
				APNsConfig{TeamID: "T", KeyID: "K", BundleID: "b", PrivateKey: pemContent, APIBase: srv.URL})
			require.NoError(t, err)

			_, err = svc.Send(context.Background(), &notify.Message{Target: "gone", Title: "t", Content: "c"})
			require.ErrorIs(t, err, ErrInvalidToken)
		})
	}
}

func TestAPNsPeerRejection(t *testing.T) {
	_, pemContent := newTestAPNsKeyPEM(t)

	srv := newH2TLSServer(t, http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("apns-error-reason", "PayloadTooLarge")
		w.WriteHeader(http.StatusRequestEntityTooLarge)
	}))
	t.Cleanup(srv.Close)

	transport := srv.Client().Transport.(*http.Transport)
	transport.ForceAttemptHTTP2 = true
	svc, err := NewAPNs(&http.Client{Transport: transport},
		APNsConfig{TeamID: "T", KeyID: "K", BundleID: "b", PrivateKey: pemContent, APIBase: srv.URL})
	require.NoError(t, err)

	_, err = svc.Send(context.Background(), &notify.Message{Target: "tok", Title: "t", Content: "c"})
	require.Error(t, err)
	require.False(t, errors.Is(err, ErrInvalidToken), "非失效类拒绝是投递失败，不是令牌失效")
	require.Contains(t, err.Error(), "PayloadTooLarge")
}
