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
	"net/http"
	"testing"
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

func routeGetAddresses(tb testing.TB, w http.ResponseWriter, r *http.Request) string {
	Ok(tb, checkMethodAndPath(r, "GET", "/addresses"))
	Ok(tb, isAuthReq(r, testUID, testAccessToken))
	return "addresses/get_response.json"
}

func TestAddressList(t *testing.T) {
	input := "1"
	addr := testAddressList.ByID(input)
	if addr != testAddressList[0] {
		t.Errorf("ById(%s) expected:\n%v\n but have:\n%v\n", input, testAddressList[0], addr)
	}

	input = "42"
	addr = testAddressList.ByID(input)
	if addr != nil {
		t.Errorf("ById expected nil for %s but have : %v\n", input, addr)
	}

	input = "root@protonmail.com"
	addr = testAddressList.ByEmail(input)
	if addr != testAddressList[2] {
		t.Errorf("ByEmail(%s) expected:\n%v\n but have:\n%v\n", input, testAddressList[2], addr)
	}

	input = "idontexist@protonmail.com"
	addr = testAddressList.ByEmail(input)
	if addr != nil {
		t.Errorf("ByEmail expected nil for %s but have : %v\n", input, addr)
	}

	addr = testAddressList.Main()
	if addr != testAddressList[1] {
		t.Errorf("Main() expected:\n%v\n but have:\n%v\n", testAddressList[1], addr)
	}
}
