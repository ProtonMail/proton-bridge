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

package smtp

import (
	"crypto/tls"
	"fmt"

	"github.com/ProtonMail/proton-bridge/internal/bridge"
	"github.com/ProtonMail/proton-bridge/internal/events"
	"github.com/ProtonMail/proton-bridge/pkg/listener"
	"github.com/emersion/go-sasl"
	goSMTP "github.com/emersion/go-smtp"
	"github.com/sirupsen/logrus"
)

type smtpServer struct {
	server        *goSMTP.Server
	eventListener listener.Listener
	useSSL        bool
}

// NewSMTPServer returns an SMTP server configured with the given options.
func NewSMTPServer(debug bool, port int, useSSL bool, tls *tls.Config, smtpBackend goSMTP.Backend, eventListener listener.Listener) *smtpServer { //nolint[golint]
	s := goSMTP.NewServer(smtpBackend)
	s.Addr = fmt.Sprintf("%v:%v", bridge.Host, port)
	s.TLSConfig = tls
	s.Domain = bridge.Host
	s.AllowInsecureAuth = true

	if debug {
		s.Debug = logrus.
			WithField("pkg", "smtp/server").
			WriterLevel(logrus.DebugLevel)
	}

	s.EnableAuth(sasl.Login, func(conn *goSMTP.Conn) sasl.Server {
		return sasl.NewLoginServer(func(address, password string) error {
			user, err := conn.Server().Backend.Login(address, password)
			if err != nil {
				return err
			}

			conn.SetUser(user)
			return nil
		})
	})

	return &smtpServer{
		server:        s,
		eventListener: eventListener,
		useSSL:        useSSL,
	}
}

// Starts the server.
func (s *smtpServer) ListenAndServe() {
	go s.monitorDisconnectedUsers()
	l := log.WithField("useSSL", s.useSSL).WithField("address", s.server.Addr)

	l.Info("SMTP server is starting")
	var err error
	if s.useSSL {
		err = s.server.ListenAndServeTLS()
	} else {
		err = s.server.ListenAndServe()
	}
	if err != nil {
		s.eventListener.Emit(events.ErrorEvent, "SMTP failed: "+err.Error())
		l.Error("SMTP failed: ", err)
		return
	}
	defer s.server.Close()

	l.Info("SMTP server stopped")
}

// Stops the server.
func (s *smtpServer) Close() {
	s.server.Close()
}

func (s *smtpServer) monitorDisconnectedUsers() {
	ch := make(chan string)
	s.eventListener.Add(events.CloseConnectionEvent, ch)

	for address := range ch {
		log.Info("Disconnecting all open SMTP connections for ", address)
		disconnectUser := func(conn *goSMTP.Conn) {
			connUser := conn.User()
			if connUser != nil {
				_ = conn.Close()
			}
		}
		s.server.ForEachConn(disconnectUser)
	}
}
