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
	"io"
	"net"

	"github.com/sirupsen/logrus"
)

// connListener sets debug loggers on server containing fields with local
// and remote addresses right after new connection is accepted.
type connListener struct {
	net.Listener

	server Server
}

func (l *connListener) Accept() (net.Conn, error) {
	conn, err := l.Listener.Accept()

	if err == nil && (l.server.DebugServer() || l.server.DebugClient()) {
		debugLog := logrus.WithField("pkg", l.server.Protocol())
		if addr := conn.LocalAddr(); addr != nil {
			debugLog = debugLog.WithField("loc", addr.String())
		}
		if addr := conn.RemoteAddr(); addr != nil {
			debugLog = debugLog.WithField("rem", addr.String())
		}

		var localDebug, remoteDebug io.Writer
		if l.server.DebugServer() {
			localDebug = debugLog.WithField("comm", "server").WriterLevel(logrus.DebugLevel)
		}
		if l.server.DebugClient() {
			remoteDebug = debugLog.WithField("comm", "client").WriterLevel(logrus.DebugLevel)
		}

		l.server.SetLoggers(localDebug, remoteDebug)
	}

	return conn, err
}
