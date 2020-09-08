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

	"github.com/ProtonMail/proton-bridge/pkg/pmapi"
)

var systemFolderMapping = map[string]string{ //nolint[gochecknoglobals]
	"bin":       "Trash",
	"junk":      "Spam",
	"all":       "All Mail",
	"sent mail": "Sent",
	"draft":     "Drafts",
	"important": "Starred",
	// Add more translations.
}

// LeastUsedColor is intended to return color for creating a new inbox or label
func LeastUsedColor(mailboxes []Mailbox) string {
	usedColors := []string{}
	for _, m := range mailboxes {
		usedColors = append(usedColors, m.Color)
	}
	return pmapi.LeastUsedColor(usedColors)
}

// Mailbox is universal data holder of mailbox details for every provider.
type Mailbox struct {
	ID          string
	Name        string
	Color       string
	IsExclusive bool
}

// IsSystemFolder returns true when ID corresponds to PM system folder.
func (m Mailbox) IsSystemFolder() bool {
	return pmapi.IsSystemLabel(m.ID)
}

// Hash returns unique identifier to be used for matching.
func (m Mailbox) Hash() string {
	return fmt.Sprintf("%x", sha256.Sum256([]byte(m.Name)))
}

// findMatchingMailboxes returns all matching mailboxes from `mailboxes`.
// Only one exclusive mailbox is included.
func (m Mailbox) findMatchingMailboxes(mailboxes []Mailbox) []Mailbox {
	nameVariants := m.nameVariants()
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

// nameVariants returns all possible variants of the mailbox name.
// The best match (original name) is at the end of the slice.
// Variants are all in lower case. Examples:
//  * Foo/bar -> [foo, bar, foo/bar]
//  * x/Bin -> [x, trash, bin, x/bin]
//  * a|b/c -> [a, b, c, a|b/c]
func (m Mailbox) nameVariants() (nameVariants []string) {
	name := strings.ToLower(m.Name)
	if strings.Contains(name, "/") || strings.Contains(name, "|") {
		for _, slashPart := range strings.Split(name, "/") {
			for _, part := range strings.Split(slashPart, "|") {
				if mappedPart, ok := systemFolderMapping[part]; ok {
					nameVariants = append(nameVariants, strings.ToLower(mappedPart))
				}
				nameVariants = append(nameVariants, part)
			}
		}
	}
	if mappedName, ok := systemFolderMapping[name]; ok {
		nameVariants = append(nameVariants, strings.ToLower(mappedName))
	}
	return append(nameVariants, name)
}
