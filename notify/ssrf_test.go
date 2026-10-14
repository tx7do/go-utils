// SSRF 拦网表的逐点验证。
//
// 172.15 与 100.63 是刻意放的"紧邻边界外"取值：/12 与 /10 写错一位就会把它们一起拦掉，
// 而漏拦的代价（打内网）与误拦的代价（公网回调发不出去）都只能靠断言发现。
package notify

import (
	"net"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestIsBlockedIP(t *testing.T) {
	cases := []struct {
		ip         string
		wantReject bool
	}{
		{"127.0.0.1", true},
		{"10.0.0.5", true},
		{"172.16.0.1", true},
		{"192.168.1.1", true},
		{"100.64.0.1", true},      // CGNAT：Go 的 IsPrivate 认不出来
		{"169.254.169.254", true}, // 云厂商元数据端点
		{"0.0.0.0", true},
		{"::1", true},
		{"fd00::1", true}, // fc00::/7 下半段
		{"fe80::1", true},
		{"8.8.8.8", false},
		{"172.15.255.255", false},
		{"100.63.255.255", false},
		{"169.253.0.1", false},
	}

	for _, c := range cases {
		blocked, reason := IsBlockedIP(net.ParseIP(c.ip))
		require.Equal(t, c.wantReject, blocked, "%s", c.ip)
		if blocked {
			require.NotEmpty(t, reason, "拦下的必须给出原因，它要进错误文案")
		}
	}
}

func TestAllowPrivateTargetEnv(t *testing.T) {
	t.Setenv(EnvAllowPrivateOutbound, "")
	require.False(t, AllowPrivateTarget(), "空值不开防线缺口")

	t.Setenv(EnvAllowPrivateOutbound, "true")
	require.False(t, AllowPrivateTarget(), "真值判断会让垃圾值也放行，只认精确的 \"1\"")

	t.Setenv(EnvAllowPrivateOutbound, "1")
	require.True(t, AllowPrivateTarget())
}
