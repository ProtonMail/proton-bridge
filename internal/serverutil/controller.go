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
	"crypto/tls"
	"fmt"
	"net"

	"github.com/ProtonMail/proton-bridge/v2/internal/events"
	"github.com/ProtonMail/proton-bridge/v2/pkg/listener"
	"github.com/sirupsen/logrus"
)

// Controller will make sure that server is listening and serving and if needed
// users are disconnected.
type Controller interface {
	ListenAndServe()
	Close()
}

// NewController return simple server controller.
func NewController(s Server, l listener.Listener) Controller {
	log := logrus.WithField("pkg", "serverutil").WithField("protocol", s.Protocol())
	c := &controller{
		server:               s,
		signals:              l,
		log:                  log,
		closeDisconnectUsers: make(chan void),
	}

	if s.DebugServer() {
		fmt.Println("THE LOG WILL CONTAIN **DECRYPTED** MESSAGE DATA")
		log.Warning("================================================")
		log.Warning("THIS LOG WILL CONTAIN **DECRYPTED** MESSAGE DATA")
		log.Warning("================================================")
	}

	return c
}

type void struct{}

type controller struct {
	server  Server
	signals listener.Listener
	log     *logrus.Entry

	closeDisconnectUsers chan void
}

func (c *controller) Close() {
	c.closeDisconnectUsers <- void{}
	if err := c.server.StopServe(); err != nil {
		c.log.WithError(err).Error("Issue when closing server")
	}
}

// ListenAndServe starts the server and keeps it on based on internet
// availability. It also monitors and disconnect users if requested.
func (c *controller) ListenAndServe() {
	go monitorDisconnectedUsers(c.server, c.signals, c.closeDisconnectUsers)

	defer c.server.HandlePanic()

	l := c.log.WithField("useSSL", c.server.UseSSL()).
		WithField("address", c.server.Address())

	var listener net.Listener
	var err error

	if c.server.UseSSL() {
		listener, err = tls.Listen("tcp", c.server.Address(), c.server.TLSConfig())
	} else {
		listener, err = net.Listen("tcp", c.server.Address())
	}

	if err != nil {
		l.WithError(err).Error("Cannot start listner.")
		c.signals.Emit(events.ErrorEvent, string(c.server.Protocol())+" failed: "+err.Error())
		return
	}

	// When starting the Bridge, we don't want to retry to notify user
	// quickly about the issue. Very probably retry will not help anyway.
	l.Info("Starting server")
	err = c.server.Serve(&connListener{listener, c.server})
	l.WithError(err).Debug("GoSMTP not serving")
}

func monitorDisconnectedUsers(s Server, l listener.Listener, done <-chan void) {
	ch := make(chan string)
	l.Add(events.CloseConnectionEvent, ch)
	for {
		select {
		case <-done:
			return
		case address := <-ch:
			s.DisconnectUser(address)
		}
	}
}
