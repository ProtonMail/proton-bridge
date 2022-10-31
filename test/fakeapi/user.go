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

package fakeapi

import (
	"context"

	"github.com/ProtonMail/gopenpgp/v2/crypto"
	"github.com/ProtonMail/proton-bridge/v2/pkg/pmapi"
)

func (api *FakePMAPI) GetMailSettings(context.Context) (pmapi.MailSettings, error) {
	if err := api.checkAndRecordCall(GET, "/mail/v4/settings", nil); err != nil {
		return pmapi.MailSettings{}, err
	}
	return pmapi.MailSettings{}, nil
}

func (api *FakePMAPI) IsUnlocked() bool {
	return api.userKeyRing != nil
}

func (api *FakePMAPI) Unlock(_ context.Context, passphrase []byte) (err error) {
	if api.userKeyRing != nil {
		return
	}

	if api.userKeyRing, err = api.user.Keys.UnlockAll(passphrase, nil); err != nil {
		return
	}

	for _, a := range *api.addresses {
		if a.HasKeys == pmapi.MissingKeys {
			continue
		}

		if api.addrKeyRing[a.ID] != nil {
			continue
		}

		var kr *crypto.KeyRing

		if kr, err = a.Keys.UnlockAll(passphrase, api.userKeyRing); err != nil {
			return
		}

		api.addrKeyRing[a.ID] = kr
	}

	return nil
}

func (api *FakePMAPI) ReloadKeys(ctx context.Context, passphrase []byte) (err error) {
	if _, err = api.UpdateUser(ctx); err != nil {
		return
	}

	return api.Unlock(ctx, passphrase)
}

func (api *FakePMAPI) CurrentUser(ctx context.Context) (*pmapi.User, error) {
	return api.UpdateUser(ctx)
}

func (api *FakePMAPI) UpdateUser(context.Context) (*pmapi.User, error) {
	if err := api.checkAndRecordCall(GET, "/users", nil); err != nil {
		return nil, err
	}

	if err := api.checkAndRecordCall(GET, "/addresses", nil); err != nil {
		return nil, err
	}

	return api.user, nil
}

func (api *FakePMAPI) GetUser(ctx context.Context) (*pmapi.User, error) {
	if err := api.checkAndRecordCall(GET, "/users", nil); err != nil {
		return nil, err
	}

	return api.user, nil
}

func (api *FakePMAPI) GetAddresses(context.Context) (pmapi.AddressList, error) {
	if err := api.checkAndRecordCall(GET, "/addresses", nil); err != nil {
		return nil, err
	}
	return *api.addresses, nil
}

func (api *FakePMAPI) ReorderAddresses(_ context.Context, addressIDs []string) error {
	if err := api.checkAndRecordCall(PUT, "/addresses/order", nil); err != nil {
		return err
	}

	for wantedIndex, addressID := range addressIDs {
		var currentIndex int

		for i, v := range *api.addresses {
			if v.ID == addressID {
				currentIndex = i
				break
			}
		}

		(*api.addresses)[wantedIndex], (*api.addresses)[currentIndex] = (*api.addresses)[currentIndex], (*api.addresses)[wantedIndex]
		(*api.addresses)[wantedIndex].Order = wantedIndex + 1 // Starts counting from 1.
		api.addEventAddress(pmapi.EventUpdate, (*api.addresses)[wantedIndex])
	}

	return nil
}

func (api *FakePMAPI) Addresses() pmapi.AddressList {
	if api.controller.noInternetConnection {
		return nil
	}
	if api.addresses == nil {
		return pmapi.AddressList{}
	}
	return *api.addresses
}

func (api *FakePMAPI) GetUserKeyRing() (*crypto.KeyRing, error) {
	return api.userKeyRing, nil
}

func (api *FakePMAPI) KeyRingForAddressID(addrID string) (*crypto.KeyRing, error) {
	return api.addrKeyRing[addrID], nil
}
