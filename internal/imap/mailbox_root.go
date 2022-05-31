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

package imap

import (
	"errors"
	"time"

	"github.com/ProtonMail/proton-bridge/v2/internal/store"

	imap "github.com/emersion/go-imap"
)

// The mailbox containing all custom folders or labels.
// The purpose of this mailbox is to see "Folders" and "Labels"
// at the root of the mailbox tree, e.g.:
//
// 		Folders 					<< this
//			Folders/Family
//
//		Labels						<< this
//			Labels/Security
//
// This mailbox cannot be modified or read in any way.
type imapRootMailbox struct {
	isFolder bool
}

func newFoldersRootMailbox() *imapRootMailbox {
	return &imapRootMailbox{isFolder: true}
}

func newLabelsRootMailbox() *imapRootMailbox {
	return &imapRootMailbox{isFolder: false}
}

func (m *imapRootMailbox) Name() string {
	if m.isFolder {
		return store.UserFoldersMailboxName
	}
	return store.UserLabelsMailboxName
}

func (m *imapRootMailbox) Info() (info *imap.MailboxInfo, err error) {
	info = &imap.MailboxInfo{
		Attributes: []string{imap.NoSelectAttr},
		Delimiter:  store.PathDelimiter,
	}

	if m.isFolder {
		info.Name = store.UserFoldersMailboxName
	} else {
		info.Name = store.UserLabelsMailboxName
	}

	return
}

func (m *imapRootMailbox) Status(_ []imap.StatusItem) (*imap.MailboxStatus, error) {
	status := &imap.MailboxStatus{}
	if m.isFolder {
		status.Name = store.UserFoldersMailboxName
	} else {
		status.Name = store.UserLabelsMailboxName
	}
	return status, nil
}

func (m *imapRootMailbox) SetSubscribed(_ bool) error {
	return errors.New("cannot subscribe or unsubsribe to Labels or Folders mailboxes")
}

func (m *imapRootMailbox) Check() error {
	return nil
}

func (m *imapRootMailbox) ListMessages(uid bool, seqset *imap.SeqSet, items []imap.FetchItem, ch chan<- *imap.Message) error {
	close(ch)
	return nil
}

func (m *imapRootMailbox) SearchMessages(uid bool, criteria *imap.SearchCriteria) (ids []uint32, err error) {
	return
}

func (m *imapRootMailbox) CreateMessage(flags []string, t time.Time, body imap.Literal) error {
	return errors.New("cannot create a message in this mailbox")
}

func (m *imapRootMailbox) UpdateMessagesFlags(uid bool, seqset *imap.SeqSet, op imap.FlagsOp, flags []string) (err error) {
	return errors.New("cannot update message flags in this mailbox")
}

func (m *imapRootMailbox) CopyMessages(uid bool, seqset *imap.SeqSet, dest string) error {
	return nil
}

// Expunge is not used by Bridge. We delete the message once it is flagged as \Deleted.
func (m *imapRootMailbox) Expunge() error {
	return nil
}
