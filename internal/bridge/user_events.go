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
	"context"
	"fmt"

	"github.com/ProtonMail/proton-bridge/v2/internal/events"
	"github.com/ProtonMail/proton-bridge/v2/internal/safe"
	"github.com/ProtonMail/proton-bridge/v2/internal/user"
	"github.com/ProtonMail/proton-bridge/v2/internal/vault"
)

func (bridge *Bridge) handleUserEvent(ctx context.Context, user *user.User, event events.Event) error {
	switch event := event.(type) {
	case events.UserAddressCreated:
		if err := bridge.handleUserAddressCreated(ctx, user, event); err != nil {
			return fmt.Errorf("failed to handle user address created event: %w", err)
		}

	case events.UserAddressUpdated:
		if err := bridge.handleUserAddressUpdated(ctx, user, event); err != nil {
			return fmt.Errorf("failed to handle user address updated event: %w", err)
		}

	case events.UserAddressDeleted:
		if err := bridge.handleUserAddressDeleted(ctx, user, event); err != nil {
			return fmt.Errorf("failed to handle user address deleted event: %w", err)
		}

	case events.UserDeauth:
		safe.Lock(func() {
			defer delete(bridge.users, user.ID())

			bridge.logoutUser(ctx, user, false)
		}, &bridge.usersLock)
	}

	return nil
}

func (bridge *Bridge) handleUserAddressCreated(ctx context.Context, user *user.User, event events.UserAddressCreated) error {
	switch user.GetAddressMode() {
	case vault.CombinedMode:
		for addrID, gluonID := range user.GetGluonIDs() {
			if err := bridge.imapServer.RemoveUser(ctx, gluonID, false); err != nil {
				return fmt.Errorf("failed to remove user from IMAP server: %w", err)
			}

			if err := bridge.imapServer.LoadUser(ctx, user.NewIMAPConnector(addrID), gluonID, user.GluonKey()); err != nil {
				return fmt.Errorf("failed to add user to IMAP server: %w", err)
			}
		}

	case vault.SplitMode:
		gluonID, err := bridge.imapServer.AddUser(ctx, user.NewIMAPConnector(event.AddressID), user.GluonKey())
		if err != nil {
			return fmt.Errorf("failed to add user to IMAP server: %w", err)
		}

		if err := user.SetGluonID(event.AddressID, gluonID); err != nil {
			return fmt.Errorf("failed to set gluon ID: %w", err)
		}
	}

	return nil
}

// GODT-1948: Handle addresses that have been disabled!
func (bridge *Bridge) handleUserAddressUpdated(_ context.Context, user *user.User, _ events.UserAddressUpdated) error {
	switch user.GetAddressMode() {
	case vault.CombinedMode:
		return fmt.Errorf("not implemented")

	case vault.SplitMode:
		return fmt.Errorf("not implemented")
	}

	return nil
}

func (bridge *Bridge) handleUserAddressDeleted(ctx context.Context, user *user.User, event events.UserAddressDeleted) error {
	switch user.GetAddressMode() {
	case vault.CombinedMode:
		for addrID, gluonID := range user.GetGluonIDs() {
			if err := bridge.imapServer.RemoveUser(ctx, gluonID, false); err != nil {
				return fmt.Errorf("failed to remove user from IMAP server: %w", err)
			}

			if err := bridge.imapServer.LoadUser(ctx, user.NewIMAPConnector(addrID), gluonID, user.GluonKey()); err != nil {
				return fmt.Errorf("failed to add user to IMAP server: %w", err)
			}
		}

	case vault.SplitMode:
		gluonID, ok := user.GetGluonID(event.AddressID)
		if !ok {
			return fmt.Errorf("gluon ID not found for address %s", event.AddressID)
		}

		if err := bridge.imapServer.RemoveUser(ctx, gluonID, true); err != nil {
			return fmt.Errorf("failed to remove user from IMAP server: %w", err)
		}
	}

	return nil
}
