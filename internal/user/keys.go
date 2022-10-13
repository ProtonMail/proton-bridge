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

package user

import (
	"fmt"

	"github.com/ProtonMail/gopenpgp/v2/crypto"
	"gitlab.protontech.ch/go/liteapi"
)

func (user *User) withUserKR(fn func(*crypto.KeyRing) error) error {
	return user.apiUser.LoadErr(func(apiUser liteapi.User) error {
		userKR, err := apiUser.Keys.Unlock(user.vault.KeyPass(), nil)
		if err != nil {
			return fmt.Errorf("failed to unlock user keys: %w", err)
		}
		defer userKR.ClearPrivateParams()

		return fn(userKR)
	})
}

func (user *User) withAddrKR(addrID string, fn func(*crypto.KeyRing) error) error {
	return user.withUserKR(func(userKR *crypto.KeyRing) error {
		if ok, err := user.apiAddrs.GetErr(addrID, func(apiAddr liteapi.Address) error {
			addrKR, err := apiAddr.Keys.Unlock(user.vault.KeyPass(), userKR)
			if err != nil {
				return fmt.Errorf("failed to unlock address keys: %w", err)
			}
			defer userKR.ClearPrivateParams()

			return fn(addrKR)
		}); !ok {
			return fmt.Errorf("no such address %q", addrID)
		} else if err != nil {
			return err
		}

		return nil
	})
}

func (user *User) withAddrKRs(fn func(map[string]*crypto.KeyRing) error) error {
	return user.withUserKR(func(userKR *crypto.KeyRing) error {
		return user.apiAddrs.ValuesErr(func(apiAddrs []liteapi.Address) error {
			addrKRs := make(map[string]*crypto.KeyRing)

			for _, apiAddr := range apiAddrs {
				addrKR, err := apiAddr.Keys.Unlock(user.vault.KeyPass(), userKR)
				if err != nil {
					return fmt.Errorf("failed to unlock address keys: %w", err)
				}
				defer userKR.ClearPrivateParams()

				addrKRs[apiAddr.ID] = addrKR
			}

			return fn(addrKRs)
		})
	})
}
