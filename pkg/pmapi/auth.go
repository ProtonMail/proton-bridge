// Copyright (c) 2021 Proton Technologies AG
//
// This file is part of ProtonMail Bridge.
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

package pmapi

import (
	"crypto/subtle"
	"encoding/base64"
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/ProtonMail/proton-bridge/pkg/srp"
)

var ErrBad2FACode = errors.New("incorrect 2FA code")
var ErrBad2FACodeTryAgain = errors.New("incorrect 2FA code: please try again")

type AuthInfoReq struct {
	Username string
}

type U2FInfo struct {
	Challenge      string
	RegisteredKeys []struct {
		Version   string
		KeyHandle string
	}
}

type TwoFactorInfo struct {
	Enabled int // 0 for disabled, 1 for OTP, 2 for U2F, 3 for both.
	TOTP    int
	U2F     U2FInfo
}

func (twoFactor *TwoFactorInfo) hasTwoFactor() bool {
	return twoFactor.Enabled > 0
}

// AuthInfo contains data used when authenticating a user. It should be
// provided to Client.Auth(). Each AuthInfo can be used for only one login attempt.
type AuthInfo struct {
	TwoFA *TwoFactorInfo `json:"2FA,omitempty"`

	version         int
	salt            string
	modulus         string
	srpSession      string
	serverEphemeral string
}

func (a *AuthInfo) HasTwoFactor() bool {
	if a.TwoFA == nil {
		return false
	}
	return a.TwoFA.hasTwoFactor()
}

type AuthInfoRes struct {
	Res
	AuthInfo

	Modulus         string
	ServerEphemeral string
	Version         int
	Salt            string
	SRPSession      string
}

func (res *AuthInfoRes) getAuthInfo() *AuthInfo {
	info := &res.AuthInfo

	// Some fields in AuthInfo are private, so we need to copy them from AuthRes
	// (private fields cannot be populated by json).
	info.version = res.Version
	info.salt = res.Salt
	info.modulus = res.Modulus
	info.srpSession = res.SRPSession
	info.serverEphemeral = res.ServerEphemeral

	return info
}

type AuthReq struct {
	Username        string
	ClientProof     string
	ClientEphemeral string
	SRPSession      string
}

// Auth contains data after a successful authentication. It should be provided to Client.Unlock().
type Auth struct {
	accessToken  string // Read from AuthRes.
	ExpiresIn    int
	uid          string // Read from AuthRes.
	RefreshToken string
	EventID      string
	PasswordMode int
	TwoFA        *TwoFactorInfo `json:"2FA,omitempty"`
}

// UID returns the session UID from the Auth.
// Only Auths generated from the /auth route will have the UID.
// Auths generated from /auth/refresh are not required to.
func (s *Auth) UID() string {
	return s.uid
}

// GenToken generates a string token containing the session UID and refresh token.
func (s *Auth) GenToken() string {
	if s == nil {
		return ""
	}

	return fmt.Sprintf("%v:%v", s.UID(), s.RefreshToken)
}

func (s *Auth) HasTwoFactor() bool {
	if s.TwoFA == nil {
		return false
	}
	return s.TwoFA.hasTwoFactor()
}

func (s *Auth) HasMailboxPassword() bool {
	return s.PasswordMode == 2
}

type AuthRes struct {
	Res
	Auth

	AccessToken string
	TokenType   string

	// UID is the session UID. This is only present in an initial Auth (/auth), not in a refreshed Auth (/auth/refresh).
	UID string

	ServerProof string
}

func (res *AuthRes) getAuth() *Auth {
	auth := &res.Auth

	// Some fields in Auth are private, so we need to copy them from AuthRes
	// (private fields cannot be populated by json).
	auth.accessToken = res.AccessToken
	auth.uid = res.UID

	return auth
}

type Auth2FAReq struct {
	TwoFactorCode string

	// Prepared for U2F:
	// U2F U2FRequest
}

type Auth2FARes struct {
	Res
}

type AuthRefreshReq struct {
	ResponseType string
	GrantType    string
	RefreshToken string
	UID          string
	RedirectURI  string
	State        string
}

func (c *client) sendAuth(auth *Auth) {
	if auth != nil {
		c.log.WithField("auth", *auth).Debug("Client is sending auth to ClientManager")
	} else {
		c.log.Debug("Client is sending nil auth to ClientManager")
	}

	if auth != nil {
		c.uid = auth.UID()
		c.accessToken = auth.accessToken
	}

	c.cm.HandleAuth(ClientAuth{UserID: c.userID, Auth: auth})
}

// AuthInfo gets authentication info for a user.
func (c *client) AuthInfo(username string) (info *AuthInfo, err error) {
	infoReq := &AuthInfoReq{
		Username: username,
	}

	req, err := c.NewJSONRequest("POST", "/auth/info", infoReq)
	if err != nil {
		return
	}

	var infoRes AuthInfoRes
	if err = c.DoJSON(req, &infoRes); err != nil {
		return
	}

	info, err = infoRes.getAuthInfo(), infoRes.Err()

	return
}

func srpProofsFromInfo(info *AuthInfo, username, password string, fallbackVersion int) (proofs *srp.SrpProofs, err error) {
	version := info.version
	if version == 0 {
		version = fallbackVersion
	}

	srpAuth, err := srp.NewSrpAuth(version, username, password, info.salt, info.modulus, info.serverEphemeral)
	if err != nil {
		return
	}

	proofs, err = srpAuth.GenerateSrpProofs(2048)
	return
}

func (c *client) tryAuth(username, password string, info *AuthInfo, fallbackVersion int) (res *AuthRes, err error) {
	proofs, err := srpProofsFromInfo(info, username, password, fallbackVersion)
	if err != nil {
		return
	}

	authReq := &AuthReq{
		Username:        username,
		ClientEphemeral: base64.StdEncoding.EncodeToString(proofs.ClientEphemeral),
		ClientProof:     base64.StdEncoding.EncodeToString(proofs.ClientProof),
		SRPSession:      info.srpSession,
	}

	req, err := c.NewJSONRequest("POST", "/auth", authReq)
	if err != nil {
		return
	}

	var authRes AuthRes
	if err = c.DoJSON(req, &authRes); err != nil {
		return
	}

	if err = authRes.Err(); err != nil {
		return
	}

	serverProof, err := base64.StdEncoding.DecodeString(authRes.ServerProof)
	if err != nil {
		return
	}

	if subtle.ConstantTimeCompare(proofs.ExpectedServerProof, serverProof) != 1 {
		return nil, errors.New("pmapi: bad server proof")
	}

	res, err = &authRes, authRes.Err()
	return res, err
}

func (c *client) tryFullAuth(username, password string, fallbackVersion int) (info *AuthInfo, authRes *AuthRes, err error) {
	info, err = c.AuthInfo(username)
	if err != nil {
		return
	}
	authRes, err = c.tryAuth(username, password, info, fallbackVersion)
	return
}

// Auth will authenticate a user.
func (c *client) Auth(username, password string, info *AuthInfo) (auth *Auth, err error) {
	if info == nil {
		if info, err = c.AuthInfo(username); err != nil {
			return
		}
	}

	authRes, err := c.tryAuth(username, password, info, 2)
	if err != nil && info.version == 0 && srp.CleanUserName(username) != strings.ToLower(username) {
		info, authRes, err = c.tryFullAuth(username, password, 1)
	}
	if err != nil && info.version == 0 {
		_, authRes, err = c.tryFullAuth(username, password, 0)
	}
	if err != nil {
		return
	}

	auth = authRes.getAuth()
	c.sendAuth(auth)

	return auth, err
}

// Auth2FA will authenticate a user into full scope.
// `Auth` struct contains method `HasTwoFactor` deciding whether this has to be done.
func (c *client) Auth2FA(twoFactorCode string, auth *Auth) error {
	auth2FAReq := &Auth2FAReq{
		TwoFactorCode: twoFactorCode,
	}

	req, err := c.NewJSONRequest("POST", "/auth/2fa", auth2FAReq)
	if err != nil {
		return err
	}

	var auth2FARes Auth2FARes
	if err := c.DoJSON(req, &auth2FARes); err != nil {
		return err
	}

	if err := auth2FARes.Err(); err != nil {
		switch auth2FARes.StatusCode {
		case http.StatusUnauthorized:
			return ErrBad2FACode
		case http.StatusUnprocessableEntity:
			return ErrBad2FACodeTryAgain
		default:
			return err
		}
	}

	return nil
}

// AuthRefresh will refresh an expired access token.
func (c *client) AuthRefresh(uidAndRefreshToken string) (auth *Auth, err error) {
	c.refreshLocker.Lock()
	defer c.refreshLocker.Unlock()

	// If we don't yet have a saved access token, save this one in case the refresh fails!
	// That way we can try again later (see handleUnauthorizedStatus).
	c.cm.setTokenIfUnset(c.userID, uidAndRefreshToken)

	split := strings.Split(uidAndRefreshToken, ":")
	if len(split) != 2 {
		err = ErrInvalidToken
		return
	}

	refreshReq := &AuthRefreshReq{
		ResponseType: "token",
		GrantType:    "refresh_token",
		RefreshToken: split[1],
		UID:          split[0],
		RedirectURI:  "https://protonmail.ch",
		State:        "random_string",
	}

	// UID must be set for `x-pm-uid` header field, see backend-communication#11
	c.uid = split[0]

	req, err := c.NewJSONRequest("POST", "/auth/refresh", refreshReq)
	if err != nil {
		return
	}

	var res AuthRes
	if err = c.DoJSON(req, &res); err != nil {
		return
	}
	if err = res.Err(); err != nil {
		return
	}

	auth = res.getAuth()

	// Responses from /auth/refresh are not guaranteed to return the UID if it has not changed.
	// But we want to always return it.
	if auth.uid == "" {
		auth.uid = c.uid
	}

	c.sendAuth(auth)

	return auth, err
}

func (c *client) AuthSalt() (string, error) {
	salts, err := c.GetKeySalts()
	if err != nil {
		return "", err
	}

	if _, err := c.CurrentUser(); err != nil {
		return "", err
	}

	for _, s := range salts {
		if s.ID == c.user.Keys[0].ID {
			return s.KeySalt, nil
		}
	}

	return "", errors.New("no matching salt found")
}

// Logout instructs the client manager to log this client out.
func (c *client) Logout() {
	c.cm.LogoutClient(c.userID)
}

// DeleteAuth deletes the API session.
func (c *client) DeleteAuth() (err error) {
	req, err := c.NewRequest("DELETE", "/auth", nil)
	if err != nil {
		return
	}

	var res Res
	if err = c.DoJSON(req, &res); err != nil {
		return
	}

	if err = res.Err(); err != nil {
		return
	}

	return
}

// IsConnected returns whether the client is authorized to access the API.
func (c *client) IsConnected() bool {
	return c.uid != "" && c.accessToken != ""
}

// ClearData clears sensitive data from the client.
func (c *client) ClearData() {
	c.uid = ""
	c.accessToken = ""
	c.addresses = nil
	c.user = nil
	c.clearKeys()
}
