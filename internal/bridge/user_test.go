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
	"testing"

	"github.com/ProtonMail/proton-bridge/pkg/pmapi"
	gomock "github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func testNewUser(m mocks) *User {
	m.credentialsStore.EXPECT().Get("user").Return(testCredentials, nil).Times(2)
	m.credentialsStore.EXPECT().UpdateToken("user", ":reftok").Return(nil)

	m.pmapiClient.EXPECT().AuthRefresh("token").Return(testAuthRefresh, nil)
	m.credentialsStore.EXPECT().Get("user").Return(testCredentials, nil)
	m.pmapiClient.EXPECT().Unlock("pass").Return(nil, nil)
	m.pmapiClient.EXPECT().UnlockAddresses([]byte("pass")).Return(nil)

	// Expectations for initial sync (when loading existing user from credentials store).
	m.pmapiClient.EXPECT().ListLabels().Return([]*pmapi.Label{}, nil)
	m.pmapiClient.EXPECT().Addresses().Return([]*pmapi.Address{testPMAPIAddress})
	m.pmapiClient.EXPECT().CountMessages("").Return([]*pmapi.MessagesCount{}, nil)
	m.pmapiClient.EXPECT().GetEvent("").Return(testPMAPIEvent, nil).AnyTimes()
	m.pmapiClient.EXPECT().ListMessages(gomock.Any()).Return([]*pmapi.Message{}, 0, nil)
	m.pmapiClient.EXPECT().GetEvent(testPMAPIEvent.EventID).Return(testPMAPIEvent, nil).AnyTimes()

	user, err := newUser(m.PanicHandler, "user", m.eventListener, m.credentialsStore, m.clientManager, m.storeCache, "/tmp")
	assert.NoError(m.t, err)

	err = user.init(nil)
	assert.NoError(m.t, err)

	return user
}

func testNewUserForLogout(m mocks) *User {
	m.credentialsStore.EXPECT().Get("user").Return(testCredentials, nil).Times(2)
	m.credentialsStore.EXPECT().UpdateToken("user", ":reftok").Return(nil)

	m.pmapiClient.EXPECT().AuthRefresh("token").Return(testAuthRefresh, nil)
	m.credentialsStore.EXPECT().Get("user").Return(testCredentials, nil)
	m.pmapiClient.EXPECT().Unlock("pass").Return(nil, nil)
	m.pmapiClient.EXPECT().UnlockAddresses([]byte("pass")).Return(nil)

	// These may or may not be hit depending on how fast the log out happens.
	m.pmapiClient.EXPECT().ListLabels().Return([]*pmapi.Label{}, nil).AnyTimes()
	m.pmapiClient.EXPECT().Addresses().Return([]*pmapi.Address{testPMAPIAddress}).AnyTimes()
	m.pmapiClient.EXPECT().CountMessages("").Return([]*pmapi.MessagesCount{}, nil)
	m.pmapiClient.EXPECT().GetEvent("").Return(testPMAPIEvent, nil).AnyTimes()
	m.pmapiClient.EXPECT().ListMessages(gomock.Any()).Return([]*pmapi.Message{}, 0, nil).AnyTimes()
	m.pmapiClient.EXPECT().GetEvent(testPMAPIEvent.EventID).Return(testPMAPIEvent, nil).AnyTimes()

	user, err := newUser(m.PanicHandler, "user", m.eventListener, m.credentialsStore, m.clientManager, m.storeCache, "/tmp")
	assert.NoError(m.t, err)

	err = user.init(nil)
	assert.NoError(m.t, err)

	return user
}

func cleanUpUserData(u *User) {
	_ = u.clearStore()
}

func _TestNeverLongStorePath(t *testing.T) { // nolint[unused]
	assert.Fail(t, "not implemented")
}

func TestClearStoreWithStore(t *testing.T) {
	m := initMocks(t)
	defer m.ctrl.Finish()

	user := testNewUserForLogout(m)
	defer cleanUpUserData(user)

	require.Nil(t, user.store.Close())
	user.store = nil
	assert.Nil(t, user.clearStore())
}

func TestClearStoreWithoutStore(t *testing.T) {
	m := initMocks(t)
	defer m.ctrl.Finish()

	user := testNewUserForLogout(m)
	defer cleanUpUserData(user)

	assert.NotNil(t, user.store)
	assert.Nil(t, user.clearStore())
}
