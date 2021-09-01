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
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	a "github.com/stretchr/testify/assert"
	r "github.com/stretchr/testify/require"
)

func TestAutomaticAuthRefresh(t *testing.T) {
	var wantAuthRefresh = &AuthRefresh{
		UID:          "testUID",
		AccessToken:  "testAcc",
		RefreshToken: "testRef",
		ExpiresIn:    100,
	}

	mux := http.NewServeMux()

	mux.HandleFunc("/auth/refresh", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		if err := json.NewEncoder(w).Encode(wantAuthRefresh); err != nil {
			panic(err)
		}
	})

	mux.HandleFunc("/addresses", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	ts := httptest.NewServer(mux)

	var gotAuthRefresh *AuthRefresh

	c := New(Config{HostURL: ts.URL}).
		NewClient("uid", "acc", "ref", time.Now().Add(-time.Second))

	c.AddAuthRefreshHandler(func(auth *AuthRefresh) { gotAuthRefresh = auth })

	// Make a request with an access token that already expired one second ago.
	_, err := c.GetAddresses(context.Background())
	r.NoError(t, err)

	// The auth callback should have been called.
	a.Equal(t, *wantAuthRefresh, *gotAuthRefresh)

	cl := c.(*client) //nolint[forcetypeassert] we want to panic here
	a.Equal(t, wantAuthRefresh.AccessToken, cl.acc)
	a.Equal(t, wantAuthRefresh.RefreshToken, cl.ref)
	a.WithinDuration(t, expiresIn(100), cl.exp, time.Second)
}

func Test401AuthRefresh(t *testing.T) {
	var wantAuthRefresh = &AuthRefresh{
		UID:          "testUID",
		AccessToken:  "testAcc",
		RefreshToken: "testRef",
	}

	mux := http.NewServeMux()

	mux.HandleFunc("/auth/refresh", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		if err := json.NewEncoder(w).Encode(wantAuthRefresh); err != nil {
			panic(err)
		}
	})

	var call int

	mux.HandleFunc("/addresses", func(w http.ResponseWriter, r *http.Request) {
		call++

		if call == 1 {
			w.WriteHeader(http.StatusUnauthorized)
		} else {
			w.WriteHeader(http.StatusOK)
		}
	})

	ts := httptest.NewServer(mux)

	var gotAuthRefresh *AuthRefresh

	// Create a new client.
	c := New(Config{HostURL: ts.URL}).
		NewClient("uid", "acc", "ref", time.Now().Add(time.Hour))

	// Register an auth handler.
	c.AddAuthRefreshHandler(func(auth *AuthRefresh) { gotAuthRefresh = auth })

	// The first request will fail with 401, triggering a refresh and retry.
	_, err := c.GetAddresses(context.Background())
	r.NoError(t, err)

	// The auth callback should have been called.
	r.Equal(t, *wantAuthRefresh, *gotAuthRefresh)
}

func Test401RevokedAuth(t *testing.T) {
	mux := http.NewServeMux()

	mux.HandleFunc("/auth/refresh", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
	})

	mux.HandleFunc("/addresses", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
	})

	ts := httptest.NewServer(mux)

	c := New(Config{HostURL: ts.URL}).
		NewClient("uid", "acc", "ref", time.Now().Add(time.Hour))

	// The request will fail with 401, triggering a refresh.
	// The retry will also fail with 401, returning an error.
	_, err := c.GetAddresses(context.Background())
	r.EqualError(t, err, ErrUnauthorized.Error())
}

func Test401RevokedAuthTokenUpdate(t *testing.T) {
	var oldAuth = &AuthRefresh{
		UID:          "UID",
		AccessToken:  "oldAcc",
		RefreshToken: "oldRef",
		ExpiresIn:    3600,
	}

	var newAuth = &AuthRefresh{
		UID:          "UID",
		AccessToken:  "newAcc",
		RefreshToken: "newRef",
	}

	mux := http.NewServeMux()

	mux.HandleFunc("/auth/refresh", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(newAuth); err != nil {
			panic(err)
		}
	})

	mux.HandleFunc("/addresses", func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") == ("Bearer " + oldAuth.AccessToken) {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		if r.Header.Get("Authorization") == ("Bearer " + newAuth.AccessToken) {
			w.WriteHeader(http.StatusOK)
			return
		}
	})

	ts := httptest.NewServer(mux)

	c := New(Config{HostURL: ts.URL}).
		NewClient(oldAuth.UID, oldAuth.AccessToken, oldAuth.RefreshToken, time.Now().Add(time.Hour))

	// The request will fail with 401, triggering a refresh. After the refresh it should succeed.
	_, err := c.GetAddresses(context.Background())
	r.NoError(t, err)
}

func TestAuth2FA(t *testing.T) {
	twoFACode := "code"

	finish, c := newTestClientCallbacks(t,
		func(tb testing.TB, w http.ResponseWriter, req *http.Request) string {
			r.NoError(t, checkMethodAndPath(req, "POST", "/auth/2fa"))

			var twoFAreq auth2FAReq
			r.NoError(t, json.NewDecoder(req.Body).Decode(&twoFAreq))
			r.Equal(t, twoFAreq.TwoFactorCode, twoFACode)

			return "/auth/2fa/post_response.json"
		},
	)
	defer finish()

	err := c.Auth2FA(context.Background(), twoFACode)
	r.NoError(t, err)
}

func TestAuth2FA_Fail(t *testing.T) {
	finish, c := newTestClientCallbacks(t,
		func(tb testing.TB, w http.ResponseWriter, req *http.Request) string {
			r.NoError(t, checkMethodAndPath(req, "POST", "/auth/2fa"))
			return "/auth/2fa/post_401_bad_password.json"
		},
	)
	defer finish()

	err := c.Auth2FA(context.Background(), "code")
	r.Equal(t, ErrBad2FACode, err)
}

func TestAuth2FA_Retry(t *testing.T) {
	finish, c := newTestClientCallbacks(t,
		func(tb testing.TB, w http.ResponseWriter, req *http.Request) string {
			r.NoError(t, checkMethodAndPath(req, "POST", "/auth/2fa"))
			return "/auth/2fa/post_422_bad_password.json"
		},
	)
	defer finish()

	err := c.Auth2FA(context.Background(), "code")
	r.Equal(t, ErrBad2FACodeTryAgain, err)
}
