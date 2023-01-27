// Copyright (c) 2023 Proton AG
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
	"github.com/ProtonMail/gluon/imap"
	"github.com/ProtonMail/gluon/queue"
)

type flusher struct {
	updateCh *queue.QueuedChannel[imap.Update]
	updates  []*imap.MessageCreated

	maxUpdateSize int
	curChunkSize  int
}

func newFlusher(updateCh *queue.QueuedChannel[imap.Update], maxUpdateSize int) *flusher {
	return &flusher{
		updateCh:      updateCh,
		maxUpdateSize: maxUpdateSize,
	}
}

func (f *flusher) push(update *imap.MessageCreated) {
	f.updates = append(f.updates, update)

	if f.curChunkSize += len(update.Literal); f.curChunkSize >= f.maxUpdateSize {
		f.flush()
	}
}

func (f *flusher) flush() {
	if len(f.updates) > 0 {
		f.updateCh.Enqueue(imap.NewMessagesCreated(true, f.updates...))
		f.updates = nil
		f.curChunkSize = 0
	}
}
