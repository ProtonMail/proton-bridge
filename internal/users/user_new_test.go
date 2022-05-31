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
	"github.com/ProtonMail/proton-bridge/v2/internal/users/credentials"
	"github.com/ProtonMail/proton-bridge/v2/pkg/pmapi"
	gomock "github.com/golang/mock/gomock"
	r "github.com/stretchr/testify/require"
)

func TestNewUserNoCredentialsStore(t *testing.T) {
	m := initMocks(t)
	defer m.ctrl.Finish()

	m.credentialsStore.EXPECT().Get("user").Return(nil, errors.New("fail"))

	_, _, err := newUser(m.PanicHandler, "user", m.eventListener, m.credentialsStore, m.storeMaker)
	r.Error(t, err)
}

func TestNewUserUnlockFails(t *testing.T) {
	m := initMocks(t)
	defer m.ctrl.Finish()

	gomock.InOrder(
		// Init of user.
		m.credentialsStore.EXPECT().Get("user").Return(testCredentials, nil),
		m.pmapiClient.EXPECT().AddAuthRefreshHandler(gomock.Any()),
		m.pmapiClient.EXPECT().IsUnlocked().Return(false),
		m.pmapiClient.EXPECT().Unlock(gomock.Any(), testCredentials.MailboxPassword).Return(pmapi.ErrUnlockFailed{OriginalError: errors.New("bad password")}),

		// Handle of unlock error.
		m.pmapiClient.EXPECT().AuthDelete(gomock.Any()).Return(nil),
		m.credentialsStore.EXPECT().Logout("user").Return(testCredentialsDisconnected, nil),
		m.eventListener.EXPECT().Emit(events.CloseConnectionEvent, "user@pm.me"),
		m.eventListener.EXPECT().Emit(events.UserRefreshEvent, "user"),
		m.eventListener.EXPECT().Emit(events.LogoutEvent, "user"),
	)

	checkNewUserHasCredentials(m, "failed to unlock user: bad password", testCredentialsDisconnected)
}

func TestNewUser(t *testing.T) {
	m := initMocks(t)
	defer m.ctrl.Finish()

	m.credentialsStore.EXPECT().Get("user").Return(testCredentials, nil)
	mockInitConnectedUser(t, m)
	mockEventLoopNoAction(m)

	checkNewUserHasCredentials(m, "", testCredentials)
}

func checkNewUserHasCredentials(m mocks, wantErr string, wantCreds *credentials.Credentials) {
	user, _, err := newUser(m.PanicHandler, "user", m.eventListener, m.credentialsStore, m.storeMaker)
	r.NoError(m.t, err)
	defer cleanUpUserData(user)

	err = user.connect(m.pmapiClient, testCredentials)
	if wantErr == "" {
		r.NoError(m.t, err)
	} else {
		r.EqualError(m.t, err, wantErr)
	}

	r.Equal(m.t, wantCreds, user.creds)

	waitForEvents()
}
