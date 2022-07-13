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
	"strings"

	"github.com/ProtonMail/proton-bridge/v2/pkg/pmapi"
	"github.com/pkg/errors"
)

// createMailbox creates the mailbox via the API.
// The store mailbox is created later by processing an event.
func (store *Store) createMailbox(name string) error {
	defer store.eventLoop.pollNow()

	log.WithField("name", name).Debug("Creating mailbox")

	if store.hasMailbox(name) {
		return fmt.Errorf("mailbox %v already exists", name)
	}

	color := store.leastUsedColor()

	var exclusive bool
	switch {
	case strings.HasPrefix(name, UserLabelsPrefix):
		name = strings.TrimPrefix(name, UserLabelsPrefix)
		exclusive = false
	case strings.HasPrefix(name, UserFoldersPrefix):
		name = strings.TrimPrefix(name, UserFoldersPrefix)
		exclusive = true
	default:
		// Ideally we would throw an error here, but then Outlook for
		// macOS keeps trying to make an IMAP Drafts folder and popping
		// up the error to the user.
		store.log.WithField("name", name).
			Warn("Ignoring creation of new mailbox in IMAP root")
		return nil
	}

	_, err := store.client().CreateLabel(exposeContextForIMAP(), &pmapi.Label{
		Name:      name,
		Color:     color,
		Exclusive: pmapi.Boolean(exclusive),
		Type:      pmapi.LabelTypeMailBox,
	})
	return err
}

// allAddressesHaveMailbox returns whether each address has a mailbox with the given labelID.
func (store *Store) allAddressesHaveMailbox(labelID string) bool {
	store.lock.RLock()
	defer store.lock.RUnlock()

	for _, a := range store.addresses {
		addressHasMailbox := false
		for _, m := range a.mailboxes {
			if m.labelID == labelID {
				addressHasMailbox = true
				break
			}
		}
		if !addressHasMailbox {
			return false
		}
	}
	return true
}

// hasMailbox returns whether there is at least one address which has a mailbox with the given name.
func (store *Store) hasMailbox(name string) bool {
	mailbox, _ := store.getMailbox(name)
	return mailbox != nil
}

// getMailbox returns the first mailbox with the given name.
func (store *Store) getMailbox(name string) (*Mailbox, error) {
	store.lock.RLock()
	defer store.lock.RUnlock()

	for _, a := range store.addresses {
		for _, m := range a.mailboxes {
			if m.labelName == name {
				return m, nil
			}
		}
	}
	return nil, fmt.Errorf("mailbox %s does not exist", name)
}

// leastUsedColor returns the least used color to be used for a newly created folder or label.
func (store *Store) leastUsedColor() string {
	store.lock.RLock()
	defer store.lock.RUnlock()

	colors := []string{}
	for _, a := range store.addresses {
		for _, m := range a.mailboxes {
			colors = append(colors, m.color)
		}
	}

	return pmapi.LeastUsedColor(colors)
}

// updateMailbox updates the mailbox via the API.
// The store mailbox is updated later by processing an event.
func (store *Store) updateMailbox(labelID, newName, color string) error {
	defer store.eventLoop.pollNow()

	_, err := store.client().UpdateLabel(exposeContextForIMAP(), &pmapi.Label{
		ID:    labelID,
		Name:  newName,
		Color: color,
	})
	return err
}

// deleteMailbox deletes the mailbox via the API.
// The store mailbox is deleted later by processing an event.
func (store *Store) deleteMailbox(labelID, addressID string) error {
	defer store.eventLoop.pollNow()

	if pmapi.IsSystemLabel(labelID) {
		var err error
		switch labelID {
		case pmapi.SpamLabel:
			err = store.client().EmptyFolder(exposeContextForIMAP(), pmapi.SpamLabel, addressID)
		case pmapi.TrashLabel:
			err = store.client().EmptyFolder(exposeContextForIMAP(), pmapi.TrashLabel, addressID)
		default:
			err = fmt.Errorf("cannot empty mailbox %v", labelID)
		}
		return err
	}
	return store.client().DeleteLabel(exposeContextForIMAP(), labelID)
}

func (store *Store) createLabelsIfMissing(affectedLabelIDs map[string]bool) error {
	newLabelIDs := []string{}
	for labelID := range affectedLabelIDs {
		if pmapi.IsSystemLabel(labelID) || store.allAddressesHaveMailbox(labelID) {
			continue
		}
		newLabelIDs = append(newLabelIDs, labelID)
	}
	if len(newLabelIDs) == 0 {
		return nil
	}

	labels, err := store.client().ListLabels(exposeContextForIMAP())
	if err != nil {
		return err
	}
	for _, newLabelID := range newLabelIDs {
		for _, label := range labels {
			if label.ID != newLabelID {
				continue
			}
			if err := store.createOrUpdateMailboxEvent(label); err != nil {
				return err
			}
		}
	}
	return nil
}

// createOrUpdateMailboxEvent creates or updates the mailbox in the store.
// This is called from the event loop.
func (store *Store) createOrUpdateMailboxEvent(label *pmapi.Label) error {
	store.lock.Lock()
	defer store.lock.Unlock()

	if label.Type != pmapi.LabelTypeMailBox {
		return nil
	}

	if err := store.createOrUpdateMailboxCountsBuckets([]*pmapi.Label{label}); err != nil {
		return errors.Wrap(err, "cannot update counts")
	}

	for _, a := range store.addresses {
		if err := a.createOrUpdateMailboxEvent(label); err != nil {
			return err
		}
	}
	return nil
}

// deleteMailboxEvent deletes the mailbox in the store.
// This is called from the event loop.
func (store *Store) deleteMailboxEvent(labelID string) error {
	store.lock.Lock()
	defer store.lock.Unlock()

	if err := store.removeMailboxCount(labelID); err != nil {
		log.WithError(err).Warn("Problem to remove mailbox counts while deleting mailbox")
	}

	for _, a := range store.addresses {
		if err := a.deleteMailboxEvent(labelID); err != nil {
			return err
		}
	}
	return nil
}
