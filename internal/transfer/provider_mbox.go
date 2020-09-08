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
	"path/filepath"
	"strings"
)

// MBOXProvider implements import and export to/from MBOX structure.
type MBOXProvider struct {
	root string
}

func NewMBOXProvider(root string) *MBOXProvider {
	return &MBOXProvider{
		root: root,
	}
}

// ID is used for generating transfer ID by combining source and target ID.
// We want to keep the same rules for import from or export to local files
// no matter exact path, therefore it returns constant. The same as EML.
func (p *MBOXProvider) ID() string {
	return "local" //nolint[goconst]
}

// Mailboxes returns all available folder names from root of EML files.
// In case the same folder name is used more than once (for example root/a/foo
// and root/b/foo), it's treated as the same folder.
func (p *MBOXProvider) Mailboxes(includeEmpty, includeAllMail bool) ([]Mailbox, error) {
	filePaths, err := getFilePathsWithSuffix(p.root, "mbox")
	if err != nil {
		return nil, err
	}

	mailboxes := []Mailbox{}
	for _, filePath := range filePaths {
		fileName := filepath.Base(filePath)
		mailboxName := strings.TrimSuffix(fileName, ".mbox")

		mailboxes = append(mailboxes, Mailbox{
			ID:          "",
			Name:        mailboxName,
			Color:       "",
			IsExclusive: false,
		})
	}

	return mailboxes, nil
}
