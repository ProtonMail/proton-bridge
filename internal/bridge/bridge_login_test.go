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

func TestBridgeFinishLoginBadMailboxPassword(t *testing.T) {
	m := initMocks(t)
	defer m.ctrl.Finish()

	err := errors.New("bad password")
	gomock.InOrder(
		// Init bridge with no user from keychain.
		m.credentialsStore.EXPECT().List().Return([]string{}, nil),

		// Set up mocks for FinishLogin.
		m.pmapiClient.EXPECT().Unlock(testCredentials.MailboxPassword).Return(nil, err),
		m.pmapiClient.EXPECT().DeleteAuth(),
		m.pmapiClient.EXPECT().Logout(),
	)

	checkBridgeFinishLogin(t, m, testAuth, testCredentials.MailboxPassword, "", err)
}

func TestBridgeFinishLoginUpgradeApplication(t *testing.T) {
	m := initMocks(t)
	defer m.ctrl.Finish()

	err := errors.New("Cannot logout when upgrade needed")
	gomock.InOrder(
		// Init bridge with no user from keychain.
		m.credentialsStore.EXPECT().List().Return([]string{}, nil),

		// Set up mocks for FinishLogin.
		m.pmapiClient.EXPECT().Unlock(testCredentials.MailboxPassword).Return(nil, pmapi.ErrUpgradeApplication),

		m.eventListener.EXPECT().Emit(events.UpgradeApplicationEvent, ""),
		m.pmapiClient.EXPECT().DeleteAuth().Return(err),
		m.pmapiClient.EXPECT().Logout(),
	)

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

	// Basically every call client has get client manager
	m.clientManager.EXPECT().GetClient("user").Return(m.pmapiClient).MinTimes(1)

	gomock.InOrder(
		// Bridge finds no users in the keychain.
		m.credentialsStore.EXPECT().List().Return([]string{}, nil),

		// Get user to be able to setup new client with proper userID.
		m.pmapiClient.EXPECT().Unlock(testCredentials.MailboxPassword).Return(nil, nil),
		m.pmapiClient.EXPECT().CurrentUser().Return(testPMAPIUser, nil),

		// bridge.Bridge.addNewUser(()
		m.pmapiClient.EXPECT().AuthRefresh(":tok").Return(refreshWithToken("afterLogin"), nil),
		m.pmapiClient.EXPECT().CurrentUser().Return(testPMAPIUser, nil),
		m.pmapiClient.EXPECT().Addresses().Return([]*pmapi.Address{testPMAPIAddress}),

		// bridge.newUser()
		m.credentialsStore.EXPECT().Add("user", "username", ":afterLogin", testCredentials.MailboxPassword, []string{testPMAPIAddress.Email}),
		m.credentialsStore.EXPECT().Get("user").Return(credentialsWithToken(":afterLogin"), nil).Times(2),

		// bridge.User.init()
		m.pmapiClient.EXPECT().AuthRefresh(":afterLogin").Return(refreshWithToken("afterCredentials"), nil),
		//TODO m.credentialsStore.EXPECT().UpdateToken("user", ":afterCredentials").Return(nil),
		//TODO m.credentialsStore.EXPECT().Get("user").Return(credentialsWithToken("afterCredentials"), nil),

		// authorize if necessary
		m.pmapiClient.EXPECT().Unlock(testCredentials.MailboxPassword).Return(nil, nil),
		m.pmapiClient.EXPECT().UnlockAddresses([]byte(testCredentials.MailboxPassword)).Return(nil),

		// Set up mocks for creating the user's store (in store.New).
		m.pmapiClient.EXPECT().ListLabels().Return([]*pmapi.Label{}, nil),
		m.pmapiClient.EXPECT().CountMessages("").Return([]*pmapi.MessagesCount{}, nil),
		m.pmapiClient.EXPECT().Addresses().Return([]*pmapi.Address{testPMAPIAddress}),

		// Emit event for new user and send metrics.
		m.clientManager.EXPECT().GetAnonymousClient().Return(m.pmapiClient),
		m.pmapiClient.EXPECT().SendSimpleMetric(string(metrics.Setup), string(metrics.NewUser), string(metrics.NoLabel)),
		m.pmapiClient.EXPECT().Logout(),

		// Reload account list in GUI.
		m.eventListener.EXPECT().Emit(events.UserRefreshEvent, "user"),
		// defer logout anonymous
		m.pmapiClient.EXPECT().Logout(),
	)

	mockEventLoopNoAction(m)

	checkBridgeFinishLogin(t, m, testAuth, testCredentials.MailboxPassword, "user", nil)
}

func TestBridgeFinishLoginExistingDisconnectedUser(t *testing.T) {
	m := initMocks(t)
	defer m.ctrl.Finish()

	loggedOutCreds := *testCredentials
	loggedOutCreds.APIToken = ""
	loggedOutCreds.MailboxPassword = ""

	m.clientManager.EXPECT().GetClient("user").Return(m.pmapiClient).MinTimes(1)

	gomock.InOrder(
		// Bridge finds one logged out user in the keychain.
		m.credentialsStore.EXPECT().List().Return([]string{"user"}, nil),
		// New user
		m.credentialsStore.EXPECT().Get("user").Return(&loggedOutCreds, nil),
		// Init user
		m.credentialsStore.EXPECT().Get("user").Return(&loggedOutCreds, nil),
		m.pmapiClient.EXPECT().ListLabels().Return(nil, pmapi.ErrInvalidToken),
		m.pmapiClient.EXPECT().Addresses().Return(nil),

		// Get user to be able to setup new client with proper userID.
		m.pmapiClient.EXPECT().Unlock(testCredentials.MailboxPassword).Return(nil, nil),
		m.pmapiClient.EXPECT().CurrentUser().Return(testPMAPIUser, nil),

		// bridge.Bridge.connectExistingUser
		m.credentialsStore.EXPECT().UpdatePassword("user", testCredentials.MailboxPassword).Return(nil),
		m.pmapiClient.EXPECT().AuthRefresh(":tok").Return(refreshWithToken("afterLogin"), nil),
		m.credentialsStore.EXPECT().UpdateToken("user", ":afterLogin").Return(nil),

		// bridge.User.init()
		m.credentialsStore.EXPECT().Get("user").Return(credentialsWithToken(":afterLogin"), nil),
		m.pmapiClient.EXPECT().AuthRefresh(":afterLogin").Return(refreshWithToken("afterCredentials"), nil),

		// authorize if necessary
		m.pmapiClient.EXPECT().Unlock(testCredentials.MailboxPassword).Return(nil, nil),
		m.pmapiClient.EXPECT().UnlockAddresses([]byte(testCredentials.MailboxPassword)).Return(nil),

		/* TODO
		// Set up mocks for authorising the new user (in user.init).
		m.credentialsStore.EXPECT().Add("user", "username", ":afterLogin", testCredentials.MailboxPassword, []string{testPMAPIAddress.Email}),
		m.credentialsStore.EXPECT().Get("user").Return(credentialsWithToken("afterCredentials"), nil),

		m.credentialsStore.EXPECT().UpdateToken("user", ":afterCredentials").Return(nil),
		*/

		// Set up mocks for creating the user's store (in store.New).
		m.pmapiClient.EXPECT().ListLabels().Return([]*pmapi.Label{}, nil),
		m.pmapiClient.EXPECT().CountMessages("").Return([]*pmapi.MessagesCount{}, nil),
		m.pmapiClient.EXPECT().Addresses().Return([]*pmapi.Address{testPMAPIAddress}),

		// Reload account list in GUI.
		m.eventListener.EXPECT().Emit(events.UserRefreshEvent, "user"),
		// defer logout anonymous
		m.pmapiClient.EXPECT().Logout(),
	)

	mockEventLoopNoAction(m)

	checkBridgeFinishLogin(t, m, testAuth, testCredentials.MailboxPassword, "user", nil)
}

func TestBridgeFinishLoginConnectedUser(t *testing.T) {
	m := initMocks(t)
	defer m.ctrl.Finish()

	m.clientManager.EXPECT().GetClient("user").Return(m.pmapiClient).MinTimes(1)
	m.credentialsStore.EXPECT().List().Return([]string{"user"}, nil)

	mockConnectedUser(m)
	mockEventLoopNoAction(m)

	bridge := testNewBridge(t, m)
	defer cleanUpBridgeUserData(bridge)

	// Then, try to log in again...
	gomock.InOrder(
		m.pmapiClient.EXPECT().Unlock(testCredentials.MailboxPassword).Return(nil, nil),
		m.pmapiClient.EXPECT().CurrentUser().Return(testPMAPIUser, nil),
		m.pmapiClient.EXPECT().DeleteAuth(),
		m.pmapiClient.EXPECT().Logout(),
	)

	_, err := bridge.FinishLogin(m.pmapiClient, testAuth, testCredentials.MailboxPassword)
	assert.Equal(t, "user is already connected", err.Error())
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
