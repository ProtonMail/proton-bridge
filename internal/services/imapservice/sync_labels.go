// Copyright (c) 2023 Proton AG
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

package imapservice

import (
	"context"
	"fmt"

	"github.com/ProtonMail/gluon/imap"
	"github.com/ProtonMail/go-proton-api"
)

// nolint:exhaustive
func syncLabels(ctx context.Context, apiLabels map[string]proton.Label, updatePublishers ...updatePublisher) error {
	var updates []imap.Update

	// Create placeholder Folders/Labels mailboxes with the \Noselect attribute.
	for _, prefix := range []string{folderPrefix, labelPrefix} {
		for _, updateCh := range updatePublishers {
			update := newPlaceHolderMailboxCreatedUpdate(prefix)
			updateCh.publishUpdate(ctx, update)
			updates = append(updates, update)
		}
	}

	// Sync the user's labels.
	for labelID, label := range apiLabels {
		if !WantLabel(label) {
			continue
		}

		switch label.Type {
		case proton.LabelTypeSystem:
			for _, updateCh := range updatePublishers {
				update := newSystemMailboxCreatedUpdate(imap.MailboxID(label.ID), label.Name)
				updateCh.publishUpdate(ctx, update)
				updates = append(updates, update)
			}

		case proton.LabelTypeFolder, proton.LabelTypeLabel:
			for _, updateCh := range updatePublishers {
				update := newMailboxCreatedUpdate(imap.MailboxID(labelID), GetMailboxName(label))
				updateCh.publishUpdate(ctx, update)
				updates = append(updates, update)
			}

		default:
			return fmt.Errorf("unknown label type: %d", label.Type)
		}
	}

	// Wait for all label updates to be applied.
	for _, update := range updates {
		err, ok := update.WaitContext(ctx)
		if ok && err != nil {
			return fmt.Errorf("failed to apply label create update in gluon %v: %w", update.String(), err)
		}
	}

	return nil
}
