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

package fakeapi

import (
	"strings"

	"github.com/ProtonMail/proton-bridge/pkg/pmapi"
)

func (api *FakePMAPI) SetAuths(auths chan<- *pmapi.Auth) {
	api.auths = auths
}

func (api *FakePMAPI) AuthInfo(username string) (*pmapi.AuthInfo, error) {
	if err := api.checkInternetAndRecordCall(POST, "/auth/info", &pmapi.AuthInfoReq{
		Username: username,
	}); err != nil {
		return nil, err
	}
	authInfo := &pmapi.AuthInfo{}
	user, ok := api.controller.usersByUsername[username]
	if !ok {
		// If username is wrong, API server will return empty but
		// positive response
		return authInfo, nil
	}
	authInfo.TwoFA = user.get2FAInfo()
	return authInfo, nil
}

func (api *FakePMAPI) Auth(username, password string, authInfo *pmapi.AuthInfo) (*pmapi.Auth, error) {
	if err := api.checkInternetAndRecordCall(POST, "/auth", &pmapi.AuthReq{
		Username: username,
	}); err != nil {
		return nil, err
	}

	session, err := api.controller.createSessionIfAuthorized(username, password)
	if err != nil {
		return nil, err
	}
	api.setUID(session.uid)

	if err := api.setUser(username); err != nil {
		return nil, err
	}

	user := api.controller.usersByUsername[username]
	auth := &pmapi.Auth{
		TwoFA:        user.get2FAInfo(),
		RefreshToken: session.refreshToken,
		ExpiresIn:    86400, // seconds
	}
	auth.DANGEROUSLYSetUID(session.uid)

	api.sendAuth(auth)

	return auth, nil
}

func (api *FakePMAPI) Auth2FA(twoFactorCode string, auth *pmapi.Auth) error {
	if err := api.checkInternetAndRecordCall(POST, "/auth/2fa", &pmapi.Auth2FAReq{
		TwoFactorCode: twoFactorCode,
	}); err != nil {
		return err
	}

	if api.uid == "" {
		return pmapi.ErrInvalidToken
	}

	session, ok := api.controller.sessionsByUID[api.uid]
	if !ok {
		return pmapi.ErrInvalidToken
	}

	session.hasFullScope = true

	return nil
}

func (api *FakePMAPI) AuthRefresh(token string) (*pmapi.Auth, error) {
	if api.lastToken == "" {
		api.lastToken = token
	}

	split := strings.Split(token, ":")
	if len(split) != 2 {
		return nil, pmapi.ErrInvalidToken
	}

	if err := api.checkInternetAndRecordCall(POST, "/auth/refresh", &pmapi.AuthRefreshReq{
		ResponseType: "token",
		GrantType:    "refresh_token",
		UID:          split[0],
		RefreshToken: split[1],
		RedirectURI:  "https://protonmail.ch",
		State:        "random_string",
	}); err != nil {
		return nil, err
	}

	session, ok := api.controller.sessionsByUID[split[0]]
	if !ok || session.refreshToken != split[1] {
		api.log.WithField("token", token).
			WithField("session", session).
			Warn("Refresh token failed")
		// The API server will respond normal error not 401 (check api)
		// i.e. should not use `sendAuth(nil)`
		api.setUID("")
		return nil, pmapi.ErrInvalidToken
	}
	api.setUID(split[0])

	if err := api.setUser(session.username); err != nil {
		return nil, err
	}
	api.controller.refreshTheTokensForSession(session)
	api.lastToken = split[0] + ":" + session.refreshToken

	auth := &pmapi.Auth{
		RefreshToken: session.refreshToken,
		ExpiresIn:    86400,
	}
	auth.DANGEROUSLYSetUID(session.uid)

	api.sendAuth(auth)

	return auth, nil
}

func (api *FakePMAPI) AuthSalt() (string, error) {
	if err := api.checkInternetAndRecordCall(GET, "/keys/salts", nil); err != nil {
		return "", err
	}

	return "", nil
}

func (api *FakePMAPI) Logout() {
	api.controller.clientManager.LogoutClient(api.userID)
}

func (api *FakePMAPI) IsConnected() bool {
	return api.uid != "" && api.lastToken != ""
}

func (api *FakePMAPI) DeleteAuth() error {
	if err := api.checkAndRecordCall(DELETE, "/auth", nil); err != nil {
		return err
	}
	api.controller.deleteSession(api.uid)
	return nil
}

func (api *FakePMAPI) ClearData() {
	if api.userKeyRing != nil {
		api.userKeyRing.ClearPrivateParams()
		api.userKeyRing = nil
	}

	for addrID, addr := range api.addrKeyRing {
		if addr != nil {
			addr.ClearPrivateParams()
			delete(api.addrKeyRing, addrID)
		}
	}

	api.unsetUser()
}
