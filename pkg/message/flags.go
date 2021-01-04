// Copyright (c) 2021 Proton Technologies AG
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

package message

import (
	"github.com/ProtonMail/proton-bridge/pkg/pmapi"
	"github.com/emersion/go-imap"
)

//nolint[gochecknoglobals]
var (
	AppleMailJunkFlag      = imap.CanonicalFlag("$Junk")
	ThunderbirdJunkFlag    = imap.CanonicalFlag("Junk")
	ThunderbirdNonJunkFlag = imap.CanonicalFlag("NonJunk")
)

func GetFlags(m *pmapi.Message) (flags []string) {
	if m.Unread == 0 {
		flags = append(flags, imap.SeenFlag)
	}
	if !m.Has(pmapi.FlagSent) && !m.Has(pmapi.FlagReceived) {
		flags = append(flags, imap.DraftFlag)
	}
	if m.Has(pmapi.FlagReplied) || m.Has(pmapi.FlagRepliedAll) {
		flags = append(flags, imap.AnsweredFlag)
	}

	hasSpam := false

	for _, l := range m.LabelIDs {
		if l == pmapi.StarredLabel {
			flags = append(flags, imap.FlaggedFlag)
		}
		if l == pmapi.SpamLabel {
			flags = append(flags, AppleMailJunkFlag, ThunderbirdJunkFlag)
			hasSpam = true
		}
	}

	if !hasSpam {
		flags = append(flags, ThunderbirdNonJunkFlag)
	}

	return
}

func ParseFlags(m *pmapi.Message, flags []string) {
	// Consider to use ComputeMessageFlagsByLabels to keep logic in one place.
	if (m.Flags & pmapi.FlagSent) == 0 {
		m.Flags |= pmapi.FlagReceived
	}
	m.Unread = 1
	for _, f := range flags {
		switch f {
		case imap.SeenFlag:
			m.Unread = 0
		case imap.DraftFlag:
			m.Flags &= ^pmapi.FlagSent
			m.Flags &= ^pmapi.FlagReceived
			m.LabelIDs = append(m.LabelIDs, pmapi.DraftLabel)
		case imap.FlaggedFlag:
			m.LabelIDs = append(m.LabelIDs, pmapi.StarredLabel)
		case imap.AnsweredFlag:
			m.Flags |= pmapi.FlagReplied
		case AppleMailJunkFlag, ThunderbirdJunkFlag:
			m.LabelIDs = append(m.LabelIDs, pmapi.SpamLabel)
		}
	}
}
