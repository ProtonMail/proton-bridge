package logging

import (
	"github.com/sirupsen/logrus"
)

// SMTPErrorLogger implements go-smtp/logger interface.
type SMTPErrorLogger struct {
	l *logrus.Entry
}

func NewSMTPLogger() *SMTPErrorLogger {
	return &SMTPErrorLogger{l: logrus.WithField("pkg", "SMTP")}
}

func (s *SMTPErrorLogger) Printf(format string, args ...interface{}) {
	s.l.Errorf(format, args...)
}

func (s *SMTPErrorLogger) Println(args ...interface{}) {
	s.l.Errorln(args...)
}

// SMTPDebugLogger implements the writer interface for debug SMTP logs
type SMTPDebugLogger struct {
	l *logrus.Entry
}

func NewSMTPDebugLogger() *SMTPDebugLogger {
	return &SMTPDebugLogger{l: logrus.WithField("pkg", "SMTP")}
}

func (l *SMTPDebugLogger) Write(p []byte) (n int, err error) {
	return l.l.WriterLevel(logrus.TraceLevel).Write(p)
}
