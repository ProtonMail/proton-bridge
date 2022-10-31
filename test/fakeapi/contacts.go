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
	"fmt"
	"net/url"
	"strconv"

	"github.com/ProtonMail/proton-bridge/v2/pkg/pmapi"
)

func (api *FakePMAPI) DecryptAndVerifyCards(cards []pmapi.Card) ([]pmapi.Card, error) {
	return cards, nil
}

func (api *FakePMAPI) GetContactEmailByEmail(_ context.Context, email string, page int, pageSize int) ([]pmapi.ContactEmail, error) {
	v := url.Values{}
	v.Set("Page", strconv.Itoa(page))
	if pageSize > 0 {
		v.Set("PageSize", strconv.Itoa(pageSize))
	}
	v.Set("Email", email)
	if err := api.checkAndRecordCall(GET, "/contacts/emails?"+v.Encode(), nil); err != nil {
		return nil, err
	}
	return []pmapi.ContactEmail{}, nil
}

func (api *FakePMAPI) GetContactByID(_ context.Context, contactID string) (pmapi.Contact, error) {
	if err := api.checkAndRecordCall(GET, "/contacts/"+contactID, nil); err != nil {
		return pmapi.Contact{}, err
	}
	return pmapi.Contact{}, fmt.Errorf("contact %s does not exist", contactID)
}
