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
	time "time"

	"github.com/ProtonMail/proton-bridge/v2/internal/events"
	"github.com/ProtonMail/proton-bridge/v2/internal/users/credentials"
	"github.com/ProtonMail/proton-bridge/v2/pkg/pmapi"
	gomock "github.com/golang/mock/gomock"
	r "github.com/stretchr/testify/require"
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

func TestNewUsersWithConnectedUser(t *testing.T) {
	m := initMocks(t)
	defer m.ctrl.Finish()

	m.credentialsStore.EXPECT().List().Return([]string{testCredentials.UserID}, nil)
	mockLoadingConnectedUser(t, m, testCredentials)
	mockEventLoopNoAction(m)
	checkUsersNew(t, m, []*credentials.Credentials{testCredentials})
}

func TestNewUsersWithDisconnectedUser(t *testing.T) {
	m := initMocks(t)
	defer m.ctrl.Finish()

	m.credentialsStore.EXPECT().List().Return([]string{testCredentialsDisconnected.UserID}, nil)
	mockLoadingDisconnectedUser(m, testCredentialsDisconnected)
	checkUsersNew(t, m, []*credentials.Credentials{testCredentialsDisconnected})
}

// Tests two users with different states and checks also the order from
// credentials store is kept also in array of users.
func TestNewUsersWithUsers(t *testing.T) {
	m := initMocks(t)
	defer m.ctrl.Finish()

	m.credentialsStore.EXPECT().List().Return([]string{testCredentialsDisconnected.UserID, testCredentials.UserID}, nil)
	mockLoadingDisconnectedUser(m, testCredentialsDisconnected)
	mockLoadingConnectedUser(t, m, testCredentials)
	mockEventLoopNoAction(m)
	checkUsersNew(t, m, []*credentials.Credentials{testCredentialsDisconnected, testCredentials})
}

func TestNewUsersWithConnectedUserWithBadToken(t *testing.T) {
	m := initMocks(t)
	defer m.ctrl.Finish()

	m.clientManager.EXPECT().NewClientWithRefresh(gomock.Any(), "uid", "acc").Return(nil, nil, pmapi.ErrAuthFailed{OriginalError: errors.New("bad token")})
	m.clientManager.EXPECT().NewClient("uid", "", "acc", time.Time{}).Return(m.pmapiClient)
	m.pmapiClient.EXPECT().AddAuthRefreshHandler(gomock.Any())
	m.pmapiClient.EXPECT().IsUnlocked().Return(false)
	m.pmapiClient.EXPECT().Unlock(gomock.Any(), testCredentials.MailboxPassword).Return(pmapi.ErrAuthFailed{OriginalError: errors.New("not authorized")})
	m.pmapiClient.EXPECT().AuthDelete(gomock.Any())

	m.credentialsStore.EXPECT().List().Return([]string{"user"}, nil)
	m.credentialsStore.EXPECT().Get("user").Return(testCredentials, nil)
	m.credentialsStore.EXPECT().Logout("user").Return(testCredentialsDisconnected, nil)

	m.eventListener.EXPECT().Emit(events.UserRefreshEvent, "user")
	m.eventListener.EXPECT().Emit(events.LogoutEvent, "user")
	m.eventListener.EXPECT().Emit(events.CloseConnectionEvent, "user@pm.me")

	checkUsersNew(t, m, []*credentials.Credentials{testCredentialsDisconnected})
}

func checkUsersNew(t *testing.T, m mocks, expectedCredentials []*credentials.Credentials) {
	users := testNewUsers(t, m)
	defer cleanUpUsersData(users)

	r.Equal(m.t, len(expectedCredentials), len(users.GetUsers()))

	credentials := []*credentials.Credentials{}
	for _, user := range users.users {
		credentials = append(credentials, user.creds)
	}

	r.Equal(m.t, expectedCredentials, credentials)
}
