// Copyright (c) 2025 Proton AG
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
	"strings"

	"github.com/ProtonMail/gluon/imap"
	"github.com/ProtonMail/go-proton-api"
)

func newSystemMailboxCreatedUpdate(labelID imap.MailboxID, labelName string) *imap.MailboxCreated {
	if strings.EqualFold(labelName, imap.Inbox) {
		labelName = imap.Inbox
	}

	attrs := imap.NewFlagSet(imap.AttrNoInferiors)
	permanentFlags := defaultMailboxPermanentFlags()
	flags := defaultMailboxFlags()

	switch labelID {
	case proton.TrashLabel:
		attrs = attrs.Add(imap.AttrTrash)

	case proton.SpamLabel:
		attrs = attrs.Add(imap.AttrJunk)

	case proton.AllMailLabel:
		attrs = attrs.Add(imap.AttrAll)
		flags = imap.NewFlagSet(imap.FlagSeen, imap.FlagFlagged)
		permanentFlags = imap.NewFlagSet(imap.FlagSeen, imap.FlagFlagged)

	case proton.ArchiveLabel:
		attrs = attrs.Add(imap.AttrArchive)

	case proton.SentLabel:
		attrs = attrs.Add(imap.AttrSent)

	case proton.DraftsLabel:
		attrs = attrs.Add(imap.AttrDrafts)

	case proton.StarredLabel:
		attrs = attrs.Add(imap.AttrFlagged)

	case proton.AllScheduledLabel:
		labelName = "Scheduled" // API actual name is "All Scheduled"
	}

	return imap.NewMailboxCreated(imap.Mailbox{
		ID:             labelID,
		Name:           []string{labelName},
		Flags:          flags,
		PermanentFlags: permanentFlags,
		Attributes:     attrs,
	})
}

func waitOnIMAPUpdates(ctx context.Context, updates []imap.Update) error {
	for _, update := range updates {
		if err, ok := update.WaitContext(ctx); ok && err != nil {
			return fmt.Errorf("failed to apply gluon update %v: %w", update.String(), err)
		}
	}

	return nil
}

func newPlaceHolderMailboxCreatedUpdate(labelName string) *imap.MailboxCreated {
	return imap.NewMailboxCreated(imap.Mailbox{
		ID:             imap.MailboxID(labelName),
		Name:           []string{labelName},
		Flags:          defaultMailboxFlags(),
		PermanentFlags: defaultMailboxPermanentFlags(),
		Attributes:     imap.NewFlagSet(imap.AttrNoSelect),
	})
}

func newMailboxCreatedUpdate(labelID imap.MailboxID, labelName []string) *imap.MailboxCreated {
	return imap.NewMailboxCreated(imap.Mailbox{
		ID:             labelID,
		Name:           labelName,
		Flags:          defaultMailboxFlags(),
		PermanentFlags: defaultMailboxPermanentFlags(),
		Attributes:     imap.NewFlagSet(),
	})
}

func newMailboxUpdatedOrCreated(labelID imap.MailboxID, labelName []string) *imap.MailboxUpdatedOrCreated {
	return imap.NewMailboxUpdatedOrCreated(imap.Mailbox{
		ID:             labelID,
		Name:           labelName,
		Flags:          defaultMailboxFlags(),
		PermanentFlags: defaultMailboxPermanentFlags(),
		Attributes:     imap.NewFlagSet(),
	})
}

func GetMailboxName(label proton.Label) []string {
	var name []string

	switch label.Type {
	case proton.LabelTypeFolder:
		name = append([]string{folderPrefix}, label.Path...)

	case proton.LabelTypeLabel:
		name = append([]string{labelPrefix}, label.Path...)

	case proton.LabelTypeContactGroup:
		fallthrough
	case proton.LabelTypeSystem:
		fallthrough
	default:
		name = label.Path
	}

	return name
}

func nameWithTempPrefix(path []string) []string {
	path[len(path)-1] = "tmp_" + path[len(path)-1]
	return path
}

func getMailboxNameWithTempPrefix(label proton.Label) []string {
	return nameWithTempPrefix(GetMailboxName(label))
}
