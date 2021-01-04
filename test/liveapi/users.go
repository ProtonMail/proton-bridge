// Copyright (c) 2021 Proton Technologies AG
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
	"github.com/ProtonMail/proton-bridge/pkg/pmapi"
	"github.com/cucumber/godog"
	"github.com/pkg/errors"
)

func (ctl *Controller) AddUser(user *pmapi.User, addresses *pmapi.AddressList, password string, twoFAEnabled bool) error {
	if twoFAEnabled {
		return godog.ErrPending
	}

	client := ctl.clientManager.GetClient(user.ID)

	authInfo, err := client.AuthInfo(user.Name)
	if err != nil {
		return errors.Wrap(err, "failed to get auth info")
	}

	_, err = client.Auth(user.Name, password, authInfo)
	if err != nil {
		return errors.Wrap(err, "failed to auth user")
	}

	salt, err := client.AuthSalt()
	if err != nil {
		return errors.Wrap(err, "failed to get salt")
	}

	mailboxPassword, err := pmapi.HashMailboxPassword(password, salt)
	if err != nil {
		return errors.Wrap(err, "failed to hash mailbox password")
	}

	if err := client.Unlock([]byte(mailboxPassword)); err != nil {
		return errors.Wrap(err, "failed to unlock user")
	}

	if err := cleanup(client, addresses); err != nil {
		return errors.Wrap(err, "failed to clean user")
	}

	ctl.pmapiByUsername[user.Name] = client

	return nil
}

func (ctl *Controller) ReorderAddresses(user *pmapi.User, addressIDs []string) error {
	client := ctl.clientManager.GetClient(user.ID)

	return client.ReorderAddresses(addressIDs)
}
