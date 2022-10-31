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
	"errors"
	"testing"

	"github.com/ProtonMail/proton-bridge/v2/internal/events"
	gomock "github.com/golang/mock/gomock"
	r "github.com/stretchr/testify/require"
)

func TestDeleteUser(t *testing.T) {
	m := initMocks(t)
	defer m.ctrl.Finish()

	users := testNewUsersWithUsers(t, m)
	defer cleanUpUsersData(users)

	gomock.InOrder(
		m.pmapiClient.EXPECT().AuthDelete(gomock.Any()).Return(nil),
		m.credentialsStore.EXPECT().Logout("user").Return(testCredentialsDisconnected, nil),
		m.credentialsStore.EXPECT().Delete("user").Return(nil),
	)
	m.eventListener.EXPECT().Emit(events.UserRefreshEvent, "user")
	m.eventListener.EXPECT().Emit(events.CloseConnectionEvent, "user@pm.me")
	m.eventListener.EXPECT().Emit(events.UserRefreshEvent, "user")

	err := users.DeleteUser("user", true)
	r.NoError(t, err)
	r.Equal(t, 1, len(users.users))
}

// Even when logout fails, delete is done.
func TestDeleteUserWithFailingLogout(t *testing.T) {
	m := initMocks(t)
	defer m.ctrl.Finish()

	users := testNewUsersWithUsers(t, m)
	defer cleanUpUsersData(users)

	gomock.InOrder(
		m.pmapiClient.EXPECT().AuthDelete(gomock.Any()).Return(nil),
		m.credentialsStore.EXPECT().Logout("user").Return(nil, errors.New("logout failed")),
		// Once called from user.Logout after failed creds.Logout as fallback, and once at the end of users.Logout.
		m.credentialsStore.EXPECT().Delete("user").Return(nil),
		m.credentialsStore.EXPECT().Delete("user").Return(nil),
	)

	m.eventListener.EXPECT().Emit(events.UserRefreshEvent, "user")
	m.eventListener.EXPECT().Emit(events.CloseConnectionEvent, "user@pm.me")
	m.eventListener.EXPECT().Emit(events.UserRefreshEvent, "user")

	err := users.DeleteUser("user", true)
	r.NoError(t, err)
	r.Equal(t, 1, len(users.users))
}
