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
	"net/http"
	"testing"

	"github.com/ProtonMail/gopenpgp/v2/crypto"
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	r "github.com/stretchr/testify/require"
)

var testCurrentUser = &User{
	ID:         "MJLke8kWh1BBvG95JBIrZvzpgsZ94hNNgjNHVyhXMiv4g9cn6SgvqiIFR5cigpml2LD_iUk_3DkV29oojTt3eA==",
	Name:       "jason",
	UsedSpace:  96691332,
	Currency:   "USD",
	Role:       2,
	Subscribed: 1,
	Services:   1,
	MaxSpace:   10737418240,
	MaxUpload:  26214400,
	Private:    1,
	Keys:       *loadPMKeys(readTestFile("keyring_userKey_JSON", false)),
}

func routeGetUsers(tb testing.TB, w http.ResponseWriter, req *http.Request) string {
	r.NoError(tb, checkMethodAndPath(req, "GET", "/users"))
	r.NoError(tb, isAuthReq(req, testUID, testAccessToken))
	return "users/get_response.json"
}

func TestClient_CurrentUser(t *testing.T) {
	finish, c := newTestClientCallbacks(t,
		routeGetUsers,
		routeGetAddresses,
	)
	defer finish()

	user, err := c.CurrentUser(context.Background())
	r.Nil(t, err)

	// Ignore KeyRings during the check because they have unexported fields and cannot be compared
	r.True(t, cmp.Equal(user, testCurrentUser, cmpopts.IgnoreTypes(&crypto.Key{})))

	r.Nil(t, c.Unlock(context.Background(), []byte(testMailboxPassword)))
}
