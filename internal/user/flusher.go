package user

import (
	"sync"
	"time"

	"github.com/ProtonMail/gluon/imap"
	"github.com/ProtonMail/gluon/queue"
	"github.com/ProtonMail/proton-bridge/v2/internal/events"
)

type flusher struct {
	userID   string
	updateCh *queue.QueuedChannel[imap.Update]
	eventCh  *queue.QueuedChannel[events.Event]

	updates      []*imap.MessageCreated
	maxChunkSize int
	curChunkSize int

	count int
	total int
	start time.Time

	pushLock sync.Mutex
}

func newFlusher(
	userID string,
	updateCh *queue.QueuedChannel[imap.Update],
	eventCh *queue.QueuedChannel[events.Event],
	total, maxChunkSize int,
) *flusher {
	return &flusher{
		userID:   userID,
		updateCh: updateCh,
		eventCh:  eventCh,

		maxChunkSize: maxChunkSize,

		total: total,
		start: time.Now(),
	}
}

func (f *flusher) push(update *imap.MessageCreated) {
	f.pushLock.Lock()
	defer f.pushLock.Unlock()

	f.updates = append(f.updates, update)

	if f.curChunkSize += len(update.Literal); f.curChunkSize >= f.maxChunkSize {
		f.flush()
	}
}

func (f *flusher) flush() {
	if len(f.updates) == 0 {
		return
	}

	f.count += len(f.updates)
	f.updateCh.Enqueue(imap.NewMessagesCreated(f.updates...))
	f.eventCh.Enqueue(newSyncProgress(f.userID, f.count, f.total, f.start))
	f.updates = nil
	f.curChunkSize = 0
}

func newSyncProgress(userID string, count, total int, start time.Time) events.SyncProgress {
	return events.SyncProgress{
		UserID:    userID,
		Progress:  float64(count) / float64(total),
		Elapsed:   time.Since(start),
		Remaining: time.Since(start) * time.Duration(total-count) / time.Duration(count),
	}
}
