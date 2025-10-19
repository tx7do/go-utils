package geolite

import (
	"net"
)

// IsPrivateIP 判断 IP（支持 IPv4 和 IPv6）是否为内网地址
func IsPrivateIP(ip net.IP) bool {
	if ip == nil {
		return false
	}

	// 定义所有内网网段（包含 IPv4 和 IPv6）
	privateNetworks := []string{
		// IPv4 内网地址
		"10.0.0.0/8",
		"172.16.0.0/12",
		"192.168.0.0/16",
		"127.0.0.0/8",    // 回环地址
		"169.254.0.0/16", // 链路本地地址

		// IPv6 内网地址
		"fc00::/7",  // 唯一本地地址（ULA）
		"fe80::/10", // 链路本地地址
		"::1/128",   // 回环地址
		"fec0::/10", // 站点本地地址（已废弃，兼容保留）
	}

	// 检查 IP 是否属于任一内网网段
	for _, cidr := range privateNetworks {
		_, network, err := net.ParseCIDR(cidr)
		if err != nil {
			continue // 忽略无效 CIDR（理论上不会出现）
		}
		if network.Contains(ip) {
			return true
		}
	}
	return false
}
