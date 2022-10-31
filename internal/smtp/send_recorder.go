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

package smtp

import (
	"context"
	"crypto/sha256"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/ProtonMail/proton-bridge/v2/pkg/pmapi"
)

type messageGetter interface {
	GetMessage(context.Context, string) (*pmapi.Message, error)
}

type sendRecorderValue struct {
	messageID string
	time      time.Time
}

type sendRecorder struct {
	lock   *sync.RWMutex
	hashes map[string]sendRecorderValue
}

func newSendRecorder() *sendRecorder {
	return &sendRecorder{
		lock:   &sync.RWMutex{},
		hashes: map[string]sendRecorderValue{},
	}
}

func (q *sendRecorder) getMessageHash(message *pmapi.Message) string {
	// Outlook Calendar updates has only headers (no body) and thus have always
	// the same hash. If the message is type of calendar, the "is sending"
	// check to avoid potential duplicates is skipped. Duplicates should not
	// be a problem in this case as calendar updates are small.
	contentType := message.Header.Get("Content-Type")
	if strings.HasPrefix(contentType, "text/calendar") {
		return ""
	}

	h := sha256.New()
	_, _ = h.Write([]byte(message.AddressID + message.Subject))
	if message.Sender != nil {
		_, _ = h.Write([]byte(message.Sender.Address))
	}
	for _, to := range message.ToList {
		_, _ = h.Write([]byte(to.Address))
	}
	for _, to := range message.CCList {
		_, _ = h.Write([]byte(to.Address))
	}
	for _, to := range message.BCCList {
		_, _ = h.Write([]byte(to.Address))
	}
	_, _ = h.Write([]byte(message.Body))
	for _, att := range message.Attachments {
		_, _ = h.Write([]byte(att.Name + att.MIMEType + fmt.Sprintf("%d", att.Size)))
	}
	return fmt.Sprintf("%x", h.Sum(nil))
}

func (q *sendRecorder) addMessage(hash string) {
	q.lock.Lock()
	defer q.lock.Unlock()

	q.deleteExpiredKeys()
	q.hashes[hash] = sendRecorderValue{
		time: time.Now(),
	}
}

func (q *sendRecorder) removeMessage(hash string) {
	q.lock.Lock()
	defer q.lock.Unlock()

	q.deleteExpiredKeys()
	delete(q.hashes, hash)
}

func (q *sendRecorder) setMessageID(hash, messageID string) {
	q.lock.Lock()
	defer q.lock.Unlock()

	if val, ok := q.hashes[hash]; ok {
		val.messageID = messageID
		q.hashes[hash] = val
	}
}

func (q *sendRecorder) isSendingOrSent(client messageGetter, hash string) (isSending bool, wasSent bool) {
	q.lock.Lock()
	defer q.lock.Unlock()

	if hash == "" {
		return false, false
	}

	q.deleteExpiredKeys()
	value, ok := q.hashes[hash]
	if !ok {
		return
	}

	// If we have a value but don't yet have a messageID, we are in the process of uploading the draft.
	if value.messageID == "" {
		return true, false
	}

	message, err := client.GetMessage(context.TODO(), value.messageID)
	// Message could be deleted or there could be an internet issue or whatever,
	// so let's assume the message was not sent.
	if err != nil {
		return
	}
	if message.IsDraft() {
		// If message is in draft for a long time, let's assume there is
		// some problem and message will not be sent anymore.
		if time.Since(time.Unix(message.Time, 0)).Minutes() > 10 {
			return
		}
		isSending = true
	}
	// Message can be in Inbox and Sent when message was sent to myself.
	if message.Has(pmapi.FlagSent) {
		wasSent = true
	}

	return isSending, wasSent
}

func (q *sendRecorder) deleteExpiredKeys() {
	for key, value := range q.hashes {
		// It's hard to find a good expiration time.
		// On the one hand, a user could set up some cron job sending the same message over and over again (heartbeat).
		// On the the other, a user could put the device into sleep mode while sending.
		// Changing the expiration time will always make one of the edge cases worse.
		// But both edge cases are something we don't care much about. Important thing is we don't send the same message many times.
		if time.Since(value.time) > 30*time.Minute {
			delete(q.hashes, key)
		}
	}
}
