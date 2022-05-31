// Copyright (c) 2022 Proton AG
//
// This file is part of Proton Mail Bridge.Bridge.
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

package liveapi

import (
	"context"

	"github.com/ProtonMail/proton-bridge/v2/pkg/pmapi"
	"github.com/ProtonMail/proton-bridge/v2/test/accounts"
	"github.com/cucumber/godog"
	"github.com/pkg/errors"
)

func (ctl *Controller) AddUser(account *accounts.TestAccount) error {
	if account.IsTwoFAEnabled() {
		return godog.ErrPending
	}

	client, err := addPersistentClient(account.User().Name, account.Password(), account.MailboxPassword())
	if err != nil {
		return errors.Wrap(err, "failed to add persistent client")
	}

	if err := cleanup(client, account.Addresses()); err != nil {
		return errors.Wrap(err, "failed to clean user")
	}

	return nil
}

func (ctl *Controller) ReorderAddresses(user *pmapi.User, addressIDs []string) error {
	client, err := getPersistentClient(user.Name)
	if err != nil {
		return err
	}
	return client.ReorderAddresses(context.Background(), addressIDs)
}

func (ctl *Controller) GetAuthClient(username string) pmapi.Client {
	client, err := getPersistentClient(username)
	if err != nil {
		ctl.log.WithError(err).
			WithField("username", username).
			Fatal("Cannot get authenticated client")
	}
	return client
}

func (ctl *Controller) RevokeSession(username string) error {
	return errors.New("revoke live session not implemented")
}
