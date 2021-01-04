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

package transfer

// EMLProvider implements import and export to/from EML file structure.
type EMLProvider struct {
	root string
}

// NewEMLProvider creates EMLProvider.
func NewEMLProvider(root string) *EMLProvider {
	return &EMLProvider{
		root: root,
	}
}

// ID is used for generating transfer ID by combining source and target ID.
// We want to keep the same rules for import from or export to local files
// no matter exact path, therefore it returns constant. The same as EML.
func (p *EMLProvider) ID() string {
	return "local" //nolint[goconst]
}

// Mailboxes returns all available folder names from root of EML files.
// In case the same folder name is used more than once (for example root/a/foo
// and root/b/foo), it's treated as the same folder.
func (p *EMLProvider) Mailboxes(includeEmpty, includeAllMail bool) (mailboxes []Mailbox, err error) {
	// Special case for exporting--we don't know the path before setup if finished.
	if p.root == "" {
		return
	}

	var folderNames []string
	if includeEmpty {
		folderNames, err = getFolderNames(p.root)
	} else {
		folderNames, err = getFolderNamesWithFileSuffix(p.root, ".eml")
	}
	if err != nil {
		return nil, err
	}

	for _, folderName := range folderNames {
		mailboxes = append(mailboxes, Mailbox{
			ID:          "",
			Name:        folderName,
			Color:       "",
			IsExclusive: false,
		})
	}

	return mailboxes, nil
}
