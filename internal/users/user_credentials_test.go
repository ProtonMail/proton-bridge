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

package users

import (
	"testing"

	"github.com/ProtonMail/proton-bridge/internal/events"
	"github.com/ProtonMail/proton-bridge/pkg/pmapi"
	gomock "github.com/golang/mock/gomock"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
)

func TestUpdateUser(t *testing.T) {
	m := initMocks(t)
	defer m.ctrl.Finish()

	user := testNewUser(m)
	defer cleanUpUserData(user)

	gomock.InOrder(
		m.pmapiClient.EXPECT().IsUnlocked().Return(false),
		m.pmapiClient.EXPECT().Unlock([]byte("pass")).Return(nil),

		m.pmapiClient.EXPECT().UpdateUser().Return(nil, nil),
		m.pmapiClient.EXPECT().ReloadKeys([]byte(testCredentials.MailboxPassword)).Return(nil),
		m.pmapiClient.EXPECT().Addresses().Return([]*pmapi.Address{testPMAPIAddress}),

		m.credentialsStore.EXPECT().UpdateEmails("user", []string{testPMAPIAddress.Email}),
		m.credentialsStore.EXPECT().Get("user").Return(testCredentials, nil),
	)

	gomock.InOrder(
		m.pmapiClient.EXPECT().GetEvent(testPMAPIEvent.EventID).Return(testPMAPIEvent, nil).MaxTimes(1),
		m.pmapiClient.EXPECT().ListMessages(gomock.Any()).Return([]*pmapi.Message{}, 0, nil).MaxTimes(1),
	)

	assert.NoError(t, user.UpdateUser())

	waitForEvents()
}

func TestUserSwitchAddressMode(t *testing.T) {
	m := initMocks(t)
	defer m.ctrl.Finish()

	user := testNewUser(m)
	defer cleanUpUserData(user)

	assert.True(t, user.store.IsCombinedMode())
	assert.True(t, user.creds.IsCombinedAddressMode)
	assert.True(t, user.IsCombinedAddressMode())
	waitForEvents()

	gomock.InOrder(
		m.eventListener.EXPECT().Emit(events.CloseConnectionEvent, "user@pm.me"),
		m.pmapiClient.EXPECT().ListLabels().Return([]*pmapi.Label{}, nil),
		m.pmapiClient.EXPECT().CountMessages("").Return([]*pmapi.MessagesCount{}, nil),
		m.pmapiClient.EXPECT().Addresses().Return([]*pmapi.Address{testPMAPIAddress}),

		m.credentialsStore.EXPECT().SwitchAddressMode("user").Return(nil),
		m.credentialsStore.EXPECT().Get("user").Return(testCredentialsSplit, nil),
	)

	assert.NoError(t, user.SwitchAddressMode())
	assert.False(t, user.store.IsCombinedMode())
	assert.False(t, user.creds.IsCombinedAddressMode)
	assert.False(t, user.IsCombinedAddressMode())

	gomock.InOrder(
		m.eventListener.EXPECT().Emit(events.CloseConnectionEvent, "users@pm.me"),
		m.eventListener.EXPECT().Emit(events.CloseConnectionEvent, "anotheruser@pm.me"),
		m.eventListener.EXPECT().Emit(events.CloseConnectionEvent, "alsouser@pm.me"),
		m.pmapiClient.EXPECT().ListLabels().Return([]*pmapi.Label{}, nil),
		m.pmapiClient.EXPECT().CountMessages("").Return([]*pmapi.MessagesCount{}, nil),
		m.pmapiClient.EXPECT().Addresses().Return([]*pmapi.Address{testPMAPIAddress}),

		m.credentialsStore.EXPECT().SwitchAddressMode("user").Return(nil),
		m.credentialsStore.EXPECT().Get("user").Return(testCredentials, nil),
	)
	m.pmapiClient.EXPECT().ListMessages(gomock.Any()).Return([]*pmapi.Message{}, 0, nil).AnyTimes()

	assert.NoError(t, user.SwitchAddressMode())
	assert.True(t, user.store.IsCombinedMode())
	assert.True(t, user.creds.IsCombinedAddressMode)
	assert.True(t, user.IsCombinedAddressMode())

	waitForEvents()
}

func TestLogoutUser(t *testing.T) {
	m := initMocks(t)
	defer m.ctrl.Finish()

	user := testNewUserForLogout(m)
	defer cleanUpUserData(user)

	gomock.InOrder(
		m.pmapiClient.EXPECT().Logout().Return(),
		m.credentialsStore.EXPECT().Logout("user").Return(nil),
		m.credentialsStore.EXPECT().Get("user").Return(testCredentialsDisconnected, nil),
	)

	m.eventListener.EXPECT().Emit(events.CloseConnectionEvent, "user@pm.me")

	err := user.Logout()

	waitForEvents()

	assert.NoError(t, err)
}

func TestLogoutUserFailsLogout(t *testing.T) {
	m := initMocks(t)
	defer m.ctrl.Finish()

	user := testNewUserForLogout(m)
	defer cleanUpUserData(user)

	gomock.InOrder(
		m.pmapiClient.EXPECT().Logout().Return(),
		m.credentialsStore.EXPECT().Logout("user").Return(errors.New("logout failed")),
		m.credentialsStore.EXPECT().Delete("user").Return(nil),
		m.credentialsStore.EXPECT().Get("user").Return(testCredentialsDisconnected, nil),
	)
	m.eventListener.EXPECT().Emit(events.CloseConnectionEvent, "user@pm.me")

	err := user.Logout()
	waitForEvents()
	assert.NoError(t, err)
}

func TestCheckBridgeLoginOK(t *testing.T) {
	m := initMocks(t)
	defer m.ctrl.Finish()

	user := testNewUser(m)
	defer cleanUpUserData(user)

	gomock.InOrder(
		m.pmapiClient.EXPECT().IsUnlocked().Return(false),
		m.pmapiClient.EXPECT().Unlock([]byte("pass")).Return(nil),
	)

	err := user.CheckBridgeLogin(testCredentials.BridgePassword)

	waitForEvents()

	assert.NoError(t, err)
}

func TestCheckBridgeLoginTwiceOK(t *testing.T) {
	m := initMocks(t)
	defer m.ctrl.Finish()

	user := testNewUser(m)
	defer cleanUpUserData(user)

	gomock.InOrder(
		m.pmapiClient.EXPECT().IsUnlocked().Return(false),
		m.pmapiClient.EXPECT().Unlock([]byte("pass")).Return(nil),
		m.pmapiClient.EXPECT().IsUnlocked().Return(true),
	)

	err := user.CheckBridgeLogin(testCredentials.BridgePassword)
	waitForEvents()
	assert.NoError(t, err)

	err = user.CheckBridgeLogin(testCredentials.BridgePassword)
	waitForEvents()
	assert.NoError(t, err)
}

func TestCheckBridgeLoginUpgradeApplication(t *testing.T) {
	m := initMocks(t)
	defer m.ctrl.Finish()

	user := testNewUser(m)
	defer cleanUpUserData(user)

	m.eventListener.EXPECT().Emit(events.UpgradeApplicationEvent, "")

	isApplicationOutdated = true

	err := user.CheckBridgeLogin("any-pass")
	waitForEvents()
	assert.Equal(t, pmapi.ErrUpgradeApplication, err)

	isApplicationOutdated = false
}

func TestCheckBridgeLoginLoggedOut(t *testing.T) {
	m := initMocks(t)
	defer m.ctrl.Finish()

	m.credentialsStore.EXPECT().Get("user").Return(testCredentialsDisconnected, nil)

	user, err := newUser(m.PanicHandler, "user", m.eventListener, m.credentialsStore, m.clientManager, m.storeMaker)
	assert.NoError(t, err)

	m.clientManager.EXPECT().GetClient(gomock.Any()).Return(m.pmapiClient).MinTimes(1)
	gomock.InOrder(
		m.credentialsStore.EXPECT().Get("user").Return(testCredentialsDisconnected, nil),
		m.pmapiClient.EXPECT().ListLabels().Return(nil, errors.New("ErrUnauthorized")),
		m.pmapiClient.EXPECT().Addresses().Return(nil),
	)

	err = user.init(nil)
	assert.Error(t, err)

	defer cleanUpUserData(user)

	m.eventListener.EXPECT().Emit(events.LogoutEvent, "user")

	err = user.CheckBridgeLogin(testCredentialsDisconnected.BridgePassword)
	waitForEvents()
	assert.Equal(t, ErrLoggedOutUser, err)
}

func TestCheckBridgeLoginBadPassword(t *testing.T) {
	m := initMocks(t)
	defer m.ctrl.Finish()

	user := testNewUser(m)
	defer cleanUpUserData(user)

	gomock.InOrder(
		m.pmapiClient.EXPECT().IsUnlocked().Return(false),
		m.pmapiClient.EXPECT().Unlock([]byte("pass")).Return(nil),
	)

	err := user.CheckBridgeLogin("wrong!")
	waitForEvents()
	assert.Equal(t, "backend/credentials: incorrect password", err.Error())
}
