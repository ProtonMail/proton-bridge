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
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestSendHasher_Insert(t *testing.T) {
	h := NewSendRecorder(SendEntryExpiry)

	// Insert a message into the hasher.
	srdID1, hash1, ok, err := testTryInsert(h, literal1, time.Now().Add(time.Second))
	require.NoError(t, err)
	require.True(t, ok)
	require.NotEmpty(t, hash1)

	// Simulate successfully sending the message.
	h.SignalMessageSent(hash1, srdID1, "abc")

	// Inserting a message with the same hash should return false.
	srdID2, _, ok, err := testTryInsert(h, literal1, time.Now().Add(time.Second))
	require.NoError(t, err)
	require.False(t, ok)
	require.Equal(t, srdID1, srdID2)

	// Inserting a message with a different hash should return true.
	srdID3, hash2, ok, err := testTryInsert(h, literal2, time.Now().Add(time.Second))
	require.NoError(t, err)
	require.True(t, ok)
	require.NotEmpty(t, hash2)
	require.NotEqual(t, srdID3, srdID1)
}

func TestSendHasher_Insert_Expired(t *testing.T) {
	h := NewSendRecorder(time.Second)

	// Insert a message into the hasher.
	srID1, hash1, ok, err := testTryInsert(h, literal1, time.Now().Add(time.Second))
	require.NoError(t, err)
	require.True(t, ok)
	require.NotEmpty(t, hash1)

	// Simulate successfully sending the message.
	h.SignalMessageSent(hash1, srID1, "abc")

	// Wait for the entry to expire.
	time.Sleep(time.Second)

	// Inserting a message with the same hash should return true because the previous entry has since expired.
	srID2, hash2, ok, err := testTryInsert(h, literal1, time.Now().Add(time.Second))
	require.NoError(t, err)
	require.True(t, ok)

	// The hashes should be the same.
	require.Equal(t, hash1, hash2)

	// Send IDs should differ
	require.NotEqual(t, srID2, srID1)
}

func TestSendHasher_Insert_DifferentToList(t *testing.T) {
	h := NewSendRecorder(time.Second)

	// Insert a message into the hasher.
	srID1, hash1, ok, err := testTryInsert(h, literal1, time.Now().Add(time.Second), []string{"abc", "def"}...)
	require.NoError(t, err)
	require.True(t, ok)
	require.NotEmpty(t, hash1)

	// Insert the same message into the hasher but with a different to list.
	srID2, hash2, ok, err := testTryInsert(h, literal1, time.Now().Add(time.Second), []string{"abc", "def", "ghi"}...)
	require.NoError(t, err)
	require.True(t, ok)
	require.NotEmpty(t, hash2)
	require.NotEqual(t, srID1, srID2)

	// Insert the same message into the hasher but with the same to list.
	_, _, ok, err = testTryInsert(h, literal1, time.Now().Add(time.Second), []string{"abc", "def", "ghi"}...)
	require.Error(t, err)
	require.False(t, ok)
}

func TestSendHasher_Wait_SendSuccess(t *testing.T) {
	h := NewSendRecorder(SendEntryExpiry)

	// Insert a message into the hasher.
	srID1, hash, ok, err := testTryInsert(h, literal1, time.Now().Add(time.Second))
	require.NoError(t, err)
	require.True(t, ok)
	require.NotEmpty(t, hash)

	// Simulate successfully sending the message after half a second.
	go func() {
		time.Sleep(time.Millisecond * 500)
		h.SignalMessageSent(hash, srID1, "abc")
	}()

	// Inserting a message with the same hash should fail.
	srID2, _, ok, err := testTryInsert(h, literal1, time.Now().Add(time.Second))
	require.NoError(t, err)
	require.False(t, ok)
	require.Equal(t, srID1, srID2)
}

func TestSendHasher_Wait_SendFail(t *testing.T) {
	h := NewSendRecorder(SendEntryExpiry)

	// Insert a message into the hasher.
	srID1, hash, ok, err := testTryInsert(h, literal1, time.Now().Add(time.Second))
	require.NoError(t, err)
	require.True(t, ok)
	require.NotEmpty(t, hash)

	// Simulate failing to send the message after half a second.
	go func() {
		time.Sleep(time.Millisecond * 500)
		h.RemoveOnFail(hash, srID1)
	}()

	// Inserting a message with the same hash should succeed because the first message failed to send.
	srID2, hash2, ok, err := testTryInsert(h, literal1, time.Now().Add(time.Second))
	require.NoError(t, err)
	require.True(t, ok)
	require.NotEqual(t, srID2, srID1)

	// The hashes should be the same.
	require.Equal(t, hash, hash2)
}

func TestSendHasher_Wait_Timeout(t *testing.T) {
	h := NewSendRecorder(SendEntryExpiry)

	// Insert a message into the hasher.
	_, hash, ok, err := testTryInsert(h, literal1, time.Now().Add(time.Second))
	require.NoError(t, err)
	require.True(t, ok)
	require.NotEmpty(t, hash)

	// We should fail to insert because the message is not sent within the timeout period.
	_, _, _, err = testTryInsert(h, literal1, time.Now().Add(time.Second))
	require.Error(t, err)
}

func TestSendHasher_HasEntry(t *testing.T) {
	h := NewSendRecorder(SendEntryExpiry)

	// Insert a message into the hasher.
	srID1, hash, ok, err := testTryInsert(h, literal1, time.Now().Add(time.Second))
	require.NoError(t, err)
	require.True(t, ok)
	require.NotEmpty(t, hash)

	// Simulate successfully sending the message.
	h.SignalMessageSent(hash, srID1, "abc")

	// The message was already sent; we should find it in the hasher.
	messageID, ok, err := testHasEntry(h, literal1, time.Now().Add(time.Second))
	require.NoError(t, err)
	require.True(t, ok)
	require.Equal(t, "abc", messageID)
}

func TestSendHasher_HasEntry_SendSuccess(t *testing.T) {
	h := NewSendRecorder(SendEntryExpiry)

	// Insert a message into the hasher.
	srID1, hash, ok, err := testTryInsert(h, literal1, time.Now().Add(time.Second))
	require.NoError(t, err)
	require.True(t, ok)
	require.NotEmpty(t, hash)

	// Simulate successfully sending the message after half a second.
	go func() {
		time.Sleep(time.Millisecond * 500)
		h.SignalMessageSent(hash, srID1, "abc")
	}()

	// The message was already sent; we should find it in the hasher.
	messageID, ok, err := testHasEntry(h, literal1, time.Now().Add(time.Second))
	require.NoError(t, err)
	require.True(t, ok)
	require.Equal(t, "abc", messageID)
}

func TestSendHasher_DualAddDoesNotCauseCrash(t *testing.T) {
	// There may be a rare case where one 2 smtp connections attempt to send the same message, but if the first message
	// is stuck long enough for it to expire, the second connection will remove it from the list and cause it to be
	// inserted as a new entry. The two clients end up sending the message twice and calling the `SignalMessageSent` x2,
	// resulting in a crash.
	h := NewSendRecorder(SendEntryExpiry)

	// Insert a message into the hasher.
	srID1, hash, ok, err := testTryInsert(h, literal1, time.Now().Add(time.Second))
	require.NoError(t, err)
	require.True(t, ok)
	require.NotEmpty(t, hash)

	// Simulate successfully sending the message. We call this method twice as it possible for multiple SMTP connections
	// to attempt to send the same message.
	h.SignalMessageSent(hash, srID1, "abc")
	h.SignalMessageSent(hash, srID1, "abc")

	// The message was already sent; we should find it in the hasher.
	messageID, ok, err := testHasEntry(h, literal1, time.Now().Add(time.Second))
	require.NoError(t, err)
	require.True(t, ok)
	require.Equal(t, "abc", messageID)
}

func TestSendHashed_MessageWithSameHasButDifferentRecipientsIsInserted(t *testing.T) {
	h := NewSendRecorder(SendEntryExpiry)

	// Insert a message into the hasher.
	srID1, hash, ok, err := testTryInsert(h, literal1, time.Now().Add(time.Second), "Receiver <receiver@pm.me>")
	require.NoError(t, err)
	require.True(t, ok)
	require.NotEmpty(t, hash)

	srID2, hash2, ok, err := testTryInsert(h, literal1, time.Now().Add(time.Second), "Receiver <receiver@pm.me>", "Receiver2 <receiver2@pm.me>")
	require.NoError(t, err)
	require.True(t, ok)
	require.NotEmpty(t, hash2)
	require.Equal(t, hash, hash2)

	// Should map to different requests
	require.NotEqual(t, srID2, srID1)
}

func TestSendHashed_SameMessageWIthDifferentToListShouldWaitSuccessfullyAfterSend(t *testing.T) {
	// Check that if we send the same message twice with different recipients and the second message is somehow
	// sent before the first, ensure that we check if the message was sent we wait on the correct object.
	h := NewSendRecorder(SendEntryExpiry)

	// Insert a message into the hasher.
	_, hash, ok, err := testTryInsert(h, literal1, time.Now().Add(time.Minute), "Receiver <receiver@pm.me>")
	require.NoError(t, err)
	require.True(t, ok)
	require.NotEmpty(t, hash)

	srID2, hash2, ok, err := testTryInsert(h, literal1, time.Now().Add(time.Microsecond), "Receiver <receiver@pm.me>", "Receiver2 <receiver2@pm.me>")
	require.NoError(t, err)
	require.True(t, ok)
	require.NotEmpty(t, hash2)
	require.Equal(t, hash, hash2)

	// simulate message sent
	h.SignalMessageSent(hash2, srID2, "newID")

	// Simulate Wait on message 2
	_, ok, err = h.HasEntryWait(context.Background(), hash2, time.Now().Add(time.Second), []string{"Receiver <receiver@pm.me>", "Receiver2 <receiver2@pm.me>"})
	require.NoError(t, err)
	require.True(t, ok)
}

func TestSendHasher_HasEntry_SendFail(t *testing.T) {
	h := NewSendRecorder(SendEntryExpiry)

	// Insert a message into the hasher.
	srID1, hash, ok, err := testTryInsert(h, literal1, time.Now().Add(time.Second))
	require.NoError(t, err)
	require.True(t, ok)
	require.NotEmpty(t, hash)

	// Simulate failing to send the message after half a second.
	go func() {
		time.Sleep(time.Millisecond * 500)
		h.RemoveOnFail(hash, srID1)
	}()

	// The message failed to send; we should not find it in the hasher.
	_, ok, err = testHasEntry(h, literal1, time.Now().Add(time.Second))
	require.NoError(t, err)
	require.False(t, ok)
}

func TestSendHasher_HasEntry_Timeout(t *testing.T) {
	h := NewSendRecorder(SendEntryExpiry)

	// Insert a message into the hasher.
	_, hash, ok, err := testTryInsert(h, literal1, time.Now().Add(time.Second))
	require.NoError(t, err)
	require.True(t, ok)
	require.NotEmpty(t, hash)

	// The message is never sent; we should not find it in the hasher.
	_, ok, err = testHasEntry(h, literal1, time.Now().Add(time.Second))
	require.NoError(t, err)
	require.False(t, ok)
}

func TestSendHasher_HasEntry_Expired(t *testing.T) {
	h := NewSendRecorder(time.Second)

	// Insert a message into the hasher.
	srID1, hash, ok, err := testTryInsert(h, literal1, time.Now().Add(time.Second))
	require.NoError(t, err)
	require.True(t, ok)
	require.NotEmpty(t, hash)

	// Simulate successfully sending the message.
	h.SignalMessageSent(hash, srID1, "abc")

	// Wait for the entry to expire.
	time.Sleep(time.Second)

	// The entry has expired; we should not find it in the hasher.
	_, ok, err = testHasEntry(h, literal1, time.Now().Add(time.Second))
	require.NoError(t, err)
	require.False(t, ok)
}

const literal1 = `From: Sender <sender@pm.me>
To: Receiver <receiver@pm.me>
Content-Type: multipart/mixed; boundary=longrandomstring

--longrandomstring

body
--longrandomstring
Content-Disposition: attachment; filename="attname.txt"

attachment
--longrandomstring--
`
const literal2 = `From: Sender <sender@pm.me>
To: Receiver <receiver@pm.me>
Content-Type: multipart/mixed; boundary=longrandomstring

--longrandomstring

body
--longrandomstring
Content-Disposition: attachment; filename="attname2.txt"

attachment
--longrandomstring--
`

func TestGetMessageHash(t *testing.T) {
	tests := []struct {
		name       string
		lit1, lit2 []byte
		wantEqual  bool
	}{
		{
			name:      "empty",
			lit1:      []byte{},
			lit2:      []byte{},
			wantEqual: true,
		},
		{
			name:      "same to",
			lit1:      []byte("To: someone@pm.me\r\n\r\nHello world!"),
			lit2:      []byte("To: someone@pm.me\r\n\r\nHello world!"),
			wantEqual: true,
		},
		{
			name:      "different to",
			lit1:      []byte("To: someone@pm.me\r\n\r\nHello world!"),
			lit2:      []byte("To: another@pm.me\r\n\r\nHello world!"),
			wantEqual: false,
		},
		{
			name:      "same from",
			lit1:      []byte("From: someone@pm.me\r\n\r\nHello world!"),
			lit2:      []byte("From: someone@pm.me\r\n\r\nHello world!"),
			wantEqual: true,
		},
		{
			name:      "different from",
			lit1:      []byte("From: someone@pm.me\r\n\r\nHello world!"),
			lit2:      []byte("From: another@pm.me\r\n\r\nHello world!"),
			wantEqual: false,
		},
		{
			name:      "same subject",
			lit1:      []byte("Subject: Hello world!\r\n\r\nHello world!"),
			lit2:      []byte("Subject: Hello world!\r\n\r\nHello world!"),
			wantEqual: true,
		},
		{
			name:      "different subject",
			lit1:      []byte("Subject: Hello world!\r\n\r\nHello world!"),
			lit2:      []byte("Subject: Goodbye world!\r\n\r\nHello world!"),
			wantEqual: false,
		},
		{
			name:      "same plaintext body",
			lit1:      []byte("To: someone@pm.me\r\nContent-Type: text/plain\r\n\r\nHello world!"),
			lit2:      []byte("To: someone@pm.me\r\nContent-Type: text/plain\r\n\r\nHello world!"),
			wantEqual: true,
		},
		{
			name:      "different plaintext body",
			lit1:      []byte("To: someone@pm.me\r\nContent-Type: text/plain\r\n\r\nHello world!"),
			lit2:      []byte("To: someone@pm.me\r\nContent-Type: text/plain\r\n\r\nGoodbye world!"),
			wantEqual: false,
		},
		{
			name:      "different attachment filenames",
			lit1:      []byte(literal1),
			lit2:      []byte(literal2),
			wantEqual: false,
		},
		{
			name:      "different date and message ID should still match",
			lit1:      []byte("To: a@b.c\r\nDate: Fri, 13 Aug 1982\r\nMessage-Id: 1@b.c\r\n\r\nHello"),
			lit2:      []byte("To: a@b.c\r\nDate: Sat, 14 Aug 1982\r\nMessage-Id: 2@b.c\r\n\r\nHello"),
			wantEqual: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hash1, err := GetMessageHash(tt.lit1)
			require.NoError(t, err)

			hash2, err := GetMessageHash(tt.lit2)
			require.NoError(t, err)

			if tt.wantEqual {
				require.Equal(t, hash1, hash2)
			} else {
				require.NotEqual(t, hash1, hash2)
			}
		})
	}
}

func testTryInsert(h *SendRecorder, literal string, deadline time.Time, toList ...string) (ID, string, bool, error) { //nolint:unparam
	hash, err := GetMessageHash([]byte(literal))
	if err != nil {
		return 0, "", false, err
	}

	srID, ok, err := h.TryInsertWait(context.Background(), hash, toList, deadline)
	if err != nil {
		return 0, "", false, err
	}

	return srID, hash, ok, nil
}

func testHasEntry(h *SendRecorder, literal string, deadline time.Time, toList ...string) (string, bool, error) { //nolint:unparam
	hash, err := GetMessageHash([]byte(literal))
	if err != nil {
		return "", false, err
	}

	return h.HasEntryWait(context.Background(), hash, deadline, toList)
}
