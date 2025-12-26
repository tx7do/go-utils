package eventloop

import (
	"io"
	"log"
	"os"
)

// Logger 是通用日志接口，使用格式化输出
type Logger interface {
	Debugf(format string, v ...interface{})
	Infof(format string, v ...interface{})
	Warnf(format string, v ...interface{})
	Errorf(format string, v ...interface{})
}

// NoopLogger 什么也不做，适合默认/测试场景
type NoopLogger struct{}

func (NoopLogger) Debugf(format string, v ...interface{}) {}
func (NoopLogger) Infof(format string, v ...interface{})  {}
func (NoopLogger) Warnf(format string, v ...interface{})  {}
func (NoopLogger) Errorf(format string, v ...interface{}) {}

// StdLogger 使用标准库 log.Logger，并在消息前加级别前缀
type StdLogger struct {
	l *log.Logger
}

// NewStdLogger 创建 StdLogger，writer 可传 os.Stdout/os.Stderr 等
func NewStdLogger(writer io.Writer, prefix string, flag int) *StdLogger {
	if writer == nil {
		writer = os.Stdout
	}
	return &StdLogger{
		l: log.New(writer, prefix, flag),
	}
}

func (s *StdLogger) Debugf(format string, v ...interface{}) {
	s.l.Printf("[DEBUG] "+format, v...)
}
func (s *StdLogger) Infof(format string, v ...interface{}) {
	s.l.Printf("[INFO] "+format, v...)
}
func (s *StdLogger) Warnf(format string, v ...interface{}) {
	s.l.Printf("[WARN] "+format, v...)
}
func (s *StdLogger) Errorf(format string, v ...interface{}) {
	s.l.Printf("[ERROR] "+format, v...)
}
