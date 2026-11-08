package netutil

import (
	"net"
	"net/http"
	"os"
	"strings"
	"sync"
)

// 客户端真实 IP 相关的请求头。
const (
	headerXForwardedFor = "X-Forwarded-For"
	headerXRealIP       = "X-Real-IP"
)

// ============================================================================
// 可信代理配置（fail-closed：防 XFF 伪造绕过限流/审计）
//
// 默认不信任任何代理：X-Forwarded-For / X-Real-IP 一律不采信，直接返回
// RemoteAddr——直连部署下客户端自带的 XFF 任何条目都不可信。仅当通过
// SetTrustedProxies / SetTrustedProxiesFromEnvVar 显式配置了可信代理链，
// 且请求的直连对端（RemoteAddr）本身属于可信代理时，才沿 XFF 从右向左
// 解析：条目只能由可信代理追加，第一个不属于可信代理的条目即客户端。
//
// 解析语义上不存在"安全的默认网段"：内网网段同样是潜在客户端，把 RFC1918
// 默认视为可信会让任意内网主机伪造 XFF（历史上这正是伪造绕过审计的入口）。
// ============================================================================

// trustedProxies 可信代理链（进程级，启动期经 SetTrustedProxies 配置一次）。
// 零值（未配置）= 不信任任何代理。
var (
	trustedProxiesMu sync.RWMutex
	trustedProxies   []*net.IPNet
)

// normalizeTrustedProxyEntry 归一单个可信代理条目：去首尾空白、跳过空串、
// 裸 IP 按单地址补码（IPv4 → /32，IPv6 → /128）。非法条目返回 ok=false。
func normalizeTrustedProxyEntry(entry string) (string, bool) {
	entry = strings.TrimSpace(entry)
	if entry == "" {
		return "", false
	}
	// 允许不带掩码的单 IP（按 /32 或 /128 处理）
	if !strings.Contains(entry, "/") {
		ip := net.ParseIP(entry)
		if ip == nil {
			return "", false
		}
		if ip.To4() != nil {
			entry += "/32"
		} else {
			entry += "/128"
		}
	}
	if _, _, err := net.ParseCIDR(entry); err != nil {
		return "", false
	}
	return entry, true
}

// ParseTrustedProxies 把逗号分隔的网段串（env 配置形态）解析为合法条目。
// 条目支持 CIDR 或裸 IP；解析失败的条目跳过；全空/全非法返回 nil。
func ParseTrustedProxies(raw string) []string {
	if strings.TrimSpace(raw) == "" {
		return nil
	}
	parts := strings.Split(raw, ",")
	out := make([]string, 0, len(parts))
	for _, part := range parts {
		if normalized, ok := normalizeTrustedProxyEntry(part); ok {
			out = append(out, normalized)
		}
	}
	if len(out) == 0 {
		return nil
	}
	return out
}

// SetTrustedProxies 设置可信代理 CIDR 列表，应在服务启动阶段（main）调用一次。
// 条目支持 CIDR 或裸 IP；非法条目跳过。
// 传入空切片或 nil 表示不信任任何代理（始终采用 RemoteAddr，fail-closed）。
func SetTrustedProxies(cidrs []string) {
	trustedProxiesMu.Lock()
	defer trustedProxiesMu.Unlock()
	trustedProxies = parseTrustedProxyCIDRs(cidrs)
}

// SetTrustedProxiesFromEnvVar 从指定环境变量读取逗号分隔的网段配置并生效。
// 变量未设置或为空时清空可信代理表（回到"不信任任何 XFF"）。
// 环境变量名由调用方传入（如应用侧自定义 GOWIND_TRUSTED_PROXIES）。
func SetTrustedProxiesFromEnvVar(name string) {
	SetTrustedProxies(ParseTrustedProxies(os.Getenv(name)))
}

func parseTrustedProxyCIDRs(cidrs []string) []*net.IPNet {
	nets := make([]*net.IPNet, 0, len(cidrs))
	for _, c := range cidrs {
		if normalized, ok := normalizeTrustedProxyEntry(c); ok {
			// normalize 已保证 CIDR 合法，ParseCIDR 不会失败
			_, n, _ := net.ParseCIDR(normalized)
			nets = append(nets, n)
		}
	}
	return nets
}

// isTrustedProxy 判断 IP 是否属于可信代理网段。未配置可信代理时恒为 false。
func isTrustedProxy(ip net.IP) bool {
	trustedProxiesMu.RLock()
	defer trustedProxiesMu.RUnlock()
	if len(trustedProxies) == 0 {
		return false
	}
	for _, n := range trustedProxies {
		if n.Contains(ip) {
			return true
		}
	}
	return false
}

// ClientIPFromRequest 从 *http.Request 解析客户端真实 IP。
//
// 解析策略（防 XFF 伪造，fail-closed）：
//  0. 未配置可信代理 → 直接返回对端 IP，完全忽略一切转发头。
//  1. 取 RemoteAddr 的直连对端 IP（TCP 连接的真实来源，无法伪造）。
//  2. 若对端 IP 不是可信代理（或无法解析）→ 直接返回对端 IP，完全忽略 XFF。
//  3. 若对端是可信代理 → 从 X-Forwarded-For 链**从右向左**跳过可信跳，
//     第一个非可信 IP 即为客户端；全部条目皆可信（XFF 耗尽）取最左条目。
//  4. XFF 为空或全部非法 → 采信 X-Real-IP（仍仅在可信代理前提下）。
//  5. 兜底：对端 IP（可信代理本身，如直连 nginx）。
//
// 注意：X-Real-IP 仅在对端为可信代理时才采信，且优先级低于 XFF 链解析。
func ClientIPFromRequest(req *http.Request) string {
	if req == nil {
		return ""
	}

	// 1. 直连对端 IP（TCP 连接的真实来源，无法伪造）
	peerIP := ipFromRemoteAddr(req.RemoteAddr)
	peer := net.ParseIP(peerIP)

	// 2. 对端非可信代理（含未配置可信代理的情形）→ 不信任任何转发头，直接用对端 IP
	if peer == nil || !isTrustedProxy(peer) {
		return peerIP
	}

	// 3. 对端是可信代理 → 从 XFF 链从右向左解析
	if xff := req.Header.Get(headerXForwardedFor); xff != "" {
		ips := splitAndTrim(xff)
		// 从右向左跳过可信跳，第一个非可信 IP 即客户端
		for i := len(ips) - 1; i >= 0; i-- {
			ipStr := ips[i]
			ip := net.ParseIP(ipStr)
			if ip == nil {
				continue
			}
			if !isTrustedProxy(ip) {
				return ipStr
			}
			// 若整条链都是可信跳，循环到最后会返回最左侧（i==0 时即使可信也返回）
			if i == 0 {
				return ipStr
			}
		}
	}

	// 4. XFF 为空或全部非法 → 采信 X-Real-IP（仅在可信代理前提下）
	if xri := req.Header.Get(headerXRealIP); xri != "" {
		if ip := net.ParseIP(strings.TrimSpace(xri)); ip != nil {
			return strings.TrimSpace(xri)
		}
	}

	// 5. 兜底：对端 IP（可信代理本身，如直连 nginx）
	return peerIP
}

// splitAndTrim 按逗号切分并去空格，过滤空串。
func splitAndTrim(s string) []string {
	parts := strings.Split(s, ",")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p != "" {
			out = append(out, p)
		}
	}
	return out
}

// ipFromRemoteAddr 从 RemoteAddr（host:port）中提取 host。
func ipFromRemoteAddr(remoteAddr string) string {
	if remoteAddr == "" {
		return ""
	}
	// 处理可能无端口的情形
	if !strings.Contains(remoteAddr, ":") {
		return remoteAddr
	}
	host, _, err := net.SplitHostPort(strings.TrimSpace(remoteAddr))
	if err != nil {
		return remoteAddr
	}
	if net.ParseIP(host) != nil {
		return host
	}
	return host
}
