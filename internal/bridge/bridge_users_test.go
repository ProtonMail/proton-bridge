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

package bridge

import (
	"errors"
	"testing"

	"github.com/ProtonMail/proton-bridge/internal/events"
	gomock "github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
)

func TestGetNoUser(t *testing.T) {
	m := initMocks(t)
	defer m.ctrl.Finish()

	m.clientManager.EXPECT().GetClient("user").Return(m.pmapiClient).MinTimes(1)
	m.clientManager.EXPECT().GetClient("users").Return(m.pmapiClient).MinTimes(1)

	checkBridgeGetUser(t, m, "nouser", -1, "user nouser not found")
}

func TestGetUserByID(t *testing.T) {
	m := initMocks(t)
	defer m.ctrl.Finish()

	m.clientManager.EXPECT().GetClient("user").Return(m.pmapiClient).MinTimes(1)
	m.clientManager.EXPECT().GetClient("users").Return(m.pmapiClient).MinTimes(1)

	checkBridgeGetUser(t, m, "user", 0, "")
	checkBridgeGetUser(t, m, "users", 1, "")
}

func TestGetUserByName(t *testing.T) {
	m := initMocks(t)
	defer m.ctrl.Finish()

	m.clientManager.EXPECT().GetClient("user").Return(m.pmapiClient).MinTimes(1)
	m.clientManager.EXPECT().GetClient("users").Return(m.pmapiClient).MinTimes(1)

	checkBridgeGetUser(t, m, "username", 0, "")
	checkBridgeGetUser(t, m, "usersname", 1, "")
}

func TestGetUserByEmail(t *testing.T) {
	m := initMocks(t)
	defer m.ctrl.Finish()

	m.clientManager.EXPECT().GetClient("user").Return(m.pmapiClient).MinTimes(1)
	m.clientManager.EXPECT().GetClient("users").Return(m.pmapiClient).MinTimes(1)

	checkBridgeGetUser(t, m, "user@pm.me", 0, "")
	checkBridgeGetUser(t, m, "users@pm.me", 1, "")
	checkBridgeGetUser(t, m, "anotheruser@pm.me", 1, "")
	checkBridgeGetUser(t, m, "alsouser@pm.me", 1, "")
}

func TestDeleteUser(t *testing.T) {
	m := initMocks(t)
	defer m.ctrl.Finish()

	m.clientManager.EXPECT().GetClient("user").Return(m.pmapiClient).MinTimes(1)
	m.clientManager.EXPECT().GetClient("users").Return(m.pmapiClient).MinTimes(1)

	bridge := testNewBridgeWithUsers(t, m)
	defer cleanUpBridgeUserData(bridge)

	gomock.InOrder(
		m.pmapiClient.EXPECT().Logout().Return(),
		m.credentialsStore.EXPECT().Logout("user").Return(nil),
		m.credentialsStore.EXPECT().Get("user").Return(testCredentialsDisconnected, nil),
		m.credentialsStore.EXPECT().Delete("user").Return(nil),
	)

	m.eventListener.EXPECT().Emit(events.CloseConnectionEvent, "user@pm.me")

	err := bridge.DeleteUser("user", true)
	assert.NoError(t, err)
	assert.Equal(t, 1, len(bridge.users))
}

// Even when logout fails, delete is done.
func TestDeleteUserWithFailingLogout(t *testing.T) {
	m := initMocks(t)
	defer m.ctrl.Finish()

	m.clientManager.EXPECT().GetClient("user").Return(m.pmapiClient).MinTimes(1)
	m.clientManager.EXPECT().GetClient("users").Return(m.pmapiClient).MinTimes(1)

	bridge := testNewBridgeWithUsers(t, m)
	defer cleanUpBridgeUserData(bridge)

	gomock.InOrder(
		m.pmapiClient.EXPECT().Logout().Return(),
		m.credentialsStore.EXPECT().Logout("user").Return(errors.New("logout failed")),
		m.credentialsStore.EXPECT().Delete("user").Return(nil),
		m.credentialsStore.EXPECT().Get("user").Return(nil, errors.New("no such user")),
		m.credentialsStore.EXPECT().Delete("user").Return(nil),
		//TODO m.credentialsStore.EXPECT().Delete("user").Return(errors.New("no such user")),
	)

	m.eventListener.EXPECT().Emit(events.CloseConnectionEvent, "user@pm.me")

	err := bridge.DeleteUser("user", true)
	assert.NoError(t, err)
	assert.Equal(t, 1, len(bridge.users))
}

func checkBridgeGetUser(t *testing.T, m mocks, query string, index int, expectedError string) {
	bridge := testNewBridgeWithUsers(t, m)
	defer cleanUpBridgeUserData(bridge)

	user, err := bridge.GetUser(query)
	waitForEvents()

	if expectedError != "" {
		assert.Equal(m.t, expectedError, err.Error())
	} else {
		assert.NoError(m.t, err)
	}

	var expectedUser *User
	if index >= 0 {
		expectedUser = bridge.users[index]
	}

	assert.Equal(m.t, expectedUser, user)
}
