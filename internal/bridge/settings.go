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

package bridge

import "github.com/ProtonMail/proton-bridge/v2/internal/config/settings"

func (b *Bridge) Get(key settings.Key) string {
	return b.settings.Get(key)
}

func (b *Bridge) Set(key settings.Key, value string) {
	b.settings.Set(key, value)
}

func (b *Bridge) GetBool(key settings.Key) bool {
	return b.settings.GetBool(key)
}

func (b *Bridge) SetBool(key settings.Key, value bool) {
	b.settings.SetBool(key, value)
}

func (b *Bridge) GetInt(key settings.Key) int {
	return b.settings.GetInt(key)
}

func (b *Bridge) SetInt(key settings.Key, value int) {
	b.settings.SetInt(key, value)
}
