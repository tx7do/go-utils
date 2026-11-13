// Package mailer 提供基于 net/smtp 的邮件发送能力。
// 支持 STARTTLS（587/25）与隐式 SSL/TLS（465）两种加密方式，
// 认证使用 SMTP PLAIN 机制（覆盖常见邮箱服务商的授权码模式）。
package mailer

import (
	"context"
	"crypto/tls"
	"fmt"
	"mime"
	"net"
	"net/mail"
	"net/smtp"
	"strconv"
	"strings"
	"time"
)

// SmtpConfig SMTP 连接配置
type SmtpConfig struct {
	Host     string // 服务器地址
	Port     uint32 // 端口
	Username string // 用户名
	Password string // 密码/授权码
	From     string // 发件人地址
	// FromName 发件人显示名（可选）。非空时 From 头为 `名称 <地址>`，
	// 非 ASCII 名称按 RFC 2047 编码——收件客户端直接显示这个名，而不是裸地址。
	FromName string
	TlsMode  string // NONE / START_TLS / SSL
}

// 加密方式在库里是可空字符串列，而值域只有三个（见 SmtpConfig.TlsMode）：
// 演示数据/手工改库都能写出 "SSL_TLS" 这类枚举外的串。拨号前把它判出来比拨号失败
// 更有用——"配置值不认识"是配置问题，不该和"SMTP 服务端报错"混成同一种失败。
type tlsMode int

const (
	tlsUnrecognized tlsMode = iota
	tlsPlaintext            // NONE / START_TLS / 空串：明文建连，服务端支持则升级
	tlsImplicit             // SSL：全程 TLS（通常 465）
)

func parseTlsMode(mode string) tlsMode {
	switch strings.ToUpper(mode) {
	case "SSL":
		return tlsImplicit
	case "NONE", "START_TLS", "":
		return tlsPlaintext
	default:
		return tlsUnrecognized
	}
}

// IsSupportedTlsMode 报告 mode 是否为 SendMail 能识别的加密方式。
// 渠道配置在投递前用它自检，免得枚举外的值一路走到拨号才失败。
func IsSupportedTlsMode(mode string) bool {
	return parseTlsMode(mode) != tlsUnrecognized
}

// sendOptions 聚合 SendMail 的可选行为开关。
type sendOptions struct {
	html bool // 正文按 text/html 声明
}

// SendOption 定制 SendMail 的报文行为。
type SendOption func(*sendOptions)

// WithHTML 将正文 Content-Type 声明为 text/html; charset=UTF-8。
// 缺省为 text/plain。
func WithHTML() SendOption {
	return func(o *sendOptions) { o.html = true }
}

// nowFunc 报文 Date 头的时间源；测试冻结它以获得确定性报文。
var nowFunc = time.Now

// SendMail 通过 SMTP 发送一封邮件（缺省纯文本，WithHTML 可切 HTML）。
// to 可为多个收件人。
//
// 报文头自动处理两件容易踩坑的事：
//   - Subject 含非 ASCII（如中文）时按 RFC 2047 编码，ASCII 主题原样穿透；
//   - 按 RFC 5322 补 Date 头（部分反垃圾网关会因缺 Date 直接拒收）。
//
// ctx 贯穿到 TCP 建连：SMTP 服务端不可达时 net.Dial 会一直卡到操作系统级的
// 连接超时（Windows 上可达 20s+），而发送验证码邮件是登录链路里的同步调用，
// 没有 ctx 就没有任何一层能把它打断。
func SendMail(ctx context.Context, cfg SmtpConfig, to []string, subject, body string, opts ...SendOption) error {
	var o sendOptions
	for _, f := range opts {
		f(&o)
	}

	if cfg.Host == "" || cfg.Port == 0 {
		return fmt.Errorf("smtp host/port is not configured")
	}
	if len(to) == 0 {
		return fmt.Errorf("recipient is empty")
	}
	if cfg.From == "" {
		cfg.From = cfg.Username
	}

	addr := cfg.Host + ":" + strconv.Itoa(int(cfg.Port))
	from := strings.TrimSpace(cfg.From)

	fromHeader := from
	if cfg.FromName != "" {
		// mail.Address.String() 会对非 ASCII 名称做 RFC 2047 Q 编码，
		// 且地址含特殊字符时自动加尖括号——手拼引号会在这两处踩坑。
		fromHeader = (&mail.Address{Name: cfg.FromName, Address: from}).String()
	}

	msg := buildMessageWith(fromHeader, to, subject, body, o)

	var client *smtp.Client
	var err error

	switch parseTlsMode(cfg.TlsMode) {
	case tlsImplicit:
		client, err = dialSSL(ctx, addr, cfg.Host)
	case tlsPlaintext:
		// NONE：明文连接（内网调试 SMTP 常见）；服务器支持 STARTTLS 时自动升级
		client, err = dialStartTLS(ctx, addr, cfg.Host)
	default:
		return fmt.Errorf("unsupported tls mode: %s", cfg.TlsMode)
	}
	if err != nil {
		return fmt.Errorf("connect smtp server failed: %w", err)
	}
	defer client.Close()

	if cfg.Username != "" {
		auth := smtp.PlainAuth("", cfg.Username, cfg.Password, cfg.Host)
		if err = client.Auth(auth); err != nil {
			return fmt.Errorf("smtp auth failed: %w", err)
		}
	}

	if err = client.Mail(from); err != nil {
		return fmt.Errorf("smtp MAIL failed: %w", err)
	}
	for _, rcpt := range to {
		if err = client.Rcpt(rcpt); err != nil {
			return fmt.Errorf("smtp RCPT failed for %s: %w", rcpt, err)
		}
	}
	w, err := client.Data()
	if err != nil {
		return fmt.Errorf("smtp DATA failed: %w", err)
	}
	if _, err = w.Write(msg); err != nil {
		return fmt.Errorf("write message failed: %w", err)
	}
	if err = w.Close(); err != nil {
		return fmt.Errorf("close message failed: %w", err)
	}

	return client.Quit()
}

func dialSSL(ctx context.Context, addr, host string) (*smtp.Client, error) {
	conn, err := (&tls.Dialer{Config: &tls.Config{ServerName: host}}).DialContext(ctx, "tcp", addr)
	if err != nil {
		return nil, err
	}
	return smtp.NewClient(conn, host)
}

func dialStartTLS(ctx context.Context, addr, host string) (*smtp.Client, error) {
	conn, err := (&net.Dialer{}).DialContext(ctx, "tcp", addr)
	if err != nil {
		return nil, err
	}
	client, err := smtp.NewClient(conn, host)
	if err != nil {
		_ = conn.Close()
		return nil, err
	}
	if ok, _ := client.Extension("STARTTLS"); ok {
		if err = client.StartTLS(&tls.Config{ServerName: host}); err != nil {
			_ = client.Close()
			return nil, fmt.Errorf("STARTTLS failed: %w", err)
		}
	}
	// 服务器不支持 STARTTLS 时按明文继续（内网调试 SMTP 常见）
	return client, nil
}

// buildMessage 缺省选项（纯文本）的报文拼装；供既有调用方与测试使用。
func buildMessage(from string, to []string, subject, body string) []byte {
	return buildMessageWith(from, to, subject, body, sendOptions{})
}

func buildMessageWith(from string, to []string, subject, body string, o sendOptions) []byte {
	// QEncoding.Encode 对纯 ASCII 输入原样返回，非 ASCII 才编码——无条件套用安全。
	subject = mime.QEncoding.Encode("UTF-8", subject)

	contentType := "text/plain; charset=UTF-8"
	if o.html {
		contentType = "text/html; charset=UTF-8"
	}

	header := make([]byte, 0, 512+len(subject)+len(body))
	header = append(header, "From: "...)
	header = append(header, from...)
	header = append(header, "\r\n"...)
	header = append(header, "To: "...)
	header = append(header, strings.Join(to, ", ")...)
	header = append(header, "\r\n"...)
	header = append(header, "Subject: "...)
	header = append(header, subject...)
	header = append(header, "\r\n"...)
	header = append(header, "Date: "...)
	header = append(header, nowFunc().Format(time.RFC1123Z)...)
	header = append(header, "\r\nMIME-Version: 1.0\r\nContent-Type: "...)
	header = append(header, contentType...)
	header = append(header, "\r\n\r\n"...)
	header = append(header, body...)
	header = append(header, "\r\n"...)
	return header
}
