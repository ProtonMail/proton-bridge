// Copyright (c) 2021 Proton Technologies AG
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
	"net"
	"strings"
	"sync/atomic"
	"time"

	imapid "github.com/ProtonMail/go-imap-id"
	"github.com/ProtonMail/proton-bridge/internal/bridge"
	"github.com/ProtonMail/proton-bridge/internal/config/useragent"
	"github.com/ProtonMail/proton-bridge/internal/events"
	"github.com/ProtonMail/proton-bridge/internal/imap/id"
	"github.com/ProtonMail/proton-bridge/internal/imap/idle"
	"github.com/ProtonMail/proton-bridge/internal/imap/uidplus"
	"github.com/ProtonMail/proton-bridge/internal/serverutil"
	"github.com/ProtonMail/proton-bridge/pkg/listener"
	"github.com/emersion/go-imap"
	imapappendlimit "github.com/emersion/go-imap-appendlimit"
	imapmove "github.com/emersion/go-imap-move"
	imapquota "github.com/emersion/go-imap-quota"
	imapunselect "github.com/emersion/go-imap-unselect"
	"github.com/emersion/go-imap/backend"
	imapserver "github.com/emersion/go-imap/server"
	"github.com/emersion/go-sasl"
	"github.com/sirupsen/logrus"
)

type imapServer struct {
	panicHandler  panicHandler
	server        *imapserver.Server
	userAgent     *useragent.UserAgent
	eventListener listener.Listener
	debugClient   bool
	debugServer   bool
	port          int
	isRunning     atomic.Value
}

// NewIMAPServer constructs a new IMAP server configured with the given options.
func NewIMAPServer(panicHandler panicHandler, debugClient, debugServer bool, port int, tls *tls.Config, imapBackend backend.Backend, userAgent *useragent.UserAgent, eventListener listener.Listener) *imapServer { // nolint[golint]
	s := imapserver.New(imapBackend)
	s.Addr = fmt.Sprintf("%v:%v", bridge.Host, port)
	s.TLSConfig = tls
	s.AllowInsecureAuth = true
	s.ErrorLog = newServerErrorLogger("server-imap")
	s.AutoLogout = 30 * time.Minute

	if debugServer {
		fmt.Println("THE LOG WILL CONTAIN **DECRYPTED** MESSAGE DATA")
		log.Warning("================================================")
		log.Warning("THIS LOG WILL CONTAIN **DECRYPTED** MESSAGE DATA")
		log.Warning("================================================")
	}

	serverID := imapid.ID{
		imapid.FieldName:       "ProtonMail Bridge",
		imapid.FieldVendor:     "Proton Technologies AG",
		imapid.FieldSupportURL: "https://protonmail.com/support",
	}

	s.EnableAuth(sasl.Login, func(conn imapserver.Conn) sasl.Server {
		return sasl.NewLoginServer(func(address, password string) error {
			user, err := conn.Server().Backend.Login(nil, address, password)
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
		idle.NewExtension(),
		imapmove.NewExtension(),
		id.NewExtension(serverID, userAgent),
		imapquota.NewExtension(),
		imapappendlimit.NewExtension(),
		imapunselect.NewExtension(),
		uidplus.NewExtension(),
	)

	server := &imapServer{
		panicHandler:  panicHandler,
		server:        s,
		userAgent:     userAgent,
		eventListener: eventListener,
		debugClient:   debugClient,
		debugServer:   debugServer,
		port:          port,
	}
	server.isRunning.Store(false)
	return server
}

func (s *imapServer) HandlePanic()    { s.panicHandler.HandlePanic() }
func (s *imapServer) IsRunning() bool { return s.isRunning.Load().(bool) }
func (s *imapServer) Port() int       { return s.port }

// ListenAndServe starts the server and keeps it on based on internet
// availability.
func (s *imapServer) ListenAndServe() {
	serverutil.ListenAndServe(s, s.eventListener)
}

// ListenRetryAndServe will start listener. If port is occupied it will try
// again after coolDown time. Once listener is OK it will serve.
func (s *imapServer) ListenRetryAndServe(retries int, retryAfter time.Duration) {
	if s.IsRunning() {
		return
	}
	s.isRunning.Store(true)

	l := log.WithField("address", s.server.Addr)
	l.Info("IMAP server is starting")
	listener, err := net.Listen("tcp", s.server.Addr)
	if err != nil {
		s.isRunning.Store(false)
		if retries > 0 {
			l.WithError(err).WithField("retries", retries).Warn("IMAP listener failed")
			time.Sleep(retryAfter)
			s.ListenRetryAndServe(retries-1, retryAfter)
			return
		}

		l.WithError(err).Error("IMAP listener failed")
		s.eventListener.Emit(events.ErrorEvent, "IMAP failed: "+err.Error())
		return
	}

	err = s.server.Serve(&connListener{
		Listener:  listener,
		server:    s,
		userAgent: s.userAgent,
	})
	// Serve returns error every time, even after closing the server.
	// User shouldn't be notified about error if server shouldn't be running,
	// but it should in case it was not closed by `s.Close()`.
	if err != nil && s.IsRunning() {
		s.isRunning.Store(false)
		l.WithError(err).Error("IMAP server failed")
		s.eventListener.Emit(events.ErrorEvent, "IMAP failed: "+err.Error())
		return
	}
	defer s.server.Close() //nolint[errcheck]

	l.Info("IMAP server stopped")
}

// Stops the server.
func (s *imapServer) Close() {
	if !s.IsRunning() {
		return
	}
	s.isRunning.Store(false)

	log.Info("Closing IMAP server")
	if err := s.server.Close(); err != nil {
		log.WithError(err).Error("Failed to close the connection")
	}
}

func (s *imapServer) DisconnectUser(address string) {
	log.Info("Disconnecting all open IMAP connections for ", address)
	s.server.ForEachConn(func(conn imapserver.Conn) {
		connUser := conn.Context().User
		if connUser != nil && strings.EqualFold(connUser.Username(), address) {
			if err := conn.Close(); err != nil {
				log.WithError(err).Error("Failed to close the connection")
			}
		}
	})
}

// connListener sets debug loggers on server containing fields with local
// and remote addresses right after new connection is accepted.
type connListener struct {
	net.Listener

	server    *imapServer
	userAgent *useragent.UserAgent
}

func (l *connListener) Accept() (net.Conn, error) {
	conn, err := l.Listener.Accept()

	if err == nil && (l.server.debugServer || l.server.debugClient) {
		debugLog := log
		if addr := conn.LocalAddr(); addr != nil {
			debugLog = debugLog.WithField("loc", addr.String())
		}
		if addr := conn.RemoteAddr(); addr != nil {
			debugLog = debugLog.WithField("rem", addr.String())
		}

		var localDebug, remoteDebug io.Writer
		if l.server.debugServer {
			localDebug = debugLog.WithField("pkg", "imap/server").WriterLevel(logrus.DebugLevel)
		}
		if l.server.debugClient {
			remoteDebug = debugLog.WithField("pkg", "imap/client").WriterLevel(logrus.DebugLevel)
		}

		l.server.server.Debug = imap.NewDebugWriter(localDebug, remoteDebug)
	}

	if !l.userAgent.HasClient() {
		l.userAgent.SetClient("UnknownClient", "0.0.1")
	}

	return conn, err
}

// serverErrorLogger implements go-imap/logger interface.
type serverErrorLogger struct {
	tag string
}

func newServerErrorLogger(tag string) *serverErrorLogger {
	return &serverErrorLogger{tag}
}

func (s *serverErrorLogger) Printf(format string, args ...interface{}) {
	err := fmt.Sprintf(format, args...)
	log.WithField("pkg", s.tag).Error(err)
}

func (s *serverErrorLogger) Println(args ...interface{}) {
	err := fmt.Sprintln(args...)
	log.WithField("pkg", s.tag).Error(err)
}
