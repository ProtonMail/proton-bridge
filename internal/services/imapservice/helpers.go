// Copyright (c) 2024 Proton AG
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
	"fmt"
	"net/mail"
	"time"

	"github.com/ProtonMail/gluon/imap"
	"github.com/ProtonMail/gluon/rfc5322"
	"github.com/ProtonMail/gluon/rfc822"
	"github.com/ProtonMail/go-proton-api"
	"github.com/bradenaw/juniper/xslices"
)

func toIMAPMailbox(label proton.Label, flags, permFlags, attrs imap.FlagSet) imap.Mailbox {
	if label.Type == proton.LabelTypeLabel {
		label.Path = append([]string{labelPrefix}, label.Path...)
	} else if label.Type == proton.LabelTypeFolder {
		label.Path = append([]string{folderPrefix}, label.Path...)
	}

	return imap.Mailbox{
		ID:             imap.MailboxID(label.ID),
		Name:           label.Path,
		Flags:          flags,
		PermanentFlags: permFlags,
		Attributes:     attrs,
	}
}

func isAllMailOrScheduled(mailboxID imap.MailboxID) bool {
	return (mailboxID == proton.AllMailLabel) || (mailboxID == proton.AllScheduledLabel)
}

func BuildFlagSetFromMessageMetadata(message proton.MessageMetadata) imap.FlagSet {
	flags := imap.NewFlagSet()

	if message.Seen() {
		flags.AddToSelf(imap.FlagSeen)
	}

	if message.Starred() {
		flags.AddToSelf(imap.FlagFlagged)
	}

	if message.IsDraft() {
		flags.AddToSelf(imap.FlagDraft)
	}

	if message.IsRepliedAll == true || message.IsReplied == true { //nolint: gosimple
		flags.AddToSelf(imap.FlagAnswered)
	}

	if message.IsForwarded {
		flags.AddToSelf(imap.ForwardFlagList...)
	}

	return flags
}

func getLiteralToList(literal []byte) ([]string, error) {
	headerLiteral, _ := rfc822.Split(literal)

	header, err := rfc822.NewHeader(headerLiteral)
	if err != nil {
		return nil, err
	}

	var result []string

	parseAddress := func(field string) error {
		if fieldAddr, ok := header.GetChecked(field); ok {
			addr, err := rfc5322.ParseAddressList(fieldAddr)
			if err != nil {
				return fmt.Errorf("failed to parse addresses for '%v': %w", field, err)
			}

			result = append(result, xslices.Map(addr, func(addr *mail.Address) string {
				return addr.Address
			})...)

			return nil
		}

		return nil
	}

	if err := parseAddress("To"); err != nil {
		return nil, err
	}

	if err := parseAddress("Cc"); err != nil {
		return nil, err
	}

	if err := parseAddress("Bcc"); err != nil {
		return nil, err
	}

	return result, nil
}

func toIMAPMessage(message proton.MessageMetadata) imap.Message {
	flags := BuildFlagSetFromMessageMetadata(message)

	var date time.Time

	if message.Time > 0 {
		date = time.Unix(message.Time, 0)
	} else {
		date = time.Now()
	}

	return imap.Message{
		ID:    imap.MessageID(message.ID),
		Flags: flags,
		Date:  date,
	}
}

func WantLabel(label proton.Label) bool {
	if label.Type != proton.LabelTypeSystem {
		return label.Type != proton.LabelTypeContactGroup
	}

	// nolint:exhaustive
	switch label.ID {
	case proton.InboxLabel:
		return true

	case proton.TrashLabel:
		return true

	case proton.SpamLabel:
		return true

	case proton.AllMailLabel:
		return true

	case proton.ArchiveLabel:
		return true

	case proton.SentLabel:
		return true

	case proton.DraftsLabel:
		return true

	case proton.StarredLabel:
		return true

	case proton.AllScheduledLabel:
		return true

	default:
		return false
	}
}

func wantLabels(apiLabels map[string]proton.Label, labelIDs []string) []string {
	return xslices.Filter(labelIDs, func(labelID string) bool {
		apiLabel, ok := apiLabels[labelID]
		if !ok {
			return false
		}

		return WantLabel(apiLabel)
	})
}
