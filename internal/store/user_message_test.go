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
	"io"
	"net/mail"
	"net/textproto"
	"strings"
	"testing"

	pkgMsg "github.com/ProtonMail/proton-bridge/v2/pkg/message"
	"github.com/ProtonMail/proton-bridge/v2/pkg/pmapi"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/require"
	bolt "go.etcd.io/bbolt"
)

func TestGetAllMessageIDs(t *testing.T) {
	m, clear := initMocks(t)
	defer clear()

	m.newStoreNoEvents(t, true)

	insertMessage(t, m, "msg1", "Test message 1", addrID1, false, []string{pmapi.AllMailLabel, pmapi.InboxLabel})
	insertMessage(t, m, "msg2", "Test message 2", addrID1, false, []string{pmapi.AllMailLabel, pmapi.ArchiveLabel})
	insertMessage(t, m, "msg3", "Test message 3", addrID1, false, []string{pmapi.AllMailLabel, pmapi.InboxLabel})
	insertMessage(t, m, "msg4", "Test message 4", addrID1, false, []string{})

	checkAllMessageIDs(t, m, []string{"msg1", "msg2", "msg3", "msg4"})
}

func TestGetMessageFromDB(t *testing.T) {
	m, clear := initMocks(t)
	defer clear()

	m.newStoreNoEvents(t, true)
	insertMessage(t, m, "msg1", "Test message 1", addrID1, false, []string{pmapi.AllMailLabel})

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
	r := require.New(t)
	m, clear := initMocks(t)
	defer clear()

	m.newStoreNoEvents(t, true)
	insertMessage(t, m, "msg1", "Test message 1", addrID1, false, []string{pmapi.AllMailLabel})

	metadata, err := m.store.getMessageFromDB("msg1")
	r.NoError(err)

	msg := &Message{msg: metadata, store: m.store, storeMailbox: nil}

	// Check non-meta and calculated data are cleared/empty.
	r.Equal("", metadata.Body)
	r.Equal([]*pmapi.Attachment(nil), metadata.Attachments)
	r.Equal("", metadata.MIMEType)
	r.Equal(make(mail.Header), metadata.Header)

	wantHeader, wantSize := putBodystructureAndSizeToDB(m, "msg1")

	// Check cached data.
	haveHeader, err := msg.GetMIMEHeader()
	r.NoError(err)
	r.Equal(wantHeader, haveHeader)

	haveSize, err := msg.GetRFC822Size()
	r.NoError(err)
	r.Equal(wantSize, haveSize)

	// Check cached data are not overridden by reinsert.
	insertMessage(t, m, "msg1", "Test message 1", addrID1, false, []string{pmapi.AllMailLabel})

	haveHeader, err = msg.GetMIMEHeader()
	r.NoError(err)
	r.Equal(wantHeader, haveHeader)

	haveSize, err = msg.GetRFC822Size()
	r.NoError(err)
	r.Equal(wantSize, haveSize)
}

func TestDeleteMessage(t *testing.T) {
	m, clear := initMocks(t)
	defer clear()

	m.newStoreNoEvents(t, true)
	insertMessage(t, m, "msg1", "Test message 1", addrID1, false, []string{pmapi.AllMailLabel})
	insertMessage(t, m, "msg2", "Test message 2", addrID1, false, []string{pmapi.AllMailLabel})

	require.Nil(t, m.store.deleteMessageEvent("msg1"))

	checkAllMessageIDs(t, m, []string{"msg2"})
	checkMailboxMessageIDs(t, m, pmapi.AllMailLabel, []wantID{{"msg2", 2}})
}

func insertMessage(t *testing.T, m *mocksForStore, id, subject, sender string, unread bool, labelIDs []string) { //nolint:unparam
	require.Nil(t, m.store.createOrUpdateMessageEvent(getTestMessage(id, subject, sender, unread, labelIDs)))
}

func getTestMessage(id, subject, sender string, unread bool, labelIDs []string) *pmapi.Message {
	address := &mail.Address{Address: sender}
	return &pmapi.Message{
		ID:       id,
		Subject:  subject,
		Unread:   pmapi.Boolean(unread),
		Sender:   address,
		Flags:    pmapi.FlagReceived,
		ToList:   []*mail.Address{address},
		LabelIDs: labelIDs,
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

	m.newStoreNoEvents(t, false)
	m.client.EXPECT().CurrentUser(gomock.Any()).Return(&pmapi.User{
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

	m.newStoreNoEvents(t, false)
	m.client.EXPECT().CurrentUser(gomock.Any()).Return(&pmapi.User{
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

func putBodystructureAndSizeToDB(m *mocksForStore, msgID string) (header textproto.MIMEHeader, size uint32) {
	size = uint32(42)

	require.NoError(m.tb, m.store.db.Update(func(tx *bolt.Tx) error {
		return tx.Bucket(sizeBucket).Put([]byte(msgID), itob(size))
	}))

	header = textproto.MIMEHeader{
		"Key": []string{"value"},
	}

	bs := pkgMsg.BodyStructure{
		"": &pkgMsg.SectionInfo{
			Header: []byte("Key: value\r\n\r\n"),
			Start:  0,
			BSize:  int(size - 11),
			Size:   int(size),
			Lines:  3,
		},
	}

	raw, err := bs.Serialize()
	require.NoError(m.tb, err)

	require.NoError(m.tb, m.store.db.Update(func(tx *bolt.Tx) error {
		return tx.Bucket(bodystructureBucket).Put([]byte(msgID), raw)
	}))

	return header, size
}
