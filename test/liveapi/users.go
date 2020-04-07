// Copyright (c) 2020 Proton Technologies AG
//
// This file is part of ProtonMail Bridge.Bridge.
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

package liveapi

import (
	"github.com/ProtonMail/bridge/pkg/pmapi"
	"github.com/cucumber/godog"
	"github.com/pkg/errors"
)

func (cntrl *Controller) AddUser(user *pmapi.User, addresses *pmapi.AddressList, password string, twoFAEnabled bool) error {
	if twoFAEnabled {
		return godog.ErrPending
	}

	client := cntrl.clientManager.GetClient(user.ID)

	authInfo, err := client.AuthInfo(user.Name)
	if err != nil {
		return errors.Wrap(err, "failed to get auth info")
	}
	auth, err := client.Auth(user.Name, password, authInfo)
	if err != nil {
		return errors.Wrap(err, "failed to auth user")
	}

	mailboxPassword, err := pmapi.HashMailboxPassword(password, auth.KeySalt)
	if err != nil {
		return errors.Wrap(err, "failed to hash mailbox password")
	}
	if _, err := client.Unlock(mailboxPassword); err != nil {
		return errors.Wrap(err, "failed to unlock user")
	}
	if err := client.UnlockAddresses([]byte(mailboxPassword)); err != nil {
		return errors.Wrap(err, "failed to unlock addresses")
	}

	if err := cleanup(client); err != nil {
		return errors.Wrap(err, "failed to clean user")
	}

	return nil
}
