package netutil

// 本文件锁定 fail-closed 默认语义与 env 解析入口：
// 未配置可信代理时，一切转发头（XFF / X-Real-IP）都不采信，
// 即使对端是内网地址——内网主机同样是潜在客户端，历史上
// "默认信任 RFC1918"正是 XFF 伪造绕过审计的入口。
//
// 另覆盖 ParseTrustedProxies（逗号串 → 合法条目）与
// SetTrustedProxiesFromEnvVar（env 注入）的解析规则。

import (
	"net"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestClientIP_FailClosedDefault 未配置可信代理时，内网对端携带
// 伪造 XFF / X-Real-IP 也必须返回对端 IP 本身。
func TestClientIP_FailClosedDefault(t *testing.T) {
	SetTrustedProxies(nil) // 显式清空：模拟"未配置可信代理"的默认态
	req := buildReq("10.0.0.1:12345", map[string]string{
		"X-Forwarded-For": "203.0.113.7",
		"X-Real-IP":       "203.0.113.9",
	})
	assert.Equal(t, "10.0.0.1", ClientIPFromRequest(req),
		"未配置可信代理时，内网对端也不得采信 XFF/X-Real-IP")
}

// TestSetTrustedProxies_NilClearsTrust 传 nil 与传空切片语义一致：清空。
func TestSetTrustedProxies_NilClearsTrust(t *testing.T) {
	t.Cleanup(func() { SetTrustedProxies(trustedV4CIDRs()) })

	SetTrustedProxies(trustedV4CIDRs())
	require.True(t, isTrustedProxy(net.ParseIP("10.0.0.1")))

	SetTrustedProxies(nil)
	assert.False(t, isTrustedProxy(net.ParseIP("10.0.0.1")),
		"传 nil 必须清空可信代理表（fail-closed）")
}

// TestParseTrustedProxies 逗号串解析：合法条目保留（裸 IP 补码），
// 非法条目跳过，全空/全非法返回 nil。
func TestParseTrustedProxies(t *testing.T) {
	t.Run("mixed valid and invalid", func(t *testing.T) {
		got := ParseTrustedProxies(" 10.0.0.0/8 , 8.8.8.8 , garbage , 999.1.1.1 , 2001:db8::1 ")
		assert.Equal(t, []string{"10.0.0.0/8", "8.8.8.8/32", "2001:db8::1/128"}, got)
	})
	t.Run("empty and blank", func(t *testing.T) {
		assert.Nil(t, ParseTrustedProxies(""))
		assert.Nil(t, ParseTrustedProxies("   "))
	})
	t.Run("all invalid", func(t *testing.T) {
		assert.Nil(t, ParseTrustedProxies("not-an-ip, 1.2.3.4/99, "))
	})
}

// TestSetTrustedProxiesFromEnvVar env 注入：合法条目生效；空 env 清空。
func TestSetTrustedProxiesFromEnvVar(t *testing.T) {
	t.Cleanup(func() { SetTrustedProxies(trustedV4CIDRs()) })

	t.Run("valid entries applied", func(t *testing.T) {
		t.Setenv("TEST_TRUSTED_PROXIES", "172.16.0.0/12,8.8.8.8")
		SetTrustedProxiesFromEnvVar("TEST_TRUSTED_PROXIES")
		assert.True(t, isTrustedProxy(net.ParseIP("172.16.5.5")))
		assert.True(t, isTrustedProxy(net.ParseIP("8.8.8.8")))
		assert.False(t, isTrustedProxy(net.ParseIP("10.0.0.1")),
			"未列入配置的网段不得成为可信代理")
	})

	t.Run("empty env clears trust", func(t *testing.T) {
		SetTrustedProxies(trustedV4CIDRs())
		require.True(t, isTrustedProxy(net.ParseIP("10.0.0.1")))

		t.Setenv("TEST_TRUSTED_PROXIES", "")
		SetTrustedProxiesFromEnvVar("TEST_TRUSTED_PROXIES")
		assert.False(t, isTrustedProxy(net.ParseIP("10.0.0.1")),
			"env 为空必须清空可信代理表（fail-closed）")
	})
}
