# mailer

[English](README.md) | 简体中文

基于标准库 `net/smtp` 的轻量邮件发送包：一封纯文本邮件、一种配置结构、一个函数调用。

## 功能特性

- **两种加密方式**：STARTTLS（587/25，服务端支持时自动升级）与隐式 SSL/TLS（465，全程 TLS）；
- **SMTP PLAIN 认证**：覆盖常见邮箱服务商的授权码模式；`Username` 为空时跳过认证（匿名投递，内网 relay 场景）；
- **`ctx` 贯穿到 TCP 建连**：SMTP 服务端不可达时不会卡到操作系统级连接超时（Windows 上可达 20s+）——登录验证码这类同步链路必须有超时打断能力；
- **拨号前自检加密方式**：`IsSupportedTlsMode` 把"配置值不认识"（配置问题）与"SMTP 服务端报错"（投递问题）分开，不被混成同一种失败；
- **发件人显示名**：`FromName` 非空时 From 头为 `名称 <地址>`，非 ASCII 名称按 RFC 2047 编码（经 `net/mail.Address` 构造，避免手拼引号踩坑）。

## 安装

独立模块，不引入其它 go-utils 包：

```bash
go get github.com/tx7do/go-utils/mailer@latest
```

```go
import "github.com/tx7do/go-utils/mailer"
```

## 快速开始

```go
package main

import (
	"context"
	"log"

	"github.com/tx7do/go-utils/mailer"
)

func main() {
	err := mailer.SendMail(context.Background(),
		mailer.SmtpConfig{
			Host:     "smtp.example.com",
			Port:     465,
			Username: "no-reply@example.com",
			Password: "auth-code",           // 服务商授权码
			From:     "no-reply@example.com",
			FromName: "Example App",         // 可选：收件客户端显示的发件人名
			TlsMode:  "SSL",
		},
		[]string{"to@example.com"},          // 支持多个收件人
		"登录验证码",
		"您的验证码是 123456，5 分钟内有效。",
	)
	if err != nil {
		log.Fatal(err)
	}
}
```

## 配置

`SmtpConfig` 字段一览（与 `smtp.go` 定义逐字对齐）：

| 字段 | 类型 | 说明 |
|---|---|---|
| `Host` | `string` | SMTP 服务器地址，必填 |
| `Port` | `uint32` | 端口，必填（465 用 SSL；587/25 用 START_TLS 或 NONE） |
| `Username` | `string` | 用户名；为空则跳过认证 |
| `Password` | `string` | 密码 / 授权码 |
| `From` | `string` | 发件人地址；为空时兜底为 `Username` |
| `FromName` | `string` | 发件人显示名（可选），非空时 From 头为 `名称 <地址>`，非 ASCII 名称按 RFC 2047 编码 |
| `TlsMode` | `string` | 加密方式：`NONE` / `START_TLS` / `SSL`，大小写不敏感；空串按 `NONE` 处理 |

`TlsMode` 三种取值的行为：

- **`SSL`**：拨号即 TLS（隐式 TLS，通常 465）；
- **`START_TLS` / `NONE`**：先明文建连，服务端宣告 `STARTTLS` 扩展则升级加密；服务端不支持时按明文继续（内网调试 SMTP 常见）；
- **其它值**：`SendMail` 直接报 `unsupported tls mode`，不会走到拨号。`IsSupportedTlsMode` 可在投递前用它自检同一份配置。

## API 参考

```go
// 通过 SMTP 发送一封纯文本邮件；to 可为多个收件人。
func SendMail(ctx context.Context, cfg SmtpConfig, to []string, subject, body string) error

// 报告 mode 是否为 SendMail 能识别的加密方式（NONE / START_TLS / SSL，大小写不敏感，空串合法）。
func IsSupportedTlsMode(mode string) bool
```

正文固定为 `text/plain; charset=UTF-8`；`Subject` 不做额外编码（含非 ASCII 字符时部分客户端可能显示异常，建议自行用 RFC 2047 编码后传入）。

## Notes / 注意事项

- **错误包装格式**：`SendMail` 返回的错误按阶段包装，可用 `strings.HasPrefix` 或错误文案定位失败点：
  - `connect smtp server failed: ...` —— TCP 建连 / STARTTLS 升级失败（`smtp.go`）；
  - `smtp auth failed: ...` —— 认证被拒；
  - `smtp MAIL failed: ...` / `smtp RCPT failed for <addr>: ...` / `smtp DATA failed: ...` —— 会话阶段被服务端拒绝；
  - `unsupported tls mode: <值>` —— 配置值不在枚举内；
  - `smtp host/port is not configured` / `recipient is empty` —— 入参校验。
- **From 兜底**：`From` 为空时自动改用 `Username` 作为发件地址（`smtp.go`）；
- **STARTTLS 缺失不报错**：`START_TLS`/`NONE` 模式下，服务端不支持 STARTTLS 时按明文继续发（`smtp.go`）——对安全有要求的场景请用 `SSL` 模式；
- **单次调用**：每次 `SendMail` 独立建连、发信、`Quit`，无连接复用；高频发送请在上层自行做队列与节流。
