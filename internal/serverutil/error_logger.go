// Copyright (c) 2022 Proton AG
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
// along with Proton Mail Bridge. If not, see <https://www.gnu.org/licenses/>.

package serverutil

import (
	"github.com/sirupsen/logrus"
)

// ServerErrorLogger implements go-imap/logger interface.
type ServerErrorLogger struct {
	l *logrus.Entry
}

func NewServerErrorLogger(protocol Protocol) *ServerErrorLogger {
	return &ServerErrorLogger{l: logrus.WithField("protocol", protocol)}
}

func (s *ServerErrorLogger) Printf(format string, args ...interface{}) {
	s.l.Errorf(format, args...)
}

func (s *ServerErrorLogger) Println(args ...interface{}) {
	s.l.Errorln(args...)
}
