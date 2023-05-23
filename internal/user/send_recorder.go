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

const sendEntryExpiry = 30 * time.Minute

type sendRecorder struct {
	expiry time.Duration

	entries     map[string][]*sendEntry
	entriesLock sync.Mutex
}

func newSendRecorder(expiry time.Duration) *sendRecorder {
	return &sendRecorder{
		expiry:  expiry,
		entries: make(map[string][]*sendEntry),
	}
}

type sendEntry struct {
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

// tryInsertWait tries to insert the given message into the send recorder.
// If an entry already exists but it was not sent yet, it waits.
// It returns whether an entry could be inserted and an error if it times out while waiting.
func (h *sendRecorder) tryInsertWait(
	ctx context.Context,
	hash string,
	toList []string,
	deadline time.Time,
) (bool, error) {
	// If we successfully inserted the hash, we can return true.
	if h.tryInsert(hash, toList) {
		return true, nil
	}

	// A message with this hash is already being sent; wait for it.
	_, wasSent, err := h.wait(ctx, hash, deadline)
	if err != nil {
		return false, fmt.Errorf("failed to wait for message to be sent: %w", err)
	}

	// If the message failed to send, try to insert it again.
	if !wasSent {
		return h.tryInsertWait(ctx, hash, toList, deadline)
	}

	return false, nil
}

// hasEntryWait returns whether the given message already exists in the send recorder.
// If it does, it waits for its ID to be known, then returns it and true.
// If no entry exists, or it times out while waiting for its ID to be known, it returns false.
func (h *sendRecorder) hasEntryWait(ctx context.Context, hash string, deadline time.Time) (string, bool, error) {
	if !h.hasEntry(hash) {
		return "", false, nil
	}

	messageID, wasSent, err := h.wait(ctx, hash, deadline)
	if errors.Is(err, context.DeadlineExceeded) {
		return "", false, nil
	} else if err != nil {
		return "", false, fmt.Errorf("failed to wait for message to be sent: %w", err)
	}

	if wasSent {
		return messageID, true, nil
	}

	return h.hasEntryWait(ctx, hash, deadline)
}

func (h *sendRecorder) removeExpiredUnsafe() {
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

func (h *sendRecorder) tryInsert(hash string, toList []string) bool {
	h.entriesLock.Lock()
	defer h.entriesLock.Unlock()

	h.removeExpiredUnsafe()

	entries, ok := h.entries[hash]
	if ok {
		for _, entry := range entries {
			if matchToList(entry.toList, toList) {
				return false
			}
		}
	}

	h.entries[hash] = append(entries, &sendEntry{
		exp:    time.Now().Add(h.expiry),
		toList: toList,
		waitCh: make(chan struct{}),
	})

	return true
}

func (h *sendRecorder) hasEntry(hash string) bool {
	h.entriesLock.Lock()
	defer h.entriesLock.Unlock()

	h.removeExpiredUnsafe()

	if _, ok := h.entries[hash]; ok {
		return true
	}

	return false
}

// signalMessageSent should be called after a message has been successfully sent.
func (h *sendRecorder) signalMessageSent(hash, msgID string, toList []string) {
	h.entriesLock.Lock()
	defer h.entriesLock.Unlock()

	entries, ok := h.entries[hash]
	if ok {
		for _, entry := range entries {
			if matchToList(entry.toList, toList) {
				entry.msgID = msgID
				entry.closeWaitChannel()
				return
			}
		}
	}

	logrus.Warn("Cannot add message ID to send hash entry, it may have expired")
}

func (h *sendRecorder) removeOnFail(hash string, toList []string) {
	h.entriesLock.Lock()
	defer h.entriesLock.Unlock()

	entries, ok := h.entries[hash]
	if !ok {
		return
	}

	for idx, entry := range entries {
		if entry.msgID == "" && matchToList(entry.toList, toList) {
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

func (h *sendRecorder) wait(ctx context.Context, hash string, deadline time.Time) (string, bool, error) {
	ctx, cancel := context.WithDeadline(ctx, deadline)
	defer cancel()

	waitCh, ok := h.getWaitCh(hash)
	if !ok {
		return "", false, nil
	}

	select {
	case <-ctx.Done():
		return "", false, ctx.Err()

	case <-waitCh:
		// ...
	}

	h.entriesLock.Lock()
	defer h.entriesLock.Unlock()

	if entry, ok := h.entries[hash]; ok {
		return entry[0].msgID, true, nil
	}

	return "", false, nil
}

func (h *sendRecorder) getWaitCh(hash string) (<-chan struct{}, bool) {
	h.entriesLock.Lock()
	defer h.entriesLock.Unlock()

	if entry, ok := h.entries[hash]; ok {
		return entry[0].waitCh, true
	}

	return nil, false
}

// getMessageHash returns the hash of the given message.
// This takes into account:
// - the Subject header,
// - the From/To/Cc headers,
// - the Content-Type header of each (leaf) part,
// - the Content-Disposition header of each (leaf) part,
// - the (decoded) body of each part.
func getMessageHash(b []byte) (string, error) {
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
