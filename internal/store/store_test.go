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

package store

import (
	"context"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"runtime"
	"testing"
	"time"

	"github.com/ProtonMail/gopenpgp/v2/crypto"
	"github.com/ProtonMail/proton-bridge/v2/internal/store/cache"
	storemocks "github.com/ProtonMail/proton-bridge/v2/internal/store/mocks"
	"github.com/ProtonMail/proton-bridge/v2/pkg/message"
	"github.com/ProtonMail/proton-bridge/v2/pkg/pmapi"
	pmapimocks "github.com/ProtonMail/proton-bridge/v2/pkg/pmapi/mocks"
	tests "github.com/ProtonMail/proton-bridge/v2/test"
	"github.com/golang/mock/gomock"

	"github.com/stretchr/testify/require"
)

const (
	addr1   = "niceaddress@pm.me"
	addrID1 = "niceaddressID"

	addr2   = "jamesandmichalarecool@pm.me"
	addrID2 = "jamesandmichalarecool"

	testPrivateKeyPassword = "apple"
	testPrivateKey         = `-----BEGIN PGP PRIVATE KEY BLOCK-----
Version: OpenPGP.js v0.7.1
Comment: http://openpgpjs.org

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
qqGRQm3MxoTdgQUShAwbNwNNQR9cInfMnA==
=2wIY
-----END PGP PRIVATE KEY BLOCK-----
`
)

var testPrivateKeyRing *crypto.KeyRing

func init() {
	privKey, err := crypto.NewKeyFromArmored(testPrivateKey)
	if err != nil {
		panic(err)
	}

	privKeyUnlocked, err := privKey.Unlock([]byte(testPrivateKeyPassword))
	if err != nil {
		panic(err)
	}

	if testPrivateKeyRing, err = crypto.NewKeyRing(privKeyUnlocked); err != nil {
		panic(err)
	}
}

type mocksForStore struct {
	tb testing.TB

	ctrl           *gomock.Controller
	events         *storemocks.MockListener
	user           *storemocks.MockBridgeUser
	client         *pmapimocks.MockClient
	panicHandler   *storemocks.MockPanicHandler
	changeNotifier *storemocks.MockChangeNotifier
	store          *Store

	tmpDir string
	cache  *Events
}

func initMocks(tb testing.TB) (*mocksForStore, func()) {
	ctrl := gomock.NewController(tb)
	mocks := &mocksForStore{
		tb:             tb,
		ctrl:           ctrl,
		events:         storemocks.NewMockListener(ctrl),
		user:           storemocks.NewMockBridgeUser(ctrl),
		client:         pmapimocks.NewMockClient(ctrl),
		panicHandler:   storemocks.NewMockPanicHandler(ctrl),
		changeNotifier: storemocks.NewMockChangeNotifier(ctrl),
	}

	// Called during clean-up.
	mocks.panicHandler.EXPECT().HandlePanic().AnyTimes()

	var err error
	mocks.tmpDir, err = ioutil.TempDir("", "store-test")
	require.NoError(tb, err)

	cacheFile := filepath.Join(mocks.tmpDir, "cache.json")
	mocks.cache = NewEvents(cacheFile)

	return mocks, func() {
		if err := recover(); err != nil {
			panic(err)
		}
		if mocks.store != nil {
			require.Nil(tb, mocks.store.Close())
		}
		ctrl.Finish()
		require.NoError(tb, os.RemoveAll(mocks.tmpDir))
	}
}

func (mocks *mocksForStore) newStoreNoEvents(t *testing.T, combinedMode bool, msgs ...*pmapi.Message) { //nolint:unparam
	mocks.user.EXPECT().ID().Return("userID").AnyTimes()
	mocks.user.EXPECT().IsConnected().Return(true)
	mocks.user.EXPECT().IsCombinedAddressMode().Return(combinedMode)

	mocks.user.EXPECT().GetClient().AnyTimes().Return(mocks.client)

	testUserKeyring := tests.MakeKeyRing(t)
	mocks.client.EXPECT().GetUserKeyRing().Return(testUserKeyring, nil).AnyTimes()
	mocks.client.EXPECT().Addresses().Return(pmapi.AddressList{
		{ID: addrID1, Email: addr1, Type: pmapi.OriginalAddress, Receive: true},
		{ID: addrID2, Email: addr2, Type: pmapi.AliasAddress, Receive: true},
	})
	mocks.client.EXPECT().ListLabels(gomock.Any()).AnyTimes()
	mocks.client.EXPECT().CountMessages(gomock.Any(), "")

	// Call to get latest event ID and then to process first event.
	eventAfterSyncRequested := make(chan struct{})
	mocks.client.EXPECT().GetEvent(gomock.Any(), "").Return(&pmapi.Event{
		EventID: "firstEventID",
	}, nil)
	mocks.client.EXPECT().GetEvent(gomock.Any(), "firstEventID").DoAndReturn(func(_ context.Context, _ string) (*pmapi.Event, error) {
		close(eventAfterSyncRequested)
		return &pmapi.Event{
			EventID: "latestEventID",
		}, nil
	})

	mocks.client.EXPECT().ListMessages(gomock.Any(), gomock.Any()).Return(msgs, len(msgs), nil).AnyTimes()
	for _, msg := range msgs {
		mocks.client.EXPECT().GetMessage(gomock.Any(), msg.ID).Return(msg, nil).AnyTimes()
	}

	var err error
	mocks.store, err = New(
		nil, // Sentry reporter is not used under unit tests.
		mocks.panicHandler,
		mocks.user,
		mocks.events,
		cache.NewInMemoryCache(1<<20),
		message.NewBuilder(runtime.NumCPU(), runtime.NumCPU()),
		filepath.Join(mocks.tmpDir, "mailbox-test.db"),
		mocks.cache,
	)
	require.NoError(mocks.tb, err)

	require.NoError(mocks.tb, mocks.store.UnlockCache(testUserKeyring))

	// We want to wait until first sync has finished.
	// Checking that event after sync was reuested is not the best way to
	// do the check, because sync could take more time, but sync is going
	// in background and if there is no message to wait for, we don't have
	// anything better.
	select {
	case <-eventAfterSyncRequested:
	case <-time.After(5 * time.Second):
	}
	require.Eventually(mocks.tb, func() bool {
		for _, msg := range msgs {
			_, err := mocks.store.getMessageFromDB(msg.ID)
			if err != nil {
				// To see in test result the latest error for debugging.
				fmt.Println("Sync wait error:", err)
				return false
			}
		}
		return true
	}, 5*time.Second, 10*time.Millisecond)
}
