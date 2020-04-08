// Copyright (c) 2020 Proton Technologies AG
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

// Package autoconfig provides automatic config of IMAP and SMTP.
// For now only for Apple Mail.
package autoconfig

import "github.com/ProtonMail/proton-bridge/internal/frontend/types"

type AutoConfig interface {
	Name() string
	Configure(imapPort int, smtpPort int, imapSSl, smtpSSL bool, user types.BridgeUser, addressIndex int) error
}

var available []AutoConfig //nolint[gochecknoglobals]

func Available() []AutoConfig {
	return available
}
