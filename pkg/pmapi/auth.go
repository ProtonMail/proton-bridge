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
	"crypto/rand"
	"encoding/base64"
	"errors"
	"io"
	"net/http"
	"time"

	"github.com/go-resty/resty/v2"
)

type AuthModulus struct {
	Modulus   string
	ModulusID string
}

type GetAuthInfoReq struct {
	Username string
}

type AuthInfo struct {
	Version         int
	Modulus         string
	ServerEphemeral string
	Salt            string
	SRPSession      string
}

type TwoFAInfo struct {
	Enabled TwoFAStatus
}

func (twoFAInfo TwoFAInfo) hasTwoFactor() bool {
	return twoFAInfo.Enabled > TwoFADisabled
}

type TwoFAStatus int

const (
	TwoFADisabled TwoFAStatus = iota
	TOTPEnabled
	U2FEnabled
	TOTPAndU2FEnabled
)

type PasswordMode int

const (
	OnePasswordMode PasswordMode = iota + 1
	TwoPasswordMode
)

type AuthReq struct {
	Username        string
	ClientProof     string
	ClientEphemeral string
	SRPSession      string
}

type AuthRefresh struct {
	UID          string
	AccessToken  string
	RefreshToken string
	ExpiresIn    int64
	Scopes       []string
}

type Auth struct {
	AuthRefresh

	UserID       string
	ServerProof  string
	PasswordMode PasswordMode
	TwoFA        *TwoFAInfo `json:"2FA,omitempty"`
}

func (a Auth) HasTwoFactor() bool {
	if a.TwoFA == nil {
		return false
	}
	return a.TwoFA.hasTwoFactor()
}

func (a Auth) HasMailboxPassword() bool {
	return a.PasswordMode == TwoPasswordMode
}

type auth2FAReq struct {
	TwoFactorCode string
}

type authRefreshReq struct {
	UID          string
	RefreshToken string
	ResponseType string
	GrantType    string
	RedirectURI  string
	State        string
}

func (c *client) Auth2FA(ctx context.Context, twoFactorCode string) error {
	// 2FA is called during login procedure during which refresh token should
	// be valid, therefore, no refresh is needed if there is an error.
	ctx = ContextWithoutAuthRefresh(ctx)

	if res, err := c.do(ctx, func(r *resty.Request) (*resty.Response, error) {
		return r.SetBody(auth2FAReq{TwoFactorCode: twoFactorCode}).Post("/auth/2fa")
	}); err != nil {
		if res != nil {
			switch res.StatusCode() {
			case http.StatusUnauthorized:
				return ErrBad2FACode
			case http.StatusUnprocessableEntity:
				return ErrBad2FACodeTryAgain
			}
		}
		return err
	}

	return nil
}

func (c *client) AuthDelete(ctx context.Context) error {
	if _, err := c.do(ctx, func(r *resty.Request) (*resty.Response, error) {
		return r.Delete("/auth")
	}); err != nil {
		return err
	}

	c.uid, c.acc, c.ref, c.exp = "", "", "", time.Time{}
	c.sendAuthRefresh(nil)
	return nil
}

func (c *client) AuthSalt(ctx context.Context) (string, error) {
	salts, err := c.GetKeySalts(ctx)
	if err != nil {
		return "", err
	}

	if _, err := c.CurrentUser(ctx); err != nil {
		return "", err
	}

	for _, s := range salts {
		if s.ID == c.user.Keys[0].ID {
			return s.KeySalt, nil
		}
	}

	return "", errors.New("no matching salt found")
}

func (c *client) AddAuthRefreshHandler(handler AuthRefreshHandler) {
	c.authHandlers = append(c.authHandlers, handler)
}

func (c *client) authRefresh(ctx context.Context) error {
	c.authLocker.Lock()
	defer c.authLocker.Unlock()

	if c.ref == "" {
		return ErrUnauthorized
	}

	auth, err := c.manager.authRefresh(ctx, c.uid, c.ref)
	if err != nil {
		if IsFailedAuth(err) {
			c.sendAuthRefresh(nil)
		}
		return err
	}

	c.acc = auth.AccessToken
	c.ref = auth.RefreshToken
	c.exp = expiresIn(auth.ExpiresIn)

	c.sendAuthRefresh(auth)
	return nil
}

func (c *client) sendAuthRefresh(auth *AuthRefresh) {
	for _, handler := range c.authHandlers {
		go handler(auth)
	}
	if auth == nil {
		c.authHandlers = []AuthRefreshHandler{}
	}
}

func randomString(length int) string {
	noise := make([]byte, length)

	if _, err := io.ReadFull(rand.Reader, noise); err != nil {
		panic(err)
	}

	return base64.StdEncoding.EncodeToString(noise)[:length]
}

func (c *client) GetCurrentAuth() *Auth {
	return &Auth{
		UserID: c.user.ID,
		AuthRefresh: AuthRefresh{
			UID:          c.uid,
			RefreshToken: c.ref,
		},
	}
}
