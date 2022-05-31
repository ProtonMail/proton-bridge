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
	"fmt"

	"github.com/ProtonMail/proton-bridge/v2/pkg/pmapi"
)

// ListMailboxes returns all mailboxes.
func (storeAddress *Address) ListMailboxes() []*Mailbox {
	storeAddress.store.lock.RLock()
	defer storeAddress.store.lock.RUnlock()

	mailboxes := make([]*Mailbox, 0, len(storeAddress.mailboxes))
	for _, m := range storeAddress.mailboxes {
		mailboxes = append(mailboxes, m)
	}
	return mailboxes
}

// GetMailbox returns mailbox with the given IMAP name.
func (storeAddress *Address) GetMailbox(name string) (*Mailbox, error) {
	storeAddress.store.lock.RLock()
	defer storeAddress.store.lock.RUnlock()

	for _, m := range storeAddress.mailboxes {
		if m.Name() == name {
			return m, nil
		}
	}

	return nil, fmt.Errorf("mailbox %v does not exist", name)
}

// CreateMailbox creates the mailbox by calling an API.
// Mailbox is created in the structure by processing event.
func (storeAddress *Address) CreateMailbox(name string) error {
	return storeAddress.store.createMailbox(name)
}

// updateMailbox updates the mailbox by calling an API.
// Mailbox is updated in the structure by processing event.
func (storeAddress *Address) updateMailbox(labelID, newName, color string) error {
	return storeAddress.store.updateMailbox(labelID, newName, color)
}

// deleteMailbox deletes the mailbox by calling an API.
// Mailbox is deleted in the structure by processing event.
func (storeAddress *Address) deleteMailbox(labelID string) error {
	return storeAddress.store.deleteMailbox(labelID, storeAddress.addressID)
}

// createOrUpdateMailboxEvent creates or updates the mailbox in the structure.
// This is called from the event loop.
func (storeAddress *Address) createOrUpdateMailboxEvent(label *pmapi.Label) error {
	prefix := getLabelPrefix(label)
	mailbox, ok := storeAddress.mailboxes[label.ID]
	if !ok {
		mailbox, err := newMailbox(storeAddress, label.ID, prefix, label.Path, label.Color)
		if err != nil {
			return err
		}
		storeAddress.mailboxes[label.ID] = mailbox
		mailbox.store.notifyMailboxCreated(storeAddress.address, mailbox.labelName)
	} else {
		mailbox.labelName = prefix + label.Path
		mailbox.color = label.Color
	}
	return nil
}

// deleteMailboxEvent deletes the mailbox in the structure.
// This is called from the event loop.
func (storeAddress *Address) deleteMailboxEvent(labelID string) error {
	storeMailbox, ok := storeAddress.mailboxes[labelID]
	if !ok {
		log.WithField("labelID", labelID).Warn("Could not find mailbox to delete")
		return nil
	}
	delete(storeAddress.mailboxes, labelID)
	return storeMailbox.deleteMailboxEvent()
}

func (storeAddress *Address) getMailboxByID(labelID string) (*Mailbox, error) {
	storeMailbox, ok := storeAddress.mailboxes[labelID]
	if !ok {
		return nil, fmt.Errorf("mailbox with id %q does not exist", labelID)
	}
	return storeMailbox, nil
}
