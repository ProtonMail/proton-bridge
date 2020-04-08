// Copyright (c) 2020 Proton Technologies AG
//
// This file is part of ProtonMail Bridge.Bridge.
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

package context

import (
	"fmt"
	"time"

	"github.com/ProtonMail/proton-bridge/internal/bridge"
	"github.com/ProtonMail/proton-bridge/internal/imap"
	"github.com/ProtonMail/proton-bridge/internal/preferences"
	"github.com/ProtonMail/proton-bridge/pkg/config"
	"github.com/ProtonMail/proton-bridge/test/mocks"
	"github.com/stretchr/testify/require"
)

// GetIMAPClient gets the imap client by name; if it doesn't exist yet it creates it.
func (ctx *TestContext) GetIMAPClient(handle string) *mocks.IMAPClient {
	if client, ok := ctx.imapClients[handle]; ok {
		return client
	}
	return ctx.newIMAPClient(handle)
}

func (ctx *TestContext) newIMAPClient(handle string) *mocks.IMAPClient {
	ctx.withIMAPServer()

	client := mocks.NewIMAPClient(ctx.t, handle, ctx.imapAddr)
	ctx.imapClients[handle] = client
	ctx.addCleanup(client.Close, "Closing IMAP client")
	return client
}

// withIMAPServer starts an imap server and connects it to the bridge instance.
// Every TestContext has this by default and thus this doesn't need to be exported.
func (ctx *TestContext) withIMAPServer() {
	if ctx.imapServer != nil {
		return
	}

	ph := newPanicHandler(ctx.t)
	pref := preferences.New(ctx.cfg)
	port := pref.GetInt(preferences.IMAPPortKey)
	tls, _ := config.GetTLSConfig(ctx.cfg)

	backend := imap.NewIMAPBackend(ph, ctx.listener, ctx.cfg, ctx.bridge)
	server := imap.NewIMAPServer(true, true, port, tls, backend, ctx.listener)

	go server.ListenAndServe()
	require.NoError(ctx.t, waitForPort(port, 5*time.Second))

	ctx.imapServer = server
	ctx.imapAddr = fmt.Sprintf("%v:%v", bridge.Host, port)
	ctx.addCleanup(ctx.imapServer.Close, "Closing IMAP server")
}

// SetIMAPLastResponse sets the last IMAP response that was received.
func (ctx *TestContext) SetIMAPLastResponse(handle string, resp *mocks.IMAPResponse) {
	ctx.imapLastResponses[handle] = resp
}

// GetIMAPLastResponse returns the last IMAP response that was received.
func (ctx *TestContext) GetIMAPLastResponse(handle string) *mocks.IMAPResponse {
	return ctx.imapLastResponses[handle]
}
