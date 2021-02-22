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

package users

import (
	"errors"
	"testing"
	time "time"

	"github.com/ProtonMail/proton-bridge/internal/users/credentials"
	gomock "github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
)

func TestNewUsersNoKeychain(t *testing.T) {
	m := initMocks(t)
	defer m.ctrl.Finish()

	m.credentialsStore.EXPECT().List().Return([]string{}, errors.New("no keychain"))

	checkUsersNew(t, m, []*credentials.Credentials{})
}

func TestNewUsersWithoutUsersInCredentialsStore(t *testing.T) {
	m := initMocks(t)
	defer m.ctrl.Finish()

	m.credentialsStore.EXPECT().List().Return([]string{}, nil)

	checkUsersNew(t, m, []*credentials.Credentials{})
}

func TestNewUsersWithDisconnectedUser(t *testing.T) {
	m := initMocks(t)
	defer m.ctrl.Finish()

	gomock.InOrder(
		m.credentialsStore.EXPECT().List().Return([]string{"user"}, nil),
		m.credentialsStore.EXPECT().Get("user").Return(testCredentialsDisconnected, nil),
		m.clientManager.EXPECT().NewClient("", "", "", time.Time{}).Return(m.pmapiClient),
		m.pmapiClient.EXPECT().AddAuthHandler(gomock.Any()),
		m.pmapiClient.EXPECT().ListLabels(gomock.Any()).Return(nil, errors.New("ErrUnauthorized")),
		m.pmapiClient.EXPECT().Addresses().Return(nil),
	)

	checkUsersNew(t, m, []*credentials.Credentials{testCredentialsDisconnected})
}

/*
func TestNewUsersWithConnectedUserWithBadToken(t *testing.T) {
	m := initMocks(t)
	defer m.ctrl.Finish()

	m.clientManager.EXPECT().GetClient("user").Return(m.pmapiClient).MinTimes(1)

	m.credentialsStore.EXPECT().List().Return([]string{"user"}, nil)
	m.credentialsStore.EXPECT().Get("user").Return(testCredentials, nil).Times(2)

	m.credentialsStore.EXPECT().Logout("user").Return(nil)
	m.pmapiClient.EXPECT().AuthRefresh("token").Return(nil, errors.New("bad token"))

	m.eventListener.EXPECT().Emit(events.LogoutEvent, "user")
	m.eventListener.EXPECT().Emit(events.UserRefreshEvent, "user")
	m.pmapiClient.EXPECT().Logout()
	m.credentialsStore.EXPECT().Logout("user").Return(nil)
	m.credentialsStore.EXPECT().Get("user").Return(testCredentialsDisconnected, nil)
	m.eventListener.EXPECT().Emit(events.CloseConnectionEvent, "user@pm.me")

	checkUsersNew(t, m, []*credentials.Credentials{testCredentialsDisconnected})
}

func TestNewUsersWithConnectedUser(t *testing.T) {
	m := initMocks(t)
	defer m.ctrl.Finish()

	m.clientManager.EXPECT().GetClient("user").Return(m.pmapiClient).MinTimes(1)
	m.credentialsStore.EXPECT().List().Return([]string{"user"}, nil)

	mockConnectedUser(m)
	mockEventLoopNoAction(m)

	checkUsersNew(t, m, []*credentials.Credentials{testCredentials})
}

// Tests two users with different states and checks also the order from
// credentials store is kept also in array of users.
func TestNewUsersWithUsers(t *testing.T) {
	m := initMocks(t)
	defer m.ctrl.Finish()

	m.clientManager.EXPECT().GetClient("user").Return(m.pmapiClient).MinTimes(1)
	m.credentialsStore.EXPECT().List().Return([]string{"userDisconnected", "user"}, nil)

	gomock.InOrder(
		m.credentialsStore.EXPECT().Get("userDisconnected").Return(testCredentialsDisconnected, nil),
		m.credentialsStore.EXPECT().Get("userDisconnected").Return(testCredentialsDisconnected, nil),
		// Set up mocks for store initialisation for the unauth user.
		m.clientManager.EXPECT().GetClient("userDisconnected").Return(m.pmapiClient),
		m.pmapiClient.EXPECT().ListLabels().Return(nil, errors.New("ErrUnauthorized")),
		m.clientManager.EXPECT().GetClient("userDisconnected").Return(m.pmapiClient),
		m.pmapiClient.EXPECT().Addresses().Return(nil),
	)

	mockConnectedUser(m)

	mockEventLoopNoAction(m)

	checkUsersNew(t, m, []*credentials.Credentials{testCredentialsDisconnected, testCredentials})
}

func TestNewUsersFirstStart(t *testing.T) {
	m := initMocks(t)
	defer m.ctrl.Finish()

	m.credentialsStore.EXPECT().List().Return([]string{}, nil)

	testNewUsers(t, m)
}
*/

func checkUsersNew(t *testing.T, m mocks, expectedCredentials []*credentials.Credentials) {
	users := testNewUsers(t, m)
	defer cleanUpUsersData(users)

	assert.Equal(m.t, len(expectedCredentials), len(users.GetUsers()))

	credentials := []*credentials.Credentials{}
	for _, user := range users.users {
		credentials = append(credentials, user.creds)
	}

	assert.Equal(m.t, expectedCredentials, credentials)
}
