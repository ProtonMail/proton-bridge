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
	"context"
	"testing"

	"github.com/ProtonMail/proton-bridge/v2/internal/events"
	"github.com/ProtonMail/proton-bridge/v2/pkg/pmapi"
	gomock "github.com/golang/mock/gomock"
	"github.com/pkg/errors"
	r "github.com/stretchr/testify/require"
)

func TestUpdateUser(t *testing.T) {
	m := initMocks(t)
	defer m.ctrl.Finish()

	user := testNewUser(t, m)
	defer cleanUpUserData(user)

	gomock.InOrder(
		m.pmapiClient.EXPECT().UpdateUser(gomock.Any()).Return(testPMAPIUser, nil),
		m.pmapiClient.EXPECT().ReloadKeys(gomock.Any(), testCredentials.MailboxPassword).Return(nil),
		m.pmapiClient.EXPECT().Addresses().Return([]*pmapi.Address{testPMAPIAddress}),

		m.credentialsStore.EXPECT().UpdateEmails("user", []string{testPMAPIAddress.Email}).Return(testCredentials, nil),
		m.eventListener.EXPECT().Emit(events.UserRefreshEvent, "user"),
	)

	r.NoError(t, user.UpdateUser(context.Background()))
}

func TestUserSwitchAddressMode(t *testing.T) {
	m := initMocks(t)
	defer m.ctrl.Finish()

	user := testNewUser(t, m)
	defer cleanUpUserData(user)

	// Ignore any sync on background.
	m.pmapiClient.EXPECT().ListMessages(gomock.Any(), gomock.Any()).Return([]*pmapi.Message{}, 0, nil).AnyTimes()

	// Check initial state.
	r.True(t, user.store.IsCombinedMode())
	r.True(t, user.creds.IsCombinedAddressMode)
	r.True(t, user.IsCombinedAddressMode())

	// Mock change to split mode.
	gomock.InOrder(
		m.eventListener.EXPECT().Emit(events.CloseConnectionEvent, "user@pm.me"),
		m.pmapiClient.EXPECT().ListLabels(gomock.Any()).Return([]*pmapi.Label{}, nil),
		m.pmapiClient.EXPECT().CountMessages(gomock.Any(), "").Return([]*pmapi.MessagesCount{}, nil),
		m.pmapiClient.EXPECT().Addresses().Return([]*pmapi.Address{testPMAPIAddress}),
		m.credentialsStore.EXPECT().SwitchAddressMode("user").Return(testCredentialsSplit, nil),
		m.eventListener.EXPECT().Emit(events.UserRefreshEvent, "user"),
	)

	// Check switch to split mode.
	r.NoError(t, user.SwitchAddressMode())
	r.False(t, user.store.IsCombinedMode())
	r.False(t, user.creds.IsCombinedAddressMode)
	r.False(t, user.IsCombinedAddressMode())

	// Mock change to combined mode.
	gomock.InOrder(
		m.eventListener.EXPECT().Emit(events.CloseConnectionEvent, "users@pm.me"),
		m.eventListener.EXPECT().Emit(events.CloseConnectionEvent, "anotheruser@pm.me"),
		m.eventListener.EXPECT().Emit(events.CloseConnectionEvent, "alsouser@pm.me"),
		m.pmapiClient.EXPECT().ListLabels(gomock.Any()).Return([]*pmapi.Label{}, nil),
		m.pmapiClient.EXPECT().CountMessages(gomock.Any(), "").Return([]*pmapi.MessagesCount{}, nil),
		m.pmapiClient.EXPECT().Addresses().Return([]*pmapi.Address{testPMAPIAddress}),
		m.credentialsStore.EXPECT().SwitchAddressMode("user").Return(testCredentials, nil),
		m.eventListener.EXPECT().Emit(events.UserRefreshEvent, "user"),
	)

	// Check switch to combined mode.
	r.NoError(t, user.SwitchAddressMode())
	r.True(t, user.store.IsCombinedMode())
	r.True(t, user.creds.IsCombinedAddressMode)
	r.True(t, user.IsCombinedAddressMode())
}

func TestLogoutUser(t *testing.T) {
	m := initMocks(t)
	defer m.ctrl.Finish()

	user := testNewUser(t, m)
	defer cleanUpUserData(user)

	gomock.InOrder(
		m.pmapiClient.EXPECT().AuthDelete(gomock.Any()).Return(nil),
		m.credentialsStore.EXPECT().Logout("user").Return(testCredentialsDisconnected, nil),
		m.eventListener.EXPECT().Emit(events.CloseConnectionEvent, "user@pm.me"),
		m.eventListener.EXPECT().Emit(events.UserRefreshEvent, "user"),
	)

	err := user.Logout()
	r.NoError(t, err)
}

func TestLogoutUserFailsLogout(t *testing.T) {
	m := initMocks(t)
	defer m.ctrl.Finish()

	user := testNewUser(t, m)
	defer cleanUpUserData(user)

	gomock.InOrder(
		m.pmapiClient.EXPECT().AuthDelete(gomock.Any()).Return(nil),
		m.credentialsStore.EXPECT().Logout("user").Return(nil, errors.New("logout failed")),
		m.credentialsStore.EXPECT().Delete("user").Return(nil),
		m.eventListener.EXPECT().Emit(events.CloseConnectionEvent, "user@pm.me"),
		m.eventListener.EXPECT().Emit(events.UserRefreshEvent, "user"),
	)

	err := user.Logout()
	r.NoError(t, err)
}

func TestCheckBridgeLogin(t *testing.T) {
	m := initMocks(t)
	defer m.ctrl.Finish()

	user := testNewUser(t, m)
	defer cleanUpUserData(user)

	err := user.CheckBridgeLogin(testCredentials.BridgePassword)
	r.NoError(t, err)
}

func TestCheckBridgeLoginUpgradeApplication(t *testing.T) {
	m := initMocks(t)
	defer m.ctrl.Finish()

	user := testNewUser(t, m)
	defer cleanUpUserData(user)

	m.eventListener.EXPECT().Emit(events.UpgradeApplicationEvent, "")

	isApplicationOutdated = true

	err := user.CheckBridgeLogin("any-pass")
	r.Equal(t, pmapi.ErrUpgradeApplication, err)

	isApplicationOutdated = false
}

func TestCheckBridgeLoginLoggedOut(t *testing.T) {
	m := initMocks(t)
	defer m.ctrl.Finish()

	gomock.InOrder(
		// Mock init of user.
		m.credentialsStore.EXPECT().Get("user").Return(testCredentialsDisconnected, nil),
		m.pmapiClient.EXPECT().AddAuthRefreshHandler(gomock.Any()),
		m.pmapiClient.EXPECT().ListLabels(gomock.Any()).Return(nil, pmapi.ErrUnauthorized),
		m.pmapiClient.EXPECT().Addresses().Return(nil),

		// Mock CheckBridgeLogin.
		m.eventListener.EXPECT().Emit(events.LogoutEvent, "user"),
	)

	user, _, err := newUser(m.PanicHandler, "user", m.eventListener, m.credentialsStore, m.storeMaker)
	r.NoError(t, err)

	err = user.connect(m.pmapiClient, testCredentialsDisconnected)
	r.Error(t, err)
	defer cleanUpUserData(user)

	err = user.CheckBridgeLogin(testCredentialsDisconnected.BridgePassword)
	r.Equal(t, ErrLoggedOutUser, err)
}

func TestCheckBridgeLoginBadPassword(t *testing.T) {
	m := initMocks(t)
	defer m.ctrl.Finish()

	user := testNewUser(t, m)
	defer cleanUpUserData(user)

	err := user.CheckBridgeLogin("wrong!")
	r.EqualError(t, err, "backend/credentials: incorrect password")
}
