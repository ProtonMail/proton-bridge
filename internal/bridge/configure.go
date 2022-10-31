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

import (
	"strings"

	"github.com/ProtonMail/proton-bridge/v2/internal/clientconfig"
	"github.com/ProtonMail/proton-bridge/v2/internal/config/settings"
	"github.com/ProtonMail/proton-bridge/v2/internal/config/useragent"
)

func (b *Bridge) ConfigureAppleMail(userID, address string) (bool, error) {
	user, err := b.GetUser(userID)
	if err != nil {
		return false, err
	}

	if address == "" {
		address = user.GetPrimaryAddress()
	}

	username := address
	addresses := address

	if user.IsCombinedAddressMode() {
		username = user.GetPrimaryAddress()
		addresses = strings.Join(user.GetAddresses(), ",")
	}

	var (
		restart = false
		smtpSSL = b.settings.GetBool(settings.SMTPSSLKey)
	)

	// If configuring apple mail for Catalina or newer, users should use SSL.
	if useragent.IsCatalinaOrNewer() && !smtpSSL {
		smtpSSL = true
		restart = true
		b.settings.SetBool(settings.SMTPSSLKey, true)
	}

	if err := (&clientconfig.AppleMail{}).Configure(
		Host,
		b.settings.GetInt(settings.IMAPPortKey),
		b.settings.GetInt(settings.SMTPPortKey),
		false, smtpSSL,
		username, addresses,
		user.GetBridgePassword(),
	); err != nil {
		return false, err
	}

	return restart, nil
}
