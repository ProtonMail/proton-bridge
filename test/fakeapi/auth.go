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

package fakeapi

import (
	"context"

	"github.com/ProtonMail/proton-bridge/pkg/pmapi"
)

func (api *FakePMAPI) Auth2FA(_ context.Context, req pmapi.Auth2FAReq) error {
	if err := api.checkAndRecordCall(POST, "/auth/2fa", req); err != nil {
		return err
	}

	if api.uid == "" {
		return pmapi.ErrUnauthorized
	}

	session, ok := api.controller.sessionsByUID[api.uid]
	if !ok {
		return pmapi.ErrUnauthorized
	}

	session.hasFullScope = true

	return nil
}

func (api *FakePMAPI) AuthSalt(_ context.Context) (string, error) {
	if err := api.checkAndRecordCall(GET, "/keys/salts", nil); err != nil {
		return "", err
	}

	return "", nil
}

func (api *FakePMAPI) AddAuthHandler(handler pmapi.AuthHandler) {
	api.authHandlers = append(api.authHandlers, handler)
}

func (api *FakePMAPI) AuthDelete(_ context.Context) error {
	if err := api.checkAndRecordCall(DELETE, "/auth", nil); err != nil {
		return err
	}

	api.controller.deleteSession(api.uid)

	return nil
}
