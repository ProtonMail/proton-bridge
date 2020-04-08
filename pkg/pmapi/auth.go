// Copyright (c) 2020 Proton Technologies AG
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
	"net/http"
	"strings"
	"time"

	pmcrypto "github.com/ProtonMail/gopenpgp/crypto"
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
	Scope        string
	uid          string // Read from AuthRes.
	RefreshToken string
	KeySalt      string
	EventID      string
	PasswordMode int
	TwoFA        *TwoFactorInfo `json:"2FA,omitempty"`
}

func (s *Auth) UID() string {
	return s.uid
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

func (s *Auth) hasFullScope() bool {
	return strings.Contains(s.Scope, "full")
}

type AuthRes struct {
	Res
	Auth

	AccessToken string
	TokenType   string
	UID         string

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

type Auth2FA struct {
	Scope string
}

type Auth2FARes struct {
	Res

	Scope string
}

func (res *Auth2FARes) getAuth2FA() *Auth2FA {
	return &Auth2FA{
		Scope: res.Scope,
	}
}

type AuthRefreshReq struct {
	ResponseType string
	GrantType    string
	RefreshToken string
	UID          string
	RedirectURI  string
	State        string
}

// SetAuths sets auths channel.
func (c *Client) SetAuths(auths chan<- *Auth) {
	c.auths = auths
}

// AuthInfo gets authentication info for a user.
func (c *Client) AuthInfo(username string) (info *AuthInfo, err error) {
	infoReq := &AuthInfoReq{
		Username: username,
	}

	req, err := NewJSONRequest("POST", "/auth/info", infoReq)
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

func (c *Client) tryAuth(username, password string, info *AuthInfo, fallbackVersion int) (res *AuthRes, err error) {
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

	req, err := NewJSONRequest("POST", "/auth", authReq)
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

func (c *Client) tryFullAuth(username, password string, fallbackVersion int) (info *AuthInfo, authRes *AuthRes, err error) {
	info, err = c.AuthInfo(username)
	if err != nil {
		return
	}
	authRes, err = c.tryAuth(username, password, info, fallbackVersion)
	return
}

// Auth will authenticate a user.
func (c *Client) Auth(username, password string, info *AuthInfo) (auth *Auth, err error) {
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
	c.uid = auth.UID()
	c.accessToken = auth.accessToken

	if c.auths != nil {
		c.auths <- auth
	}

	if c.tokenManager != nil {
		c.tokenManager.SetToken(c.userID, c.uid+":"+auth.RefreshToken)
		c.log.Info("Set token from auth " + c.uid + ":" + auth.RefreshToken)
	}

	// Auth has to be fully unlocked to get key salt. During `Auth` it can happen
	// only to accounts without 2FA. For 2FA accounts, it's done in `Auth2FA`.
	if auth.hasFullScope() {
		err = c.setKeySaltToAuth(auth)
		if err != nil {
			return nil, err
		}
	}

	c.expiresAt = time.Now().Add(time.Duration(auth.ExpiresIn) * time.Second)
	return auth, err
}

// Auth2FA will authenticate a user into full scope.
// `Auth` struct contains method `HasTwoFactor` deciding whether this has to be done.
func (c *Client) Auth2FA(twoFactorCode string, auth *Auth) (*Auth2FA, error) {
	auth2FAReq := &Auth2FAReq{
		TwoFactorCode: twoFactorCode,
	}

	req, err := NewJSONRequest("POST", "/auth/2fa", auth2FAReq)
	if err != nil {
		return nil, err
	}

	var auth2FARes Auth2FARes
	if err := c.DoJSON(req, &auth2FARes); err != nil {
		return nil, err
	}

	if err := auth2FARes.Err(); err != nil {
		switch auth2FARes.StatusCode {
		case http.StatusUnauthorized:
			return nil, ErrBad2FACode
		case http.StatusUnprocessableEntity:
			return nil, ErrBad2FACodeTryAgain
		default:
			return nil, err
		}
	}

	if err := c.setKeySaltToAuth(auth); err != nil {
		return nil, err
	}

	return auth2FARes.getAuth2FA(), nil
}

func (c *Client) setKeySaltToAuth(auth *Auth) error {
	// KeySalt already set up, no need to do it again.
	if auth.KeySalt != "" {
		return nil
	}

	user, err := c.CurrentUser()
	if err != nil {
		return err
	}
	salts, err := c.GetKeySalts()
	if err != nil {
		return err
	}
	for _, s := range salts {
		if s.ID == user.KeyRing().FirstKeyID {
			auth.KeySalt = s.KeySalt
			break
		}
	}
	return nil
}

// Unlock decrypts the key ring.
// If the password is invalid, IsUnlockError(err) will return true.
func (c *Client) Unlock(password string) (kr *pmcrypto.KeyRing, err error) {
	if _, err = c.CurrentUser(); err != nil {
		return
	}

	c.keyLocker.Lock()
	defer c.keyLocker.Unlock()

	kr = c.user.KeyRing()
	if err = unlockKeyRingNoErrorWhenAlreadyUnlocked(kr, []byte(password)); err != nil {
		return
	}

	c.kr = kr
	return kr, err
}

// AuthRefresh will refresh an expired access token.
func (c *Client) AuthRefresh(uidAndRefreshToken string) (auth *Auth, err error) {
	// If we don't yet have a saved access token, save this one in case the refresh fails!
	// That way we can try again later (see handleUnauthorizedStatus).
	if c.tokenManager != nil {
		currentAccessToken := c.tokenManager.GetToken(c.userID)
		if currentAccessToken == "" {
			c.log.WithField("token", uidAndRefreshToken).
				Info("Currently have no access token, setting given one")
			c.tokenManager.SetToken(c.userID, uidAndRefreshToken)
		}
	}

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

	req, err := NewJSONRequest("POST", "/auth/refresh", refreshReq)
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
	// UID should never change after auth, see backend-communication#11
	auth.uid = c.uid
	if c.auths != nil {
		c.auths <- auth
	}

	c.uid = auth.UID()
	c.accessToken = auth.accessToken

	if c.tokenManager != nil {
		c.tokenManager.SetToken(c.userID, c.uid+":"+res.RefreshToken)
		c.log.Info("Set token from auth refresh " + c.uid + ":" + res.RefreshToken)
	}

	c.expiresAt = time.Now().Add(time.Duration(auth.ExpiresIn) * time.Second)
	return auth, err
}

// Logout logs the current user out.
func (c *Client) Logout() (err error) {
	req, err := NewRequest("DELETE", "/auth", nil)
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

	// This can trigger a deadlock! We don't want to do it if the above requests failed (GODT-154).
	// That's why it's not in the deferred statement above.
	if c.auths != nil {
		c.auths <- nil
	}

	// This should ideally be deferred at the top of this method so that it is executed
	// regardless of what happens, but we currently don't have a way to prevent ourselves
	// from using a logged out client. So for now, it's down here, as it was in Charles release.
	// defer func() {
	c.uid = ""
	c.accessToken = ""
	c.kr = nil
	// c.addresses = nil
	c.user = nil
	if c.tokenManager != nil {
		c.tokenManager.SetToken(c.userID, "")
	}
	// }()

	return err
}
