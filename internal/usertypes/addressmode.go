// Copyright (c) 2023 Proton AG
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

package usertypes

import "github.com/ProtonMail/proton-bridge/v3/internal/vault"

type AddressMode int

const (
	AddressModeCombined AddressMode = iota
	AddressModeSplit
)

func VaultToAddressMode(mode vault.AddressMode) AddressMode {
	var smtpAddressMode AddressMode

	if mode == vault.SplitMode {
		smtpAddressMode = AddressModeSplit
	} else {
		smtpAddressMode = AddressModeCombined
	}

	return smtpAddressMode
}

func AddressModeToVault(mode AddressMode) vault.AddressMode {
	if mode == AddressModeCombined {
		return vault.CombinedMode
	}

	return vault.SplitMode
}
