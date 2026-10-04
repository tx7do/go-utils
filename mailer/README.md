# mailer

基于 `net/smtp` 的邮件发送：

- 支持 STARTTLS（587/25，服务端支持时自动升级）与隐式 SSL/TLS（465）两种加密方式；
- 认证使用 SMTP PLAIN 机制（覆盖常见邮箱服务商的授权码模式）；
- `ctx` 贯穿到 TCP 建连——SMTP 不可达时不会卡到操作系统级连接超时（Windows 上可达 20s+），
  登录验证码这类同步链路必须有超时打断能力；
- `IsSupportedTlsMode` 在拨号前自检加密方式配置，"配置值不认识"与"SMTP 服务端报错"
  不会被混成同一种失败。

```go
import "github.com/tx7do/go-utils/mailer"

err := mailer.SendMail(ctx, mailer.SmtpConfig{
	Host: "smtp.example.com", Port: 465,
	Username: "no-reply@example.com", Password: "authcode",
	From: "no-reply@example.com", TlsMode: "SSL",
}, []string{"to@example.com"}, "subject", "body")
```
