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

package bridge

import (
	"context"
	"strings"

	"github.com/ProtonMail/proton-bridge/v3/internal/clientconfig"
	"github.com/ProtonMail/proton-bridge/v3/internal/constants"
	"github.com/ProtonMail/proton-bridge/v3/internal/logging"
	"github.com/ProtonMail/proton-bridge/v3/internal/safe"
	"github.com/ProtonMail/proton-bridge/v3/internal/useragent"
	"github.com/ProtonMail/proton-bridge/v3/internal/vault"
	"github.com/sirupsen/logrus"
)

// ConfigureAppleMail configures apple mail for the given userID and address.
// If configuring apple mail for Catalina or newer, it ensures Bridge is using SSL.
func (bridge *Bridge) ConfigureAppleMail(ctx context.Context, userID, address string) error {
	logrus.WithFields(logrus.Fields{
		"userID":  userID,
		"address": logging.Sensitive(address),
	}).Info("Configuring Apple Mail")

	return safe.RLockRet(func() error {
		user, ok := bridge.users[userID]
		if !ok {
			return ErrNoSuchUser
		}

		if address == "" {
			address = user.Emails()[0]
		}

		username := address
		addresses := address

		if user.GetAddressMode() == vault.CombinedMode {
			username = user.Emails()[0]
			addresses = strings.Join(user.Emails(), ",")
		}

		if useragent.IsCatalinaOrNewer() && !bridge.vault.GetSMTPSSL() {
			if err := bridge.SetSMTPSSL(ctx, true); err != nil {
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
	}, bridge.usersLock)
}
