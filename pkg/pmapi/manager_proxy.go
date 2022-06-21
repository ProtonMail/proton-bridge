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

package pmapi

// AllowProxy allows the client manager to switch clients over to a proxy if need be.
func (m *manager) AllowProxy() {
	if m.proxyDialer != nil {
		m.proxyDialer.AllowProxy()
	}
}

// DisallowProxy prevents the client manager from switching clients over to a proxy if need be.
func (m *manager) DisallowProxy() {
	if m.proxyDialer != nil {
		m.proxyDialer.DisallowProxy()
	}
}
