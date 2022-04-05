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
	"net/http"
	"testing"

	r "github.com/stretchr/testify/require"
)

var testAddressList = AddressList{
	&Address{
		ID:     "1",
		Email:  "root@nsa.gov",
		Send:   SecondarySendAddress,
		Status: EnabledAddress,
		Order:  2,
	},
	&Address{
		ID:     "2",
		Email:  "root@gchq.gov.uk",
		Send:   MainSendAddress,
		Status: EnabledAddress,
		Order:  1,
	},
	&Address{
		ID:     "3",
		Email:  "root@protonmail.com",
		Send:   NoSendAddress,
		Status: DisabledAddress,
		Order:  3,
	},
}

func routeGetAddresses(tb testing.TB, w http.ResponseWriter, req *http.Request) string {
	r.NoError(tb, checkMethodAndPath(req, "GET", "/addresses"))
	r.NoError(tb, isAuthReq(req, testUID, testAccessToken))
	return "addresses/get_response.json"
}

func TestAddressList(t *testing.T) {
	input := "1"
	addr := testAddressList.ByID(input)
	r.Equal(t, testAddressList[0], addr)

	input = "42"
	addr = testAddressList.ByID(input)
	r.Nil(t, addr)

	input = "root@protonmail.com"
	addr = testAddressList.ByEmail(input)
	r.Equal(t, testAddressList[2], addr)

	input = "idontexist@protonmail.com"
	addr = testAddressList.ByEmail(input)
	r.Nil(t, addr)

	addr = testAddressList.Main()
	r.Equal(t, testAddressList[1], addr)
}
