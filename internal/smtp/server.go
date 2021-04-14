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
	"github.com/ProtonMail/proton-bridge/pkg/listener"
	"github.com/ProtonMail/proton-bridge/pkg/ports"
	"github.com/emersion/go-sasl"
	goSMTP "github.com/emersion/go-smtp"
	"github.com/sirupsen/logrus"
)

type smtpServer struct {
	panicHandler  panicHandler
	server        *goSMTP.Server
	eventListener listener.Listener
	useSSL        bool
	port          int
	isRunning     atomic.Value
}

// NewSMTPServer returns an SMTP server configured with the given options.
func NewSMTPServer(panicHandler panicHandler, debug bool, port int, useSSL bool, tls *tls.Config, smtpBackend goSMTP.Backend, eventListener listener.Listener) *smtpServer { //nolint[golint]
	s := goSMTP.NewServer(smtpBackend)
	s.Addr = fmt.Sprintf("%v:%v", bridge.Host, port)
	s.TLSConfig = tls
	s.Domain = bridge.Host
	s.AllowInsecureAuth = true
	s.MaxLineLength = 2 << 16

	if debug {
		fmt.Println("THE LOG WILL CONTAIN **DECRYPTED** MESSAGE DATA")
		log.Warning("================================================")
		log.Warning("THIS LOG WILL CONTAIN **DECRYPTED** MESSAGE DATA")
		log.Warning("================================================")
	}

	if debug {
		s.Debug = logrus.
			WithField("pkg", "smtp/server").
			WriterLevel(logrus.DebugLevel)
	}

	s.EnableAuth(sasl.Login, func(conn *goSMTP.Conn) sasl.Server {
		return sasl.NewLoginServer(func(address, password string) error {
			user, err := conn.Server().Backend.Login(nil, address, password)
			if err != nil {
				return err
			}

			conn.SetSession(user)
			return nil
		})
	})

	server := &smtpServer{
		panicHandler:  panicHandler,
		server:        s,
		eventListener: eventListener,
		useSSL:        useSSL,
		port:          port,
	}
	server.isRunning.Store(false)
	return server
}

// Starts the server.
func (s *smtpServer) ListenAndServe() {
	go s.monitorDisconnectedUsers()
	go s.monitorInternetConnection()

	// When starting the Bridge, we don't want to retry to notify user
	// quickly about the issue. Very probably retry will not help anyway.
	s.listenAndServe(0)
}

func (s *smtpServer) listenAndServe(retries int) {
	if s.isRunning.Load().(bool) {
		return
	}
	s.isRunning.Store(true)

	l := log.WithField("useSSL", s.useSSL).WithField("address", s.server.Addr)
	l.Info("SMTP server is starting")
	var listener net.Listener
	var err error
	if s.useSSL {
		listener, err = tls.Listen("tcp", s.server.Addr, s.server.TLSConfig)
	} else {
		listener, err = net.Listen("tcp", s.server.Addr)
	}
	if err != nil {
		s.isRunning.Store(false)
		if retries > 0 {
			l.WithError(err).WithField("retries", retries).Warn("SMTP listener failed")
			time.Sleep(15 * time.Second)
			s.listenAndServe(retries - 1)
			return
		}

		l.WithError(err).Error("SMTP listener failed")
		s.eventListener.Emit(events.ErrorEvent, "SMTP failed: "+err.Error())
		return
	}

	err = s.server.Serve(listener)
	// Serve returns error every time, even after closing the server.
	// User shouldn't be notified about error if server shouldn't be running,
	// but it should in case it was not closed by `s.Close()`.
	if err != nil && s.isRunning.Load().(bool) {
		s.isRunning.Store(false)
		l.WithError(err).Error("SMTP server failed")
		s.eventListener.Emit(events.ErrorEvent, "SMTP failed: "+err.Error())
		return
	}
	defer s.server.Close() //nolint[errcheck]

	l.Info("SMTP server stopped")
}

// Stops the server.
func (s *smtpServer) Close() {
	if !s.isRunning.Load().(bool) {
		return
	}
	s.isRunning.Store(false)

	if err := s.server.Close(); err != nil {
		log.WithError(err).Error("Failed to close the connection")
	}
}

func (s *smtpServer) monitorInternetConnection() {
	on := make(chan string)
	s.eventListener.Add(events.InternetOnEvent, on)
	off := make(chan string)
	s.eventListener.Add(events.InternetOffEvent, off)

	for {
		var expectedIsPortFree bool
		select {
		case <-on:
			go func() {
				defer s.panicHandler.HandlePanic()
				// We had issues on Mac that from time to time something
				// blocked our port for a bit after we closed IMAP server
				// due to connection issues.
				// Restart always helped, so we do retry to not bother user.
				s.listenAndServe(10)
			}()
			expectedIsPortFree = false
		case <-off:
			s.Close()
			expectedIsPortFree = true
		}

		start := time.Now()
		for {
			if ports.IsPortFree(s.port) == expectedIsPortFree {
				break
			}
			// Safety stop if something went wrong.
			if time.Since(start) > 15*time.Second {
				log.WithField("expectedIsPortFree", expectedIsPortFree).Warn("Server start/stop check timeouted")
				break
			}
			time.Sleep(100 * time.Millisecond)
		}
	}
}

func (s *smtpServer) monitorDisconnectedUsers() {
	ch := make(chan string)
	s.eventListener.Add(events.CloseConnectionEvent, ch)

	for address := range ch {
		log.Info("Disconnecting all open SMTP connections for ", address)
		disconnectUser := func(conn *goSMTP.Conn) {
			connUser := conn.Session()
			if connUser != nil {
				if err := conn.Close(); err != nil {
					log.WithError(err).Error("Failed to close the connection")
				}
			}
		}
		s.server.ForEachConn(disconnectUser)
	}
}
