package user

import (
	"context"
	"sync"

	"github.com/ProtonMail/gluon/imap"
	"github.com/ProtonMail/gluon/queue"
)

type flusher struct {
	userID   string
	updateCh *queue.QueuedChannel[imap.Update]

	updates      []*imap.MessageCreated
	maxChunkSize int
	curChunkSize int

	pushLock sync.Mutex
}

func newFlusher(userID string, updateCh *queue.QueuedChannel[imap.Update], maxChunkSize int) *flusher {
	return &flusher{
		userID:       userID,
		updateCh:     updateCh,
		maxChunkSize: maxChunkSize,
	}
}

func (f *flusher) push(ctx context.Context, update *imap.MessageCreated) {
	f.pushLock.Lock()
	defer f.pushLock.Unlock()

	f.updates = append(f.updates, update)

	if f.curChunkSize += len(update.Literal); f.curChunkSize >= f.maxChunkSize {
		f.flush(ctx, false)
	}
}

func (f *flusher) flush(ctx context.Context, wait bool) {
	if len(f.updates) == 0 {
		return
	}

	f.updateCh.Enqueue(imap.NewMessagesCreated(f.updates...))
	f.updates = nil
	f.curChunkSize = 0

	if wait {
		update := imap.NewNoop()
		defer update.WaitContext(ctx)

		f.updateCh.Enqueue(update)
	}
}
