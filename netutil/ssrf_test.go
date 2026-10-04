package netutil

import (
	"context"
	"net"
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestIsBlockedIP 私有/保留/回环/链路本地等地址必须被拦截。
func TestIsBlockedIP(t *testing.T) {
	blocked := []string{
		// IPv4 私有
		"10.0.0.1",
		"10.255.255.255",
		"172.16.0.1",
		"172.31.255.255",
		"192.168.1.1",
		"192.168.0.0",
		// 回环
		"127.0.0.1",
		"127.255.255.255",
		// 链路本地（含云元数据 169.254.169.254）
		"169.254.169.254",
		"169.254.0.1",
		// CGNAT
		"100.64.0.1",
		// 未指定 / 本网络
		"0.0.0.0",
		"0.1.2.3",
		// 文档/测试网段
		"192.0.2.1",
		"198.51.100.1",
		"203.0.113.1",
		// 多播 / 保留
		"224.0.0.1",
		"240.0.0.1",
		// IPv6
		"::1",
		"::",
		"fe80::1",
		"fc00::1",
		"fd00::1",
		"ff00::1",
		"2001:db8::1",
	}
	for _, ip := range blocked {
		t.Run("blocked/"+ip, func(t *testing.T) {
			assert.True(t, IsBlockedIP(ip), "期望 %s 被拦截", ip)
		})
	}
}

// TestIsBlockedIP_PublicAddr 公网地址不应被拦截。
func TestIsBlockedIP_PublicAddr(t *testing.T) {
	public := []string{
		"8.8.8.8",
		"1.1.1.1",
		"114.114.114.114",
		"2606:4700:4700::1111", // Cloudflare IPv6
	}
	for _, ip := range public {
		t.Run("public/"+ip, func(t *testing.T) {
			assert.False(t, IsBlockedIP(ip), "公网地址 %s 不应被拦截", ip)
		})
	}
}

// TestIsBlockedIP_InvalidAddr 非法 IP 字符串应保守判定为拦截。
func TestIsBlockedIP_InvalidAddr(t *testing.T) {
	invalid := []string{
		"",
		"not-an-ip",
		"999.999.999.999",
		"   ",
		"192.168.1.1.1",
	}
	for _, ip := range invalid {
		t.Run("invalid/"+ip, func(t *testing.T) {
			assert.True(t, IsBlockedIP(ip), "非法输入 %q 应判定为拦截", ip)
		})
	}
}

// TestIsBlockedIPAddr_NilInput nil 入参应判定为拦截。
func TestIsBlockedIPAddr_NilInput(t *testing.T) {
	assert.True(t, IsBlockedIPAddr(nil), "nil IP 应判定为拦截")
}

// TestIsBlockedIPAddr_CoversPrivate 确认 net.IP 形式与字符串形式判定一致。
func TestIsBlockedIPAddr_CoversPrivate(t *testing.T) {
	cases := []string{"10.1.2.3", "127.0.0.1", "169.254.169.254", "192.168.0.1"}
	for _, s := range cases {
		t.Run(s, func(t *testing.T) {
			ip := net.ParseIP(s)
			assert.NotNil(t, ip)
			assert.True(t, IsBlockedIPAddr(ip), "%s 应被拦截", s)
		})
	}
}

// TestLookupAndCheckHost_IPLiteral IP 字面量形式的 host 校验。
func TestLookupAndCheckHost_IPLiteral(t *testing.T) {
	t.Run("public IP literal", func(t *testing.T) {
		ips, err := LookupAndCheckHost(context.Background(), "8.8.8.8")
		assert.NoError(t, err)
		assert.Len(t, ips, 1)
		assert.Equal(t, net.IPv4(8, 8, 8, 8), ips[0])
	})
	t.Run("blocked IP literal", func(t *testing.T) {
		_, err := LookupAndCheckHost(context.Background(), "169.254.169.254")
		assert.Error(t, err, "链路本地地址应被拦截")
	})
	t.Run("blocked IP with port", func(t *testing.T) {
		_, err := LookupAndCheckHost(context.Background(), "127.0.0.1:8080")
		assert.Error(t, err, "回环地址（带端口）应被拦截")
	})
	t.Run("loopback IP", func(t *testing.T) {
		_, err := LookupAndCheckHost(context.Background(), "127.0.0.1")
		assert.Error(t, err)
	})
}

// TestLookupAndCheckHost_Unresolvable 无法解析的域名应返回错误。
func TestLookupAndCheckHost_Unresolvable(t *testing.T) {
	_, err := LookupAndCheckHost(context.Background(), "nonexistent.invalid")
	assert.Error(t, err)
}
