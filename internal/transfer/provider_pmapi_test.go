// Copyright (c) 2020 Proton Technologies AG
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

package transfer

import (
	"bytes"
	"fmt"
	"testing"
	"time"

	"github.com/ProtonMail/proton-bridge/pkg/pmapi"
	gomock "github.com/golang/mock/gomock"
	r "github.com/stretchr/testify/require"
)

func TestPMAPIProviderMailboxes(t *testing.T) {
	m := initMocks(t)
	defer m.ctrl.Finish()

	setupPMAPIClientExpectationForExport(&m)
	provider, err := NewPMAPIProvider(m.clientManager, "user", "addressID")
	r.NoError(t, err)

	tests := []struct {
		includeEmpty   bool
		includeAllMail bool
		wantMailboxes  []Mailbox
	}{
		{true, false, []Mailbox{
			{ID: "folder1", Name: "One", Color: "red", IsExclusive: true},
			{ID: "folder2", Name: "Two", Color: "orange", IsExclusive: true},
			{ID: "label2", Name: "Bar", Color: "green", IsExclusive: false},
			{ID: "label1", Name: "Foo", Color: "blue", IsExclusive: false},
		}},
		{false, true, []Mailbox{
			{ID: pmapi.AllMailLabel, Name: "All Mail", IsExclusive: true},
			{ID: "folder1", Name: "One", Color: "red", IsExclusive: true},
			{ID: "folder2", Name: "Two", Color: "orange", IsExclusive: true},
			{ID: "label1", Name: "Foo", Color: "blue", IsExclusive: false},
		}},
	}
	for _, tc := range tests {
		tc := tc
		t.Run(fmt.Sprintf("%v-%v", tc.includeEmpty, tc.includeAllMail), func(t *testing.T) {
			mailboxes, err := provider.Mailboxes(tc.includeEmpty, tc.includeAllMail)
			r.NoError(t, err)
			r.Equal(t, []Mailbox{
				{ID: pmapi.InboxLabel, Name: "Inbox", IsExclusive: true},
				{ID: pmapi.DraftLabel, Name: "Drafts", IsExclusive: true},
				{ID: pmapi.SentLabel, Name: "Sent", IsExclusive: true},
				{ID: pmapi.StarredLabel, Name: "Starred", IsExclusive: true},
				{ID: pmapi.ArchiveLabel, Name: "Archive", IsExclusive: true},
				{ID: pmapi.SpamLabel, Name: "Spam", IsExclusive: true},
				{ID: pmapi.TrashLabel, Name: "Trash", IsExclusive: true},
			}, mailboxes[:7])
			r.Equal(t, tc.wantMailboxes, mailboxes[7:])
		})
	}
}

func TestPMAPIProviderTransferTo(t *testing.T) {
	m := initMocks(t)
	defer m.ctrl.Finish()

	setupPMAPIClientExpectationForExport(&m)
	provider, err := NewPMAPIProvider(m.clientManager, "user", "addressID")
	r.NoError(t, err)

	rules, rulesClose := newTestRules(t)
	defer rulesClose()
	setupPMAPIRules(rules)

	testTransferTo(t, rules, provider, []string{
		"0_msg1",
		"0_msg2",
	})
}

func TestPMAPIProviderTransferFrom(t *testing.T) {
	m := initMocks(t)
	defer m.ctrl.Finish()

	setupPMAPIClientExpectationForImport(&m)
	provider, err := NewPMAPIProvider(m.clientManager, "user", "addressID")
	r.NoError(t, err)

	rules, rulesClose := newTestRules(t)
	defer rulesClose()
	setupPMAPIRules(rules)

	testTransferFrom(t, rules, provider, []Message{
		{ID: "msg1", Body: getTestMsgBody("msg1"), Targets: []Mailbox{{ID: pmapi.InboxLabel}}},
		{ID: "msg2", Body: getTestMsgBody("msg2"), Targets: []Mailbox{{ID: pmapi.InboxLabel}}},
	})
}

func TestPMAPIProviderTransferFromDraft(t *testing.T) {
	m := initMocks(t)
	defer m.ctrl.Finish()

	setupPMAPIClientExpectationForImportDraft(&m)
	provider, err := NewPMAPIProvider(m.clientManager, "user", "addressID")
	r.NoError(t, err)

	rules, rulesClose := newTestRules(t)
	defer rulesClose()
	setupPMAPIRules(rules)

	testTransferFrom(t, rules, provider, []Message{
		{ID: "draft1", Body: getTestMsgBody("draft1"), Targets: []Mailbox{{ID: pmapi.DraftLabel}}},
	})
}

func TestPMAPIProviderTransferFromTo(t *testing.T) {
	m := initMocks(t)
	defer m.ctrl.Finish()

	setupPMAPIClientExpectationForExport(&m)
	setupPMAPIClientExpectationForImport(&m)

	source, err := NewPMAPIProvider(m.clientManager, "user", "addressID")
	r.NoError(t, err)
	target, err := NewPMAPIProvider(m.clientManager, "user", "addressID")
	r.NoError(t, err)

	rules, rulesClose := newTestRules(t)
	defer rulesClose()
	setupPMAPIRules(rules)

	testTransferFromTo(t, rules, source, target, 5*time.Second)
}

func setupPMAPIRules(rules transferRules) {
	_ = rules.setRule(Mailbox{ID: pmapi.InboxLabel}, []Mailbox{{ID: pmapi.InboxLabel}}, 0, 0)
}

func setupPMAPIClientExpectationForExport(m *mocks) {
	m.pmapiClient.EXPECT().KeyRingForAddressID(gomock.Any()).Return(m.keyring, nil).AnyTimes()
	m.pmapiClient.EXPECT().ListLabels().Return([]*pmapi.Label{
		{ID: "label1", Name: "Foo", Color: "blue", Exclusive: 0, Order: 2},
		{ID: "label2", Name: "Bar", Color: "green", Exclusive: 0, Order: 1},
		{ID: "folder1", Name: "One", Color: "red", Exclusive: 1, Order: 1},
		{ID: "folder2", Name: "Two", Color: "orange", Exclusive: 1, Order: 2},
	}, nil).AnyTimes()
	m.pmapiClient.EXPECT().CountMessages(gomock.Any()).Return([]*pmapi.MessagesCount{
		{LabelID: "label1", Total: 10},
		{LabelID: "label2", Total: 0},
		{LabelID: "folder1", Total: 20},
	}, nil).AnyTimes()
	m.pmapiClient.EXPECT().ListMessages(gomock.Any()).Return([]*pmapi.Message{
		{ID: "msg1"},
		{ID: "msg2"},
	}, 2, nil).AnyTimes()
	m.pmapiClient.EXPECT().GetMessage(gomock.Any()).DoAndReturn(func(msgID string) (*pmapi.Message, error) {
		return &pmapi.Message{
			ID:       msgID,
			Body:     string(getTestMsgBody(msgID)),
			MIMEType: pmapi.ContentTypeMultipartMixed,
		}, nil
	}).AnyTimes()
}

func setupPMAPIClientExpectationForImport(m *mocks) {
	m.pmapiClient.EXPECT().KeyRingForAddressID(gomock.Any()).Return(m.keyring, nil).AnyTimes()
	m.pmapiClient.EXPECT().Import(gomock.Any()).DoAndReturn(func(requests []*pmapi.ImportMsgReq) ([]*pmapi.ImportMsgRes, error) {
		r.Equal(m.t, 1, len(requests))

		request := requests[0]
		for _, msgID := range []string{"msg1", "msg2"} {
			if bytes.Contains(request.Body, []byte(msgID)) {
				return []*pmapi.ImportMsgRes{{MessageID: msgID, Error: nil}}, nil
			}
		}
		r.Fail(m.t, "No message found")
		return nil, nil
	}).Times(2)
}

func setupPMAPIClientExpectationForImportDraft(m *mocks) {
	m.pmapiClient.EXPECT().KeyRingForAddressID(gomock.Any()).Return(m.keyring, nil).AnyTimes()
	m.pmapiClient.EXPECT().CreateDraft(gomock.Any(), gomock.Any(), gomock.Any()).DoAndReturn(func(msg *pmapi.Message, parentID string, action int) (*pmapi.Message, error) {
		r.Equal(m.t, msg.Subject, "draft1")
		msg.ID = "draft1"
		return msg, nil
	})
}
