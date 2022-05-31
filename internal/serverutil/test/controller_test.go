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
	"net/http"
	"testing"
	"time"

	"github.com/ProtonMail/proton-bridge/v2/internal/events"
	"github.com/ProtonMail/proton-bridge/v2/internal/serverutil"
	"github.com/ProtonMail/proton-bridge/v2/pkg/listener"
	"github.com/stretchr/testify/require"
)

func setup(t *testing.T) (*require.Assertions, *testServer, listener.Listener, serverutil.Controller) {
	r := require.New(t)
	s := newTestServer()
	l := listener.New()
	c := serverutil.NewController(s, l)

	return r, s, l, c
}

func TestControllerListernServeClose(t *testing.T) {
	r, s, l, c := setup(t)

	errorCh := l.ProvideChannel(events.ErrorEvent)

	r.True(s.portIsFree())
	go c.ListenAndServe()
	r.Eventually(s.portIsOccupied, time.Second, 50*time.Millisecond)

	r.NoError(s.ping())

	r.Nil(s.localDebug)
	r.Nil(s.remoteDebug)

	c.Close()
	r.Eventually(s.portIsFree, time.Second, 50*time.Millisecond)

	select {
	case msg := <-errorCh:
		r.Fail("Expected no error but have %q", msg)
	case <-time.Tick(100 * time.Millisecond):
		break
	}
}

func TestControllerFailOnBusyPort(t *testing.T) {
	r, s, l, c := setup(t)

	ocupator := http.Server{Addr: s.Address()}
	defer ocupator.Close() //nolint:errcheck

	go ocupator.ListenAndServe() //nolint:errcheck
	r.Eventually(s.portIsOccupied, time.Second, 50*time.Millisecond)

	errorCh := l.ProvideChannel(events.ErrorEvent)
	go c.ListenAndServe()

	r.Eventually(s.portIsOccupied, time.Second, 50*time.Millisecond)

	select {
	case <-errorCh:
		break
	case <-time.Tick(time.Second):
		r.Fail("Expected error but have none.")
	}
}

func TestControllerCallDisconnectUser(t *testing.T) {
	r, s, l, c := setup(t)

	go c.ListenAndServe()
	r.Eventually(s.portIsOccupied, time.Second, 50*time.Millisecond)
	r.NoError(s.ping())

	l.Emit(events.CloseConnectionEvent, "")
	r.Eventually(func() bool { return s.calledDisconnected == 1 }, time.Second, 50*time.Millisecond)

	c.Close()
	r.Eventually(s.portIsFree, time.Second, 50*time.Millisecond)

	l.Emit(events.CloseConnectionEvent, "")
	r.Equal(1, s.calledDisconnected)
}

func TestDebugClient(t *testing.T) {
	r, s, _, c := setup(t)

	s.debugServer = false
	s.debugClient = true

	go c.ListenAndServe()
	r.Eventually(s.portIsOccupied, time.Second, 50*time.Millisecond)
	r.NoError(s.ping())

	r.Nil(s.localDebug)
	r.NotNil(s.remoteDebug)

	c.Close()
	r.Eventually(s.portIsFree, time.Second, 50*time.Millisecond)
}

func TestDebugServer(t *testing.T) {
	r, s, _, c := setup(t)

	s.debugServer = true
	s.debugClient = false

	go c.ListenAndServe()
	r.Eventually(s.portIsOccupied, time.Second, 50*time.Millisecond)
	r.NoError(s.ping())

	r.NotNil(s.localDebug)
	r.Nil(s.remoteDebug)

	c.Close()
	r.Eventually(s.portIsFree, time.Second, 50*time.Millisecond)
}

func TestDebugBoth(t *testing.T) {
	r, s, _, c := setup(t)

	s.debugServer = true
	s.debugClient = true

	go c.ListenAndServe()
	r.Eventually(s.portIsOccupied, time.Second, 50*time.Millisecond)
	r.NoError(s.ping())

	r.NotNil(s.localDebug)
	r.NotNil(s.remoteDebug)

	c.Close()
	r.Eventually(s.portIsFree, time.Second, 50*time.Millisecond)
}
