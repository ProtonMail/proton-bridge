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
	"crypto/sha256"
	"fmt"
	"strings"
)

// Mailbox is universal data holder of mailbox details for every provider.
type Mailbox struct {
	ID          string
	Name        string
	Color       string
	IsExclusive bool
}

// Hash returns unique identifier to be used for matching.
func (m Mailbox) Hash() string {
	return fmt.Sprintf("%x", sha256.Sum256([]byte(m.Name)))
}

// findMatchingMailboxes returns all matching mailboxes from `mailboxes`.
// Only one exclusive mailbox is returned.
func (m Mailbox) findMatchingMailboxes(mailboxes []Mailbox) []Mailbox {
	nameVariants := []string{}
	if strings.Contains(m.Name, "/") || strings.Contains(m.Name, "|") {
		for _, slashPart := range strings.Split(m.Name, "/") {
			for _, part := range strings.Split(slashPart, "|") {
				nameVariants = append(nameVariants, strings.ToLower(part))
			}
		}
	}
	nameVariants = append(nameVariants, strings.ToLower(m.Name))

	isExclusiveIncluded := false
	matches := []Mailbox{}
	for i := range nameVariants {
		nameVariant := nameVariants[len(nameVariants)-1-i]
		for _, mailbox := range mailboxes {
			if mailbox.IsExclusive && isExclusiveIncluded {
				continue
			}
			if strings.ToLower(mailbox.Name) == nameVariant {
				matches = append(matches, mailbox)
				if mailbox.IsExclusive {
					isExclusiveIncluded = true
				}
			}
		}
	}
	return matches
}
