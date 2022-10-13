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
// along with Proton Mail Bridge.  If not, see <https://www.gnu.org/licenses/>.

package user

import (
	"context"
	"sync"

	"github.com/ProtonMail/gluon/imap"
	"github.com/ProtonMail/gluon/queue"
)

type flusher struct {
	updateCh *queue.QueuedChannel[imap.Update]
	updates  []*imap.MessageCreated

	maxUpdateSize int
	curChunkSize  int

	pushLock sync.Mutex
}

func newFlusher(updateCh *queue.QueuedChannel[imap.Update], maxUpdateSize int) *flusher {
	return &flusher{
		updateCh:      updateCh,
		maxUpdateSize: maxUpdateSize,
	}
}

func (f *flusher) push(ctx context.Context, update *imap.MessageCreated) {
	f.pushLock.Lock()
	defer f.pushLock.Unlock()

	f.updates = append(f.updates, update)

	if f.curChunkSize += len(update.Literal); f.curChunkSize >= f.maxUpdateSize {
		f.flush(ctx, false)
	}
}

func (f *flusher) flush(ctx context.Context, wait bool) {
	if len(f.updates) > 0 {
		f.updateCh.Enqueue(imap.NewMessagesCreated(f.updates...))
		f.updates = nil
		f.curChunkSize = 0
	}

	if wait {
		update := imap.NewNoop()
		defer update.WaitContext(ctx)

		f.updateCh.Enqueue(update)
	}
}
