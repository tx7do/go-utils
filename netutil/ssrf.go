// Package netutil 提供网络相关的安全工具：SSRF 防护与防 XFF 伪造的客户端真实 IP 解析。
package netutil

import (
	"context"
	"fmt"
	"net"
	"net/url"
	"strings"
)

// blockedCIDRs 是禁止服务端访问的 IP 范围（私有/回环/链路本地/未指定/多播等）。
// 用于阻止 SSRF 攻击者借助服务端向内网或元数据服务（如 169.254.169.254）发起请求。
var blockedCIDRs = func() []*net.IPNet {
	cidrs := []string{
		// IPv4
		"0.0.0.0/8",       // 未指定/本网络
		"10.0.0.0/8",      // 私有 A
		"100.64.0.0/10",   // 运营商级 NAT（CGNAT）
		"127.0.0.0/8",     // 回环
		"169.254.0.0/16",  // 链路本地（含 AWS/Azure/IMDS 元数据 169.254.169.254）
		"172.16.0.0/12",   // 私有 B
		"192.0.0.0/24",    // IETF 协议分配
		"192.0.2.0/24",    // TEST-NET-1（文档用）
		"192.168.0.0/16",  // 私有 C
		"198.18.0.0/15",   // 基准测试
		"198.51.100.0/24", // TEST-NET-2
		"203.0.113.0/24",  // TEST-NET-3
		"224.0.0.0/4",     // 多播
		"240.0.0.0/4",     // 保留
		// IPv6
		"::1/128",       // 回环
		"fc00::/7",      // 唯一本地地址（ULA）
		"fe80::/10",     // 链路本地
		"ff00::/8",      // 多播
		"::/128",        // 未指定
		"2001:db8::/32", // 文档用
	}
	nets := make([]*net.IPNet, 0, len(cidrs))
	for _, c := range cidrs {
		_, n, err := net.ParseCIDR(c)
		if err != nil {
			continue
		}
		nets = append(nets, n)
	}
	return nets
}()

// IsBlockedIP 判断给定 IP 是否落入禁止访问的网段。
// 既接受 "10.1.2.3" 字符串，也接受 net.IP。
func IsBlockedIP(ipStr string) bool {
	ip := net.ParseIP(strings.TrimSpace(ipStr))
	if ip == nil {
		// 无法解析的 IP 视为非法，保守判定为禁止
		return true
	}
	return IsBlockedIPAddr(ip)
}

// IsBlockedIPAddr 同 IsBlockedIP，接受 net.IP 入参。
func IsBlockedIPAddr(ip net.IP) bool {
	if ip == nil {
		return true
	}
	if ip.IsUnspecified() || ip.IsInterfaceLocalMulticast() {
		return true
	}
	for _, n := range blockedCIDRs {
		if n.Contains(ip) {
			return true
		}
	}
	return false
}

// ValidateURL 对 SSRF 入口 URL 做静态校验：scheme 必须 http/https，host 非空。
func ValidateURL(rawURL string) (*url.URL, error) {
	if rawURL == "" {
		return nil, fmt.Errorf("empty url")
	}
	u, err := url.Parse(rawURL)
	if err != nil {
		return nil, fmt.Errorf("invalid url: %w", err)
	}
	if u.Scheme != "http" && u.Scheme != "https" {
		return nil, fmt.Errorf("unsupported scheme %q (only http/https allowed)", u.Scheme)
	}
	if u.Host == "" {
		return nil, fmt.Errorf("missing host")
	}
	// 拒绝携带用户信息（user:pass@host）的 URL，常见 SSRF 混淆手法
	if u.User != nil {
		return nil, fmt.Errorf("userinfo not allowed in url")
	}
	return u, nil
}

// LookupAndCheckHost 解析 hostname，返回所有解析到的 IP。
// 若任一 IP 落入禁止网段，返回错误（阻止 SSRF）。
// 注意：调用方在真正建连时，应使用解析得到的合法 IP 拨号（pin IP），
// 以避免 DNS rebinding 在校验通过后篡改解析结果。
func LookupAndCheckHost(ctx context.Context, host string) ([]net.IP, error) {
	// 先去掉可能存在的端口
	hostname := host
	if h, _, err := net.SplitHostPort(host); err == nil {
		hostname = h
	}

	// 如果 host 本身就是 IP 字面量，直接校验
	if ip := net.ParseIP(hostname); ip != nil {
		if IsBlockedIPAddr(ip) {
			return nil, fmt.Errorf("host %s is a blocked address", hostname)
		}
		return []net.IP{ip}, nil
	}

	ips, err := net.DefaultResolver.LookupIPAddr(ctx, hostname)
	if err != nil {
		return nil, fmt.Errorf("resolve host %q failed: %w", hostname, err)
	}
	result := make([]net.IP, 0, len(ips))
	for _, ia := range ips {
		if IsBlockedIPAddr(ia.IP) {
			return nil, fmt.Errorf("host %q resolves to blocked address %s", hostname, ia.IP)
		}
		result = append(result, ia.IP)
	}
	if len(result) == 0 {
		return nil, fmt.Errorf("host %q has no resolvable address", hostname)
	}
	return result, nil
}
