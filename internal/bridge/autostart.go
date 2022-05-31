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

// Package bridge provides core functionality of Bridge app.
package bridge

import "github.com/ProtonMail/proton-bridge/v2/internal/config/settings"

// IsAutostartEnabled checks if link file exits.
func (b *Bridge) IsAutostartEnabled() bool {
	return b.autostart.IsEnabled()
}

// EnableAutostart creates link and sets the preferences.
func (b *Bridge) EnableAutostart() error {
	b.settings.SetBool(settings.AutostartKey, true)
	return b.autostart.Enable()
}

// DisableAutostart removes link and sets the preferences.
func (b *Bridge) DisableAutostart() error {
	b.settings.SetBool(settings.AutostartKey, false)
	return b.autostart.Disable()
}
