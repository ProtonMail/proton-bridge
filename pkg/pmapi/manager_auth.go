package pmapi

import (
	"context"
	"encoding/base64"
	"time"

	"github.com/ProtonMail/proton-bridge/pkg/srp"
)

func (m *manager) NewClient(uid, acc, ref string, exp time.Time) Client {
	return newClient(m, uid).withAuth(acc, ref, exp)
}

func (m *manager) NewClientWithRefresh(ctx context.Context, uid, ref string) (Client, *Auth, error) {
	c := newClient(m, uid)

	auth, err := m.authRefresh(ctx, uid, ref)
	if err != nil {
		return nil, nil, err
	}

	return c.withAuth(auth.AccessToken, auth.RefreshToken, expiresIn(auth.ExpiresIn)), auth, nil
}

func (m *manager) NewClientWithLogin(ctx context.Context, username, password string) (Client, *Auth, error) {
	info, err := m.getAuthInfo(ctx, GetAuthInfoReq{Username: username})
	if err != nil {
		return nil, nil, err
	}

	srpAuth, err := srp.NewSrpAuth(info.Version, username, password, info.Salt, info.Modulus, info.ServerEphemeral)
	if err != nil {
		return nil, nil, err
	}

	proofs, err := srpAuth.GenerateSrpProofs(2048)
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

func (m *manager) getAuthModulus(ctx context.Context) (AuthModulus, error) {
	var res struct {
		AuthModulus
	}

	if _, err := m.r(ctx).SetResult(&res).Get("/auth/modulus"); err != nil {
		return AuthModulus{}, err
	}

	return res.AuthModulus, nil
}

func (m *manager) getAuthInfo(ctx context.Context, req GetAuthInfoReq) (*AuthInfo, error) {
	var res struct {
		*AuthInfo
	}

	if _, err := m.r(ctx).SetBody(req).SetResult(&res).Post("/auth/info"); err != nil {
		return nil, err
	}

	return res.AuthInfo, nil
}

func (m *manager) auth(ctx context.Context, req AuthReq) (*Auth, error) {
	var res struct {
		*Auth
	}

	if _, err := m.r(ctx).SetBody(req).SetResult(&res).Post("/auth"); err != nil {
		return nil, err
	}

	return res.Auth, nil
}

func (m *manager) authRefresh(ctx context.Context, uid, ref string) (*Auth, error) {
	var req = AuthRefreshReq{
		UID:          uid,
		RefreshToken: ref,
		ResponseType: "token",
		GrantType:    "refresh_token",
		RedirectURI:  "https://protonmail.ch",
		State:        randomString(32),
	}

	var res struct {
		*Auth
	}

	if _, err := m.r(ctx).SetBody(req).SetResult(&res).Post("/auth/refresh"); err != nil {
		return nil, err
	}

	return res.Auth, nil
}

func expiresIn(seconds int64) time.Time {
	return time.Now().Add(time.Duration(seconds) * time.Second)
}
