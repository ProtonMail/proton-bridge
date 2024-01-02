// Copyright (c) 2024 Proton AG
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

package sendrecorder

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/ProtonMail/gluon/rfc822"
	"github.com/bradenaw/juniper/xslices"
	"github.com/sirupsen/logrus"
	"golang.org/x/exp/slices"
)

const SendEntryExpiry = 30 * time.Minute

type ID uint64

type SendRecorder struct {
	expiry time.Duration

	entries         map[string][]*sendEntry
	entriesLock     sync.Mutex
	cancelIDCounter uint64
}

func NewSendRecorder(expiry time.Duration) *SendRecorder {
	return &SendRecorder{
		expiry:  expiry,
		entries: make(map[string][]*sendEntry),
	}
}

type sendEntry struct {
	srID         ID
	msgID        string
	toList       []string
	exp          time.Time
	waitCh       chan struct{}
	waitChClosed bool
}

func (s *sendEntry) closeWaitChannel() {
	if !s.waitChClosed {
		close(s.waitCh)
		s.waitChClosed = true
	}
}

// TryInsertWait tries to insert the given message into the send recorder.
// If an entry already exists but it was not sent yet, it waits.
// It returns whether an entry could be inserted and an error if it times out while waiting.
func (h *SendRecorder) TryInsertWait(
	ctx context.Context,
	hash string,
	toList []string,
	deadline time.Time,
) (ID, bool, error) {
	// If we successfully inserted the hash, we can return true.
	srID, waitCh, ok := h.TryInsert(hash, toList)
	if ok {
		return srID, true, nil
	}

	// A message with this hash is already being sent; wait for it.
	_, wasSent, err := h.wait(ctx, hash, waitCh, srID, deadline)
	if err != nil {
		return 0, false, fmt.Errorf("failed to wait for message to be sent: %w", err)
	}

	// If the message failed to send, try to insert it again.
	if !wasSent {
		return h.TryInsertWait(ctx, hash, toList, deadline)
	}

	return srID, false, nil
}

// HasEntryWait returns whether the given message already exists in the send recorder.
// If it does, it waits for its ID to be known, then returns it and true.
// If no entry exists, or it times out while waiting for its ID to be known, it returns false.
func (h *SendRecorder) HasEntryWait(ctx context.Context,
	hash string,
	deadline time.Time,
	toList []string,
) (string, bool, error) {
	srID, waitCh, found := h.getEntryWaitInfo(hash, toList)
	if !found {
		return "", false, nil
	}

	messageID, wasSent, err := h.wait(ctx, hash, waitCh, srID, deadline)
	if errors.Is(err, context.DeadlineExceeded) {
		return "", false, nil
	} else if err != nil {
		return "", false, fmt.Errorf("failed to wait for message to be sent: %w", err)
	}

	if wasSent {
		return messageID, true, nil
	}

	return h.HasEntryWait(ctx, hash, deadline, toList)
}

func (h *SendRecorder) removeExpiredUnsafe() {
	for hash, entry := range h.entries {
		remaining := xslices.Filter(entry, func(t *sendEntry) bool {
			return !t.exp.Before(time.Now())
		})

		if len(remaining) == 0 {
			delete(h.entries, hash)
		} else {
			h.entries[hash] = remaining
		}
	}
}

func (h *SendRecorder) TryInsert(hash string, toList []string) (ID, <-chan struct{}, bool) {
	h.entriesLock.Lock()
	defer h.entriesLock.Unlock()

	h.removeExpiredUnsafe()

	entries, ok := h.entries[hash]
	if ok {
		for _, entry := range entries {
			if matchToList(entry.toList, toList) {
				return entry.srID, entry.waitCh, false
			}
		}
	}

	cancelID := h.newSendRecorderID()
	waitCh := make(chan struct{})

	h.entries[hash] = append(entries, &sendEntry{
		srID:   cancelID,
		exp:    time.Now().Add(h.expiry),
		toList: toList,
		waitCh: waitCh,
	})

	return cancelID, waitCh, true
}

func (h *SendRecorder) getEntryWaitInfo(hash string, toList []string) (ID, <-chan struct{}, bool) {
	h.entriesLock.Lock()
	defer h.entriesLock.Unlock()

	h.removeExpiredUnsafe()

	if entries, ok := h.entries[hash]; ok {
		for _, e := range entries {
			if matchToList(e.toList, toList) {
				return e.srID, e.waitCh, true
			}
		}
	}

	return 0, nil, false
}

// SignalMessageSent should be called after a message has been successfully sent.
func (h *SendRecorder) SignalMessageSent(hash string, srID ID, msgID string) {
	h.entriesLock.Lock()
	defer h.entriesLock.Unlock()

	entries, ok := h.entries[hash]
	if ok {
		for _, entry := range entries {
			if entry.srID == srID {
				entry.msgID = msgID
				entry.closeWaitChannel()
				return
			}
		}
	}

	logrus.Warn("Cannot add message ID to send hash entry, it may have expired")
}

func (h *SendRecorder) RemoveOnFail(hash string, id ID) {
	h.entriesLock.Lock()
	defer h.entriesLock.Unlock()

	entries, ok := h.entries[hash]
	if !ok {
		return
	}

	for idx, entry := range entries {
		if entry.srID == id && entry.msgID == "" {
			entry.closeWaitChannel()

			remaining := xslices.Remove(entries, idx, 1)
			if len(remaining) != 0 {
				h.entries[hash] = remaining
			} else {
				delete(h.entries, hash)
			}
		}
	}
}

func (h *SendRecorder) wait(
	ctx context.Context,
	hash string,
	waitCh <-chan struct{},
	srID ID,
	deadline time.Time,
) (string, bool, error) {
	ctx, cancel := context.WithDeadline(ctx, deadline)
	defer cancel()

	select {
	case <-ctx.Done():
		return "", false, ctx.Err()

	case <-waitCh:
		// ...
	}

	h.entriesLock.Lock()
	defer h.entriesLock.Unlock()

	if entry, ok := h.entries[hash]; ok {
		for _, e := range entry {
			if e.srID == srID {
				return e.msgID, true, nil
			}
		}
	}

	return "", false, nil
}

func (h *SendRecorder) newSendRecorderID() ID {
	h.cancelIDCounter++
	return ID(h.cancelIDCounter)
}

// GetMessageHash returns the hash of the given message.
// This takes into account:
// - the Subject header,
// - the From/To/Cc headers,
// - the Content-Type header of each (leaf) part,
// - the Content-Disposition header of each (leaf) part,
// - the (decoded) body of each part.
func GetMessageHash(b []byte) (string, error) {
	return rfc822.GetMessageHash(b)
}

func matchToList(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}

	for i := range a {
		if !slices.Contains(b, a[i]) {
			return false
		}
	}

	for i := range b {
		if !slices.Contains(a, b[i]) {
			return false
		}
	}

	return true
}
