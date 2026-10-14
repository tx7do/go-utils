package email

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/tx7do/go-utils/notify"
)

// 邮件渠道是 mailer 的一层映射：传输层的行为（TLS 模式、会话错误分支）已由
// go-utils/mailer 的测试钉死，这里只钉映射本身与入口校验。
func TestServiceMapping(t *testing.T) {
	svc := New(Config{Host: "smtp.example.com", Port: 465, Username: "u", Password: "p", From: "noreply@example.com", TlsMode: "SSL"})
	require.Equal(t, "email", svc.Name())

	_, err := svc.Send(context.Background(), nil)
	require.Error(t, err, "nil 消息要报错而不是 panic")

	_, err = svc.Send(context.Background(), &notify.Message{Title: "t"})
	require.Error(t, err, "没有收件人的投递在拨号之前就拒绝")
}
