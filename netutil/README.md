# netutil

网络相关的安全工具：

- **SSRF 防护**（`ssrf.go`）：私有/回环/链路本地/元数据网段封禁（`IsBlockedIP`）、
  URL 静态校验（`ValidateURL`：仅 http/https、拒绝 userinfo）、
  域名解析校验（`LookupAndCheckHost`：任一解析结果落禁投网段即报错；
  建连时请用返回的 IP 拨号以规避 DNS rebinding）。
- **客户端真实 IP 解析**（`client_ip.go`）：`ClientIPFromRequest` 按可信代理链
  从 `X-Forwarded-For` 从右向左解析，防 XFF 伪造绕过限流/审计。
  默认信任 RFC1918 + 回环网段，部署时用 `SetTrustedProxies` 指定实际反代网段。

```go
import "github.com/tx7do/go-utils/netutil"

netutil.SetTrustedProxies([]string{"10.0.0.0/8"})
ip := netutil.ClientIPFromRequest(req)

u, err := netutil.ValidateURL(rawURL)
ips, err := netutil.LookupAndCheckHost(ctx, u.Host)
```
