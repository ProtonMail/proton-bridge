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

package store

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"
	"time"

	storemocks "github.com/ProtonMail/proton-bridge/internal/store/mocks"
	"github.com/ProtonMail/proton-bridge/pkg/pmapi"
	pmapimocks "github.com/ProtonMail/proton-bridge/pkg/pmapi/mocks"
	"github.com/golang/mock/gomock"

	"github.com/stretchr/testify/require"
)

const (
	addr1   = "niceaddress@pm.me"
	addrID1 = "niceaddressID"

	addr2   = "jamesandmichalarecool@pm.me"
	addrID2 = "jamesandmichalarecool"
)

type mocksForStore struct {
	tb testing.TB

	ctrl           *gomock.Controller
	events         *storemocks.MockListener
	user           *storemocks.MockBridgeUser
	client         *pmapimocks.MockClient
	clientManager  *storemocks.MockClientManager
	panicHandler   *storemocks.MockPanicHandler
	changeNotifier *storemocks.MockChangeNotifier
	store          *Store

	tmpDir string
	cache  *Cache
}

func initMocks(tb testing.TB) (*mocksForStore, func()) {
	ctrl := gomock.NewController(tb)
	mocks := &mocksForStore{
		tb:             tb,
		ctrl:           ctrl,
		events:         storemocks.NewMockListener(ctrl),
		user:           storemocks.NewMockBridgeUser(ctrl),
		client:         pmapimocks.NewMockClient(ctrl),
		clientManager:  storemocks.NewMockClientManager(ctrl),
		panicHandler:   storemocks.NewMockPanicHandler(ctrl),
		changeNotifier: storemocks.NewMockChangeNotifier(ctrl),
	}

	// Called during clean-up.
	mocks.panicHandler.EXPECT().HandlePanic().AnyTimes()

	var err error
	mocks.tmpDir, err = ioutil.TempDir("", "store-test")
	require.NoError(tb, err)

	cacheFile := filepath.Join(mocks.tmpDir, "cache.json")
	mocks.cache = NewCache(cacheFile)

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

func (mocks *mocksForStore) newStoreNoEvents(combinedMode bool, msgs ...*pmapi.Message) { //nolint[unparam]
	mocks.user.EXPECT().ID().Return("userID").AnyTimes()
	mocks.user.EXPECT().IsConnected().Return(true)
	mocks.user.EXPECT().IsCombinedAddressMode().Return(combinedMode)

	mocks.clientManager.EXPECT().GetClient("userID").AnyTimes().Return(mocks.client)

	mocks.client.EXPECT().Addresses().Return(pmapi.AddressList{
		{ID: addrID1, Email: addr1, Type: pmapi.OriginalAddress, Receive: pmapi.CanReceive},
		{ID: addrID2, Email: addr2, Type: pmapi.AliasAddress, Receive: pmapi.CanReceive},
	})
	mocks.client.EXPECT().ListLabels()
	mocks.client.EXPECT().CountMessages("")

	// Call to get latest event ID and then to process first event.
	mocks.client.EXPECT().GetEvent("").Return(&pmapi.Event{
		EventID: "firstEventID",
	}, nil)
	mocks.client.EXPECT().GetEvent("firstEventID").Return(&pmapi.Event{
		EventID: "latestEventID",
	}, nil)

	mocks.client.EXPECT().ListMessages(gomock.Any()).Return(msgs, len(msgs), nil).AnyTimes()
	for _, msg := range msgs {
		mocks.client.EXPECT().GetMessage(msg.ID).Return(msg, nil).AnyTimes()
	}

	var err error
	mocks.store, err = New(
		mocks.panicHandler,
		mocks.user,
		mocks.clientManager,
		mocks.events,
		filepath.Join(mocks.tmpDir, "mailbox-test.db"),
		mocks.cache,
	)
	require.NoError(mocks.tb, err)

	// We want to wait until first sync has finished.
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
