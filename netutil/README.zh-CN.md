# netutil

网络相关的安全工具包：SSRF（服务端请求伪造）防护与防 XFF 伪造的客户端真实 IP 解析。

[English](README.md) | 简体中文

## 特性

- **SSRF 防护**（`ssrf.go`）
  - 私有/回环/链路本地/云元数据等危险网段封禁：`IsBlockedIP` / `IsBlockedIPAddr`
  - URL 静态校验：`ValidateURL`（仅允许 http/https、host 非空、拒绝 `user:pass@host` 形式的 userinfo）
  - 域名解析校验：`LookupAndCheckHost`（任一解析结果落入禁投网段即报错，返回通过校验的 IP 列表）
- **客户端真实 IP 解析**（`client_ip.go`）
  - `ClientIPFromRequest` 按可信代理白名单从 `X-Forwarded-For` 链从右向左解析，防止攻击者伪造 XFF 头绕过基于 IP 的限流/审计
  - `SetTrustedProxies` 在启动时一次性指定实际反向代理网段，覆盖默认值

## 安装

netutil 是独立 Go 模块（`netutil/go.mod`）：

```bash
go get github.com/tx7do/go-utils/netutil
```

```go
import "github.com/tx7do/go-utils/netutil"
```

## 快速开始

### 1. 启动时设置可信代理（调用一次即可）

```go
package main

import "github.com/tx7do/go-utils/netutil"

func main() {
	// 用实际的反向代理网段覆盖默认值；放在 main / 服务初始化阶段，调用一次
	netutil.SetTrustedProxies([]string{"10.0.0.0/8", "192.168.1.2"})

	// 传入空切片表示不信任任何代理（始终使用 TCP 直连对端 IP）
	// netutil.SetTrustedProxies([]string{})
}
```

### 2. 在 HTTP 处理器中解析客户端真实 IP

```go
package main

import (
	"net/http"

	"github.com/tx7do/go-utils/netutil"
)

func handler(w http.ResponseWriter, r *http.Request) {
	clientIP := netutil.ClientIPFromRequest(r)
	_ = clientIP // 用于限流、审计等
}
```

### 3. SSRF 防护：校验 URL 并固定 IP 拨号（防 DNS rebinding）

只做"解析时校验"是不够的：攻击者可借助 DNS rebinding，让域名在校验时解析到公网 IP、
真正建连时解析到内网 IP。因此校验通过后，必须用返回的合法 IP 固定（pin）拨号：

```go
package main

import (
	"context"
	"fmt"
	"net"
	"syscall"

	"github.com/tx7do/go-utils/netutil"
)

// safeDial 校验入口 URL 后，用 net.Dialer.Control 在操作系统实际建连前
// 对目标地址做最终校验，并固定到已通过校验的 IP，防御 DNS rebinding。
func safeDial(ctx context.Context, rawURL string) (net.Conn, error) {
	// 1. 静态校验：仅 http/https、host 非空、拒绝 userinfo
	u, err := netutil.ValidateURL(rawURL)
	if err != nil {
		return nil, err
	}

	// 2. 解析并校验：任一解析结果落禁投网段即报错，ips 为通过校验的 IP 列表
	ips, err := netutil.LookupAndCheckHost(ctx, u.Host)
	if err != nil {
		return nil, err
	}

	// 3. 拆出主机与端口（缺省端口按 scheme 补全）
	host, port := u.Hostname(), u.Port()
	if port == "" {
		if u.Scheme == "https" {
			port = "443"
		} else {
			port = "80"
		}
	}

	// 4. Control 回调在真正建连前执行：此时 address 是操作系统拿到的实际地址。
	//    若 DNS 在校验后被篡改（rebinding），这里会拒绝连接。
	var d net.Dialer
	d.Control = func(network, address string, _ syscall.RawConn) error {
		h, _, err := net.SplitHostPort(address)
		if err != nil {
			return err
		}
		ip := net.ParseIP(h)
		if ip == nil || netutil.IsBlockedIPAddr(ip) {
			return fmt.Errorf("blocked address: %s", h)
		}
		for _, allowed := range ips {
			if allowed.Equal(ip) {
				return nil // 与校验阶段解析到的合法 IP 一致，放行
			}
		}
		return fmt.Errorf("address %s was not validated", h)
	}

	return d.DialContext(ctx, "tcp", net.JoinHostPort(host, port))
}
```

## API 参考

### SSRF 防护（`ssrf.go`）

| 函数 | 说明 |
|------|------|
| `IsBlockedIP(ipStr string) bool` | 判断 IP 字符串是否落入禁止访问的网段。无法解析的字符串保守判定为禁止（返回 `true`） |
| `IsBlockedIPAddr(ip net.IP) bool` | 同上，接受 `net.IP` 入参；`nil` 判定为禁止 |
| `ValidateURL(rawURL string) (*url.URL, error)` | 静态校验 SSRF 入口 URL：scheme 必须 http/https、host 非空、拒绝携带 userinfo（`user:pass@host`） |
| `LookupAndCheckHost(ctx context.Context, host string) ([]net.IP, error)` | 解析域名并校验：IP 字面量直接校验；域名经 `net.DefaultResolver` 解析后逐个校验，任一 IP 落禁投网段即返回错误；通过则返回全部合法 IP。host 允许带端口（自动剥离） |

被禁止的网段（`ssrf.go` 中 `blockedCIDRs`）：

- IPv4：`0.0.0.0/8`、`10.0.0.0/8`、`100.64.0.0/10`（CGNAT）、`127.0.0.0/8`、
  `169.254.0.0/16`（链路本地，含 AWS/Azure/IMDS 元数据 `169.254.169.254`）、
  `172.16.0.0/12`、`192.0.0.0/24`、`192.0.2.0/24`、`192.168.0.0/16`、
  `198.18.0.0/15`、`198.51.100.0/24`、`203.0.113.0/24`、`224.0.0.0/4`（多播）、`240.0.0.0/4`
- IPv6：`::1/128`、`fc00::/7`（ULA）、`fe80::/10`、`ff00::/8`、`::/128`、`2001:db8::/32`

另外 `IsBlockedIPAddr` 对未指定地址与接口本地多播地址（`ip.IsUnspecified()` /
`ip.IsInterfaceLocalMulticast()`）也直接判定为禁止。

### 客户端真实 IP（`client_ip.go`）

| 函数 | 说明 |
|------|------|
| `SetTrustedProxies(cidrs []string)` | 设置可信代理 CIDR 列表（覆盖默认值）。应在服务启动阶段调用一次。**空切片表示不信任任何代理**（始终使用 RemoteAddr）。不带掩码的单 IP 按 `/32`（IPv4）或 `/128`（IPv6）处理；非法 CIDR 会被跳过忽略 |
| `ClientIPFromRequest(req *http.Request) string` | 解析客户端真实 IP；`req` 为 `nil` 时返回 `""` |

#### 默认可信网段

未调用 `SetTrustedProxies` 时，默认信任以下网段（`client_ip.go` 的 `defaultTrustedProxyCIDRs`）：

- RFC1918 私有地址：`10.0.0.0/8`、`172.16.0.0/12`、`192.168.0.0/16`
- 回环：`127.0.0.0/8`、`::1/128`
- IPv6 唯一本地地址（ULA）：`fc00::/7`

生产部署应通过 `SetTrustedProxies` 精确指定实际反向代理网段。

#### ClientIPFromRequest 解析策略

1. 取 `RemoteAddr` 的直连对端 IP（TCP 连接的真实来源，无法伪造）。
2. 对端 IP 不是可信代理（或无法解析）→ 直接返回对端 IP，完全忽略 XFF。
3. 对端是可信代理 → 从 `X-Forwarded-For` 链**从右向左**跳过可信跳，第一个非可信
   IP 即为客户端真实 IP；链中无法解析的条目跳过；若整条链都是可信跳，取最左侧 IP。
4. XFF 为空或全部非法 → 才采信 `X-Real-IP`（前提仍是对端为可信代理）。
5. 兜底：对端 IP（如直连 nginx 的场景）。

## 注意事项

- `SetTrustedProxies` 应在服务启动阶段调用一次；`ClientIPFromRequest` 每次调用都会
  在首次使用时以默认值完成初始化（`sync.Once`），但生产环境不应依赖默认值。
- 始终部署在内网反代（nginx 等）之后时，将可信代理收敛为实际反代网段，避免内网
  任意主机被当作可信代理。
- `LookupAndCheckHost` 的校验结果与真正建连之间存在时间窗口，务必配合"固定 IP 拨号"
  （见快速开始第 3 例）使用，否则 DNS rebinding 可以绕过校验。
- SSRF 封禁策略是保守的"黑名单 + 保守判定"：无法解析的输入一律视为禁止。

## 参考资料

- [OWASP SSRF Prevention Cheat Sheet](https://cheatsheetseries.owasp.org/cheatsheets/Server_Side_Request_Forgery_Prevention_Cheat_Sheet.html)
- [Go net.Dialer.Control 文档](https://pkg.go.dev/net#Dialer)
