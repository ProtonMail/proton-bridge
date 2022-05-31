// Copyright (c) 2022 Proton AG
//
// This file is part of Proton Mail Bridge.Bridge.
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

package context

import (
	"fmt"
	"time"

	"github.com/ProtonMail/proton-bridge/v2/internal/bridge"
	"github.com/ProtonMail/proton-bridge/v2/internal/config/settings"
	"github.com/ProtonMail/proton-bridge/v2/internal/config/tls"
	"github.com/ProtonMail/proton-bridge/v2/internal/imap"
	"github.com/ProtonMail/proton-bridge/v2/test/mocks"
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

	settingsPath, _ := ctx.locations.ProvideSettingsPath()
	ph := newPanicHandler(ctx.t)
	port := ctx.settings.GetInt(settings.IMAPPortKey)
	tls, _ := tls.New(settingsPath).GetConfig()

	backend := imap.NewIMAPBackend(ph, ctx.listener, ctx.cache, ctx.settings, ctx.bridge)
	server := imap.NewIMAPServer(ph, true, true, port, tls, backend, ctx.userAgent, ctx.listener)

	go server.ListenAndServe()
	require.NoError(ctx.t, waitForPort(port, 5*time.Second))

	ctx.imapServer = server
	ctx.imapAddr = fmt.Sprintf("%v:%v", bridge.Host, port)
	ctx.addCleanup(ctx.imapServer.Close, "Closing IMAP server")
}

// SetIMAPLastResponse sets the last IMAP response that was received.
func (ctx *TestContext) SetIMAPLastResponse(handle string, resp *mocks.IMAPResponse) {
	ctx.imapResponseLocker.Lock()
	defer ctx.imapResponseLocker.Unlock()

	ctx.imapLastResponses[handle] = resp
}

// GetIMAPLastResponse returns the last IMAP response that was received.
func (ctx *TestContext) GetIMAPLastResponse(handle string) *mocks.IMAPResponse {
	ctx.imapResponseLocker.Lock()
	defer ctx.imapResponseLocker.Unlock()

	return ctx.imapLastResponses[handle]
}
