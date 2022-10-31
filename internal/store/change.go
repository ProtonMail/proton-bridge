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
	"github.com/ProtonMail/proton-bridge/v2/pkg/pmapi"
)

type ChangeNotifier interface {
	Notice(address, notice string)
	UpdateMessage(
		address, mailboxName string,
		uid, sequenceNumber uint32,
		msg *pmapi.Message, hasDeletedFlag bool)
	DeleteMessage(address, mailboxName string, sequenceNumber uint32)
	MailboxCreated(address, mailboxName string)
	MailboxStatus(address, mailboxName string, total, unread, unreadSeqNum uint32)

	CanDelete(mailboxID string) (bool, func())
}

// SetChangeNotifier sets notifier to be called once mailbox or message changes.
func (store *Store) SetChangeNotifier(notifier ChangeNotifier) {
	store.notifier = notifier
}

func (store *Store) notifyNotice(address, notice string) {
	if store.notifier == nil {
		return
	}
	store.notifier.Notice(address, notice)
}

func (store *Store) notifyUpdateMessage(address, mailboxName string, uid, sequenceNumber uint32, msg *pmapi.Message, hasDeletedFlag bool) {
	if store.notifier == nil {
		return
	}
	store.notifier.UpdateMessage(address, mailboxName, uid, sequenceNumber, msg, hasDeletedFlag)
}

func (store *Store) notifyDeleteMessage(address, mailboxName string, sequenceNumber uint32) {
	if store.notifier == nil {
		return
	}
	store.notifier.DeleteMessage(address, mailboxName, sequenceNumber)
}

func (store *Store) notifyMailboxCreated(address, mailboxName string) {
	if store.notifier == nil {
		return
	}
	store.notifier.MailboxCreated(address, mailboxName)
}

func (store *Store) notifyMailboxStatus(address, mailboxName string, total, unread, unreadSeqNum uint) {
	if store.notifier == nil {
		return
	}
	store.notifier.MailboxStatus(address, mailboxName, uint32(total), uint32(unread), uint32(unreadSeqNum))
}
