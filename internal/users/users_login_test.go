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
	"testing"

	"github.com/ProtonMail/proton-bridge/internal/events"
	"github.com/ProtonMail/proton-bridge/internal/metrics"
	"github.com/ProtonMail/proton-bridge/pkg/pmapi"
	gomock "github.com/golang/mock/gomock"
	"github.com/pkg/errors"
	r "github.com/stretchr/testify/require"
)

func TestUsersFinishLoginBadMailboxPassword(t *testing.T) {
	m := initMocks(t)
	defer m.ctrl.Finish()

	// Init users with no user from keychain.
	m.credentialsStore.EXPECT().List().Return([]string{}, nil)

	// Set up mocks for FinishLogin.
	m.pmapiClient.EXPECT().AuthSalt(gomock.Any()).Return("", nil)
	m.pmapiClient.EXPECT().Unlock(gomock.Any(), []byte(testCredentials.MailboxPassword)).Return(errors.New("no keys could be unlocked"))

	checkUsersFinishLogin(t, m, testAuthRefresh, testCredentials.MailboxPassword, "", ErrWrongMailboxPassword)
}

func TestUsersFinishLoginNewUser(t *testing.T) {
	m := initMocks(t)
	defer m.ctrl.Finish()

	// Init users with no user from keychain.
	m.credentialsStore.EXPECT().List().Return([]string{}, nil)

	mockAddingConnectedUser(m)
	mockEventLoopNoAction(m)

	m.clientManager.EXPECT().SendSimpleMetric(gomock.Any(), string(metrics.Setup), string(metrics.NewUser), string(metrics.NoLabel))
	m.eventListener.EXPECT().Emit(events.UserRefreshEvent, testCredentials.UserID)

	checkUsersFinishLogin(t, m, testAuthRefresh, testCredentials.MailboxPassword, testCredentials.UserID, nil)
}

func TestUsersFinishLoginExistingDisconnectedUser(t *testing.T) {
	m := initMocks(t)
	defer m.ctrl.Finish()

	// Mock loading disconnected user.
	m.credentialsStore.EXPECT().List().Return([]string{testCredentialsDisconnected.UserID}, nil)
	mockLoadingDisconnectedUser(m, testCredentialsDisconnected)

	// Mock process of FinishLogin of already added user.
	gomock.InOrder(
		m.pmapiClient.EXPECT().AuthSalt(gomock.Any()).Return("", nil),
		m.pmapiClient.EXPECT().Unlock(gomock.Any(), []byte(testCredentials.MailboxPassword)).Return(nil),
		m.pmapiClient.EXPECT().CurrentUser(gomock.Any()).Return(testPMAPIUserDisconnected, nil),
		m.credentialsStore.EXPECT().UpdateToken(testCredentialsDisconnected.UserID, testAuthRefresh.UID, testAuthRefresh.RefreshToken).Return(testCredentials, nil),
		m.credentialsStore.EXPECT().UpdatePassword(testCredentialsDisconnected.UserID, testCredentials.MailboxPassword).Return(testCredentials, nil),
	)
	mockInitConnectedUser(m)
	mockEventLoopNoAction(m)
	m.eventListener.EXPECT().Emit(events.UserRefreshEvent, testCredentialsDisconnected.UserID)

	authRefresh := &pmapi.Auth{
		UserID: testCredentialsDisconnected.UserID,
		AuthRefresh: pmapi.AuthRefresh{
			UID:          "uid",
			AccessToken:  "acc",
			RefreshToken: "ref",
		},
	}
	checkUsersFinishLogin(t, m, authRefresh, testCredentials.MailboxPassword, testCredentialsDisconnected.UserID, nil)
}

func TestUsersFinishLoginConnectedUser(t *testing.T) {
	m := initMocks(t)
	defer m.ctrl.Finish()

	// Mock loading connected user.
	m.credentialsStore.EXPECT().List().Return([]string{testCredentials.UserID}, nil)
	mockLoadingConnectedUser(m, testCredentials)
	mockEventLoopNoAction(m)

	// Mock process of FinishLogin of already connected user.
	gomock.InOrder(
		m.pmapiClient.EXPECT().AuthSalt(gomock.Any()).Return("", nil),
		m.pmapiClient.EXPECT().Unlock(gomock.Any(), []byte(testCredentials.MailboxPassword)).Return(nil),
		m.pmapiClient.EXPECT().CurrentUser(gomock.Any()).Return(testPMAPIUser, nil),
		m.pmapiClient.EXPECT().AuthDelete(gomock.Any()).Return(nil),
	)

	users := testNewUsers(t, m)
	defer cleanUpUsersData(users)

	_, err := users.FinishLogin(m.pmapiClient, testAuthRefresh, testCredentials.MailboxPassword)
	r.EqualError(t, err, "user is already connected")
}

func checkUsersFinishLogin(t *testing.T, m mocks, auth *pmapi.Auth, mailboxPassword string, expectedUserID string, expectedErr error) {
	users := testNewUsers(t, m)
	defer cleanUpUsersData(users)

	user, err := users.FinishLogin(m.pmapiClient, auth, mailboxPassword)

	r.Equal(t, expectedErr, err)

	if expectedUserID != "" {
		r.Equal(t, expectedUserID, user.ID())
		r.Equal(t, 1, len(users.users))
		r.Equal(t, expectedUserID, users.users[0].ID())
	} else {
		r.Equal(t, (*User)(nil), user)
		r.Equal(t, 0, len(users.users))
	}
}
