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
	"errors"
	"fmt"
	"net/mail"
	"testing"
	"time"

	"github.com/ProtonMail/proton-bridge/v2/pkg/pmapi"
	"github.com/stretchr/testify/assert"
)

type testSendRecorderGetMessageMock struct {
	message *pmapi.Message
	err     error
}

func (m *testSendRecorderGetMessageMock) GetMessage(_ context.Context, messageID string) (*pmapi.Message, error) {
	return m.message, m.err
}

func TestSendRecorder_getMessageHash(t *testing.T) {
	q := newSendRecorder()

	message := &pmapi.Message{
		AddressID: "address123",
		Subject:   "Subject #1",
		Sender: &mail.Address{
			Address: "from@pm.me",
		},
		ToList: []*mail.Address{
			{Address: "to@pm.me"},
		},
		CCList:  []*mail.Address{},
		BCCList: []*mail.Address{},
		Body:    "body",
		Attachments: []*pmapi.Attachment{
			{
				Name:     "att1",
				MIMEType: "image/png",
				Size:     12345,
			},
		},
	}
	hash := q.getMessageHash(message)

	testCases := []struct {
		message     *pmapi.Message
		expectEqual bool
	}{
		{
			message,
			true,
		},
		{
			&pmapi.Message{},
			false,
		},
		{ // Different AddressID
			&pmapi.Message{
				AddressID: "...",
				Subject:   "Subject #1",
				Sender: &mail.Address{
					Address: "from@pm.me",
				},
				ToList: []*mail.Address{
					{Address: "to@pm.me"},
				},
				CCList:  []*mail.Address{},
				BCCList: []*mail.Address{},
				Body:    "body",
				Attachments: []*pmapi.Attachment{
					{
						Name:     "att1",
						MIMEType: "image/png",
						Size:     12345,
					},
				},
			},
			false,
		},
		{ // Different subject
			&pmapi.Message{
				AddressID: "address123",
				Subject:   "Subject #1.",
				Sender: &mail.Address{
					Address: "from@pm.me",
				},
				ToList: []*mail.Address{
					{Address: "to@pm.me"},
				},
				CCList:  []*mail.Address{},
				BCCList: []*mail.Address{},
				Body:    "body",
				Attachments: []*pmapi.Attachment{
					{
						Name:     "att1",
						MIMEType: "image/png",
						Size:     12345,
					},
				},
			},
			false,
		},
		{ // Different sender
			&pmapi.Message{
				AddressID: "address123",
				Subject:   "Subject #1",
				Sender: &mail.Address{
					Address: "sender@pm.me",
				},
				ToList: []*mail.Address{
					{Address: "to@pm.me"},
				},
				CCList:  []*mail.Address{},
				BCCList: []*mail.Address{},
				Body:    "body",
				Attachments: []*pmapi.Attachment{
					{
						Name:     "att1",
						MIMEType: "image/png",
						Size:     12345,
					},
				},
			},
			false,
		},
		{ // Different ToList - changed address
			&pmapi.Message{
				AddressID: "address123",
				Subject:   "Subject #1",
				Sender: &mail.Address{
					Address: "from@pm.me",
				},
				ToList: []*mail.Address{
					{Address: "other@pm.me"},
				},
				CCList:  []*mail.Address{},
				BCCList: []*mail.Address{},
				Body:    "body",
				Attachments: []*pmapi.Attachment{
					{
						Name:     "att1",
						MIMEType: "image/png",
						Size:     12345,
					},
				},
			},
			false,
		},
		{ // Different ToList - more addresses
			&pmapi.Message{
				AddressID: "address123",
				Subject:   "Subject #1",
				Sender: &mail.Address{
					Address: "from@pm.me",
				},
				ToList: []*mail.Address{
					{Address: "to@pm.me"},
					{Address: "another@pm.me"},
				},
				CCList:  []*mail.Address{},
				BCCList: []*mail.Address{},
				Body:    "body",
				Attachments: []*pmapi.Attachment{
					{
						Name:     "att1",
						MIMEType: "image/png",
						Size:     12345,
					},
				},
			},
			false,
		},
		{ // Different CCList
			&pmapi.Message{
				AddressID: "address123",
				Subject:   "Subject #1",
				Sender: &mail.Address{
					Address: "from@pm.me",
				},
				ToList: []*mail.Address{
					{Address: "to@pm.me"},
				},
				CCList: []*mail.Address{
					{Address: "to@pm.me"},
				},
				BCCList: []*mail.Address{},
				Body:    "body",
				Attachments: []*pmapi.Attachment{
					{
						Name:     "att1",
						MIMEType: "image/png",
						Size:     12345,
					},
				},
			},
			false,
		},
		{ // Different BCCList
			&pmapi.Message{
				AddressID: "address123",
				Subject:   "Subject #1",
				Sender: &mail.Address{
					Address: "from@pm.me",
				},
				ToList: []*mail.Address{
					{Address: "to@pm.me"},
				},
				CCList: []*mail.Address{},
				BCCList: []*mail.Address{
					{Address: "to@pm.me"},
				},
				Body: "body",
				Attachments: []*pmapi.Attachment{
					{
						Name:     "att1",
						MIMEType: "image/png",
						Size:     12345,
					},
				},
			},
			false,
		},
		{ // Different body
			&pmapi.Message{
				AddressID: "address123",
				Subject:   "Subject #1",
				Sender: &mail.Address{
					Address: "from@pm.me",
				},
				ToList: []*mail.Address{
					{Address: "to@pm.me"},
				},
				CCList:  []*mail.Address{},
				BCCList: []*mail.Address{},
				Body:    "body.",
				Attachments: []*pmapi.Attachment{
					{
						Name:     "att1",
						MIMEType: "image/png",
						Size:     12345,
					},
				},
			},
			false,
		},
		{ // Different attachment - no attachment
			&pmapi.Message{
				AddressID: "address123",
				Subject:   "Subject #1",
				Sender: &mail.Address{
					Address: "from@pm.me",
				},
				ToList: []*mail.Address{
					{Address: "to@pm.me"},
				},
				CCList:      []*mail.Address{},
				BCCList:     []*mail.Address{},
				Body:        "body",
				Attachments: []*pmapi.Attachment{},
			},
			false,
		},
		{ // Different attachment - name
			&pmapi.Message{
				AddressID: "address123",
				Subject:   "Subject #1",
				Sender: &mail.Address{
					Address: "from@pm.me",
				},
				ToList: []*mail.Address{
					{Address: "to@pm.me"},
				},
				CCList:  []*mail.Address{},
				BCCList: []*mail.Address{},
				Body:    "body",
				Attachments: []*pmapi.Attachment{
					{
						Name:     "...",
						MIMEType: "image/png",
						Size:     12345,
					},
				},
			},
			false,
		},
		{ // Different attachment - MIMEType
			&pmapi.Message{
				AddressID: "address123",
				Subject:   "Subject #1",
				Sender: &mail.Address{
					Address: "from@pm.me",
				},
				ToList: []*mail.Address{
					{Address: "to@pm.me"},
				},
				CCList:  []*mail.Address{},
				BCCList: []*mail.Address{},
				Body:    "body",
				Attachments: []*pmapi.Attachment{
					{
						Name:     "att1",
						MIMEType: "image/jpeg",
						Size:     12345,
					},
				},
			},
			false,
		},
		{ // Different attachment - Size
			&pmapi.Message{
				AddressID: "address123",
				Subject:   "Subject #1",
				Sender: &mail.Address{
					Address: "from@pm.me",
				},
				ToList: []*mail.Address{
					{Address: "to@pm.me"},
				},
				CCList:  []*mail.Address{},
				BCCList: []*mail.Address{},
				Body:    "body",
				Attachments: []*pmapi.Attachment{
					{
						Name:     "att1",
						MIMEType: "image/png",
						Size:     42,
					},
				},
			},
			false,
		},
		{ // Different content type - calendar
			&pmapi.Message{
				Header: mail.Header{
					"Content-Type": []string{"text/calendar"},
				},
				AddressID: "address123",
				Subject:   "Subject #1",
				Sender: &mail.Address{
					Address: "from@pm.me",
				},
				ToList: []*mail.Address{
					{Address: "to@pm.me"},
				},
				CCList:  []*mail.Address{},
				BCCList: []*mail.Address{},
				Body:    "body",
				Attachments: []*pmapi.Attachment{
					{
						Name:     "att1",
						MIMEType: "image/png",
						Size:     12345,
					},
				},
			},
			false,
		},
	}
	for i, tc := range testCases {
		tc := tc // bind
		t.Run(fmt.Sprintf("%d / %v", i, tc.message), func(t *testing.T) {
			newHash := q.getMessageHash(tc.message)
			if tc.expectEqual {
				assert.Equal(t, hash, newHash)
			} else {
				assert.NotEqual(t, hash, newHash)
			}
		})
	}
}

func TestSendRecorder_isSendingOrSent(t *testing.T) {
	q := newSendRecorder()
	q.addMessage("hash")
	q.setMessageID("hash", "messageID")

	draftFlag := pmapi.FlagInternal | pmapi.FlagE2E
	selfSent := pmapi.FlagSent | pmapi.FlagReceived

	testCases := []struct {
		hash          string
		message       *pmapi.Message
		err           error
		wantIsSending bool
		wantWasSent   bool
	}{
		{"badhash", &pmapi.Message{Flags: draftFlag}, nil, false, false},
		{"hash", nil, errors.New("message not found"), false, false},
		{"hash", &pmapi.Message{Flags: pmapi.FlagReceived}, nil, false, false},
		{"hash", &pmapi.Message{Flags: draftFlag, Time: time.Now().Add(-20 * time.Minute).Unix()}, nil, false, false},
		{"hash", &pmapi.Message{Flags: draftFlag, Time: time.Now().Unix()}, nil, true, false},
		{"hash", &pmapi.Message{Flags: pmapi.FlagSent}, nil, false, true},
		{"hash", &pmapi.Message{Flags: selfSent}, nil, false, true},
		{"", &pmapi.Message{Flags: selfSent}, nil, false, false},
	}
	for i, tc := range testCases {
		tc := tc // bind
		t.Run(fmt.Sprintf("%d / %v / %v / %v", i, tc.hash, tc.message, tc.err), func(t *testing.T) {
			messageGetter := &testSendRecorderGetMessageMock{message: tc.message, err: tc.err}
			isSending, wasSent := q.isSendingOrSent(messageGetter, tc.hash)
			assert.Equal(t, tc.wantIsSending, isSending, "isSending does not match")
			assert.Equal(t, tc.wantWasSent, wasSent, "wasSent does not match")
		})
	}
}

func TestSendRecorder_deleteExpiredKeys(t *testing.T) {
	q := newSendRecorder()

	q.hashes["hash1"] = sendRecorderValue{
		messageID: "msg1",
		time:      time.Now(),
	}
	q.hashes["hash2"] = sendRecorderValue{
		messageID: "msg2",
		time:      time.Now().Add(-31 * time.Minute),
	}

	q.deleteExpiredKeys()

	_, ok := q.hashes["hash1"]
	assert.True(t, ok)
	_, ok = q.hashes["hash2"]
	assert.False(t, ok)
}
