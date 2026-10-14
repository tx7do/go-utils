// Package notify 多渠道通知投递的统一抽象：Service 接口 + 渠道注册表 + SSRF 防线。
//
// 这里只管"怎么发"：给定一个已经解析好的渠道配置（构造 Service 时注入）和一条
// 正文就绪的 Message，把消息发出去并拿回执。渠道配置从哪里来（数据库/环境变量）、
// 失败怎么记账、要不要重试，都不归本包——那是调用方的编排层。
//
// 各渠道实现：
//   - email/   SMTP 邮件（走同库的 mailer 子模块）
//   - webhook/ HTTP 回调（钉钉/飞书/企微签名风格 + HMAC 自有方案 + SSRF 防线）
//   - sms/     短信（阿里云 Dysmsapi / 腾讯云 SMS，裸 HTTP 实现，零 SDK）
package notify

import (
	"context"
	"errors"
)

// ErrNotConfigured 渠道不可用的哨兵：没配 / 配错 / 停用 / 目标不合法等
// "一次都没联系到对端"的错误都包在它里面。
//
// 为什么要单独一个哨兵：调用方的台账要能区分两种失败——
// "没尝试过投递"（SKIPPED：配置问题，重试也不会变好）与
// "投递被拒"（FAILED：拨过号/递过单，对端说了不）。
// 实现准则：凡是发生在联系对端**之前**的拒绝，一律包 ErrNotConfigured。
var ErrNotConfigured = errors.New("notification channel is not configured")

// Message 一次投递意图。
//
// 不含模板与路由：正文由调用方备好；发给谁、怎么记账也是调用方的事。
type Message struct {
	// Target 投递目标，按渠道解释：email 为收件地址，sms 为手机号，
	// webhook 为回调 URL。
	Target string
	// Title 通知标题（邮件为 Subject；短信等无标题渠道忽略）。
	Title string
	// Content 通知正文。
	Content string
	// Params 渠道原生参数，按渠道解释：
	//   - sms：模板参数（阿里云按参数名传 JSON，腾讯云按参数名排序取值传位置参数）
	//   - webhook：payload 模板的额外占位符变量
	Params map[string]string
}

// Receipt 渠道回执。
type Receipt struct {
	// ProviderID 服务商侧的消息/请求 ID（阿里云 RequestId、腾讯云 SerialNo 等），
	// 拿得到就带，方便对账；多数渠道没有就留空。
	ProviderID string
}

// Service 一种投递渠道。构造时持有自己的配置，Send 只管发。
//
// Send 返回 error 即视为投递失败；配置类失败必须包 ErrNotConfigured（见其注释）。
// 返回的 error 非 nil 时 Receipt 仍可带回已确定的事实（如服务商侧 ID）。
type Service interface {
	// Name 渠道名（"email" / "webhook" / "sms"），供注册表与路由分派。
	Name() string
	Send(ctx context.Context, msg *Message) (*Receipt, error)
}

// Registry 渠道注册表：按 Name 注册与取用。
//
// Register 覆盖同名旧实现：装配期一次性写入，运行期只读。
type Registry struct {
	services map[string]Service
}

func NewRegistry() *Registry {
	return &Registry{services: make(map[string]Service)}
}

func (r *Registry) Register(s Service) {
	r.services[s.Name()] = s
}

// Service 取渠道实现；未注册返回 ok=false。
func (r *Registry) Service(name string) (Service, bool) {
	s, ok := r.services[name]
	return s, ok
}
