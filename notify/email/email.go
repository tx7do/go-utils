// Package email SMTP 邮件渠道：把 notify.Message 映射到同库 mailer 的一次投递。
package email

import (
	"context"
	"errors"

	"github.com/tx7do/go-utils/mailer"
	"github.com/tx7do/go-utils/notify"
)

// Config SMTP 连接配置。
type Config = mailer.SmtpConfig

// Service 邮件渠道。无状态：配置在构造时注入，可并发使用。
type Service struct {
	cfg Config
}

// New 构造邮件渠道。
func New(cfg Config) *Service {
	return &Service{cfg: cfg}
}

// Name 渠道名。
func (s *Service) Name() string { return "email" }

// Send 投递一封邮件：Target 为收件地址（多个收件人的场景直接用 mailer.SendMail），
// Title 为 Subject，Content 为纯文本正文。
func (s *Service) Send(ctx context.Context, msg *notify.Message) (*notify.Receipt, error) {
	if msg == nil || msg.Target == "" {
		return nil, errors.New("email target is empty")
	}
	if err := mailer.SendMail(ctx, s.cfg, []string{msg.Target}, msg.Title, msg.Content); err != nil {
		return nil, err
	}
	return &notify.Receipt{}, nil
}
