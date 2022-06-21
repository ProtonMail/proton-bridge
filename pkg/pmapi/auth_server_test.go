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
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/stretchr/testify/require"
)

type testRefreshResponse struct {
	Code         int
	AccessToken  string
	ExpiresIn    int
	TokenType    string
	Scope        string
	Scopes       []string
	UID          string
	RefreshToken string
	LocalID      int

	r *require.Assertions
}

var tokenID = 0

func newTestRefreshToken(r *require.Assertions) testRefreshResponse {
	tokenID++
	scopes := []string{
		"full",
		"self",
		"parent",
		"user",
		"loggedin",
		"paid",
		"nondelinquent",
		"mail",
		"verified",
	}
	return testRefreshResponse{
		Code:         1000,
		AccessToken:  fmt.Sprintf("acc%d", tokenID),
		ExpiresIn:    3600,
		TokenType:    "Bearer",
		Scope:        strings.Join(scopes, " "),
		Scopes:       scopes,
		UID:          fmt.Sprintf("uid%d", tokenID),
		RefreshToken: fmt.Sprintf("ref%d", tokenID),
		r:            r,
	}
}

func (r *testRefreshResponse) isCorrectRefreshToken(body io.ReadCloser) int {
	request := authRefreshReq{}
	err := json.NewDecoder(body).Decode(&request)
	r.r.NoError(body.Close())
	r.r.NoError(err)

	if r.UID != request.UID {
		return http.StatusUnprocessableEntity
	}
	if r.RefreshToken != request.RefreshToken {
		return http.StatusBadRequest
	}
	return http.StatusOK
}

func (r *testRefreshResponse) handleAuthRefresh(response http.ResponseWriter, request *http.Request) {
	if code := r.isCorrectRefreshToken(request.Body); code != http.StatusOK {
		response.WriteHeader(code)
		return
	}

	tokenID++
	r.AccessToken = fmt.Sprintf("acc%d", tokenID)
	r.RefreshToken = fmt.Sprintf("ref%d", tokenID)

	response.Header().Set("Content-Type", "application/json")
	response.WriteHeader(http.StatusOK)
	r.r.NoError(json.NewEncoder(response).Encode(r))
}

func (r *testRefreshResponse) wantAuthRefresh() AuthRefresh {
	return AuthRefresh{
		UID:          r.UID,
		AccessToken:  r.AccessToken,
		RefreshToken: r.RefreshToken,
		ExpiresIn:    int64(r.ExpiresIn),
		Scopes:       r.Scopes,
	}
}

func (r *testRefreshResponse) isAuthorized(header http.Header) bool {
	return header.Get("x-pm-uid") == r.UID && header.Get("Authorization") == "Bearer "+r.AccessToken
}

func (r *testRefreshResponse) handleAuthCheckOnly(response http.ResponseWriter, request *http.Request) {
	if r.isAuthorized(request.Header) {
		response.WriteHeader(http.StatusOK)
	} else {
		response.WriteHeader(http.StatusUnauthorized)
	}
}
