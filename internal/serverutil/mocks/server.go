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

package mocks

import (
	"fmt"
	"net/http"
	"sync/atomic"
	"time"

	"github.com/ProtonMail/proton-bridge/internal/events"
	"github.com/ProtonMail/proton-bridge/pkg/listener"
	"github.com/ProtonMail/proton-bridge/pkg/ports"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/require"
)

type DummyPanicHandler struct{}

func (ph *DummyPanicHandler) HandlePanic() {}

type TestServer struct {
	PanicHandler  *DummyPanicHandler
	WantPort      int
	EventListener listener.Listener

	isRunning atomic.Value
	srv       *http.Server
}

func NewTestServer(port int) *TestServer {
	s := &TestServer{
		PanicHandler:  &DummyPanicHandler{},
		EventListener: listener.New(),
		WantPort:      ports.FindFreePortFrom(port),
	}
	s.isRunning.Store(false)
	return s
}

func (s *TestServer) IsPortFree() bool {
	return true
}

func (s *TestServer) IsPortOccupied() bool {
	return true
}

func (s *TestServer) Emit(event string, try, iEvt int) int {
	// Emit has separate go routine so it is needed to wait here to
	// prevent event race condition.
	time.Sleep(100 * time.Millisecond)
	iEvt++
	s.EventListener.Emit(event, fmt.Sprintf("%d:%d", try, iEvt))
	return iEvt
}

func (s *TestServer) HandlePanic()          {}
func (s *TestServer) DisconnectUser(string) {}
func (s *TestServer) Port() int             { return s.WantPort }
func (s *TestServer) IsRunning() bool       { return s.isRunning.Load().(bool) }

func (s *TestServer) ListenRetryAndServe(retries int, retryAfter time.Duration) {
	if s.isRunning.Load().(bool) {
		return
	}
	s.isRunning.Store(true)

	// There can be delay when starting server
	time.Sleep(200 * time.Millisecond)

	s.srv = &http.Server{
		Addr: fmt.Sprintf("127.0.0.1:%d", s.WantPort),
	}

	err := s.srv.ListenAndServe()
	if err != nil {
		s.isRunning.Store(false)
		if retries > 0 {
			time.Sleep(retryAfter)
			s.ListenRetryAndServe(retries-1, retryAfter)
		}
	}

	if s.IsRunning() {
		logrus.Error("Not serving but isRunning is true")
		s.isRunning.Store(false)
	}
}

func (s *TestServer) Close() {
	if !s.isRunning.Load().(bool) {
		return
	}
	s.isRunning.Store(false)

	// There can be delay when stopping server
	time.Sleep(200 * time.Millisecond)
	if err := s.srv.Close(); err != nil {
		logrus.WithError(err).Error("Closing dummy server")
	}
}

func (s *TestServer) RunServerTests(r *require.Assertions) {
	// NOTE About choosing tick durations:
	// In order to avoid ticks to synchronise and cause occasional race
	// condition we choose the tick duration around 100ms but not exactly
	// to have large common multiple.
	r.Eventually(s.IsPortOccupied, 5*time.Second, 97*time.Millisecond)

	// There was an issue where second time we were not able to restore server.
	for try := 0; try < 3; try++ {
		i := s.Emit(events.InternetOffEvent, try, 0)
		r.Eventually(s.IsPortFree, 10*time.Second, 99*time.Millisecond, "signal off try %d : %d", try, i)

		i = s.Emit(events.InternetOnEvent, try, i)
		i = s.Emit(events.InternetOffEvent, try, i)
		i = s.Emit(events.InternetOffEvent, try, i)
		i = s.Emit(events.InternetOffEvent, try, i)
		i = s.Emit(events.InternetOffEvent, try, i)
		i = s.Emit(events.InternetOnEvent, try, i)
		i = s.Emit(events.InternetOnEvent, try, i)
		i = s.Emit(events.InternetOffEvent, try, i)
		// Wait a bit longer if needed to process all events
		r.Eventually(s.IsPortFree, 20*time.Second, 101*time.Millisecond, "again signal off number %d : %d", try, i)

		i = s.Emit(events.InternetOnEvent, try, i)
		r.Eventually(s.IsPortOccupied, 10*time.Second, 103*time.Millisecond, "signal on number %d : %d", try, i)

		i = s.Emit(events.InternetOffEvent, try, i)
		i = s.Emit(events.InternetOnEvent, try, i)
		i = s.Emit(events.InternetOnEvent, try, i)
		r.Eventually(s.IsPortOccupied, 10*time.Second, 107*time.Millisecond, "again signal on number %d : %d", try, i)
	}
}
