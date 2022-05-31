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

package smtp

import (
	"crypto/tls"
	"fmt"
	"io"
	"net"

	"github.com/ProtonMail/proton-bridge/v2/internal/bridge"
	"github.com/ProtonMail/proton-bridge/v2/internal/serverutil"
	"github.com/ProtonMail/proton-bridge/v2/pkg/listener"
	"github.com/emersion/go-sasl"
	goSMTP "github.com/emersion/go-smtp"
)

// Server is Bridge SMTP server implementation.
type Server struct {
	panicHandler panicHandler
	backend      goSMTP.Backend
	debug        bool
	useSSL       bool
	port         int
	tls          *tls.Config

	server     *goSMTP.Server
	controller serverutil.Controller
}

// NewSMTPServer returns an SMTP server configured with the given options.
func NewSMTPServer(
	panicHandler panicHandler,
	debug bool, port int, useSSL bool,
	tls *tls.Config,
	smtpBackend goSMTP.Backend,
	eventListener listener.Listener,
) *Server {
	server := &Server{
		panicHandler: panicHandler,
		backend:      smtpBackend,
		debug:        debug,
		useSSL:       useSSL,
		port:         port,
		tls:          tls,
	}

	server.server = newGoSMTPServer(server)
	server.controller = serverutil.NewController(server, eventListener)
	return server
}

func newGoSMTPServer(s *Server) *goSMTP.Server {
	newSMTP := goSMTP.NewServer(s.backend)
	newSMTP.Addr = s.Address()
	newSMTP.TLSConfig = s.tls
	newSMTP.Domain = bridge.Host
	newSMTP.ErrorLog = serverutil.NewServerErrorLogger(serverutil.SMTP)
	newSMTP.AllowInsecureAuth = true
	newSMTP.MaxLineLength = 1 << 16

	newSMTP.EnableAuth(sasl.Login, func(conn *goSMTP.Conn) sasl.Server {
		return sasl.NewLoginServer(func(address, password string) error {
			user, err := conn.Server().Backend.Login(nil, address, password)
			if err != nil {
				return err
			}

			conn.SetSession(user)
			return nil
		})
	})
	return newSMTP
}

// ListenAndServe will run server and all monitors.
func (s *Server) ListenAndServe() { s.controller.ListenAndServe() }

// Close turns off server and monitors.
func (s *Server) Close() { s.controller.Close() }

// Implements servertutil.Server interface.

func (Server) Protocol() serverutil.Protocol { return serverutil.SMTP }
func (s *Server) UseSSL() bool               { return s.useSSL }
func (s *Server) Address() string            { return fmt.Sprintf("%s:%d", bridge.Host, s.port) }
func (s *Server) TLSConfig() *tls.Config     { return s.tls }
func (s *Server) HandlePanic()               { s.panicHandler.HandlePanic() }

func (s *Server) DebugServer() bool { return s.debug }
func (s *Server) DebugClient() bool { return s.debug }

func (s *Server) SetLoggers(localDebug, remoteDebug io.Writer) { s.server.Debug = localDebug }

func (s *Server) DisconnectUser(address string) {
	log.Info("Disconnecting all open SMTP connections for ", address)
	s.server.ForEachConn(func(conn *goSMTP.Conn) {
		connUser := conn.Session()
		if connUser != nil {
			if err := conn.Close(); err != nil {
				log.WithError(err).Error("Failed to close the connection")
			}
		}
	})
}

func (s *Server) Serve(l net.Listener) error { return s.server.Serve(l) }
func (s *Server) StopServe() error           { return s.server.Close() }
