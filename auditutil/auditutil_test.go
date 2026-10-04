package auditutil

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestIPFromRemoteAddr(t *testing.T) {
	assert.Equal(t, "::1", IPFromRemoteAddr("[::1]:7788"))
	assert.Equal(t, "127.0.0.1", IPFromRemoteAddr("127.0.0.1:7788"))
	assert.Equal(t, "127.0.0.1", IPFromRemoteAddr("127.0.0.1"))
	assert.Equal(t, "::1", IPFromRemoteAddr("::1"))
	assert.Equal(t, "127.0.0.1", IPFromRemoteAddr("127.0.0.1:12456"))
	assert.Equal(t, "192.0.2.1", IPFromRemoteAddr("192.0.2.1:5566"))
	assert.Equal(t, "2001:db8::68", IPFromRemoteAddr("2001:db8::68"))

	assert.Equal(t, "", IPFromRemoteAddr("192.0.2"))
}

func TestClientRealIP(t *testing.T) {
	build := func(remote string, headers map[string]string) *http.Request {
		req := httptest.NewRequest("GET", "/", nil)
		req.RemoteAddr = remote
		for k, v := range headers {
			req.Header.Set(k, v)
		}
		return req
	}

	t.Run("nil request", func(t *testing.T) {
		assert.Equal(t, "", ClientRealIP(nil))
	})

	t.Run("xff first valid ip wins", func(t *testing.T) {
		req := build("10.0.0.1:1000", map[string]string{
			"X-Forwarded-For": "8.8.8.8, 10.0.0.2",
		})
		assert.Equal(t, "8.8.8.8", ClientRealIP(req))
	})

	t.Run("xff invalid entries skipped", func(t *testing.T) {
		req := build("10.0.0.1:1000", map[string]string{
			"X-Forwarded-For": "garbage, 8.8.8.8",
		})
		assert.Equal(t, "8.8.8.8", ClientRealIP(req))
	})

	t.Run("x-real-ip fallback", func(t *testing.T) {
		req := build("10.0.0.1:1000", map[string]string{
			"X-Real-IP": "203.0.113.5",
		})
		assert.Equal(t, "203.0.113.5", ClientRealIP(req))
	})

	t.Run("remote addr fallback", func(t *testing.T) {
		req := build("192.0.2.1:5566", nil)
		assert.Equal(t, "192.0.2.1", ClientRealIP(req))
	})
}

func TestIsPrivateIP(t *testing.T) {
	for _, ip := range []string{"10.1.2.3", "172.16.0.9", "192.168.1.1", "127.0.0.1", "169.254.1.1", "::1", "fc00::1"} {
		assert.True(t, IsPrivateIP(ip), "%s 应判定为内网", ip)
	}
	for _, ip := range []string{"8.8.8.8", "192.0.2.1", "not-an-ip", ""} {
		assert.False(t, IsPrivateIP(ip), "%s 不应判定为内网", ip)
	}
}

func TestRequestID(t *testing.T) {
	assert.Equal(t, "", RequestID(nil))

	req := httptest.NewRequest("GET", "/", nil)
	req.Header.Set("X-Request-ID", "req-1")
	assert.Equal(t, "req-1", RequestID(req))

	req.Header.Del("X-Request-ID")
	req.Header.Set("X-Correlation-ID", "corr-1")
	assert.Equal(t, "corr-1", RequestID(req))

	req.Header.Del("X-Correlation-ID")
	req.Header.Set("x-fc-request-id", "fc-1")
	assert.Equal(t, "fc-1", RequestID(req))

	req.Header.Del("x-fc-request-id")
	generated := RequestID(req)
	assert.Regexp(t, `^[0-9a-f]{32}$`, generated, "无请求头时应生成 GUID")
}

func TestParseUsernameFromBytes(t *testing.T) {
	t.Run("json body", func(t *testing.T) {
		u, err := ParseUsernameFromBytes([]byte(`{"username":"alice","password":"x"}`))
		require.NoError(t, err)
		assert.Equal(t, "alice", u)
	})

	t.Run("form body", func(t *testing.T) {
		u, err := ParseUsernameFromBytes([]byte(`username=bob&password=x`))
		require.NoError(t, err)
		assert.Equal(t, "bob", u)
	})

	t.Run("line break injection stripped", func(t *testing.T) {
		u, err := ParseUsernameFromBytes([]byte(`{"username":"evil\r\nINJECTED"}`))
		require.NoError(t, err)
		assert.False(t, strings.ContainsAny(u, "\r\n"), "CR/LF 必须被剥离")
	})

	t.Run("missing username", func(t *testing.T) {
		_, err := ParseUsernameFromBytes([]byte(`{"other":"x"}`))
		assert.Error(t, err)
	})
}

func TestExtractUsernameFromRequest(t *testing.T) {
	req := httptest.NewRequest("POST", "/login", strings.NewReader(`{"username":"carol"}`))
	u, err := ExtractUsernameFromRequest(req)
	require.NoError(t, err)
	assert.Equal(t, "carol", u)

	// Body 必须已恢复，可重复读取
	body, err := io.ReadAll(req.Body)
	require.NoError(t, err)
	assert.Contains(t, string(body), "carol")
}

func TestDetectPlatformFromUA(t *testing.T) {
	cases := map[string]string{
		"":                                        PlatformOther,
		"okhttp/4.9.0":                            PlatformAndroidApp,
		"Mozilla/5.0 (Linux; Android 10; wv)":     PlatformAndroidApp,
		"com.example.app/1.0 (Android)":           PlatformAndroidApp,
		"MyApp/1.0 CFNetwork/1206 Darwin/20.0.0":  PlatformiOSApp,
		"Mozilla/5.0 (iPhone; CPU iPhone OS 15_0": PlatformiOSApp,
		"Mozilla/5.0 (Windows NT 10.0) Electron/9.0": PlatformDesktopWindows,
		"Mozilla/5.0 (Macintosh) Electron/9.0":       PlatformDesktopMac,
		"Mozilla/5.0 (X11; Linux) Electron/9.0":      PlatformDesktopLinux,
		"Mozilla/5.0 (Windows NT 10.0; Win64; x64) Chrome/120.0": PlatformWeb,
	}
	for ua, want := range cases {
		assert.Equal(t, want, DetectPlatformFromUA(ua), "UA %q", ua)
	}
}
