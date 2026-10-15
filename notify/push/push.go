// Package push 移动端厂商推送渠道：APNs / FCM / 华为 / 荣耀 / 小米 / OPPO / vivo。
//
// 全部裸 HTTP（APNs 为标准库 HTTP/2 + ES256 Token 鉴权），零厂商 SDK。
// 每个 Service 发送**单台设备**（notify.Message.Target = 设备令牌/regId），
// 多设备的遍历、失效令牌清理与结果聚合是调用方编排层的事。
//
// 错误语义沿 notify 惯例分三类：
//   - 配置类（notify.ErrNotConfigured）：凭据缺失，拨号之前拦下；
//   - 令牌失效（ErrInvalidToken）：厂商明确说这个 regId/token 已经作废
//     （应用卸载/实例删除）——调用方应删设备行自清理，不算投递失败重试；
//   - 其余 error：拨过号被拒或链路故障，按投递失败处理。
package push

import (
	"fmt"
	"time"

	"github.com/tx7do/go-utils/notify"
)

// ErrInvalidToken 厂商侧明确判定的设备令牌失效。
//
// v0.0.3 时是本包定义的哨兵；随跨渠道通用化上移到 notify 根包，
// 这里保留别名，既有 `errors.Is(push.ErrInvalidToken, ...)` 用法不受影响。
var ErrInvalidToken = notify.ErrInvalidToken

// pushTimeout 单台设备出站请求的总超时。
const pushTimeout = 10 * time.Second

// answerLimit 读对端应答的上限：结论都是几百字节的小 JSON/文本。
const answerLimit = 8 << 10

// errSnippet 进错误文案的响应片段截断。
func errSnippet(body []byte) string {
	const limit = 256
	if len(body) > limit {
		return string(body[:limit])
	}
	return string(body)
}

// truncateRunes rune 安全截断（厂商标题有 rune 数上限，超了直接 400）。
func truncateRunes(s string, max int) string {
	runes := []rune(s)
	if len(runes) <= max {
		return s
	}
	return string(runes[:max])
}

// notConfigured 配置类失败的统一包装。
func notConfigured(format string, args ...any) error {
	return fmt.Errorf("%w: "+format, append([]any{notify.ErrNotConfigured}, args...)...)
}
