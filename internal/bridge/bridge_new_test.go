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

	credentials "github.com/ProtonMail/proton-bridge/internal/bridge/credentials"
	"github.com/ProtonMail/proton-bridge/internal/events"
	"github.com/ProtonMail/proton-bridge/internal/metrics"
	"github.com/ProtonMail/proton-bridge/internal/preferences"
	"github.com/ProtonMail/proton-bridge/pkg/pmapi"
	gomock "github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
)

func TestNewBridgeNoKeychain(t *testing.T) {
	m := initMocks(t)
	defer m.ctrl.Finish()

	m.credentialsStore.EXPECT().List().Return([]string{}, errors.New("no keychain"))

	checkBridgeNew(t, m, []*credentials.Credentials{})
}

func TestNewBridgeWithoutUsersInCredentialsStore(t *testing.T) {
	m := initMocks(t)
	defer m.ctrl.Finish()

	m.credentialsStore.EXPECT().List().Return([]string{}, nil)

	checkBridgeNew(t, m, []*credentials.Credentials{})
}

func TestNewBridgeWithDisconnectedUser(t *testing.T) {
	m := initMocks(t)
	defer m.ctrl.Finish()

	m.credentialsStore.EXPECT().List().Return([]string{"user"}, nil)
	m.credentialsStore.EXPECT().Get("user").Return(testCredentialsDisconnected, nil).Times(2)
	m.pmapiClient.EXPECT().ListLabels().Return(nil, errors.New("ErrUnauthorized"))
	m.pmapiClient.EXPECT().Addresses().Return(nil)
	m.clientManager.EXPECT().GetClient("user").Return(m.pmapiClient)

	checkBridgeNew(t, m, []*credentials.Credentials{testCredentialsDisconnected})
}

func TestNewBridgeWithConnectedUserWithBadToken(t *testing.T) {
	m := initMocks(t)
	defer m.ctrl.Finish()

	m.credentialsStore.EXPECT().List().Return([]string{"user"}, nil)
	m.credentialsStore.EXPECT().Get("user").Return(testCredentials, nil).Times(2)
	m.clientManager.EXPECT().GetClient("user").Return(m.pmapiClient).MinTimes(1)

	m.credentialsStore.EXPECT().Logout("user").Return(nil)
	m.pmapiClient.EXPECT().AuthRefresh("token").Return(nil, errors.New("bad token"))

	m.eventListener.EXPECT().Emit(events.LogoutEvent, "user")
	m.eventListener.EXPECT().Emit(events.UserRefreshEvent, "user")
	m.pmapiClient.EXPECT().Logout()
	m.credentialsStore.EXPECT().Logout("user").Return(nil)
	m.credentialsStore.EXPECT().Get("user").Return(testCredentialsDisconnected, nil)
	m.eventListener.EXPECT().Emit(events.CloseConnectionEvent, "user@pm.me")

	checkBridgeNew(t, m, []*credentials.Credentials{testCredentialsDisconnected})
}

func TestNewBridgeWithConnectedUser(t *testing.T) {
	m := initMocks(t)
	defer m.ctrl.Finish()

	m.credentialsStore.EXPECT().List().Return([]string{"user"}, nil)
	m.credentialsStore.EXPECT().Get("user").Return(testCredentials, nil).Times(2)
	m.credentialsStore.EXPECT().UpdateToken("user", ":reftok").Return(nil)

	m.pmapiClient.EXPECT().AuthRefresh("token").Return(testAuthRefresh, nil)
	m.pmapiClient.EXPECT().Unlock(testCredentials.MailboxPassword).Return(nil, nil)
	m.pmapiClient.EXPECT().UnlockAddresses([]byte(testCredentials.MailboxPassword)).Return(nil)
	m.clientManager.EXPECT().GetClient("user").Return(m.pmapiClient).MinTimes(1)

	// Set up mocks for store initialisation for the authorized user.
	m.pmapiClient.EXPECT().ListLabels().Return([]*pmapi.Label{}, nil)
	m.pmapiClient.EXPECT().Addresses().Return([]*pmapi.Address{testPMAPIAddress})
	m.pmapiClient.EXPECT().CountMessages("").Return([]*pmapi.MessagesCount{}, nil)
	m.pmapiClient.EXPECT().GetEvent("").Return(testPMAPIEvent, nil)
	m.pmapiClient.EXPECT().ListMessages(gomock.Any()).Return([]*pmapi.Message{}, 0, nil).AnyTimes()
	m.pmapiClient.EXPECT().GetEvent(testPMAPIEvent.EventID).Return(testPMAPIEvent, nil)

	checkBridgeNew(t, m, []*credentials.Credentials{testCredentials})
}

// Tests two users with different states and checks also the order from
// credentials store is kept also in array of Bridge users.
func TestNewBridgeWithUsers(t *testing.T) {
	m := initMocks(t)
	defer m.ctrl.Finish()

	m.pmapiClient.EXPECT().AuthRefresh("token").Return(testAuthRefresh, nil)
	m.pmapiClient.EXPECT().Unlock(testCredentials.MailboxPassword).Return(nil, nil)
	m.pmapiClient.EXPECT().UnlockAddresses([]byte(testCredentials.MailboxPassword)).Return(nil)

	m.credentialsStore.EXPECT().List().Return([]string{"user", "user"}, nil)
	m.credentialsStore.EXPECT().Get("user").Return(testCredentialsDisconnected, nil).Times(2)
	m.credentialsStore.EXPECT().Get("user").Return(testCredentials, nil).Times(2)
	m.credentialsStore.EXPECT().UpdateToken("user", ":reftok").Return(nil)

	// Set up mocks for store initialisation for the unauth user.
	m.pmapiClient.EXPECT().ListLabels().Return(nil, errors.New("ErrUnauthorized"))
	m.pmapiClient.EXPECT().Addresses().Return(nil)

	// Set up mocks for store initialisation for the authorized user.
	m.credentialsStore.EXPECT().Get("user").Return(testCredentials, nil)
	m.pmapiClient.EXPECT().ListLabels().Return([]*pmapi.Label{}, nil)
	m.pmapiClient.EXPECT().Addresses().Return([]*pmapi.Address{testPMAPIAddress})
	m.pmapiClient.EXPECT().CountMessages("").Return([]*pmapi.MessagesCount{}, nil)
	m.pmapiClient.EXPECT().GetEvent("").Return(testPMAPIEvent, nil)
	m.pmapiClient.EXPECT().ListMessages(gomock.Any()).Return([]*pmapi.Message{}, 0, nil).AnyTimes()
	m.pmapiClient.EXPECT().GetEvent(testPMAPIEvent.EventID).Return(testPMAPIEvent, nil)

	checkBridgeNew(t, m, []*credentials.Credentials{testCredentialsDisconnected, testCredentials})
}

func TestNewBridgeFirstStart(t *testing.T) {
	m := initMocks(t)
	defer m.ctrl.Finish()

	m.prefProvider.EXPECT().GetBool(preferences.FirstStartKey).Return(true)
	m.credentialsStore.EXPECT().List().Return([]string{}, nil)
	m.pmapiClient.EXPECT().SendSimpleMetric(string(metrics.Setup), string(metrics.FirstStart), gomock.Any())

	testNewBridge(t, m)
}

func checkBridgeNew(t *testing.T, m mocks, expectedCredentials []*credentials.Credentials) {
	bridge := testNewBridge(t, m)
	defer cleanUpBridgeUserData(bridge)

	assert.Equal(m.t, len(expectedCredentials), len(bridge.GetUsers()))

	credentials := []*credentials.Credentials{}
	for _, user := range bridge.users {
		credentials = append(credentials, user.creds)
	}

	assert.Equal(m.t, expectedCredentials, credentials)
}
