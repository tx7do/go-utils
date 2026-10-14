package notify

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/http"
	"os"
	"time"
)

const (
	// guardedClientTimeout 单次出站请求的总超时（拨号 + 写请求 + 读响应）。
	// 出站这件事不需要每个渠道各定一个数，10s 是个对人对己都体面的默认。
	guardedClientTimeout = 10 * time.Second

	// guardedDialTimeout 拨号阶段上限（掐在 TCP 握手，不影响慢但对端活着的响应）。
	guardedDialTimeout = 5 * time.Second
)

// ErrBlockedTarget SSRF 防线主动拒绝拨号时垫底的哨兵。
//
// 为什么要单独一个错误而不是一句普通文本：拦网发生在拨号之前，一次都没有联系对端，
// 调用方必须能把它归成"地址不合法"（SKIPPED）而不是"发了但被拒"（FAILED）——
// 再重试四次也不会让这条 DNS 记录改口。
// http.Client 会把 dial 错误包成 url.Error，只能靠 errors.Is 沿链找到它。
var ErrBlockedTarget = errors.New("notification target blocked by the ssrf guard")

// EnvAllowPrivateOutbound 放行内网/环回地址的环境变量名，仅供本机联调。
// 值 "1" 时整条 SSRF 防线关闭（连环回都能发），生产部署不得设置。
const EnvAllowPrivateOutbound = "NOTIFY_ALLOW_PRIVATE_TARGET"

// AllowPrivateTarget 读环境变量决定要不要关掉 SSRF 防线。
// 只认精确的 "1"：真值判断（strconv.ParseBool）会让 "false"、"0" 之外的垃圾值也放行，
// 而这一条开关放行的内容是"本进程可向内网任意地址发请求"。
func AllowPrivateTarget() bool {
	return os.Getenv(EnvAllowPrivateOutbound) == "1"
}

// blockedNets SSRF 拦网的地址段。
//
// 为什么是一张显式 CIDR 表而不是 net.IP 的一组布尔判断：布尔法会漏掉 100.64.0.0/10
// （CGNAT，Go 的 IsPrivate 只认 RFC1918 三段，且其文档明写"不得用于访问控制"），
// 而这张表里每一项都是"本进程真能打到东西"的段：环回、RFC1918、CGNAT、
// 链路本地（含 169.254.169.254 这块云元数据端点）、组播与保留段。
// 表还有一个好处：想知道防线是什么，读这一处就够，不必在脑内合成五个布尔判断的并集。
var blockedNets = mustParseNets(
	"0.0.0.0/8", // "本网络"：部分实现把 0.x 当本地
	"10.0.0.0/8",
	"100.64.0.0/10", // CGNAT：K8s 服务网段常落在这里
	"127.0.0.0/8",
	"169.254.0.0/16", // 链路本地 + 云厂商元数据端点
	"172.16.0.0/12",
	"192.0.0.0/24",
	"192.168.0.0/16",
	"198.18.0.0/15", // 基准测试段
	"224.0.0.0/4",
	"240.0.0.0/4",
	"::1/128",
	"fc00::/7", // IPv6 ULA
	"fe80::/10",
)

// NewGuardedClient 造一个"拨号前先把地址过一遍内网拦网"的 http.Client。
//
// 为什么防线在 dial 而不是在 url 校验：校验 hostname 只挡得住字面量（127.0.0.1、
// 169.254.169.254），挡不住 `internal.example.com → 10.0.0.5` 这种指到内网的记录。
// 这里自己解析、自己挑一个放行的 IP、并且**直接 dial 那个 IP**，
// 于是"校验的地址"与"连接的地址"是同一个，TOCTOU 的窗口只剩 DNS 重绑定这一种
// 需要在本进程解析之后再改记录的形式，而解析结果已经决定了连接目标。
func NewGuardedClient(allowPrivate bool) *http.Client {
	dialer := &net.Dialer{Timeout: guardedDialTimeout, KeepAlive: 15 * time.Second}

	return &http.Client{
		Timeout: guardedClientTimeout,
		Transport: &http.Transport{
			DialContext: func(ctx context.Context, network, addr string) (net.Conn, error) {
				if allowPrivate {
					return dialer.DialContext(ctx, network, addr)
				}
				return dialGuarded(ctx, dialer, network, addr)
			},
		},
		// 不跟随重定向：Location 指到哪就等于绕过了"调用方指定的这条地址"，
		// 台账/日志里那行 target 也成了假话。3xx 一律当失败处理。
		CheckRedirect: func(*http.Request, []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}
}

// dialGuarded 解析 addr 的 hostname，挑一个不在拦网表里的 IP 直连。
func dialGuarded(ctx context.Context, dialer *net.Dialer, network, addr string) (net.Conn, error) {
	host, port, err := net.SplitHostPort(addr)
	if err != nil {
		return nil, fmt.Errorf("bad dial address %q: %w", addr, err)
	}

	// "ip" 而非 network：这里要的是"两种族都查"，而 network 传的是 tcp/tcp4/tcp6 这类协议名。
	ips, err := net.DefaultResolver.LookupIPAddr(ctx, host)
	if err != nil {
		return nil, fmt.Errorf("resolve %q failed: %w", host, err)
	}
	if len(ips) == 0 {
		return nil, fmt.Errorf("%q resolves to no address", host)
	}

	for _, ip := range ips {
		// 一个名字只要有一条记录指向内网就整条拒掉：允许"混着指"等于把
		// "这一次到底打到了哪"交给 DNS 的返回顺序。
		if blocked, reason := IsBlockedIP(ip.IP); blocked {
			return nil, fmt.Errorf("%w: target %q resolves to %s which is %s — off-host requests must point at a public address",
				ErrBlockedTarget, host, ip.IP.String(), reason)
		}
		conn, dialErr := dialer.DialContext(ctx, network, net.JoinHostPort(ip.IP.String(), port))
		if dialErr != nil {
			// 这个 IP 试不通就接着试下一个：多 A 记录（v4/v6 双栈）是常态，
			// 拿第一条的连接失败当整次投递的失败会误判。
			err = dialErr
			continue
		}
		return conn, nil
	}

	return nil, fmt.Errorf("dial %q failed: %w", addr, err)
}

// IsBlockedIP 报告地址是否落在拦网表里。
//
// 返回的 reason 会进错误文案，所以要能看懂"为什么这条被拦"。
func IsBlockedIP(ip net.IP) (bool, string) {
	if ip == nil {
		return true, "not an IP address"
	}
	for _, parsed := range blockedNets {
		if parsed.Contains(ip) {
			return true, "inside the blocked range " + parsed.String()
		}
	}
	return false, ""
}

// mustParseNets 解析拦网表；写错在包初始化期 panic，不留到第一次投递才报错。
func mustParseNets(cidrs ...string) []*net.IPNet {
	nets := make([]*net.IPNet, 0, len(cidrs))
	for _, c := range cidrs {
		_, parsed, err := net.ParseCIDR(c)
		if err != nil {
			panic(fmt.Sprintf("notify: bad blocked CIDR %q: %v", c, err))
		}
		nets = append(nets, parsed)
	}
	return nets
}
