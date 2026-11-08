// context.go —— kratos transport 上下文提取助手。
//
// 依赖 go-kratos：service 层经由这三个函数读取客户端真实 IP / 请求头 /
// HttpOnly cookie，无需直接依赖具体中间件包。非 HTTP 形态（非 server
// context、gRPC 传输、request 为 nil）一律返回零值而非 panic。
package netutil

import (
	"context"
	"net/http"

	"github.com/go-kratos/kratos/v2/transport"
	khttp "github.com/go-kratos/kratos/v2/transport/http"
)

// ClientIPFromContext 从 kratos context 中提取客户端真实 IP。
// 若上下文非 HTTP 传输或取不到 request，返回空串。
func ClientIPFromContext(ctx context.Context) string {
	return ClientIPFromRequest(requestFromContext(ctx))
}

// HeaderFromContext 从 kratos context 中提取 HTTP 请求头。
// 用于 service 层读取通过 header 传递的参数（如验证码 id/value）。
// 若上下文非 HTTP 传输或取不到 request，返回 nil。
func HeaderFromContext(ctx context.Context) http.Header {
	req := requestFromContext(ctx)
	if req == nil {
		return nil
	}
	return req.Header
}

// CookieFromContext 从 kratos context 中提取具名 HTTP cookie 值。
// 用于刷新令牌这类接口从 HttpOnly cookie 中读取令牌。
// 若上下文非 HTTP 传输、取不到 request 或无此 cookie，返回空串。
func CookieFromContext(ctx context.Context, name string) string {
	req := requestFromContext(ctx)
	if req == nil {
		return ""
	}
	c, err := req.Cookie(name)
	if err != nil {
		return ""
	}
	return c.Value
}

// requestFromContext 从 kratos server context 中取出 *http.Request。
func requestFromContext(ctx context.Context) *http.Request {
	tr, ok := transport.FromServerContext(ctx)
	if !ok {
		return nil
	}
	htr, ok := tr.(khttp.Transporter)
	if !ok {
		return nil
	}
	return htr.Request()
}
