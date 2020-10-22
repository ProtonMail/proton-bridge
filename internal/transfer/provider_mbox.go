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
	"os"
	"path/filepath"
	"strings"

	"github.com/pkg/errors"
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
	filePaths, err := getAllPathsWithSuffix(p.root, ".mbox")
	if err != nil {
		return nil, err
	}

	mailboxNames := map[string]bool{}
	for _, filePath := range filePaths {
		fileName := filepath.Base(filePath)
		filePath, err := p.handleAppleMailMBOXStructure(filePath)
		if err != nil {
			log.WithError(err).Warn("Failed to handle MBOX structure")
			continue
		}

		mailboxName := strings.TrimSuffix(fileName, ".mbox")
		mailboxNames[mailboxName] = true

		labels, err := getGmailLabelsFromMboxFile(filepath.Join(p.root, filePath))
		if err != nil {
			log.WithError(err).Error("Failed to get gmail labels from mbox file")
			continue
		}
		for label := range labels {
			mailboxNames[label] = true
		}
	}

	mailboxes := []Mailbox{}
	for mailboxName := range mailboxNames {
		mailboxes = append(mailboxes, Mailbox{
			ID:          "",
			Name:        mailboxName,
			Color:       "",
			IsExclusive: false,
		})
	}
	return mailboxes, nil
}

// handleAppleMailMBOXStructure changes the path of mailbox directory to
// the path of mbox file. Apple Mail MBOX exports has this structure:
// `Folder.mbox` directory with `mbox` file inside.
// Example: `Folder.mbox/mbox` (and this function converts `Folder.mbox`
// to `Folder.mbox/mbox`).
func (p *MBOXProvider) handleAppleMailMBOXStructure(filePath string) (string, error) {
	if info, err := os.Stat(filepath.Join(p.root, filePath)); err == nil && info.IsDir() {
		if _, err := os.Stat(filepath.Join(p.root, filePath, "mbox")); err != nil {
			return "", errors.Wrap(err, "wrong mbox structure")
		}
		return filepath.Join(filePath, "mbox"), nil
	}
	return filePath, nil
}
