package netutil

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

// trustedV4CIDRs 测试用可信代理网段（私有地址）。
// 与 defaultTrustedProxyCIDRs 一致，这里显式设置以保证测试可重复。
func trustedV4CIDRs() []string {
	return []string{
		"10.0.0.0/8",
		"172.16.0.0/12",
		"192.168.0.0/16",
		"127.0.0.0/8",
	}
}

// buildReq 构造一个带指定 RemoteAddr 与 header 的 *http.Request。
func buildReq(remoteAddr string, headers map[string]string) *http.Request {
	req := httptest.NewRequest("GET", "/", nil)
	req.RemoteAddr = remoteAddr
	for k, v := range headers {
		req.Header.Set(k, v)
	}
	return req
}

// TestClientIP_NonTrustedPeerIgnoresXFF 对端非可信代理时，必须忽略 XFF，
// 直接返回对端 IP（防止攻击者伪造 XFF 头绕过基于 IP 的限流/审计）。
func TestClientIP_NonTrustedPeerIgnoresXFF(t *testing.T) {
	SetTrustedProxies(trustedV4CIDRs())
	// 对端 8.8.8.8 是公网 IP（非可信代理），即使携带伪造的 XFF 也应忽略
	req := buildReq("8.8.8.8:12345", map[string]string{
		"X-Forwarded-For": "10.0.0.1, 8.8.8.8",
		"X-Real-IP":       "127.0.0.1",
	})
	assert.Equal(t, "8.8.8.8", ClientIPFromRequest(req))
}

// TestClientIP_TrustedProxyParsesXFF 对端是可信代理时，从 XFF 链从右向左
// 跳过可信跳，取第一个非可信 IP 作为客户端真实 IP。
func TestClientIP_TrustedProxyParsesXFF(t *testing.T) {
	SetTrustedProxies(trustedV4CIDRs())
	// 对端 10.0.0.1 是可信代理；XFF 链 [8.8.8.8, 10.0.0.2]
	// 从右向左：10.0.0.2 是可信跳（跳过），8.8.8.8 是非可信 → 返回 8.8.8.8
	req := buildReq("10.0.0.1:12345", map[string]string{
		"X-Forwarded-For": "8.8.8.8, 10.0.0.2",
	})
	assert.Equal(t, "8.8.8.8", ClientIPFromRequest(req))
}

// TestClientIP_AllTrustedHopsReturnsLeftmost XFF 链全是可信跳时，
// 取最左侧 IP（按实现约定）。
func TestClientIP_AllTrustedHopsReturnsLeftmost(t *testing.T) {
	SetTrustedProxies(trustedV4CIDRs())
	// 对端 10.0.0.1 可信；XFF [10.0.0.2, 10.0.0.3] 全可信 → 返回最左侧 10.0.0.2
	req := buildReq("10.0.0.1:12345", map[string]string{
		"X-Forwarded-For": "10.0.0.2, 10.0.0.3",
	})
	assert.Equal(t, "10.0.0.2", ClientIPFromRequest(req))
}

// TestClientIP_NoXFFFallsBackToXRealIP 对端是可信代理且无 XFF 时，
// 采信 X-Real-IP（仅在可信代理前提下）。
func TestClientIP_NoXFFFallsBackToXRealIP(t *testing.T) {
	SetTrustedProxies(trustedV4CIDRs())
	req := buildReq("10.0.0.1:12345", map[string]string{
		"X-Real-IP": "203.0.113.5",
	})
	assert.Equal(t, "203.0.113.5", ClientIPFromRequest(req))
}

// TestClientIP_TrustedProxyNoHeadersReturnsPeer 对端是可信代理但无任何转发头，
// 兜底返回对端 IP（可信代理本身，如直连 nginx）。
func TestClientIP_TrustedProxyNoHeadersReturnsPeer(t *testing.T) {
	SetTrustedProxies(trustedV4CIDRs())
	req := buildReq("10.0.0.1:12345", nil)
	assert.Equal(t, "10.0.0.1", ClientIPFromRequest(req))
}

// TestClientIP_NonTrustedPeerNoHeadersReturnsPeer 对端非可信代理且无转发头，
// 返回对端 IP。
func TestClientIP_NonTrustedPeerNoHeadersReturnsPeer(t *testing.T) {
	SetTrustedProxies(trustedV4CIDRs())
	req := buildReq("114.114.114.114:12345", nil)
	assert.Equal(t, "114.114.114.114", ClientIPFromRequest(req))
}

// TestClientIP_NilRequest nil 请求返回空串。
func TestClientIP_NilRequest(t *testing.T) {
	assert.Equal(t, "", ClientIPFromRequest(nil))
}

// TestClientIP_EmptyRemoteAddr 空的 RemoteAddr 返回空串。
func TestClientIP_EmptyRemoteAddr(t *testing.T) {
	SetTrustedProxies(trustedV4CIDRs())
	req := buildReq("", nil)
	assert.Equal(t, "", ClientIPFromRequest(req))
}

// TestClientIP_XFFWithInvalidEntriesSkipped XFF 链中混入非法条目时，
// 解析应跳过非法项，仍从合法项中取第一个非可信 IP。
func TestClientIP_XFFWithInvalidEntriesSkipped(t *testing.T) {
	SetTrustedProxies(trustedV4CIDRs())
	// 链：["not-an-ip", "8.8.8.8", "10.0.0.2"]
	// 从右向左：10.0.0.2 可信跳（跳过）→ 8.8.8.8 非可信 → 返回 8.8.8.8
	// "not-an-ip" 因 net.ParseIP 失败被跳过
	req := buildReq("10.0.0.1:12345", map[string]string{
		"X-Forwarded-For": "not-an-ip, 8.8.8.8, 10.0.0.2",
	})
	assert.Equal(t, "8.8.8.8", ClientIPFromRequest(req))
}

// TestClientIP_NoTrustedProxiesAlwaysPeer 配置为不信任任何代理时，
// 无论是否有 XFF，都返回对端 IP。
func TestClientIP_NoTrustedProxiesAlwaysPeer(t *testing.T) {
	SetTrustedProxies([]string{}) // 空切片表示不信任任何代理
	req := buildReq("10.0.0.1:12345", map[string]string{
		"X-Forwarded-For": "8.8.8.8",
	})
	assert.Equal(t, "10.0.0.1", ClientIPFromRequest(req), "不信任任何代理时应返回对端 IP")
}
