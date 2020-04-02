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
	"encoding/json"
	"math/rand"
	"net/http"
	"testing"
	"time"

	pmcrypto "github.com/ProtonMail/gopenpgp/crypto"
	"github.com/ProtonMail/proton-bridge/pkg/srp"

	"github.com/sirupsen/logrus"

	a "github.com/stretchr/testify/assert"
	r "github.com/stretchr/testify/require"
)

var aLongTimeAgo = time.Unix(233431200, 0)

var testIdentity = &pmcrypto.Identity{
	Name:  "UserID",
	Email: "",
}

const (
	testUsername    = "jason"
	testAPIPassword = "apple"

	testUID             = "729ad6012421d67ad26950dc898bebe3a6e3caa2" //nolint[gosec]
	testAccessToken     = "de0423049b44243afeec7d9c1d99be7b46da1e8a" //nolint[gosec]
	testAccessTokenOld  = "feb3159ac63fb05119bcf4480d939278aa746926" //nolint[gosec]
	testRefreshToken    = "a49b98256745bb497bec20e9b55f5de16f01fb52" //nolint[gosec]
	testRefreshTokenNew = "b894b4c4f20003f12d486900d8b88c7d68e67235" //nolint[gosec]
)

var testAuthInfo = &AuthInfo{
	TwoFA: &TwoFactorInfo{TOTP: 1},

	version:         4,
	salt:            "yKlc5/CvObfoiw==",
	modulus:         "-----BEGIN PGP SIGNED MESSAGE-----\nHash: SHA256\n\nW2z5HBi8RvsfYzZTS7qBaUxxPhsfHJFZpu3Kd6s1JafNrCCH9rfvPLrfuqocxWPgWDH2R8neK7PkNvjxto9TStuY5z7jAzWRvFWN9cQhAKkdWgy0JY6ywVn22+HFpF4cYesHrqFIKUPDMSSIlWjBVmEJZ/MusD44ZT29xcPrOqeZvwtCffKtGAIjLYPZIEbZKnDM1Dm3q2K/xS5h+xdhjnndhsrkwm9U9oyA2wxzSXFL+pdfj2fOdRwuR5nW0J2NFrq3kJjkRmpO/Genq1UW+TEknIWAb6VzJJJA244K/H8cnSx2+nSNZO3bbo6Ys228ruV9A8m6DhxmS+bihN3ttQ==\n-----BEGIN PGP SIGNATURE-----\nVersion: ProtonMail\nComment: https://protonmail.com\n\nwl4EARYIABAFAlwB1j0JEDUFhcTpUY8mAAD8CgEAnsFnF4cF0uSHKkXa1GIa\nGO86yMV4zDZEZcDSJo0fgr8A/AlupGN9EdHlsrZLmTA1vhIx+rOgxdEff28N\nkvNM7qIK\n=q6vu\n-----END PGP SIGNATURE-----\n",
	srpSession:      "9b2946bbd9055f17c34940abdce0c3d3",
	serverEphemeral: "5tfigcLKoM0DPWYB+EqYE7QlqsiT63iOVlO5ZX0lTMEILSsrRdVCYrN8L3zkinsAjUZ/cx5wIS7N05k66uZb+ZE3lFOJS2s1BkqLvCrGxYL0e3n5YAnzHYlvCCJKXw/sK57ntfF1OOoblBXX6dw5LjeeDglEep2/DaE0TjD8WUpq4Ls2HlQGn9wrC7dFO2lJXsMhRffxKghiOsdvCLXDmwXginzn/LFezA8KrDsWOBSEGntwpg3s1xFj5h8BqtRHvC0igmoscqgw+3GCMTJ0NZAQ/L+5aJ/0ccL0WBK208ltCNl+/X6Sz0kpyvOP4RqFJhC1auVDJ9AjZQYSYZ1NEQ==",
}

// testAuth has default values which are adjusted in each test.
var testAuth = &Auth{
	EventID:      "NcKPtU5eMNPMrDkIMbEJrgMtC9yQ7Xc5ZBT-tB3UtV1rZ324RWfCIdBI758q0UnsfywS8CkNenIQlWLIX_dUng==",
	ExpiresIn:    86400,
	RefreshToken: "feb3159ac63fb05119bcf4480d939278aa746926",
	Scope:        "full mail payments reset keys",

	accessToken: testAccessToken,
	uid:         testUID,
}

var testAuth2FA = &Auth2FA{
	Scope: "full mail payments reset keys",
}

var testAuthRefreshReq = AuthRefreshReq{
	ResponseType: "token",
	GrantType:    "refresh_token",
	RefreshToken: testRefreshToken,
	UID:          testUID,
	RedirectURI:  "https://protonmail.ch",
	State:        "random_string",
}

var testAuthReq = AuthReq{
	Username:        testUsername,
	ClientProof:     "axfvYdl9iXZjY6zQ+hBYmY7X3TDc/9JtSvrmyZXhDxjxkXB3Hro27t1KItmFIJloItY5sLZDs0eEEZJI34oFZD4ViSG0kfB7ZXcCZ9Jse+U5OFu4vdnPTGolnSofRMEs1NR6ePXzH7mQ10qoq43ity3ve2vmhQNuJNlHAPynKf2WqKOgxq7mmkBzEpXES4mIhwwgVbOygKcUSvguz5E5g13ATF0ZX2d9SJWAbZ262Tks+h99Cdk/dOfgLQhr0nO/r0cpwP84W2RWU2Q34LNkKuuQHkjmxelgBleGq54tCbhoCAYPP6vapgrQjNoVAC/dkjIIAoNL9bJSIynFM5znAA==",
	ClientEphemeral: "mK+eSMosfZO/Cs5s+vcbjpsN7F8UAObwlKKnCy/z9FpoMRM2PfTe5ywLBgffmLYaapPq7XOxaqaj08kcZLHcM1fIA2JQZZTKPnESN1qAQztJ3/YHMI0op6yBgzx9803OjIznjCD2B3XBSMOHIG4oG0UwocsIX32hiMnYlMMkt8NGrityPlnmEbxpRna3fu9LEZ+v0uo6PjKCrO7+9E3uaMi64HadXBfyx2raBFFwA+yh7FvE7U+hl3AJclEre4d8pmfhMdxXze1soJI8fMuqaa07rY0r0rF5mLLTuqTIGRFkU1qG9loq9+IMsSwgkt1P3ghW63JK7Y6LWdDy0d6cAg==",
	SRPSession:      "9b2946bbd9055f17c34940abdce0c3d3",
}

var testAuth2FAReq = Auth2FAReq{
	TwoFactorCode: "424242",
}

func init() {
	logrus.SetLevel(logrus.DebugLevel)
	srp.RandReader = rand.New(rand.NewSource(42))
}

func TestClient_AuthInfo(t *testing.T) {
	finish, c := newTestServerCallbacks(t,
		func(tb testing.TB, w http.ResponseWriter, r *http.Request) string {
			Ok(t, checkMethodAndPath(r, "POST", "/auth/info"))

			var infoReq AuthInfoReq
			Ok(t, json.NewDecoder(r.Body).Decode(&infoReq))
			Equals(t, infoReq.Username, testUsername)

			return "/auth/info/post_response.json"
		},
	)
	defer finish()

	info, err := c.AuthInfo(testCurrentUser.Name)
	Ok(t, err)
	Equals(t, testAuthInfo, info)
}

// TestClient_Auth reflects changes from proton/backend-communcation#3.
func TestClient_Auth(t *testing.T) {
	srp.RandReader = rand.New(rand.NewSource(42))
	finish, c := newTestServerCallbacks(t,
		func(tb testing.TB, w http.ResponseWriter, req *http.Request) string {
			a.Nil(t, checkMethodAndPath(req, "POST", "/auth"))

			var authReq AuthReq
			r.Nil(t, json.NewDecoder(req.Body).Decode(&authReq))
			r.Equal(t, testAuthReq, authReq)

			return "/auth/post_response.json"
		},
		routeGetUsers,
		routeGetAddresses,
		routeGetSalts,
	)
	defer finish()

	auth, err := c.Auth(testUsername, testAPIPassword, testAuthInfo)
	r.Nil(t, err)

	r.True(t, c.user.KeyRing().FirstKeyID != "", "Parsing First key ID issue")

	exp := &Auth{}
	*exp = *testAuth
	exp.accessToken = testAccessToken
	exp.RefreshToken = testRefreshToken
	exp.KeySalt = "abc"
	a.Equal(t, exp, auth)
}

func TestClient_Auth2FA(t *testing.T) {
	finish, c := newTestServerCallbacks(t,
		func(tb testing.TB, w http.ResponseWriter, r *http.Request) string {
			Ok(t, checkMethodAndPath(r, "POST", "/auth/2fa"))

			var info2FAReq Auth2FAReq
			Ok(t, json.NewDecoder(r.Body).Decode(&info2FAReq))
			Equals(t, info2FAReq.TwoFactorCode, testAuth2FAReq.TwoFactorCode)

			return "/auth/2fa/post_response.json"
		},
		routeGetUsers,
		routeGetAddresses,
		routeGetSalts,
	)
	defer finish()

	c.uid = testUID
	c.accessToken = testAccessToken
	auth2FA, err := c.Auth2FA(testAuth2FAReq.TwoFactorCode, testAuth)
	Ok(t, err)

	Equals(t, testAuth2FA, auth2FA)
}

func TestClient_Auth2FA_Fail(t *testing.T) {
	finish, c := newTestServerCallbacks(t,
		func(tb testing.TB, w http.ResponseWriter, r *http.Request) string {
			Ok(t, checkMethodAndPath(r, "POST", "/auth/2fa"))

			var info2FAReq Auth2FAReq
			Ok(t, json.NewDecoder(r.Body).Decode(&info2FAReq))
			Equals(t, info2FAReq.TwoFactorCode, testAuth2FAReq.TwoFactorCode)

			return "/auth/2fa/post_401_bad_password.json"
		},
	)
	defer finish()

	c.uid = testUID
	c.accessToken = testAccessToken
	_, err := c.Auth2FA(testAuth2FAReq.TwoFactorCode, testAuth)
	Equals(t, ErrBad2FACode, err)
}

func TestClient_Auth2FA_Retry(t *testing.T) {
	finish, c := newTestServerCallbacks(t,
		func(tb testing.TB, w http.ResponseWriter, r *http.Request) string {
			Ok(t, checkMethodAndPath(r, "POST", "/auth/2fa"))

			var info2FAReq Auth2FAReq
			Ok(t, json.NewDecoder(r.Body).Decode(&info2FAReq))
			Equals(t, info2FAReq.TwoFactorCode, testAuth2FAReq.TwoFactorCode)

			return "/auth/2fa/post_422_bad_password.json"
		},
	)
	defer finish()

	c.uid = testUID
	c.accessToken = testAccessToken
	_, err := c.Auth2FA(testAuth2FAReq.TwoFactorCode, testAuth)
	Equals(t, ErrBad2FACodeTryAgain, err)
}

func TestClient_Unlock(t *testing.T) {
	finish, c := newTestServerCallbacks(t,
		routeGetUsers,
		routeGetAddresses,
	)
	defer finish()
	c.uid = testUID
	c.accessToken = testAccessToken

	_, err := c.Unlock("wrong")
	a.True(t, IsUnlockError(err), "expected error, pasword is wrong")

	_, err = c.Unlock(testMailboxPassword)
	a.Nil(t, err)
	a.Equal(t, testUID, c.uid)
	a.Equal(t, testAccessToken, c.accessToken)

	// second try should not fail because there is an unlocked key already
	_, err = c.Unlock("wrong")
	a.Nil(t, err)
}

func TestClient_Unlock_EncPrivKey(t *testing.T) {
	finish, c := newTestServerCallbacks(t,
		routeGetUsers,
		routeGetAddresses,
	)
	defer finish()
	c.uid = testUID
	c.accessToken = testAccessToken

	_, err := c.Unlock(testMailboxPassword)
	Ok(t, err)
	Equals(t, testUID, c.uid)
	Equals(t, testAccessToken, c.accessToken)
}

func routeAuthRefresh(tb testing.TB, w http.ResponseWriter, r *http.Request) string {
	Ok(tb, checkMethodAndPath(r, "POST", "/auth/refresh"))
	Ok(tb, checkHeader(r.Header, "x-pm-uid", testUID))

	var refreshReq AuthRefreshReq
	Ok(tb, json.NewDecoder(r.Body).Decode(&refreshReq))
	Equals(tb, testAuthRefreshReq, refreshReq)

	return "/auth/refresh/post_response.json"
}

// TestClient_AuthRefresh reflects changes from proton/backend-communcation#11.
func TestClient_AuthRefresh(t *testing.T) {
	finish, c := newTestServerCallbacks(t,
		routeAuthRefresh,
	)
	defer finish()
	c.uid = "" // Testing that we always send correct `x-pm-uid`.
	c.accessToken = "oldToken"

	auth, err := c.AuthRefresh(testUID + ":" + testRefreshToken)
	Ok(t, err)

	exp := &Auth{}
	*exp = *testAuth
	exp.uid = "" // AuthRefresh will not return UID (only Auth returns the UID).
	exp.accessToken = testAccessToken
	exp.KeySalt = ""
	exp.EventID = ""
	exp.ExpiresIn = 360000
	exp.RefreshToken = testRefreshTokenNew
	Equals(t, exp, auth)
}

func routeAuthRefreshHasUID(tb testing.TB, w http.ResponseWriter, r *http.Request) string {
	Ok(tb, checkMethodAndPath(r, "POST", "/auth/refresh"))
	Ok(tb, checkHeader(r.Header, "x-pm-uid", testUID))

	var refreshReq AuthRefreshReq
	Ok(tb, json.NewDecoder(r.Body).Decode(&refreshReq))
	Equals(tb, testAuthRefreshReq, refreshReq)

	return "/auth/refresh/post_resp_has_uid.json"
}

// TestClient_AuthRefresh reflects changes from proton/backend-communcation#3.
func TestClient_AuthRefresh_HasUID(t *testing.T) {
	finish, c := newTestServerCallbacks(t,
		routeAuthRefreshHasUID,
	)
	defer finish()
	c.uid = testUID
	c.accessToken = "oldToken"

	auth, err := c.AuthRefresh(testUID + ":" + testRefreshToken)
	Ok(t, err)

	exp := &Auth{}
	*exp = *testAuth
	exp.accessToken = testAccessToken
	exp.KeySalt = ""
	exp.EventID = ""
	exp.ExpiresIn = 360000
	exp.RefreshToken = testRefreshTokenNew
	Equals(t, exp, auth)
}

func TestClient_Logout(t *testing.T) {
	finish, c := newTestServerCallbacks(t,
		func(tb testing.TB, w http.ResponseWriter, r *http.Request) string {
			Ok(t, checkMethodAndPath(r, "DELETE", "/auth"))
			Ok(t, isAuthReq(r, testUID, testAccessToken))
			return "auth/delete_response.json"
		},
	)
	defer finish()

	c.uid = testUID
	c.accessToken = testAccessToken

	c.Logout()

	r.Eventually(t, func() bool {
		// TODO: Use a method like IsConnected() which returns whether the client was logged out or not.
		return c.accessToken == "" &&
			c.uid == "" &&
			c.kr == nil &&
			c.addresses == nil &&
			c.user == nil
	}, 10*time.Second, 10*time.Millisecond)
}

func TestClient_DoUnauthorized(t *testing.T) {
	finish, c := newTestServerCallbacks(t,
		func(tb testing.TB, w http.ResponseWriter, r *http.Request) string {
			Ok(t, checkMethodAndPath(r, "GET", "/"))
			return httpResponse(http.StatusUnauthorized)
		},
		routeAuthRefresh,
		func(tb testing.TB, w http.ResponseWriter, r *http.Request) string {
			Ok(t, checkMethodAndPath(r, "GET", "/"))
			Ok(t, isAuthReq(r, testUID, testAccessToken))
			return httpResponse(http.StatusOK)
		},
	)
	defer finish()

	c.uid = testUID
	c.accessToken = testAccessTokenOld
	c.cm.tokens[c.userID] = testUID + ":" + testRefreshToken

	req, err := c.NewRequest("GET", "/", nil)
	Ok(t, err)

	res, err := c.Do(req, true)
	Ok(t, err)

	defer Ok(t, res.Body.Close())
}
