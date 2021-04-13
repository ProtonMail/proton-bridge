// Copyright (c) 2021 Proton Technologies AG
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
	"io"
	"net/mail"
	"strings"
	"testing"

	"github.com/ProtonMail/proton-bridge/pkg/pmapi"
	a "github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetAllMessageIDs(t *testing.T) {
	m, clear := initMocks(t)
	defer clear()

	m.newStoreNoEvents(true)

	insertMessage(t, m, "msg1", "Test message 1", addrID1, 0, []string{pmapi.AllMailLabel, pmapi.InboxLabel})
	insertMessage(t, m, "msg2", "Test message 2", addrID1, 0, []string{pmapi.AllMailLabel, pmapi.ArchiveLabel})
	insertMessage(t, m, "msg3", "Test message 3", addrID1, 0, []string{pmapi.AllMailLabel, pmapi.InboxLabel})
	insertMessage(t, m, "msg4", "Test message 4", addrID1, 0, []string{})

	checkAllMessageIDs(t, m, []string{"msg1", "msg2", "msg3", "msg4"})
}

func TestGetMessageFromDB(t *testing.T) {
	m, clear := initMocks(t)
	defer clear()

	m.newStoreNoEvents(true)
	insertMessage(t, m, "msg1", "Test message 1", addrID1, 0, []string{pmapi.AllMailLabel})

	tests := []struct{ msgID, wantErr string }{
		{"msg1", ""},
		{"msg2", "no such api id"},
	}
	for _, tc := range tests {
		tc := tc
		t.Run(tc.msgID, func(t *testing.T) {
			msg, err := m.store.getMessageFromDB(tc.msgID)
			if tc.wantErr != "" {
				require.EqualError(t, err, tc.wantErr)
			} else {
				require.Nil(t, err)
				require.Equal(t, tc.msgID, msg.ID)
			}
		})
	}
}

func TestCreateOrUpdateMessageMetadata(t *testing.T) {
	m, clear := initMocks(t)
	defer clear()

	m.newStoreNoEvents(true)
	insertMessage(t, m, "msg1", "Test message 1", addrID1, 0, []string{pmapi.AllMailLabel})

	msg, err := m.store.getMessageFromDB("msg1")
	require.Nil(t, err)

	// Check non-meta and calculated data are cleared/empty.
	a.Equal(t, "", msg.Body)
	a.Equal(t, []*pmapi.Attachment(nil), msg.Attachments)
	a.Equal(t, int64(-1), msg.Size)
	a.Equal(t, "", msg.MIMEType)
	a.Equal(t, make(mail.Header), msg.Header)

	// Change the calculated data.
	wantSize := int64(42)
	wantMIMEType := "plain-text"
	wantHeader := mail.Header{
		"Key": []string{"value"},
	}

	storeMsg, err := m.store.addresses[addrID1].mailboxes[pmapi.AllMailLabel].GetMessage("msg1")
	require.Nil(t, err)
	require.Nil(t, storeMsg.SetSize(wantSize))
	require.Nil(t, storeMsg.SetContentTypeAndHeader(wantMIMEType, wantHeader))

	// Check calculated data.
	msg, err = m.store.getMessageFromDB("msg1")
	require.Nil(t, err)
	a.Equal(t, wantSize, msg.Size)
	a.Equal(t, wantMIMEType, msg.MIMEType)
	a.Equal(t, wantHeader, msg.Header)

	// Check calculated data are not overridden by reinsert.
	insertMessage(t, m, "msg1", "Test message 1", addrID1, 0, []string{pmapi.AllMailLabel})

	msg, err = m.store.getMessageFromDB("msg1")
	require.Nil(t, err)
	a.Equal(t, wantSize, msg.Size)
	a.Equal(t, wantMIMEType, msg.MIMEType)
	a.Equal(t, wantHeader, msg.Header)
}

func TestDeleteMessage(t *testing.T) {
	m, clear := initMocks(t)
	defer clear()

	m.newStoreNoEvents(true)
	insertMessage(t, m, "msg1", "Test message 1", addrID1, 0, []string{pmapi.AllMailLabel})
	insertMessage(t, m, "msg2", "Test message 2", addrID1, 0, []string{pmapi.AllMailLabel})

	require.Nil(t, m.store.deleteMessageEvent("msg1"))

	checkAllMessageIDs(t, m, []string{"msg2"})
	checkMailboxMessageIDs(t, m, pmapi.AllMailLabel, []wantID{{"msg2", 2}})
}

func insertMessage(t *testing.T, m *mocksForStore, id, subject, sender string, unread int, labelIDs []string) { //nolint[unparam]
	msg := getTestMessage(id, subject, sender, unread, labelIDs)
	require.Nil(t, m.store.createOrUpdateMessageEvent(msg))
}

func getTestMessage(id, subject, sender string, unread int, labelIDs []string) *pmapi.Message {
	address := &mail.Address{Address: sender}
	return &pmapi.Message{
		ID:       id,
		Subject:  subject,
		Unread:   unread,
		Sender:   address,
		ToList:   []*mail.Address{address},
		LabelIDs: labelIDs,
		Size:     12345,
		Body:     "body of message",
		Attachments: []*pmapi.Attachment{{
			ID:        "attachment1",
			MessageID: id,
			Name:      "attachment",
		}},
	}
}

func checkAllMessageIDs(t *testing.T, m *mocksForStore, wantIDs []string) {
	allIds, allErr := m.store.getAllMessageIDs()
	require.Nil(t, allErr)
	require.Equal(t, wantIDs, allIds)
}

func TestCreateDraftCheckMessageSize(t *testing.T) {
	m, clear := initMocks(t)
	defer clear()

	m.newStoreNoEvents(false)
	m.client.EXPECT().CurrentUser().Return(&pmapi.User{
		MaxUpload: 100, // Decrypted message 5 chars, encrypted 500+.
	}, nil)

	// Even small body is bloated to at least about 500 chars of basic pgp message.
	message := &pmapi.Message{
		Body: strings.Repeat("a", 5),
	}
	attachmentReaders := []io.Reader{}
	_, _, err := m.store.CreateDraft(testPrivateKeyRing, message, attachmentReaders, "", "", "")

	require.EqualError(t, err, "message is too large")
}

func TestCreateDraftCheckMessageWithAttachmentSize(t *testing.T) {
	m, clear := initMocks(t)
	defer clear()

	m.newStoreNoEvents(false)
	m.client.EXPECT().CurrentUser().Return(&pmapi.User{
		MaxUpload: 800, // Decrypted message 5 chars + 5 chars of attachment, encrypted 500+ + 300+.
	}, nil)

	// Even small body is bloated to at least about 500 chars of basic pgp message.
	message := &pmapi.Message{
		Body: strings.Repeat("a", 5),
		Attachments: []*pmapi.Attachment{
			{Name: "name"},
		},
	}
	// Even small attachment is bloated to about 300 chars of encrypted text.
	attachmentReaders := []io.Reader{
		strings.NewReader(strings.Repeat("b", 5)),
	}
	_, _, err := m.store.CreateDraft(testPrivateKeyRing, message, attachmentReaders, "", "", "")

	require.EqualError(t, err, "message is too large")
}
