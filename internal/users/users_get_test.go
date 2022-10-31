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

package users

import (
	"testing"

	r "github.com/stretchr/testify/require"
)

func TestGetNoUser(t *testing.T) {
	m := initMocks(t)
	defer m.ctrl.Finish()

	checkUsersGetUser(t, m, "nouser", -1, "user nouser not found")
}

func TestGetUserByID(t *testing.T) {
	m := initMocks(t)
	defer m.ctrl.Finish()

	checkUsersGetUser(t, m, "user", 0, "")
	checkUsersGetUser(t, m, "users", 1, "")
}

func TestGetUserByName(t *testing.T) {
	m := initMocks(t)
	defer m.ctrl.Finish()

	checkUsersGetUser(t, m, "username", 0, "")
	checkUsersGetUser(t, m, "usersname", 1, "")
}

func TestGetUserByEmail(t *testing.T) {
	m := initMocks(t)
	defer m.ctrl.Finish()

	checkUsersGetUser(t, m, "user@pm.me", 0, "")
	checkUsersGetUser(t, m, "users@pm.me", 1, "")
	checkUsersGetUser(t, m, "anotheruser@pm.me", 1, "")
	checkUsersGetUser(t, m, "alsouser@pm.me", 1, "")
}

func checkUsersGetUser(t *testing.T, m mocks, query string, index int, expectedError string) {
	users := testNewUsersWithUsers(t, m)
	defer cleanUpUsersData(users)

	user, err := users.GetUser(query)

	if expectedError != "" {
		r.EqualError(m.t, err, expectedError)
	} else {
		r.NoError(m.t, err)
	}

	var expectedUser *User
	if index >= 0 {
		expectedUser = users.users[index]
	}
	r.Equal(m.t, expectedUser, user)
}
