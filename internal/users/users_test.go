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
	"fmt"
	"io/ioutil"
	"os"
	"runtime/debug"
	"testing"
	"time"

	"github.com/ProtonMail/proton-bridge/internal/events"
	"github.com/ProtonMail/proton-bridge/internal/store"
	"github.com/ProtonMail/proton-bridge/internal/users/credentials"
	usersmocks "github.com/ProtonMail/proton-bridge/internal/users/mocks"
	"github.com/ProtonMail/proton-bridge/pkg/pmapi"
	pmapimocks "github.com/ProtonMail/proton-bridge/pkg/pmapi/mocks"
	gomock "github.com/golang/mock/gomock"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/require"
)

func TestMain(m *testing.M) {
	if os.Getenv("VERBOSITY") == "fatal" {
		logrus.SetLevel(logrus.FatalLevel)
	}
	if os.Getenv("VERBOSITY") == "trace" {
		logrus.SetLevel(logrus.TraceLevel)
	}
	os.Exit(m.Run())
}

var (
	testAuth = &pmapi.Auth{ //nolint[gochecknoglobals]
		RefreshToken: "tok",
	}
	testAuthRefresh = &pmapi.Auth{ //nolint[gochecknoglobals]
		RefreshToken: "reftok",
	}

	testCredentials = &credentials.Credentials{ //nolint[gochecknoglobals]
		UserID:                "user",
		Name:                  "username",
		Emails:                "user@pm.me",
		APIToken:              "token",
		MailboxPassword:       "pass",
		BridgePassword:        "0123456789abcdef",
		Version:               "v1",
		Timestamp:             123456789,
		IsHidden:              false,
		IsCombinedAddressMode: true,
	}
	testCredentialsSplit = &credentials.Credentials{ //nolint[gochecknoglobals]
		UserID:                "users",
		Name:                  "usersname",
		Emails:                "users@pm.me;anotheruser@pm.me;alsouser@pm.me",
		APIToken:              "token",
		MailboxPassword:       "pass",
		BridgePassword:        "0123456789abcdef",
		Version:               "v1",
		Timestamp:             123456789,
		IsHidden:              false,
		IsCombinedAddressMode: false,
	}
	testCredentialsDisconnected = &credentials.Credentials{ //nolint[gochecknoglobals]
		UserID:                "user",
		Name:                  "username",
		Emails:                "user@pm.me",
		APIToken:              "",
		MailboxPassword:       "",
		BridgePassword:        "0123456789abcdef",
		Version:               "v1",
		Timestamp:             123456789,
		IsHidden:              false,
		IsCombinedAddressMode: true,
	}

	testPMAPIUser = &pmapi.User{ //nolint[gochecknoglobals]
		ID:   "user",
		Name: "username",
	}

	testPMAPIAddress = &pmapi.Address{ //nolint[gochecknoglobals]
		ID:      "testAddressID",
		Type:    pmapi.OriginalAddress,
		Email:   "user@pm.me",
		Receive: pmapi.CanReceive,
	}

	testPMAPIAddresses = []*pmapi.Address{ //nolint[gochecknoglobals]
		{ID: "usersAddress1ID", Email: "users@pm.me", Receive: pmapi.CanReceive, Type: pmapi.OriginalAddress},
		{ID: "usersAddress2ID", Email: "anotheruser@pm.me", Receive: pmapi.CanReceive, Type: pmapi.AliasAddress},
		{ID: "usersAddress3ID", Email: "alsouser@pm.me", Receive: pmapi.CanReceive, Type: pmapi.AliasAddress},
	}

	testPMAPIEvent = &pmapi.Event{ // nolint[gochecknoglobals]
		EventID: "ACXDmTaBub14w==",
	}
)

func waitForEvents() {
	// Wait for goroutine to add listener.
	// E.g. calling login to invoke firstsync event. Functions can end sooner than
	// goroutines call the listener mock. We need to wait a little bit before the end of
	// the test to capture all event calls. This allows us to detect whether there were
	// missing calls, or perhaps whether something was called too many times.
	time.Sleep(100 * time.Millisecond)
}

type mocks struct {
	t *testing.T

	ctrl             *gomock.Controller
	config           *usersmocks.MockConfiger
	PanicHandler     *usersmocks.MockPanicHandler
	clientManager    *usersmocks.MockClientManager
	credentialsStore *usersmocks.MockCredentialsStorer
	storeMaker       *usersmocks.MockStoreMaker
	eventListener    *MockListener

	pmapiClient *pmapimocks.MockClient

	storeCache *store.Cache
}

type fullStackReporter struct {
	T testing.TB
}

func (fr *fullStackReporter) Errorf(format string, args ...interface{}) {
	fmt.Printf("err: "+format+"\n", args...)
	fr.T.Fail()
}
func (fr *fullStackReporter) Fatalf(format string, args ...interface{}) {
	debug.PrintStack()
	fmt.Printf("fail: "+format+"\n", args...)
	fr.T.FailNow()
}

func initMocks(t *testing.T) mocks {
	var mockCtrl *gomock.Controller
	if os.Getenv("VERBOSITY") == "trace" {
		mockCtrl = gomock.NewController(&fullStackReporter{t})
	} else {
		mockCtrl = gomock.NewController(t)
	}

	cacheFile, err := ioutil.TempFile("", "bridge-store-cache-*.db")
	require.NoError(t, err, "could not get temporary file for store cache")

	m := mocks{
		t: t,

		ctrl:             mockCtrl,
		config:           usersmocks.NewMockConfiger(mockCtrl),
		PanicHandler:     usersmocks.NewMockPanicHandler(mockCtrl),
		clientManager:    usersmocks.NewMockClientManager(mockCtrl),
		credentialsStore: usersmocks.NewMockCredentialsStorer(mockCtrl),
		storeMaker:       usersmocks.NewMockStoreMaker(mockCtrl),
		eventListener:    NewMockListener(mockCtrl),

		pmapiClient: pmapimocks.NewMockClient(mockCtrl),

		storeCache: store.NewCache(cacheFile.Name()),
	}

	// Called during clean-up.
	m.PanicHandler.EXPECT().HandlePanic().AnyTimes()

	// Set up store factory.
	m.storeMaker.EXPECT().New(gomock.Any()).DoAndReturn(func(user store.BridgeUser) (*store.Store, error) {
		dbFile, err := ioutil.TempFile("", "bridge-store-db-*.db")
		require.NoError(t, err, "could not get temporary file for store db")
		return store.New(m.PanicHandler, user, m.clientManager, m.eventListener, dbFile.Name(), m.storeCache)
	}).AnyTimes()
	m.storeMaker.EXPECT().Remove(gomock.Any()).AnyTimes()

	return m
}

func testNewUsersWithUsers(t *testing.T, m mocks) *Users {
	// Events are asynchronous
	m.pmapiClient.EXPECT().GetEvent("").Return(testPMAPIEvent, nil).Times(2)
	m.pmapiClient.EXPECT().GetEvent(testPMAPIEvent.EventID).Return(testPMAPIEvent, nil).Times(2)
	m.pmapiClient.EXPECT().ListMessages(gomock.Any()).Return([]*pmapi.Message{}, 0, nil).Times(2)

	gomock.InOrder(
		m.credentialsStore.EXPECT().List().Return([]string{"user", "users"}, nil),

		// Init for user.
		m.credentialsStore.EXPECT().Get("user").Return(testCredentials, nil),
		m.credentialsStore.EXPECT().Get("user").Return(testCredentials, nil),
		m.pmapiClient.EXPECT().AuthRefresh("token").Return(testAuthRefresh, nil),
		m.pmapiClient.EXPECT().Unlock([]byte("pass")).Return(nil),
		m.pmapiClient.EXPECT().ListLabels().Return([]*pmapi.Label{}, nil),
		m.pmapiClient.EXPECT().CountMessages("").Return([]*pmapi.MessagesCount{}, nil),
		m.pmapiClient.EXPECT().Addresses().Return([]*pmapi.Address{testPMAPIAddress}),

		// Init for users.
		m.credentialsStore.EXPECT().Get("users").Return(testCredentialsSplit, nil),
		m.credentialsStore.EXPECT().Get("users").Return(testCredentialsSplit, nil),
		m.pmapiClient.EXPECT().AuthRefresh("token").Return(testAuthRefresh, nil),
		m.pmapiClient.EXPECT().Unlock([]byte("pass")).Return(nil),
		m.pmapiClient.EXPECT().ListLabels().Return([]*pmapi.Label{}, nil),
		m.pmapiClient.EXPECT().CountMessages("").Return([]*pmapi.MessagesCount{}, nil),
		m.pmapiClient.EXPECT().Addresses().Return(testPMAPIAddresses),
	)

	users := testNewUsers(t, m)

	user, _ := users.GetUser("user")
	mockAuthUpdate(user, "reftok", m)

	user, _ = users.GetUser("user")
	mockAuthUpdate(user, "reftok", m)

	return users
}

func testNewUsers(t *testing.T, m mocks) *Users { //nolint[unparam]
	m.config.EXPECT().GetVersion().Return("ver").AnyTimes()
	m.eventListener.EXPECT().Add(events.UpgradeApplicationEvent, gomock.Any())
	m.clientManager.EXPECT().GetAuthUpdateChannel().Return(make(chan pmapi.ClientAuth))

	users := New(m.config, m.PanicHandler, m.eventListener, m.clientManager, m.credentialsStore, m.storeMaker, true)

	waitForEvents()

	return users
}

func cleanUpUsersData(b *Users) {
	for _, user := range b.users {
		_ = user.clearStore()
	}
}

func TestClearData(t *testing.T) {
	m := initMocks(t)
	defer m.ctrl.Finish()

	m.clientManager.EXPECT().GetClient("user").Return(m.pmapiClient).MinTimes(1)
	m.clientManager.EXPECT().GetClient("users").Return(m.pmapiClient).MinTimes(1)

	users := testNewUsersWithUsers(t, m)
	defer cleanUpUsersData(users)

	m.eventListener.EXPECT().Emit(events.CloseConnectionEvent, "user@pm.me")
	m.eventListener.EXPECT().Emit(events.CloseConnectionEvent, "users@pm.me")
	m.eventListener.EXPECT().Emit(events.CloseConnectionEvent, "anotheruser@pm.me")
	m.eventListener.EXPECT().Emit(events.CloseConnectionEvent, "alsouser@pm.me")

	m.pmapiClient.EXPECT().Logout()
	m.credentialsStore.EXPECT().Logout("user").Return(nil)
	m.credentialsStore.EXPECT().Get("user").Return(testCredentials, nil)

	m.pmapiClient.EXPECT().Logout()
	m.credentialsStore.EXPECT().Logout("users").Return(nil)
	m.credentialsStore.EXPECT().Get("users").Return(testCredentialsSplit, nil)

	m.config.EXPECT().ClearData().Return(nil)

	require.NoError(t, users.ClearData())

	waitForEvents()
}

func mockEventLoopNoAction(m mocks) {
	// Set up mocks for starting the store's event loop (in store.New).
	// The event loop runs in another goroutine so this might happen at any time.
	m.pmapiClient.EXPECT().GetEvent("").Return(testPMAPIEvent, nil).AnyTimes()
	m.pmapiClient.EXPECT().GetEvent(testPMAPIEvent.EventID).Return(testPMAPIEvent, nil).AnyTimes()
	m.pmapiClient.EXPECT().ListMessages(gomock.Any()).Return([]*pmapi.Message{}, 0, nil).AnyTimes()
}

func mockConnectedUser(m mocks) {
	gomock.InOrder(
		m.credentialsStore.EXPECT().Get("user").Return(testCredentials, nil),

		m.credentialsStore.EXPECT().Get("user").Return(testCredentials, nil),
		m.pmapiClient.EXPECT().AuthRefresh("token").Return(testAuthRefresh, nil),

		m.pmapiClient.EXPECT().Unlock([]byte(testCredentials.MailboxPassword)).Return(nil),

		// Set up mocks for store initialisation for the authorized user.
		m.pmapiClient.EXPECT().ListLabels().Return([]*pmapi.Label{}, nil),
		m.pmapiClient.EXPECT().CountMessages("").Return([]*pmapi.MessagesCount{}, nil),
		m.pmapiClient.EXPECT().Addresses().Return([]*pmapi.Address{testPMAPIAddress}),
	)
}

// mockAuthUpdate simulates users calling UpdateAuthToken on the given user.
// This would normally be done by users when it receives an auth from the ClientManager,
// but as we don't have a full users instance here, we do this manually.
func mockAuthUpdate(user *User, token string, m mocks) {
	gomock.InOrder(
		m.credentialsStore.EXPECT().UpdateToken("user", ":"+token).Return(nil),
		m.credentialsStore.EXPECT().Get("user").Return(credentialsWithToken(token), nil),
	)

	user.updateAuthToken(refreshWithToken(token))

	waitForEvents()
}
