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
	"io/ioutil"
	"os"
	"path/filepath"
	"sync"
	"testing"

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

	ctrl          *gomock.Controller
	events        *storemocks.MockListener
	user          *storemocks.MockBridgeUser
	client        *pmapimocks.MockClient
	clientManager *storemocks.MockClientManager
	panicHandler  *storemocks.MockPanicHandler
	store         *Store

	tmpDir string
	cache  *Cache
}

func initMocks(tb testing.TB) (*mocksForStore, func()) {
	ctrl := gomock.NewController(tb)
	mocks := &mocksForStore{
		tb:            tb,
		ctrl:          ctrl,
		events:        storemocks.NewMockListener(ctrl),
		user:          storemocks.NewMockBridgeUser(ctrl),
		client:        pmapimocks.NewMockClient(ctrl),
		clientManager: storemocks.NewMockClientManager(ctrl),
		panicHandler:  storemocks.NewMockPanicHandler(ctrl),
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

func (mocks *mocksForStore) newStoreNoEvents(combinedMode bool) { //nolint[unparam]
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
	mocks.client.EXPECT().GetEvent(gomock.Any()).
		Return(&pmapi.Event{
			EventID: "latestEventID",
		}, nil).AnyTimes()

	// We want to wait until first sync has finished.
	firstSyncWaiter := sync.WaitGroup{}
	firstSyncWaiter.Add(1)
	mocks.client.EXPECT().
		ListMessages(gomock.Any()).
		DoAndReturn(func(*pmapi.MessagesFilter) ([]*pmapi.Message, int, error) {
			firstSyncWaiter.Done()
			return []*pmapi.Message{}, 0, nil
		})

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

	// Wait for sync to finish.
	firstSyncWaiter.Wait()
}
