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
	"testing"

	storemocks "github.com/ProtonMail/proton-bridge/v2/internal/store/mocks"
	"github.com/golang/mock/gomock"
	"github.com/pkg/errors"
)

func withTestCacher(t *testing.T, doTest func(storer *storemocks.MockStorer, cacher *MsgCachePool)) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	// Mock storer used to build/cache messages.
	storer := storemocks.NewMockStorer(ctrl)

	// Create a new cacher pointing to the fake store.
	cacher := newMsgCachePool(storer)

	// Start the cacher and wait for it to stop.
	cacher.start()
	defer cacher.stop()

	doTest(storer, cacher)
}

func TestCacher(t *testing.T) {
	// If the message is not yet cached, we should expect to try to build and cache it.
	withTestCacher(t, func(storer *storemocks.MockStorer, cacher *MsgCachePool) {
		storer.EXPECT().IsCached("messageID").Return(false)
		storer.EXPECT().BuildAndCacheMessage(cacher.ctx, "messageID").Return(nil)
		cacher.newJob("messageID")
	})
}

func TestCacherAlreadyCached(t *testing.T) {
	// If the message is already cached, we should not try to build it.
	withTestCacher(t, func(storer *storemocks.MockStorer, cacher *MsgCachePool) {
		storer.EXPECT().IsCached("messageID").Return(true)
		cacher.newJob("messageID")
	})
}

func TestCacherFail(t *testing.T) {
	// If building the message fails, we should not try to cache it.
	withTestCacher(t, func(storer *storemocks.MockStorer, cacher *MsgCachePool) {
		storer.EXPECT().IsCached("messageID").Return(false)
		storer.EXPECT().BuildAndCacheMessage(cacher.ctx, "messageID").Return(errors.New("failed to build message"))
		cacher.newJob("messageID")
	})
}

func TestCacherStop(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	// Mock storer used to build/cache messages.
	storer := storemocks.NewMockStorer(ctrl)

	// Create a new cacher pointing to the fake store.
	cacher := newMsgCachePool(storer)

	// Start the cacher.
	cacher.start()

	// Send a job -- this should succeed.
	storer.EXPECT().IsCached("messageID").Return(false)
	storer.EXPECT().BuildAndCacheMessage(cacher.ctx, "messageID").Return(nil)
	cacher.newJob("messageID")

	// Stop the cacher.
	cacher.stop()

	// Send more jobs -- these should all be dropped.
	cacher.newJob("messageID2")
	cacher.newJob("messageID3")
	cacher.newJob("messageID4")
	cacher.newJob("messageID5")

	// Stopping the cacher multiple times is safe.
	cacher.stop()
	cacher.stop()
	cacher.stop()
	cacher.stop()
}
