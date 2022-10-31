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

package message

import (
	"github.com/ProtonMail/proton-bridge/v2/pkg/pmapi"
	"github.com/emersion/go-imap"
)

// Various client specific flags.
const (
	AppleMailJunkFlag      = "$Junk"
	ThunderbirdJunkFlag    = "Junk"
	ThunderbirdNonJunkFlag = "NonJunk"
)

// GetFlags returns imap flags from pmapi message attributes.
func GetFlags(m *pmapi.Message) (flags []string) {
	if !m.Unread {
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
