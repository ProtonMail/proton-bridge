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

package test

import (
	"crypto/tls"
	"fmt"
	"io"
	"net"
	"net/http"

	"github.com/ProtonMail/proton-bridge/v2/internal/serverutil"
	"github.com/ProtonMail/proton-bridge/v2/pkg/ports"
)

func newTestServer() *testServer {
	return &testServer{port: 11188}
}

type testServer struct {
	http http.Server

	useSSL,
	debugServer,
	debugClient bool
	calledDisconnected int

	port int
	tls  *tls.Config

	localDebug, remoteDebug io.Writer
}

func (*testServer) Protocol() serverutil.Protocol { return serverutil.HTTP }
func (s *testServer) UseSSL() bool                { return s.useSSL }
func (s *testServer) Address() string             { return fmt.Sprintf("127.0.0.1:%d", s.port) }
func (s *testServer) TLSConfig() *tls.Config      { return s.tls }
func (s *testServer) HandlePanic()                {}

func (s *testServer) DebugServer() bool { return s.debugServer }
func (s *testServer) DebugClient() bool { return s.debugClient }
func (s *testServer) SetLoggers(localDebug, remoteDebug io.Writer) {
	s.localDebug = localDebug
	s.remoteDebug = remoteDebug
}

func (s *testServer) DisconnectUser(string) {
	s.calledDisconnected++
}

func (s *testServer) Serve(l net.Listener) error {
	return s.http.Serve(l)
}

func (s *testServer) StopServe() error { return s.http.Close() }

func (s *testServer) portIsFree() bool {
	return ports.IsPortFree(s.port)
}

func (s *testServer) portIsOccupied() bool {
	return !ports.IsPortFree(s.port)
}

func (s *testServer) ping() error {
	client := &http.Client{}
	resp, err := client.Get("http://" + s.Address() + "/ping")
	if err != nil {
		return err
	}

	return resp.Body.Close()
}
