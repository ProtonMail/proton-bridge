// Copyright (c) 2024 Proton AG
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
// along with Proton Mail Bridge.  If not, see <https://www.gnu.org/licenses/>.

package bridge

import (
	"context"
	"crypto/tls"

	"github.com/ProtonMail/proton-bridge/v3/internal/identifier"
)

func (bridge *Bridge) restartSMTP(ctx context.Context) error {
	return bridge.serverManager.RestartSMTP(ctx)
}

type bridgeSMTPSettings struct {
	b *Bridge
}

func (b *bridgeSMTPSettings) TLSConfig() *tls.Config {
	return b.b.tlsConfig
}

func (b *bridgeSMTPSettings) Log() bool {
	return b.b.logSMTP
}

func (b *bridgeSMTPSettings) Port() int {
	return b.b.vault.GetSMTPPort()
}

func (b *bridgeSMTPSettings) SetPort(i int) error {
	return b.b.vault.SetSMTPPort(i)
}

func (b *bridgeSMTPSettings) UseSSL() bool {
	return b.b.vault.GetSMTPSSL()
}

func (b *bridgeSMTPSettings) Identifier() identifier.UserAgentUpdater {
	return &bridgeUserAgentUpdater{Bridge: b.b}
}
