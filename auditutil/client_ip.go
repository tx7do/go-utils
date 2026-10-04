// Package auditutil 提供 HTTP 审计日志的信息采集与防篡改辅助：
//
//   - 客户端 IP / 请求 ID 等请求信息的归一化提取；
//   - 请求体中用户名的解析与日志行注入清洗；
//   - User-Agent 平台类型启发式识别；
//   - 审计日志的确定性哈希（排除 log_hash/signature 字段）与 ECDSA 签名
//     （DER 编码），构成"hash 链 + 签名"的防篡改约定。
package auditutil

import (
	"net"
	"net/http"
	"strings"
)

// 审计信息采集涉及的请求头名。
const (
	headerXForwardedFor  = "X-Forwarded-For"
	headerXRealIP        = "X-Real-IP"
	headerXRequestID     = "X-Request-ID"
	headerXCorrelationID = "X-Correlation-ID"
	headerXFcRequestID   = "x-fc-request-id"
)

// ClientRealIP 获取客户端真实 IP。
//
// 依次检查 X-Forwarded-For 与 X-Real-IP 头（取第一个合法 IP），都取不到时
// 回退 RemoteAddr。注意：这是"按转发头如实记录"的审计口径——最外层反向代理
// 必须覆写（而非追加）X-Forwarded-For 以防伪造。若需要防伪造的安全判定
// （可信代理链从右向左解析），应使用 netutil.ClientIPFromRequest。
func ClientRealIP(request *http.Request) string {
	if request == nil {
		return ""
	}

	if xff := request.Header.Get(headerXForwardedFor); xff != "" {
		// X-Forwarded-For 是逗号分隔的 IP 列表，取第一个合法 IP
		for _, ip := range strings.Split(xff, ",") {
			ip = strings.TrimSpace(ip)
			if net.ParseIP(ip) != nil {
				return ip
			}
		}
	}

	if xri := request.Header.Get(headerXRealIP); xri != "" {
		if net.ParseIP(xri) != nil {
			return xri
		}
	}

	return IPFromRemoteAddr(request.RemoteAddr)
}

// IPFromRemoteAddr 从 RemoteAddr（host:port）中提取合法 IP。
// 无法解析出 IP 时返回空串。
func IPFromRemoteAddr(hostAddress string) string {
	if strings.Contains(hostAddress, ":") {
		host, _, err := net.SplitHostPort(strings.TrimSpace(hostAddress))
		if err == nil {
			if net.ParseIP(host) != nil {
				return host
			}
		}
	}
	if net.ParseIP(hostAddress) != nil {
		return hostAddress
	}
	return ""
}

// IsPrivateIP 检查 IP 是否属于常见内网或链路本地地址。
// 无法解析的地址返回 false（由调用方决定如何处理未知形态）。
func IsPrivateIP(ipStr string) bool {
	ip := net.ParseIP(strings.TrimSpace(ipStr))
	if ip == nil {
		return false
	}
	cidrs := []string{
		"10.0.0.0/8",
		"172.16.0.0/12",
		"192.168.0.0/16",
		"127.0.0.0/8",
		"169.254.0.0/16",
		"::1/128",
		"fc00::/7",
	}
	for _, c := range cidrs {
		_, n, _ := net.ParseCIDR(c)
		if n.Contains(ip) {
			return true
		}
	}
	return false
}
