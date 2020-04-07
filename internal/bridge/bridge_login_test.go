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

	"github.com/ProtonMail/proton-bridge/internal/bridge/credentials"
	"github.com/ProtonMail/proton-bridge/internal/events"
	"github.com/ProtonMail/proton-bridge/internal/metrics"
	"github.com/ProtonMail/proton-bridge/pkg/pmapi"
	gomock "github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
)

func TestBridgeFinishLoginBadPassword(t *testing.T) {
	m := initMocks(t)
	defer m.ctrl.Finish()

	// Init bridge with no user from keychain.
	m.credentialsStore.EXPECT().List().Return([]string{}, nil)

	// Set up mocks for FinishLogin.
	err := errors.New("bad password")
	m.pmapiClient.EXPECT().Unlock(testCredentials.MailboxPassword).Return(nil, err)
	m.pmapiClient.EXPECT().Logout()

	checkBridgeFinishLogin(t, m, testAuth, testCredentials.MailboxPassword, "", err)
}

func TestBridgeFinishLoginUpgradeApplication(t *testing.T) {
	m := initMocks(t)
	defer m.ctrl.Finish()

	// Init bridge with no user from keychain.
	m.credentialsStore.EXPECT().List().Return([]string{}, nil)

	// Set up mocks for FinishLogin.
	m.pmapiClient.EXPECT().Unlock(testCredentials.MailboxPassword).Return(nil, pmapi.ErrUpgradeApplication)

	m.eventListener.EXPECT().Emit(events.UpgradeApplicationEvent, "")
	err := errors.New("Cannot logout when upgrade needed")
	m.pmapiClient.EXPECT().Logout().Return(err)

	checkBridgeFinishLogin(t, m, testAuth, testCredentials.MailboxPassword, "", pmapi.ErrUpgradeApplication)
}

func refreshWithToken(token string) *pmapi.Auth {
	return &pmapi.Auth{
		RefreshToken: token,
		KeySalt:      "", // No salting in tests.
	}
}

func credentialsWithToken(token string) *credentials.Credentials {
	tmp := &credentials.Credentials{}
	*tmp = *testCredentials
	tmp.APIToken = token
	return tmp
}

func TestBridgeFinishLoginNewUser(t *testing.T) {
	m := initMocks(t)
	defer m.ctrl.Finish()

	// Bridge finds no users in the keychain.
	m.credentialsStore.EXPECT().List().Return([]string{}, nil)

	// Get user to be able to setup new client with proper userID.
	m.pmapiClient.EXPECT().Unlock(testCredentials.MailboxPassword).Return(nil, nil)
	m.pmapiClient.EXPECT().CurrentUser().Return(testPMAPIUser, nil)

	// Setup of new client.
	m.pmapiClient.EXPECT().AuthRefresh(":tok").Return(refreshWithToken("afterLogin"), nil)
	m.pmapiClient.EXPECT().CurrentUser().Return(testPMAPIUser, nil)
	m.pmapiClient.EXPECT().Addresses().Return([]*pmapi.Address{testPMAPIAddress})

	// Set up mocks for authorising the new user (in user.init).
	m.credentialsStore.EXPECT().Add("user", "username", ":afterLogin", testCredentials.MailboxPassword, []string{testPMAPIAddress.Email})
	m.credentialsStore.EXPECT().Get("user").Return(credentialsWithToken(":afterLogin"), nil).Times(2)
	m.pmapiClient.EXPECT().AuthRefresh(":afterLogin").Return(refreshWithToken("afterCredentials"), nil)
	m.credentialsStore.EXPECT().Get("user").Return(credentialsWithToken("afterCredentials"), nil)
	m.pmapiClient.EXPECT().Unlock(testCredentials.MailboxPassword).Return(nil, nil)
	m.pmapiClient.EXPECT().UnlockAddresses([]byte(testCredentials.MailboxPassword)).Return(nil)

	m.credentialsStore.EXPECT().UpdateToken("user", ":afterCredentials").Return(nil)

	// Set up mocks for creating the user's store (in store.New).
	m.pmapiClient.EXPECT().ListLabels().Return([]*pmapi.Label{}, nil)
	m.pmapiClient.EXPECT().Addresses().Return([]*pmapi.Address{testPMAPIAddress})
	m.pmapiClient.EXPECT().CountMessages("").Return([]*pmapi.MessagesCount{}, nil)

	// Emit event for new user and send metrics.
	m.eventListener.EXPECT().Emit(events.UserRefreshEvent, "user")
	m.pmapiClient.EXPECT().SendSimpleMetric(string(metrics.Setup), string(metrics.NewUser), string(metrics.NoLabel))

	// Set up mocks for starting the store's event loop (in store.New).
	// The event loop runs in another goroutine so this might happen at any time.
	m.pmapiClient.EXPECT().GetEvent("").Return(testPMAPIEvent, nil)
	m.pmapiClient.EXPECT().GetEvent(testPMAPIEvent.EventID).Return(testPMAPIEvent, nil)

	// Set up mocks for performing the initial store sync.
	m.pmapiClient.EXPECT().ListMessages(gomock.Any()).Return([]*pmapi.Message{}, 0, nil)

	checkBridgeFinishLogin(t, m, testAuth, testCredentials.MailboxPassword, "user", nil)
}

func TestBridgeFinishLoginExistingUser(t *testing.T) {
	m := initMocks(t)
	defer m.ctrl.Finish()

	loggedOutCreds := *testCredentials
	loggedOutCreds.APIToken = ""
	loggedOutCreds.MailboxPassword = ""

	// Bridge finds one logged out user in the keychain.
	m.credentialsStore.EXPECT().List().Return([]string{"user"}, nil)
	// New user
	m.credentialsStore.EXPECT().Get("user").Return(&loggedOutCreds, nil)
	// Init user
	m.credentialsStore.EXPECT().Get("user").Return(&loggedOutCreds, nil)
	m.pmapiClient.EXPECT().ListLabels().Return(nil, pmapi.ErrInvalidToken)
	m.pmapiClient.EXPECT().Addresses().Return(nil)

	// Get user to be able to setup new client with proper userID.
	m.pmapiClient.EXPECT().Unlock(testCredentials.MailboxPassword).Return(nil, nil)
	m.pmapiClient.EXPECT().CurrentUser().Return(testPMAPIUser, nil)

	// Setup of new client.
	m.pmapiClient.EXPECT().AuthRefresh(":tok").Return(refreshWithToken("afterLogin"), nil)
	m.pmapiClient.EXPECT().CurrentUser().Return(testPMAPIUser, nil)
	m.pmapiClient.EXPECT().Addresses().Return([]*pmapi.Address{testPMAPIAddress})

	// Set up mocks for authorising the new user (in user.init).
	m.credentialsStore.EXPECT().Add("user", "username", ":afterLogin", testCredentials.MailboxPassword, []string{testPMAPIAddress.Email})
	m.credentialsStore.EXPECT().Get("user").Return(credentialsWithToken(":afterLogin"), nil)
	m.pmapiClient.EXPECT().AuthRefresh(":afterLogin").Return(refreshWithToken("afterCredentials"), nil)
	m.credentialsStore.EXPECT().Get("user").Return(credentialsWithToken("afterCredentials"), nil)
	m.pmapiClient.EXPECT().Unlock(testCredentials.MailboxPassword).Return(nil, nil)
	m.pmapiClient.EXPECT().UnlockAddresses([]byte(testCredentials.MailboxPassword)).Return(nil)

	m.credentialsStore.EXPECT().UpdateToken("user", ":afterCredentials").Return(nil)

	// Set up mocks for creating the user's store (in store.New).
	m.pmapiClient.EXPECT().ListLabels().Return([]*pmapi.Label{}, nil)
	m.pmapiClient.EXPECT().Addresses().Return([]*pmapi.Address{testPMAPIAddress})
	m.pmapiClient.EXPECT().CountMessages("").Return([]*pmapi.MessagesCount{}, nil)

	// Reload account list in GUI.
	m.eventListener.EXPECT().Emit(events.UserRefreshEvent, "user")

	// Set up mocks for starting the store's event loop (in store.New)
	// The event loop runs in another goroutine so this might happen at any time.
	m.pmapiClient.EXPECT().GetEvent("").Return(testPMAPIEvent, nil)
	m.pmapiClient.EXPECT().GetEvent(testPMAPIEvent.EventID).Return(testPMAPIEvent, nil)

	// Set up mocks for performing the initial store sync.
	m.pmapiClient.EXPECT().ListMessages(gomock.Any()).Return([]*pmapi.Message{}, 0, nil)

	checkBridgeFinishLogin(t, m, testAuth, testCredentials.MailboxPassword, "user", nil)
}

func TestBridgeDoubleLogin(t *testing.T) {
	m := initMocks(t)
	defer m.ctrl.Finish()

	// Firstly, start bridge with existing user...

	m.credentialsStore.EXPECT().List().Return([]string{"user"}, nil)
	m.credentialsStore.EXPECT().Get("user").Return(testCredentials, nil)

	m.credentialsStore.EXPECT().Get("user").Return(testCredentials, nil)
	m.pmapiClient.EXPECT().AuthRefresh("token").Return(testAuthRefresh, nil)
	m.credentialsStore.EXPECT().Get("user").Return(testCredentials, nil)
	m.credentialsStore.EXPECT().UpdateToken("user", ":reftok").Return(nil)
	m.pmapiClient.EXPECT().Unlock(testCredentials.MailboxPassword).Return(nil, nil)
	m.pmapiClient.EXPECT().UnlockAddresses([]byte(testCredentials.MailboxPassword)).Return(nil)
	m.pmapiClient.EXPECT().ListLabels().Return([]*pmapi.Label{}, nil)
	m.pmapiClient.EXPECT().CountMessages("").Return([]*pmapi.MessagesCount{}, nil)
	m.pmapiClient.EXPECT().Addresses().Return([]*pmapi.Address{testPMAPIAddress})

	m.pmapiClient.EXPECT().GetEvent("").Return(testPMAPIEvent, nil)
	m.pmapiClient.EXPECT().ListMessages(gomock.Any()).Return([]*pmapi.Message{}, 0, nil)
	m.pmapiClient.EXPECT().GetEvent(testPMAPIEvent.EventID).Return(testPMAPIEvent, nil)

	bridge := testNewBridge(t, m)
	defer cleanUpBridgeUserData(bridge)

	// Then, try to log in again...

	m.pmapiClient.EXPECT().Unlock(testCredentials.MailboxPassword).Return(nil, nil)
	m.pmapiClient.EXPECT().CurrentUser().Return(testPMAPIUser, nil)
	m.pmapiClient.EXPECT().Logout()

	_, err := bridge.FinishLogin(m.pmapiClient, testAuth, testCredentials.MailboxPassword)
	assert.Equal(t, "user is already logged in", err.Error())
}

func checkBridgeFinishLogin(t *testing.T, m mocks, auth *pmapi.Auth, mailboxPassword string, expectedUserID string, expectedErr error) {
	bridge := testNewBridge(t, m)
	defer cleanUpBridgeUserData(bridge)

	user, err := bridge.FinishLogin(m.pmapiClient, auth, mailboxPassword)

	waitForEvents()

	assert.Equal(t, expectedErr, err)

	if expectedUserID != "" {
		assert.Equal(t, expectedUserID, user.ID())
		assert.Equal(t, 1, len(bridge.users))
		assert.Equal(t, expectedUserID, bridge.users[0].ID())
	} else {
		assert.Equal(t, (*User)(nil), user)
		assert.Equal(t, 0, len(bridge.users))
	}
}
