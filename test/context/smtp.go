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
	"github.com/ProtonMail/proton-bridge/v2/internal/smtp"
	"github.com/ProtonMail/proton-bridge/v2/test/mocks"
	"github.com/stretchr/testify/require"
)

// GetSMTPClient gets the smtp client by name; if it doesn't exist yet it creates it.
func (ctx *TestContext) GetSMTPClient(handle string) *mocks.SMTPClient {
	if client, ok := ctx.smtpClients[handle]; ok {
		return client
	}
	return ctx.newSMTPClient(handle)
}

func (ctx *TestContext) newSMTPClient(handle string) *mocks.SMTPClient {
	ctx.withSMTPServer()

	client := mocks.NewSMTPClient(ctx.t, handle, ctx.smtpAddr)
	ctx.smtpClients[handle] = client
	ctx.addCleanup(client.Close, "Closing SMTP client")
	return client
}

// withSMTPServer starts an smtp server and connects it to the bridge instance.
// Every TestContext has this by default and thus this doesn't need to be exported.
func (ctx *TestContext) withSMTPServer() {
	if ctx.smtpServer != nil {
		return
	}

	settingsPath, _ := ctx.locations.ProvideSettingsPath()
	ph := newPanicHandler(ctx.t)
	tls, _ := tls.New(settingsPath).GetConfig()
	port := ctx.settings.GetInt(settings.SMTPPortKey)
	useSSL := ctx.settings.GetBool(settings.SMTPSSLKey)

	backend := smtp.NewSMTPBackend(ph, ctx.listener, ctx.settings, ctx.bridge)
	server := smtp.NewSMTPServer(ph, true, port, useSSL, tls, backend, ctx.listener)

	go server.ListenAndServe()
	require.NoError(ctx.t, waitForPort(port, 5*time.Second))

	ctx.smtpServer = server
	ctx.smtpAddr = fmt.Sprintf("%v:%v", bridge.Host, port)
	ctx.addCleanup(ctx.smtpServer.Close, "Closing SMTP server")
}

// SetSMTPLastResponse sets the last SMTP response that was received.
func (ctx *TestContext) SetSMTPLastResponse(handle string, resp *mocks.SMTPResponse) {
	ctx.smtpResponseLocker.Lock()
	defer ctx.smtpResponseLocker.Unlock()

	ctx.smtpLastResponses[handle] = resp
}

// GetSMTPLastResponse returns the last IMAP response that was received.
func (ctx *TestContext) GetSMTPLastResponse(handle string) *mocks.SMTPResponse {
	ctx.smtpResponseLocker.Lock()
	defer ctx.smtpResponseLocker.Unlock()

	return ctx.smtpLastResponses[handle]
}
