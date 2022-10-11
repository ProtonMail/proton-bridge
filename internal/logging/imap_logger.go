package logging

import "github.com/sirupsen/logrus"

// IMAPLogger implements the writer interface for Gluon IMAP logs
type IMAPLogger struct {
	l *logrus.Entry
}

func NewIMAPLogger() *IMAPLogger {
	return &IMAPLogger{l: logrus.WithField("pkg", "IMAP")}
}

func (l *IMAPLogger) Write(p []byte) (n int, err error) {
	return l.l.WriterLevel(logrus.TraceLevel).Write(p)
}
