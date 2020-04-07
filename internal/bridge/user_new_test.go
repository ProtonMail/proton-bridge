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
	"github.com/ProtonMail/proton-bridge/pkg/pmapi"
	gomock "github.com/golang/mock/gomock"
	a "github.com/stretchr/testify/assert"
)

func TestNewUserNoCredentialsStore(t *testing.T) {
	m := initMocks(t)
	defer m.ctrl.Finish()

	m.credentialsStore.EXPECT().Get("user").Return(nil, errors.New("fail"))

	_, err := newUser(m.PanicHandler, "user", m.eventListener, m.credentialsStore, m.clientManager, m.storeCache, "/tmp")
	a.Error(t, err)
}

func TestNewUserBridgeOutdated(t *testing.T) {
	m := initMocks(t)
	defer m.ctrl.Finish()

	m.credentialsStore.EXPECT().Get("user").Return(testCredentials, nil).Times(2)
	m.credentialsStore.EXPECT().Logout("user").Return(nil).AnyTimes()
	m.pmapiClient.EXPECT().AuthRefresh("token").Return(nil, pmapi.ErrUpgradeApplication).AnyTimes()
	m.eventListener.EXPECT().Emit(events.UpgradeApplicationEvent, "").AnyTimes()
	m.pmapiClient.EXPECT().ListLabels().Return(nil, pmapi.ErrUpgradeApplication)
	m.pmapiClient.EXPECT().Addresses().Return(nil)

	checkNewUser(m)
}

func TestNewUserNoInternetConnection(t *testing.T) {
	m := initMocks(t)
	defer m.ctrl.Finish()

	m.credentialsStore.EXPECT().Get("user").Return(testCredentials, nil).Times(2)
	m.pmapiClient.EXPECT().AuthRefresh("token").Return(nil, pmapi.ErrAPINotReachable).AnyTimes()
	m.eventListener.EXPECT().Emit(events.InternetOffEvent, "").AnyTimes()

	m.pmapiClient.EXPECT().Addresses().Return(nil)
	m.pmapiClient.EXPECT().ListLabels().Return(nil, pmapi.ErrAPINotReachable)
	m.pmapiClient.EXPECT().GetEvent("").Return(nil, pmapi.ErrAPINotReachable).AnyTimes()

	checkNewUser(m)
}

func TestNewUserAuthRefreshFails(t *testing.T) {
	m := initMocks(t)
	defer m.ctrl.Finish()

	m.credentialsStore.EXPECT().Get("user").Return(testCredentials, nil).Times(2)
	m.credentialsStore.EXPECT().Logout("user").Return(nil)
	m.pmapiClient.EXPECT().AuthRefresh("token").Return(nil, errors.New("bad token")).AnyTimes()

	m.eventListener.EXPECT().Emit(events.LogoutEvent, "user")
	m.eventListener.EXPECT().Emit(events.UserRefreshEvent, "user")
	m.pmapiClient.EXPECT().Logout().Return(nil)
	m.credentialsStore.EXPECT().Logout("user").Return(nil)
	m.credentialsStore.EXPECT().Get("user").Return(testCredentialsDisconnected, nil)
	m.eventListener.EXPECT().Emit(events.CloseConnectionEvent, "user@pm.me")

	checkNewUserDisconnected(m)
}

func TestNewUserUnlockFails(t *testing.T) {
	m := initMocks(t)
	defer m.ctrl.Finish()

	m.credentialsStore.EXPECT().Get("user").Return(testCredentials, nil).Times(2)
	m.credentialsStore.EXPECT().UpdateToken("user", ":reftok").Return(nil)
	m.credentialsStore.EXPECT().Logout("user").Return(nil)

	m.pmapiClient.EXPECT().AuthRefresh("token").Return(testAuthRefresh, nil)
	m.credentialsStore.EXPECT().Get("user").Return(testCredentials, nil)
	m.pmapiClient.EXPECT().Unlock("pass").Return(nil, errors.New("bad password"))

	m.eventListener.EXPECT().Emit(events.LogoutEvent, "user")
	m.eventListener.EXPECT().Emit(events.UserRefreshEvent, "user")
	m.pmapiClient.EXPECT().Logout().Return(nil)
	m.credentialsStore.EXPECT().Logout("user").Return(nil)
	m.credentialsStore.EXPECT().Get("user").Return(testCredentialsDisconnected, nil)
	m.eventListener.EXPECT().Emit(events.CloseConnectionEvent, "user@pm.me")

	checkNewUserDisconnected(m)
}

func TestNewUserUnlockAddressesFails(t *testing.T) {
	m := initMocks(t)
	defer m.ctrl.Finish()

	m.credentialsStore.EXPECT().Get("user").Return(testCredentials, nil).Times(2)
	m.credentialsStore.EXPECT().UpdateToken("user", ":reftok").Return(nil)
	m.credentialsStore.EXPECT().Logout("user").Return(nil)

	m.pmapiClient.EXPECT().AuthRefresh("token").Return(testAuthRefresh, nil)
	m.credentialsStore.EXPECT().Get("user").Return(testCredentials, nil)
	m.pmapiClient.EXPECT().Unlock("pass").Return(nil, nil)
	m.pmapiClient.EXPECT().UnlockAddresses([]byte("pass")).Return(errors.New("bad password"))

	m.eventListener.EXPECT().Emit(events.LogoutEvent, "user")
	m.eventListener.EXPECT().Emit(events.UserRefreshEvent, "user")
	m.pmapiClient.EXPECT().Logout().Return(nil)
	m.credentialsStore.EXPECT().Logout("user").Return(nil)
	m.credentialsStore.EXPECT().Get("user").Return(testCredentialsDisconnected, nil)
	m.eventListener.EXPECT().Emit(events.CloseConnectionEvent, "user@pm.me")

	checkNewUserDisconnected(m)
}

func TestNewUser(t *testing.T) {
	m := initMocks(t)
	defer m.ctrl.Finish()

	m.credentialsStore.EXPECT().Get("user").Return(testCredentials, nil).Times(2)
	m.credentialsStore.EXPECT().UpdateToken("user", ":reftok").Return(nil)

	m.pmapiClient.EXPECT().AuthRefresh("token").Return(testAuthRefresh, nil)
	m.credentialsStore.EXPECT().Get("user").Return(testCredentials, nil)
	m.pmapiClient.EXPECT().Unlock("pass").Return(nil, nil)
	m.pmapiClient.EXPECT().UnlockAddresses([]byte("pass")).Return(nil)

	m.pmapiClient.EXPECT().Addresses().Return([]*pmapi.Address{testPMAPIAddress})
	m.pmapiClient.EXPECT().ListLabels().Return([]*pmapi.Label{}, nil)
	m.pmapiClient.EXPECT().CountMessages("").Return([]*pmapi.MessagesCount{}, nil)

	m.pmapiClient.EXPECT().GetEvent("").Return(testPMAPIEvent, nil)
	m.pmapiClient.EXPECT().ListMessages(gomock.Any()).Return([]*pmapi.Message{}, 0, nil)
	m.pmapiClient.EXPECT().GetEvent(testPMAPIEvent.EventID).Return(testPMAPIEvent, nil)

	checkNewUser(m)
}

func checkNewUser(m mocks) {
	user, _ := newUser(m.PanicHandler, "user", m.eventListener, m.credentialsStore, m.clientManager, m.storeCache, "/tmp")
	defer cleanUpUserData(user)

	_ = user.init(nil)

	waitForEvents()

	a.Equal(m.t, testCredentials, user.creds)
}

func checkNewUserDisconnected(m mocks) {
	user, _ := newUser(m.PanicHandler, "user", m.eventListener, m.credentialsStore, m.clientManager, m.storeCache, "/tmp")
	defer cleanUpUserData(user)

	_ = user.init(nil)

	waitForEvents()

	a.Equal(m.t, testCredentialsDisconnected, user.creds)
}

func _TestUserEventRefreshUpdatesAddresses(t *testing.T) { // nolint[funlen]
	a.Fail(t, "not implemented")
}
