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

import "github.com/sirupsen/logrus"

func (bridge *Bridge) GetCurrentUserAgent() string {
	return bridge.identifier.GetUserAgent()
}

func (bridge *Bridge) SetCurrentPlatform(platform string) {
	bridge.identifier.SetPlatform(platform)
}

func (bridge *Bridge) setUserAgent(name, version string) {
	currentUserAgent := bridge.identifier.GetClientString()

	bridge.identifier.SetClient(name, version)

	newUserAgent := bridge.identifier.GetClientString()

	if currentUserAgent != newUserAgent {
		if err := bridge.vault.SetLastUserAgent(newUserAgent); err != nil {
			logrus.WithError(err).Error("Failed to write new user agent to vault")
		}
	}
}

type bridgeUserAgentUpdater struct {
	*Bridge
}

func (b *bridgeUserAgentUpdater) GetUserAgent() string {
	return b.identifier.GetUserAgent()
}

func (b *bridgeUserAgentUpdater) HasClient() bool {
	return b.identifier.HasClient()
}

func (b *bridgeUserAgentUpdater) SetClient(name, version string) {
	b.identifier.SetClient(name, version)
}

func (b *bridgeUserAgentUpdater) SetPlatform(platform string) {
	b.identifier.SetPlatform(platform)
}

func (b *bridgeUserAgentUpdater) SetClientString(client string) {
	b.identifier.SetClientString(client)
}

func (b *bridgeUserAgentUpdater) GetClientString() string {
	return b.identifier.GetClientString()
}

func (b *bridgeUserAgentUpdater) SetUserAgent(name, version string) {
	b.setUserAgent(name, version)
}
