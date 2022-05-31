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
	"fmt"
	"io/ioutil"
	"os"
	"runtime"
	"runtime/debug"
	"testing"
	"time"

	"github.com/ProtonMail/proton-bridge/v2/internal/events"
	"github.com/ProtonMail/proton-bridge/v2/internal/sentry"
	"github.com/ProtonMail/proton-bridge/v2/internal/store"
	"github.com/ProtonMail/proton-bridge/v2/internal/store/cache"
	"github.com/ProtonMail/proton-bridge/v2/internal/users/credentials"
	usersmocks "github.com/ProtonMail/proton-bridge/v2/internal/users/mocks"
	"github.com/ProtonMail/proton-bridge/v2/pkg/message"
	"github.com/ProtonMail/proton-bridge/v2/pkg/pmapi"
	pmapimocks "github.com/ProtonMail/proton-bridge/v2/pkg/pmapi/mocks"
	tests "github.com/ProtonMail/proton-bridge/v2/test"
	gomock "github.com/golang/mock/gomock"
	"github.com/sirupsen/logrus"
	r "github.com/stretchr/testify/require"
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
	testAuthRefresh = &pmapi.Auth{ //nolint:gochecknoglobals
		UserID: "user",
		AuthRefresh: pmapi.AuthRefresh{
			UID:          "uid",
			AccessToken:  "acc",
			RefreshToken: "ref",
		},
	}

	testCredentials = &credentials.Credentials{ //nolint:gochecknoglobals
		UserID:                "user",
		Name:                  "username",
		Emails:                "user@pm.me",
		APIToken:              "uid:acc",
		MailboxPassword:       []byte("pass"),
		BridgePassword:        "0123456789abcdef",
		Version:               "v1",
		Timestamp:             123456789,
		IsHidden:              false,
		IsCombinedAddressMode: true,
	}

	testCredentialsSplit = &credentials.Credentials{ //nolint:gochecknoglobals
		UserID:                "users",
		Name:                  "usersname",
		Emails:                "users@pm.me;anotheruser@pm.me;alsouser@pm.me",
		APIToken:              "uid:acc",
		MailboxPassword:       []byte("pass"),
		BridgePassword:        "0123456789abcdef",
		Version:               "v1",
		Timestamp:             123456789,
		IsHidden:              false,
		IsCombinedAddressMode: false,
	}

	testCredentialsDisconnected = &credentials.Credentials{ //nolint:gochecknoglobals
		UserID:                "userDisconnected",
		Name:                  "username",
		Emails:                "user@pm.me",
		APIToken:              "",
		MailboxPassword:       []byte{},
		BridgePassword:        "0123456789abcdef",
		Version:               "v1",
		Timestamp:             123456789,
		IsHidden:              false,
		IsCombinedAddressMode: true,
	}

	testCredentialsSplitDisconnected = &credentials.Credentials{ //nolint:gochecknoglobals
		UserID:                "usersDisconnected",
		Name:                  "usersname",
		Emails:                "users@pm.me;anotheruser@pm.me;alsouser@pm.me",
		APIToken:              "",
		MailboxPassword:       []byte{},
		BridgePassword:        "0123456789abcdef",
		Version:               "v1",
		Timestamp:             123456789,
		IsHidden:              false,
		IsCombinedAddressMode: false,
	}

	usedSpace = int64(1048576)
	maxSpace  = int64(10485760)

	testPMAPIUser = &pmapi.User{ //nolint:gochecknoglobals
		ID:        "user",
		Name:      "username",
		UsedSpace: &usedSpace,
		MaxSpace:  &maxSpace,
	}

	testPMAPIUserDisconnected = &pmapi.User{ //nolint:gochecknoglobals
		ID:   "userDisconnected",
		Name: "username",
	}

	testPMAPIAddress = &pmapi.Address{ //nolint:gochecknoglobals
		ID:      "testAddressID",
		Type:    pmapi.OriginalAddress,
		Email:   "user@pm.me",
		Receive: true,
	}

	testPMAPIEvent = &pmapi.Event{ // nolint:gochecknoglobals
		EventID: "ACXDmTaBub14w==",
	}
)

type mocks struct {
	t *testing.T

	ctrl             *gomock.Controller
	locator          *usersmocks.MockLocator
	PanicHandler     *usersmocks.MockPanicHandler
	credentialsStore *usersmocks.MockCredentialsStorer
	storeMaker       *usersmocks.MockStoreMaker
	eventListener    *usersmocks.MockListener

	clientManager *pmapimocks.MockManager
	pmapiClient   *pmapimocks.MockClient

	storeCache *store.Events
}

func initMocks(t *testing.T) mocks {
	var mockCtrl *gomock.Controller
	if os.Getenv("VERBOSITY") == "trace" {
		mockCtrl = gomock.NewController(&fullStackReporter{t})
	} else {
		mockCtrl = gomock.NewController(t)
	}

	cacheFile, err := ioutil.TempFile("", "bridge-store-cache-*.db")
	r.NoError(t, err, "could not get temporary file for store cache")
	r.NoError(t, cacheFile.Close())

	m := mocks{
		t: t,

		ctrl:             mockCtrl,
		locator:          usersmocks.NewMockLocator(mockCtrl),
		PanicHandler:     usersmocks.NewMockPanicHandler(mockCtrl),
		credentialsStore: usersmocks.NewMockCredentialsStorer(mockCtrl),
		storeMaker:       usersmocks.NewMockStoreMaker(mockCtrl),
		eventListener:    usersmocks.NewMockListener(mockCtrl),

		clientManager: pmapimocks.NewMockManager(mockCtrl),
		pmapiClient:   pmapimocks.NewMockClient(mockCtrl),

		storeCache: store.NewEvents(cacheFile.Name()),
	}

	// Called during clean-up.
	m.PanicHandler.EXPECT().HandlePanic().AnyTimes()

	// Set up store factory.
	m.storeMaker.EXPECT().New(gomock.Any()).DoAndReturn(func(user store.BridgeUser) (*store.Store, error) {
		var sentryReporter *sentry.Reporter // Sentry reporter is not used under unit tests.

		dbFile, err := ioutil.TempFile(t.TempDir(), "bridge-store-db-*.db")
		r.NoError(t, err, "could not get temporary file for store db")
		r.NoError(t, dbFile.Close())

		return store.New(
			sentryReporter,
			m.PanicHandler,
			user,
			m.eventListener,
			cache.NewInMemoryCache(1<<20),
			message.NewBuilder(runtime.NumCPU(), runtime.NumCPU()),
			dbFile.Name(),
			m.storeCache,
		)
	}).AnyTimes()
	m.storeMaker.EXPECT().Remove(gomock.Any()).AnyTimes()

	return m
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

func testNewUsersWithUsers(t *testing.T, m mocks) *Users {
	m.credentialsStore.EXPECT().List().Return([]string{testCredentials.UserID, testCredentialsSplit.UserID}, nil)
	mockLoadingConnectedUser(t, m, testCredentials)
	mockLoadingConnectedUser(t, m, testCredentialsSplit)
	mockEventLoopNoAction(m)

	return testNewUsers(t, m)
}

func testNewUsers(t *testing.T, m mocks) *Users { //nolint:unparam
	m.eventListener.EXPECT().ProvideChannel(events.UpgradeApplicationEvent)
	m.eventListener.EXPECT().ProvideChannel(events.InternetConnChangedEvent)

	users := New(m.locator, m.PanicHandler, m.eventListener, m.clientManager, m.credentialsStore, m.storeMaker)

	waitForEvents()

	return users
}

func waitForEvents() {
	// Wait for goroutine to add listener.
	// E.g. calling login to invoke firstsync event. Functions can end sooner than
	// goroutines call the listener mock. We need to wait a little bit before the end of
	// the test to capture all event calls. This allows us to detect whether there were
	// missing calls, or perhaps whether something was called too many times.
	time.Sleep(100 * time.Millisecond)
}

func cleanUpUsersData(b *Users) {
	for _, user := range b.users {
		_ = user.clearStore()
	}
}

func mockAddingConnectedUser(t *testing.T, m mocks) {
	gomock.InOrder(
		// Mock of users.FinishLogin.
		m.pmapiClient.EXPECT().AuthSalt(gomock.Any()).Return("", nil),
		m.pmapiClient.EXPECT().Unlock(gomock.Any(), testCredentials.MailboxPassword).Return(nil),
		m.pmapiClient.EXPECT().CurrentUser(gomock.Any()).Return(testPMAPIUser, nil),
		m.pmapiClient.EXPECT().Addresses().Return([]*pmapi.Address{testPMAPIAddress}),
		m.credentialsStore.EXPECT().Add("user", "username", testAuthRefresh.UID, testAuthRefresh.RefreshToken, testCredentials.MailboxPassword, []string{testPMAPIAddress.Email}).Return(testCredentials, nil),
		m.credentialsStore.EXPECT().Get("user").Return(testCredentials, nil),
	)

	mockInitConnectedUser(t, m)
}

func mockLoadingConnectedUser(t *testing.T, m mocks, creds *credentials.Credentials) {
	authRefresh := &pmapi.AuthRefresh{
		UID:          "uid",
		AccessToken:  "acc",
		RefreshToken: "ref",
	}

	gomock.InOrder(
		// Mock of users.loadUsersFromCredentialsStore.
		m.credentialsStore.EXPECT().Get(creds.UserID).Return(creds, nil),
		m.clientManager.EXPECT().NewClientWithRefresh(gomock.Any(), "uid", "acc").Return(m.pmapiClient, authRefresh, nil),
		m.credentialsStore.EXPECT().UpdateToken(creds.UserID, authRefresh.UID, authRefresh.RefreshToken).Return(creds, nil),
	)

	mockInitConnectedUser(t, m)
}

func mockInitConnectedUser(t *testing.T, m mocks) {
	// Mock of user initialisation.
	m.pmapiClient.EXPECT().AddAuthRefreshHandler(gomock.Any())
	m.pmapiClient.EXPECT().IsUnlocked().Return(true).AnyTimes()
	m.pmapiClient.EXPECT().GetUser(gomock.Any()).Return(testPMAPIUser, nil) // load connected user

	// Mock of store initialisation.
	gomock.InOrder(
		m.pmapiClient.EXPECT().ListLabels(gomock.Any()).Return([]*pmapi.Label{}, nil),
		m.pmapiClient.EXPECT().CountMessages(gomock.Any(), "").Return([]*pmapi.MessagesCount{}, nil),
		m.pmapiClient.EXPECT().Addresses().Return([]*pmapi.Address{testPMAPIAddress}),
		m.pmapiClient.EXPECT().GetUserKeyRing().Return(tests.MakeKeyRing(t), nil).AnyTimes(),
	)
}

func mockLoadingDisconnectedUser(m mocks, creds *credentials.Credentials) {
	gomock.InOrder(
		// Mock of users.loadUsersFromCredentialsStore.
		m.credentialsStore.EXPECT().Get(creds.UserID).Return(creds, nil),
		m.clientManager.EXPECT().NewClient("", "", "", time.Time{}).Return(m.pmapiClient),
	)

	mockInitDisconnectedUser(m)
}

func mockInitDisconnectedUser(m mocks) {
	gomock.InOrder(
		// Mock of user initialisation.
		m.pmapiClient.EXPECT().AddAuthRefreshHandler(gomock.Any()),

		// Mock of store initialisation for the unauthorized user.
		m.pmapiClient.EXPECT().ListLabels(gomock.Any()).Return(nil, pmapi.ErrUnauthorized),
		m.pmapiClient.EXPECT().Addresses().Return(nil),
	)
}

func mockEventLoopNoAction(m mocks) {
	// Set up mocks for starting the store's event loop (in store.New).
	// The event loop runs in another goroutine so this might happen at any time.
	m.pmapiClient.EXPECT().GetEvent(gomock.Any(), "").Return(testPMAPIEvent, nil).AnyTimes()
	m.pmapiClient.EXPECT().GetEvent(gomock.Any(), testPMAPIEvent.EventID).Return(testPMAPIEvent, nil).AnyTimes()
	m.pmapiClient.EXPECT().ListMessages(gomock.Any(), gomock.Any()).Return([]*pmapi.Message{}, 0, nil).AnyTimes()
}
