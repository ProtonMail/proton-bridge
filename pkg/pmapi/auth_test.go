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
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestAutomaticAuthRefresh(t *testing.T) {
	r := require.New(t)
	mux := http.NewServeMux()

	currentTokens := newTestRefreshToken(r)
	testUID := currentTokens.UID
	testAcc := currentTokens.AccessToken
	testRef := currentTokens.RefreshToken
	currentTokens.ExpiresIn = 100

	mux.HandleFunc("/auth/refresh", currentTokens.handleAuthRefresh)
	mux.HandleFunc("/addresses", currentTokens.handleAuthCheckOnly)

	ts := httptest.NewServer(mux)

	var gotAuthRefresh *AuthRefresh

	c := New(Config{HostURL: ts.URL}).
		NewClient(testUID, testAcc, testRef, time.Now().Add(-time.Second))

	c.AddAuthRefreshHandler(func(auth *AuthRefresh) { gotAuthRefresh = auth })

	// Make a request with an access token that already expired one second ago.
	_, err := c.GetAddresses(context.Background())
	r.NoError(err)

	wantAuthRefresh := currentTokens.wantAuthRefresh()

	// The auth callback should have been called.
	r.NotNil(gotAuthRefresh)
	r.Equal(wantAuthRefresh, *gotAuthRefresh)

	cl := c.(*client) //nolint:forcetypeassert  // we want to panic here
	r.Equal(wantAuthRefresh.AccessToken, cl.acc)
	r.Equal(wantAuthRefresh.RefreshToken, cl.ref)
	r.WithinDuration(expiresIn(100), cl.exp, time.Second)
}

func Test401AuthRefresh(t *testing.T) {
	r := require.New(t)
	currentTokens := newTestRefreshToken(r)
	testUID := currentTokens.UID
	testRef := currentTokens.RefreshToken

	mux := http.NewServeMux()

	mux.HandleFunc("/auth/refresh", currentTokens.handleAuthRefresh)
	mux.HandleFunc("/addresses", currentTokens.handleAuthCheckOnly)

	ts := httptest.NewServer(mux)
	var gotAuthRefresh *AuthRefresh

	// Create a new client.
	m := New(Config{HostURL: ts.URL})
	c := m.NewClient(testUID, "oldAccToken", testRef, time.Now().Add(time.Hour))

	// Register an auth handler.
	c.AddAuthRefreshHandler(func(auth *AuthRefresh) { gotAuthRefresh = auth })

	// The first request will fail with 401, triggering a refresh and retry.
	_, err := c.GetAddresses(context.Background())
	r.NoError(err)

	// The auth callback should have been called.
	r.NotNil(gotAuthRefresh)
	r.Equal(currentTokens.wantAuthRefresh(), *gotAuthRefresh)
}

func Test401RevokedAuth(t *testing.T) {
	r := require.New(t)
	currentTokens := newTestRefreshToken(r)

	mux := http.NewServeMux()

	mux.HandleFunc("/auth/refresh", currentTokens.handleAuthRefresh)
	mux.HandleFunc("/addresses", currentTokens.handleAuthCheckOnly)

	ts := httptest.NewServer(mux)

	c := New(Config{HostURL: ts.URL}).
		NewClient("badUID", "badAcc", "badRef", time.Now().Add(time.Hour))

	// The request will fail with 401, triggering a refresh.
	// The retry will also fail with 401, returning an error.
	_, err := c.GetAddresses(context.Background())
	r.True(IsFailedAuth(err))
}

func Test401OldRefreshToken(t *testing.T) {
	r := require.New(t)
	currentTokens := newTestRefreshToken(r)

	mux := http.NewServeMux()

	mux.HandleFunc("/auth/refresh", currentTokens.handleAuthRefresh)
	mux.HandleFunc("/addresses", currentTokens.handleAuthCheckOnly)

	ts := httptest.NewServer(mux)

	c := New(Config{HostURL: ts.URL}).
		NewClient(currentTokens.UID, "oldAcc", "oldRef", time.Now().Add(time.Hour))

	// The request will fail with 401, triggering a refresh.
	// The retry will also fail with 401, returning an error.
	_, err := c.GetAddresses(context.Background())
	r.True(IsFailedAuth(err))
}

func Test401NoAccessToken(t *testing.T) {
	r := require.New(t)
	currentTokens := newTestRefreshToken(r)
	testUID := currentTokens.UID
	testRef := currentTokens.RefreshToken
	mux := http.NewServeMux()
	mux.HandleFunc("/auth/refresh", currentTokens.handleAuthRefresh)
	mux.HandleFunc("/addresses", currentTokens.handleAuthCheckOnly)

	ts := httptest.NewServer(mux)

	c := New(Config{HostURL: ts.URL}).
		NewClient(testUID, "", testRef, time.Now().Add(time.Hour))

	// The request will fail with 401, triggering a refresh. After the refresh it should succeed.
	_, err := c.GetAddresses(context.Background())
	r.NoError(err)
}

func Test401ExpiredAuthUpdateUser(t *testing.T) {
	r := require.New(t)
	mux := http.NewServeMux()
	currentTokens := newTestRefreshToken(r)
	testUID := currentTokens.UID
	testRef := currentTokens.RefreshToken

	mux.HandleFunc("/auth/refresh", currentTokens.handleAuthRefresh)

	mux.HandleFunc("/users", func(w http.ResponseWriter, r *http.Request) {
		if !currentTokens.isAuthorized(r.Header) {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		respObj := struct {
			Code int
			User *User
		}{
			Code: 1000,
			User: &User{
				ID:        "MJLke8kWh1BBvG95JBIrZvzpgsZ94hNNgjNHVyhXMiv4g9cn6SgvqiIFR5cigpml2LD_iUk_3DkV29oojTt3eA==",
				Name:      "jason",
				UsedSpace: &usedSpace,
			},
		}
		if err := json.NewEncoder(w).Encode(respObj); err != nil {
			panic(err)
		}
	})

	mux.HandleFunc("/addresses", func(w http.ResponseWriter, r *http.Request) {
		if !currentTokens.isAuthorized(r.Header) {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		respObj := struct {
			Code      int
			Addresses []*Address
		}{
			Code:      1000,
			Addresses: []*Address{},
		}
		if err := json.NewEncoder(w).Encode(respObj); err != nil {
			panic(err)
		}
	})

	ts := httptest.NewServer(mux)
	m := New(Config{HostURL: ts.URL})
	c, _, err := m.NewClientWithRefresh(context.Background(), testUID, testRef)
	r.NoError(err)

	// The request will fail with 401, triggering a refresh. After the refresh it should succeed.
	_, err = c.UpdateUser(context.Background())
	r.NoError(err)
}

func TestAuth2FA(t *testing.T) {
	r := require.New(t)
	twoFACode := "code"

	finish, c := newTestClientCallbacks(t,
		func(tb testing.TB, w http.ResponseWriter, req *http.Request) string {
			r.NoError(checkMethodAndPath(req, "POST", "/auth/2fa"))

			var twoFAreq auth2FAReq
			r.NoError(json.NewDecoder(req.Body).Decode(&twoFAreq))
			r.Equal(twoFAreq.TwoFactorCode, twoFACode)

			return "/auth/2fa/post_response.json"
		},
	)
	defer finish()

	err := c.Auth2FA(context.Background(), twoFACode)
	r.NoError(err)
}

func TestAuth2FA_Fail(t *testing.T) {
	r := require.New(t)
	finish, c := newTestClientCallbacks(t,
		func(tb testing.TB, w http.ResponseWriter, req *http.Request) string {
			r.NoError(checkMethodAndPath(req, "POST", "/auth/2fa"))
			return "/auth/2fa/post_401_bad_password.json"
		},
	)
	defer finish()

	err := c.Auth2FA(context.Background(), "code")
	r.Equal(ErrBad2FACode, err)
}

func TestAuth2FA_Retry(t *testing.T) {
	r := require.New(t)
	finish, c := newTestClientCallbacks(t,
		func(tb testing.TB, w http.ResponseWriter, req *http.Request) string {
			r.NoError(checkMethodAndPath(req, "POST", "/auth/2fa"))
			return "/auth/2fa/post_422_bad_password.json"
		},
	)
	defer finish()

	err := c.Auth2FA(context.Background(), "code")
	r.Equal(ErrBad2FACodeTryAgain, err)
}
