// Copyright (c) 2020 Proton Technologies AG
//
// This file is part of ProtonMail Bridge.
//
// ProtonMail Bridge is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// ProtonMail Bridge is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with ProtonMail Bridge.  If not, see <https://www.gnu.org/licenses/>.

package imap

import (
	"crypto/tls"
	"fmt"
	"io"
	"strings"
	"time"

	imapid "github.com/ProtonMail/go-imap-id"
	"github.com/ProtonMail/proton-bridge/internal/bridge"
	"github.com/ProtonMail/proton-bridge/internal/events"
	"github.com/ProtonMail/proton-bridge/internal/imap/uidplus"
	"github.com/ProtonMail/proton-bridge/pkg/listener"
	"github.com/emersion/go-imap"
	imapappendlimit "github.com/emersion/go-imap-appendlimit"
	imapidle "github.com/emersion/go-imap-idle"
	imapquota "github.com/emersion/go-imap-quota"
	imapspecialuse "github.com/emersion/go-imap-specialuse"
	imapunselect "github.com/emersion/go-imap-unselect"
	imapserver "github.com/emersion/go-imap/server"
	"github.com/emersion/go-sasl"
	"github.com/sirupsen/logrus"
)

type imapServer struct {
	server        *imapserver.Server
	eventListener listener.Listener
}

// NewIMAPServer constructs a new IMAP server configured with the given options.
func NewIMAPServer(debugClient, debugServer bool, port int, tls *tls.Config, imapBackend *imapBackend, eventListener listener.Listener) *imapServer { //nolint[golint]
	s := imapserver.New(imapBackend)
	s.Addr = fmt.Sprintf("%v:%v", bridge.Host, port)
	s.TLSConfig = tls
	s.AllowInsecureAuth = true
	s.ErrorLog = newServerErrorLogger("server-imap")
	s.AutoLogout = 30 * time.Minute

	if debugClient || debugServer {
		var localDebug, remoteDebug imap.WriterWithFields
		if debugClient {
			remoteDebug = &logWithFields{log: log.WithField("pkg", "imap/client"), fields: logrus.Fields{}}
		}
		if debugServer {
			localDebug = &logWithFields{log: log.WithField("pkg", "imap/server"), fields: logrus.Fields{}}
		}
		s.Debug = imap.NewDebugWithFields(localDebug, remoteDebug)
	}

	serverID := imapid.ID{
		imapid.FieldName:       "ProtonMail",
		imapid.FieldVendor:     "Proton Technologies AG",
		imapid.FieldSupportURL: "https://protonmail.com/support",
	}

	s.EnableAuth(sasl.Login, func(conn imapserver.Conn) sasl.Server {
		conn.Server().ForEachConn(func(candidate imapserver.Conn) {
			if id, ok := candidate.(imapid.Conn); ok {
				if conn.Context() == candidate.Context() {
					imapBackend.setLastMailClient(id.ID())
					return
				}
			}
		})

		return sasl.NewLoginServer(func(address, password string) error {
			user, err := conn.Server().Backend.Login(address, password)
			if err != nil {
				return err
			}

			ctx := conn.Context()
			ctx.State = imap.AuthenticatedState
			ctx.User = user
			return nil
		})
	})

	s.Enable(
		imapidle.NewExtension(),
		//imapmove.NewExtension(), // extension is not fully implemented: if UIDPLUS exists it MUST return COPYUID and EXPUNGE continuous responses
		imapspecialuse.NewExtension(),
		imapid.NewExtension(serverID),
		imapquota.NewExtension(),
		imapappendlimit.NewExtension(),
		imapunselect.NewExtension(),
		uidplus.NewExtension(),
	)

	return &imapServer{
		server:        s,
		eventListener: eventListener,
	}
}

// Starts the server.
func (s *imapServer) ListenAndServe() {
	go s.monitorDisconnectedUsers()

	log.Info("IMAP server listening at ", s.server.Addr)
	err := s.server.ListenAndServe()
	if err != nil {
		s.eventListener.Emit(events.ErrorEvent, "IMAP failed: "+err.Error())
		log.Error("IMAP failed: ", err)
		return
	}
	defer s.server.Close() //nolint[errcheck]

	log.Info("IMAP server stopped")
}

// Stops the server.
func (s *imapServer) Close() {
	_ = s.server.Close()
}

func (s *imapServer) monitorDisconnectedUsers() {
	ch := make(chan string)
	s.eventListener.Add(events.CloseConnectionEvent, ch)

	for address := range ch {
		address := address
		log.Info("Disconnecting all open IMAP connections for ", address)
		disconnectUser := func(conn imapserver.Conn) {
			connUser := conn.Context().User
			if connUser != nil && strings.EqualFold(connUser.Username(), address) {
				_ = conn.Close()
			}
		}
		s.server.ForEachConn(disconnectUser)
	}
}

// logWithFields is used for debuging with additional field.
type logWithFields struct {
	log    *logrus.Entry
	fields logrus.Fields
}

func (lf *logWithFields) Writer() io.Writer {
	w := lf.log.WithFields(lf.fields).WriterLevel(logrus.DebugLevel)
	lf.fields = logrus.Fields{}
	return w
}

func (lf *logWithFields) SetField(key, value string) {
	lf.fields[key] = value
}

// serverErrorLogger implements go-imap/logger interface.
type serverErrorLogger struct {
	tag string
}

func newServerErrorLogger(tag string) *serverErrorLogger {
	return &serverErrorLogger{tag}
}

func (s *serverErrorLogger) CheckErrorForReport(serverErr string) {
}

func (s *serverErrorLogger) Printf(format string, args ...interface{}) {
	err := fmt.Sprintf(format, args...)
	s.CheckErrorForReport(err)
	log.WithField("pkg", s.tag).Error(err)
}

func (s *serverErrorLogger) Println(args ...interface{}) {
	err := fmt.Sprintln(args...)
	s.CheckErrorForReport(err)
	log.WithField("pkg", s.tag).Error(err)
}
