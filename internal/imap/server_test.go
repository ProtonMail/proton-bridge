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
	"fmt"
	"testing"

	"github.com/ProtonMail/proton-bridge/internal/bridge"
	"github.com/ProtonMail/proton-bridge/internal/config/useragent"
	"github.com/ProtonMail/proton-bridge/internal/serverutil/mocks"
	imapserver "github.com/emersion/go-imap/server"

	"github.com/stretchr/testify/require"
)

func TestIMAPServerTurnOffAndOnAgain(t *testing.T) {
	r := require.New(t)
	ts := mocks.NewTestServer(12345)

	server := imapserver.New(nil)
	server.Addr = fmt.Sprintf("%v:%v", bridge.Host, ts.WantPort)

	s := &imapServer{
		panicHandler:  ts.PanicHandler,
		server:        server,
		port:          ts.WantPort,
		eventListener: ts.EventListener,
		userAgent:     useragent.New(),
	}
	s.isRunning.Store(false)

	r.True(ts.IsPortFree())

	go s.ListenAndServe()
	ts.RunServerTests(r)
}
