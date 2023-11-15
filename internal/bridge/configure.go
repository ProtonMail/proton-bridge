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
	"errors"
	"strings"

	"github.com/ProtonMail/proton-bridge/v3/internal/clientconfig"
	"github.com/ProtonMail/proton-bridge/v3/internal/constants"
	"github.com/ProtonMail/proton-bridge/v3/internal/logging"
	"github.com/ProtonMail/proton-bridge/v3/internal/safe"
	userpkg "github.com/ProtonMail/proton-bridge/v3/internal/user"
	"github.com/ProtonMail/proton-bridge/v3/internal/useragent"
	"github.com/ProtonMail/proton-bridge/v3/internal/vault"
	"github.com/sirupsen/logrus"
	"golang.org/x/exp/slices"
)

// ConfigureAppleMail configures Apple Mail for the given userID and address.
// If configuring Apple Mail for Catalina or newer, it ensures Bridge is using SSL.
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

		identities := user.Identities()
		if len(identities) == 0 {
			return errors.New("could not retrieve user identities")
		}

		if address == "" {
			address = identities[0].Email
		}

		var username, displayName, addresses string
		if user.GetAddressMode() == vault.CombinedMode {
			username = identities[0].Email
			displayName = identities[0].DisplayName
			addresses = strings.Join(user.Emails(), ",")
		} else {
			username = address
			addresses = address
			index := slices.IndexFunc(identities, func(identity userpkg.Identity) bool {
				return strings.EqualFold(identity.Email, address)
			})
			if index >= 0 {
				displayName = identities[index].DisplayName
			} else {
				displayName = address
			}
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
			displayName,
			addresses,
			user.BridgePass(),
		)
	}, bridge.usersLock)
}
