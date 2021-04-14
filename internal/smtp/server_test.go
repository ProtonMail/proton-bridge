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
	"fmt"
	"testing"
	"time"

	"github.com/ProtonMail/proton-bridge/internal/bridge"
	"github.com/ProtonMail/proton-bridge/internal/events"
	"github.com/ProtonMail/proton-bridge/pkg/listener"
	"github.com/ProtonMail/proton-bridge/pkg/ports"
	goSMTP "github.com/emersion/go-smtp"

	"github.com/stretchr/testify/require"
)

type testPanicHandler struct{}

func (ph *testPanicHandler) HandlePanic() {}

func TestSMTPServerTurnOffAndOnAgain(t *testing.T) {
	panicHandler := &testPanicHandler{}

	eventListener := listener.New()

	port := ports.FindFreePortFrom(12345)
	server := goSMTP.NewServer(nil)
	server.Addr = fmt.Sprintf("%v:%v", bridge.Host, port)

	s := &smtpServer{
		panicHandler:  panicHandler,
		server:        server,
		eventListener: eventListener,
	}
	s.isRunning.Store(false)

	go s.ListenAndServe()
	time.Sleep(5 * time.Second)
	require.False(t, ports.IsPortFree(port))

	eventListener.Emit(events.InternetOffEvent, "")
	time.Sleep(10 * time.Second)
	require.True(t, ports.IsPortFree(port))

	eventListener.Emit(events.InternetOnEvent, "")
	time.Sleep(10 * time.Second)
	require.False(t, ports.IsPortFree(port))
}
