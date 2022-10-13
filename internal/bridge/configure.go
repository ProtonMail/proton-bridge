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
// along with Proton Mail Bridge.  If not, see <https://www.gnu.org/licenses/>.

package bridge

import (
	"fmt"
	"strings"

	"github.com/ProtonMail/proton-bridge/v2/internal/clientconfig"
	"github.com/ProtonMail/proton-bridge/v2/internal/constants"
	"github.com/ProtonMail/proton-bridge/v2/internal/user"
	"github.com/ProtonMail/proton-bridge/v2/internal/useragent"
	"github.com/ProtonMail/proton-bridge/v2/internal/vault"
)

func (bridge *Bridge) ConfigureAppleMail(userID, address string) error {
	if ok, err := bridge.users.GetErr(userID, func(user *user.User) error {
		if address == "" {
			address = user.Emails()[0]
		}

		username := address
		addresses := address

		if user.GetAddressMode() == vault.CombinedMode {
			username = user.Emails()[0]
			addresses = strings.Join(user.Emails(), ",")
		}

		// If configuring apple mail for Catalina or newer, users should use SSL.
		if useragent.IsCatalinaOrNewer() && !bridge.vault.GetSMTPSSL() {
			if err := bridge.SetSMTPSSL(true); err != nil {
				return err
			}
		}

		return (&clientconfig.AppleMail{}).Configure(
			constants.Host,
			bridge.vault.GetIMAPPort(),
			bridge.vault.GetSMTPPort(),
			bridge.vault.GetIMAPSSL(),
			bridge.vault.GetSMTPSSL(),
			username,
			addresses,
			user.BridgePass(),
		)
	}); !ok {
		return ErrNoSuchUser
	} else if err != nil {
		return fmt.Errorf("failed to configure apple mail: %w", err)
	}

	return nil
}
