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
// along with Proton Mail Bridge. If not, see <https://www.gnu.org/licenses/>.

package pmapi

import (
	"context"
	"encoding/base64"
	"time"

	"github.com/ProtonMail/go-srp"
)

func (m *manager) NewClient(uid, acc, ref string, exp time.Time) Client {
	log.Trace("New client")

	return newClient(m, uid).withAuth(acc, ref, exp)
}

func (m *manager) NewClientWithRefresh(ctx context.Context, uid, ref string) (Client, *AuthRefresh, error) {
	log.Trace("New client with refresh")

	c := newClient(m, uid)

	auth, err := m.authRefresh(ctx, uid, ref)
	if err != nil {
		return nil, nil, err
	}

	return c.withAuth(auth.AccessToken, auth.RefreshToken, expiresIn(auth.ExpiresIn)), auth, nil
}

func (m *manager) NewClientWithLogin(ctx context.Context, username string, password []byte) (Client, *Auth, error) {
	log.Trace("New client with login")

	info, err := m.getAuthInfo(ctx, GetAuthInfoReq{Username: username})
	if err != nil {
		return nil, nil, err
	}

	srpAuth, err := srp.NewAuth(info.Version, username, password, info.Salt, info.Modulus, info.ServerEphemeral)
	if err != nil {
		return nil, nil, err
	}

	proofs, err := srpAuth.GenerateProofs(2048)
	if err != nil {
		return nil, nil, err
	}

	auth, err := m.auth(ctx, AuthReq{
		Username:        username,
		ClientProof:     base64.StdEncoding.EncodeToString(proofs.ClientProof),
		ClientEphemeral: base64.StdEncoding.EncodeToString(proofs.ClientEphemeral),
		SRPSession:      info.SRPSession,
	})
	if err != nil {
		return nil, nil, err
	}

	return newClient(m, auth.UID).withAuth(auth.AccessToken, auth.RefreshToken, expiresIn(auth.ExpiresIn)), auth, nil
}

func (m *manager) getAuthInfo(ctx context.Context, req GetAuthInfoReq) (*AuthInfo, error) {
	var res struct {
		*AuthInfo
	}

	_, err := wrapNoConnection(m.r(ctx).SetBody(req).SetResult(&res).Post("/auth/info"))
	if err != nil {
		return nil, err
	}

	return res.AuthInfo, nil
}

func (m *manager) auth(ctx context.Context, req AuthReq) (*Auth, error) {
	var res struct {
		*Auth
	}

	_, err := wrapNoConnection(m.r(ctx).SetBody(req).SetResult(&res).Post("/auth"))
	if err != nil {
		return nil, err
	}

	return res.Auth, nil
}

func (m *manager) authRefresh(ctx context.Context, uid, ref string) (*AuthRefresh, error) {
	m.refreshingAuth.Lock()
	defer m.refreshingAuth.Unlock()

	req := authRefreshReq{
		UID:          uid,
		RefreshToken: ref,
		ResponseType: "token",
		GrantType:    "refresh_token",
		RedirectURI:  "https://protonmail.ch",
		State:        randomString(32),
	}

	var res struct {
		*AuthRefresh
	}

	_, err := wrapNoConnection(m.r(ctx).SetBody(req).SetResult(&res).Post("/auth/refresh"))
	if err != nil {
		if IsBadRequest(err) || IsUnprocessableEntity(err) {
			err = ErrAuthFailed{err}
		}
		return nil, err
	}

	return res.AuthRefresh, nil
}

func expiresIn(seconds int64) time.Time {
	return time.Now().Add(time.Duration(seconds) * time.Second)
}
