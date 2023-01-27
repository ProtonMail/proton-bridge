// Copyright (c) 2023 Proton AG
//
// This file is part of Proton Mail Bridge.
//
// Proton Mail Bridge is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// Proton Mail Bridge is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with Proton Mail Bridge.  If not, see <https://www.gnu.org/licenses/>.

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

// SMTPDebugLogger implements the writer interface for debug SMTP logs.
type SMTPDebugLogger struct {
	l *logrus.Entry
}

func NewSMTPDebugLogger() *SMTPDebugLogger {
	return &SMTPDebugLogger{l: logrus.WithField("pkg", "SMTP")}
}

func (l *SMTPDebugLogger) Write(p []byte) (n int, err error) {
	return l.l.WriterLevel(logrus.TraceLevel).Write(p)
}
