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

package smtp

import (
	"crypto/tls"
	"fmt"
	"net"
	"sync/atomic"
	"time"

	"github.com/ProtonMail/proton-bridge/internal/bridge"
	"github.com/ProtonMail/proton-bridge/internal/events"
	"github.com/ProtonMail/proton-bridge/internal/serverutil"
	"github.com/ProtonMail/proton-bridge/pkg/listener"
	"github.com/emersion/go-sasl"
	goSMTP "github.com/emersion/go-smtp"
	"github.com/sirupsen/logrus"
)

// Server is Bridge SMTP server implementation.
type Server struct {
	panicHandler  panicHandler
	backend       goSMTP.Backend
	server        *goSMTP.Server
	eventListener listener.Listener
	debug         bool
	useSSL        bool
	port          int
	tls           *tls.Config
	isRunning     atomic.Value
}

// NewSMTPServer returns an SMTP server configured with the given options.
func NewSMTPServer(panicHandler panicHandler, debug bool, port int, useSSL bool, tls *tls.Config, smtpBackend goSMTP.Backend, eventListener listener.Listener) *Server {
	if debug {
		fmt.Println("THE LOG WILL CONTAIN **DECRYPTED** MESSAGE DATA")
		log.Warning("================================================")
		log.Warning("THIS LOG WILL CONTAIN **DECRYPTED** MESSAGE DATA")
		log.Warning("================================================")
	}

	server := &Server{
		panicHandler:  panicHandler,
		backend:       smtpBackend,
		eventListener: eventListener,
		debug:         debug,
		useSSL:        useSSL,
		port:          port,
		tls:           tls,
	}
	server.isRunning.Store(false)
	return server
}

func (s *Server) HandlePanic()    { s.panicHandler.HandlePanic() }
func (s *Server) IsRunning() bool { return s.isRunning.Load().(bool) }
func (s *Server) Port() int       { return s.port }

func newGoSMTPServer(debug bool, smtpBackend goSMTP.Backend, port int, tls *tls.Config) *goSMTP.Server {
	newSMTP := goSMTP.NewServer(smtpBackend)
	newSMTP.Addr = fmt.Sprintf("%v:%v", bridge.Host, port)
	newSMTP.TLSConfig = tls
	newSMTP.Domain = bridge.Host
	newSMTP.AllowInsecureAuth = true
	newSMTP.MaxLineLength = 1 << 16

	if debug {
		newSMTP.Debug = logrus.
			WithField("pkg", "smtp/server").
			WriterLevel(logrus.DebugLevel)
	}

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

// ListenAndServe starts the server and keeps it on based on internet
// availability.
func (s *Server) ListenAndServe() {
	serverutil.ListenAndServe(s, s.eventListener)
}

func (s *Server) ListenRetryAndServe(retries int, retryAfter time.Duration) {
	if s.IsRunning() {
		return
	}
	s.isRunning.Store(true)

	s.server = newGoSMTPServer(s.debug, s.backend, s.port, s.tls)

	l := log.WithField("useSSL", s.useSSL).WithField("address", s.server.Addr)
	l.Info("SMTP server is starting")

	var listener net.Listener
	var err error
	if s.useSSL {
		listener, err = tls.Listen("tcp", s.server.Addr, s.server.TLSConfig)
	} else {
		listener, err = net.Listen("tcp", s.server.Addr)
	}
	l.WithError(err).Debug("Listener for SMTP created")
	if err != nil {
		s.isRunning.Store(false)
		if retries > 0 {
			l.WithError(err).WithField("retries", retries).Warn("SMTP listener failed")
			time.Sleep(retryAfter)
			s.ListenRetryAndServe(retries-1, retryAfter)
			return
		}

		l.WithError(err).Error("SMTP listener failed")
		s.eventListener.Emit(events.ErrorEvent, "SMTP failed: "+err.Error())
		return
	}

	err = s.server.Serve(listener)
	l.WithError(err).Debug("GoSMTP not serving")
	// Serve returns error every time, even after closing the server.
	// User shouldn't be notified about error if server shouldn't be running,
	// but it should in case it was not closed by `s.Close()`.
	if err != nil && s.IsRunning() {
		s.isRunning.Store(false)
		l.WithError(err).Error("SMTP server failed")
		s.eventListener.Emit(events.ErrorEvent, "SMTP failed: "+err.Error())
		return
	}
	defer func() {
		// Go SMTP server instance can be closed only once. Otherwise
		// it returns an error. The error is not export therefore we
		// will check the string value.
		err := s.server.Close()
		if err == nil || err.Error() != "smtp: server already closed" {
			l.WithError(err).Warn("Server was not closed")
		}
	}()

	l.Info("SMTP server closed")
}

// Close stops the server.
func (s *Server) Close() {
	if !s.IsRunning() {
		return
	}
	s.isRunning.Store(false)

	if err := s.server.Close(); err != nil {
		log.WithError(err).Error("Cannot close the server")
	}
}

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
