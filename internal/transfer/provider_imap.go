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
	"net"
	"strings"

	imapClient "github.com/emersion/go-imap/client"
)

// IMAPProvider implements export from IMAP server.
type IMAPProvider struct {
	username string
	password string
	addr     string

	client *imapClient.Client
}

// NewIMAPProvider returns new IMAPProvider.
func NewIMAPProvider(username, password, host, port string) (*IMAPProvider, error) {
	p := &IMAPProvider{
		username: username,
		password: password,
		addr:     net.JoinHostPort(host, port),
	}

	if err := p.auth(); err != nil {
		return nil, err
	}

	return p, nil
}

// ID is used for generating transfer ID by combining source and target ID.
// We want to keep the same rules for import from any IMAP server, therefore
// it returns constant.
func (p *IMAPProvider) ID() string {
	return "imap"
}

// Mailboxes returns all available folder names from root of EML files.
// In case the same folder name is used more than once (for example root/a/foo
// and root/b/foo), it's treated as the same folder.
func (p *IMAPProvider) Mailboxes(includeEmpty, includeAllMail bool) ([]Mailbox, error) {
	mailboxesInfo, err := p.list()
	if err != nil {
		return nil, err
	}

	mailboxes := []Mailbox{}
	for _, mailbox := range mailboxesInfo {
		hasNoSelect := false
		for _, attrib := range mailbox.Attributes {
			if strings.ToLower(attrib) == "\\noselect" {
				hasNoSelect = true
				break
			}
		}
		if hasNoSelect {
			continue
		}

		if !includeEmpty || true {
			mailboxStatus, err := p.selectIn(mailbox.Name)
			if err != nil {
				return nil, err
			}
			if mailboxStatus.Messages == 0 {
				continue
			}
		}

		mailboxes = append(mailboxes, Mailbox{
			ID:          "",
			Name:        mailbox.Name,
			Color:       "",
			IsExclusive: false,
		})
	}
	return mailboxes, nil
}
