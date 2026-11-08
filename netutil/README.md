# netutil

Network security utilities: SSRF (Server-Side Request Forgery) protection and
client real-IP resolution that resists `X-Forwarded-For` spoofing.

English | [简体中文](README.zh-CN.md)

## Features

- **SSRF protection** (`ssrf.go`)
  - Blocks private/loopback/link-local/cloud-metadata ranges: `IsBlockedIP` / `IsBlockedIPAddr`
  - Static URL validation: `ValidateURL` (http/https only, host required, rejects `user:pass@host` userinfo)
  - Hostname resolution checks: `LookupAndCheckHost` (errors out if any resolved address falls into a blocked range, returns the validated IP list)
- **Client real-IP resolution** (`client_ip.go`, fail-closed semantics)
  - By default **no proxy is trusted**: without explicit configuration, XFF / X-Real-IP are ignored entirely — even private-range peers (historically, "trust RFC1918 by default" was exactly the spoofing entry point that bypassed auditing)
  - `ClientIPFromRequest` walks the `X-Forwarded-For` chain right-to-left through trusted proxies only, so attackers cannot spoof XFF to bypass IP-based rate limiting or auditing
  - `SetTrustedProxies` / `SetTrustedProxiesFromEnvVar` configure your actual reverse-proxy ranges once at startup
- **Kratos context helpers** (`context.go`)
  - `ClientIPFromContext` / `HeaderFromContext` / `CookieFromContext`: read the client IP / request headers / HttpOnly cookies from a Kratos server context; non-HTTP transports return zero values instead of panicking

## Installation

netutil is an independent Go module (`netutil/go.mod`):

```bash
go get github.com/tx7do/go-utils/netutil
```

```go
import "github.com/tx7do/go-utils/netutil"
```

## Quick Start

### 1. Configure trusted proxies once at startup

```go
package main

import "github.com/tx7do/go-utils/netutil"

func main() {
	// No proxy is trusted by default; specify your actual reverse-proxy ranges;
	// call once in main / service initialization.
	netutil.SetTrustedProxies([]string{"10.0.0.0/8", "192.168.1.2"})

	// Or read a comma-separated list from an environment variable
	// (unset / empty means "trust no proxy").
	netutil.SetTrustedProxiesFromEnvVar("GOWIND_TRUSTED_PROXIES")

	// An empty slice / nil likewise means "trust no proxy" (always use the TCP peer IP).
	// netutil.SetTrustedProxies(nil)
}
```

### 2. Resolve the client IP in an HTTP handler

```go
package main

import (
	"net/http"

	"github.com/tx7do/go-utils/netutil"
)

func handler(w http.ResponseWriter, r *http.Request) {
	clientIP := netutil.ClientIPFromRequest(r)
	_ = clientIP // use for rate limiting, auditing, etc.
}
```

### 3. SSRF protection: validate the URL and pin the dial IP (anti DNS rebinding)

Validation at resolution time alone is not enough: with DNS rebinding an attacker
can make a hostname resolve to a public IP during the check and to an internal IP
when the connection is actually established. After validation, dial the validated
IPs (pin the IP) as shown below:

```go
package main

import (
	"context"
	"fmt"
	"net"
	"syscall"

	"github.com/tx7do/go-utils/netutil"
)

// safeDial validates the entry URL, then uses net.Dialer.Control to re-check
// the target address right before the OS connects, pinning the dial to the
// validated IPs to defend against DNS rebinding.
func safeDial(ctx context.Context, rawURL string) (net.Conn, error) {
	// 1. Static validation: http/https only, host required, no userinfo.
	u, err := netutil.ValidateURL(rawURL)
	if err != nil {
		return nil, err
	}

	// 2. Resolve and check: errors out if any resolved address is blocked;
	//    ips is the validated IP list.
	ips, err := netutil.LookupAndCheckHost(ctx, u.Host)
	if err != nil {
		return nil, err
	}

	// 3. Split host and port (default the port by scheme when absent).
	host, port := u.Hostname(), u.Port()
	if port == "" {
		if u.Scheme == "https" {
			port = "443"
		} else {
			port = "80"
		}
	}

	// 4. The Control callback runs before the actual connect; at that point
	//    address is the real address the OS is about to dial. If DNS was
	//    re-bound after validation, the connection is rejected here.
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
				return nil // matches a validated IP from step 2, allow
			}
		}
		return fmt.Errorf("address %s was not validated", h)
	}

	return d.DialContext(ctx, "tcp", net.JoinHostPort(host, port))
}
```

## API Reference

### SSRF protection (`ssrf.go`)

| Function | Description |
|----------|-------------|
| `IsBlockedIP(ipStr string) bool` | Reports whether the IP string falls into a blocked range. Unparseable input is conservatively treated as blocked (returns `true`) |
| `IsBlockedIPAddr(ip net.IP) bool` | Same as above for `net.IP`; `nil` is treated as blocked |
| `ValidateURL(rawURL string) (*url.URL, error)` | Statically validates an SSRF entry URL: scheme must be http/https, host required, userinfo (`user:pass@host`) rejected |
| `LookupAndCheckHost(ctx context.Context, host string) ([]net.IP, error)` | Resolves the host and checks every address: IP literals are checked directly; hostnames go through `net.DefaultResolver` and every result is checked — any blocked IP aborts with an error, otherwise all valid IPs are returned. `host` may include a port (it is stripped) |

Blocked ranges (`blockedCIDRs` in `ssrf.go`):

- IPv4: `0.0.0.0/8`, `10.0.0.0/8`, `100.64.0.0/10` (CGNAT), `127.0.0.0/8`,
  `169.254.0.0/16` (link-local, incl. AWS/Azure/IMDS metadata `169.254.169.254`),
  `172.16.0.0/12`, `192.0.0.0/24`, `192.0.2.0/24`, `192.168.0.0/16`,
  `198.18.0.0/15`, `198.51.100.0/24`, `203.0.113.0/24`, `224.0.0.0/4` (multicast), `240.0.0.0/4`
- IPv6: `::1/128`, `fc00::/7` (ULA), `fe80::/10`, `ff00::/8`, `::/128`, `2001:db8::/32`

In addition, `IsBlockedIPAddr` treats unspecified and interface-local multicast
addresses (`ip.IsUnspecified()` / `ip.IsInterfaceLocalMulticast()`) as blocked.

### Client real IP (`client_ip.go`)

| Function | Description |
|----------|-------------|
| `SetTrustedProxies(cidrs []string)` | Sets the trusted proxy CIDR list. Call once at startup. **An empty slice / nil means no proxy is trusted** (RemoteAddr is always used). A bare IP without a mask is treated as `/32` (IPv4) or `/128` (IPv6); invalid entries are skipped |
| `ParseTrustedProxies(raw string) []string` | Parses a comma-separated range list (env-config form) into valid entries; invalid entries are skipped, empty/all-invalid input returns `nil` |
| `SetTrustedProxiesFromEnvVar(name string)` | Reads a comma-separated range list from the named environment variable; unset/empty clears the trusted-proxy table |
| `ClientIPFromRequest(req *http.Request) string` | Resolves the client real IP; returns `""` when `req` is `nil` |

#### Default semantics: fail-closed

Without `SetTrustedProxies`, **no proxy is trusted**: XFF / X-Real-IP are ignored
entirely, even for private-range peers. There is no "safe default range" here —
private ranges are potential clients too, and trusting RFC1918 by default lets any
internal host spoof XFF. In production, specify your actual reverse-proxy ranges
precisely via `SetTrustedProxies` / `SetTrustedProxiesFromEnvVar`.

#### ClientIPFromRequest resolution strategy

0. No trusted proxies configured → return the peer IP and ignore every forwarding header.
1. Take the direct peer IP from `RemoteAddr` (the real source of the TCP connection, cannot be spoofed).
2. If the peer is not a trusted proxy (or is unparseable) → return the peer IP and ignore XFF entirely.
3. If the peer is a trusted proxy → walk the `X-Forwarded-For` chain **right-to-left**, skipping trusted hops;
   the first non-trusted IP is the client. Unparseable entries are skipped; if every entry is trusted, the leftmost IP is returned.
4. If XFF is empty or entirely invalid → fall back to `X-Real-IP` (still only when the peer is a trusted proxy).
5. Fallback: the peer IP itself (e.g. a direct connection to nginx).

### Kratos context helpers (`context.go`)

| Function | Description |
|----------|-------------|
| `ClientIPFromContext(ctx context.Context) string` | Resolves the client real IP from a Kratos server context (equivalent to calling `ClientIPFromRequest` on the extracted `*http.Request`); returns `""` for non-HTTP transports |
| `HeaderFromContext(ctx context.Context) http.Header` | Extracts HTTP request headers from a Kratos server context; returns `nil` for non-HTTP transports |
| `CookieFromContext(ctx context.Context, name string) string` | Extracts a named cookie value from a Kratos server context; returns `""` for non-HTTP transports or missing cookies |

## Notes

- Call `SetTrustedProxies` once at startup. Without configuration the trusted-proxy
  table is empty (fail-closed) and every request is treated as a direct connection.
- When always deployed behind an internal reverse proxy (nginx etc.), narrow the trusted
  proxies to the actual proxy ranges so arbitrary internal hosts are not treated as trusted.
- There is a time-of-check-to-time-of-use window between `LookupAndCheckHost` and the
  actual connection; always pair it with IP pinning (see Quick Start example 3), otherwise
  DNS rebinding can bypass the check.
- The SSRF policy is a conservative blacklist: unparseable input is always treated as blocked.

## References

- [OWASP SSRF Prevention Cheat Sheet](https://cheatsheetseries.owasp.org/cheatsheets/Server_Side_Request_Forgery_Prevention_Cheat_Sheet.html)
- [Go net.Dialer.Control documentation](https://pkg.go.dev/net#Dialer)
