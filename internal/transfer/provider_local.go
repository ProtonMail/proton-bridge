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

// LocalProvider implements import from local EML and MBOX file structure.
type LocalProvider struct {
	root         string
	emlProvider  *EMLProvider
	mboxProvider *MBOXProvider
}

func NewLocalProvider(root string) *LocalProvider {
	return &LocalProvider{
		root:         root,
		emlProvider:  NewEMLProvider(root),
		mboxProvider: NewMBOXProvider(root),
	}
}

// ID is used for generating transfer ID by combining source and target ID.
// We want to keep the same rules for import from or export to local files
// no matter exact path, therefore it returns constant.
// The same as EML and MBOX.
func (p *LocalProvider) ID() string {
	return "local" //nolint[goconst]
}

// Mailboxes returns all available folder names from root of EML and MBOX files.
func (p *LocalProvider) Mailboxes(includeEmpty, includeAllMail bool) ([]Mailbox, error) {
	mailboxes, err := p.emlProvider.Mailboxes(includeEmpty, includeAllMail)
	if err != nil {
		return nil, err
	}

	mboxMailboxes, err := p.mboxProvider.Mailboxes(includeEmpty, includeAllMail)
	if err != nil {
		return nil, err
	}

	for _, mboxMailbox := range mboxMailboxes {
		found := false
		for _, mailboxes := range mailboxes {
			if mboxMailbox.Name == mailboxes.Name {
				found = true
				break
			}
		}
		if !found {
			mailboxes = append(mailboxes, mboxMailbox)
		}
	}
	return mailboxes, nil
}
