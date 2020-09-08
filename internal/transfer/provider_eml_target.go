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
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/hashicorp/go-multierror"
)

// DefaultMailboxes returns the default mailboxes for default rules if no other is found.
func (p *EMLProvider) DefaultMailboxes(sourceMailbox Mailbox) []Mailbox {
	return []Mailbox{{
		Name: sourceMailbox.Name,
	}}
}

// CreateMailbox does nothing. Folders are created dynamically during the import.
func (p *EMLProvider) CreateMailbox(mailbox Mailbox) (Mailbox, error) {
	return mailbox, nil
}

// TransferFrom imports messages from channel.
func (p *EMLProvider) TransferFrom(rules transferRules, progress *Progress, ch <-chan Message) {
	log.Info("Started transfer from channel to EML")
	defer log.Info("Finished transfer from channel to EML")

	err := p.createFolders(rules)
	if err != nil {
		progress.fatal(err)
		return
	}

	for msg := range ch {
		for progress.shouldStop() {
			break
		}

		err := p.writeFile(msg)
		progress.messageImported(msg.ID, "", err)
	}
}

func (p *EMLProvider) createFolders(rules transferRules) error {
	for rule := range rules.iterateActiveRules() {
		for _, mailbox := range rule.TargetMailboxes {
			path := filepath.Join(p.root, mailbox.Name)
			if err := os.MkdirAll(path, os.ModePerm); err != nil {
				return err
			}
		}
	}
	return nil
}

func (p *EMLProvider) writeFile(msg Message) error {
	fileName := filepath.Base(msg.ID)
	if filepath.Ext(fileName) != ".eml" {
		fileName += ".eml"
	}

	var err error
	for _, mailbox := range msg.Targets {
		path := filepath.Join(p.root, mailbox.Name, fileName)

		if localErr := ioutil.WriteFile(path, msg.Body, 0600); localErr != nil {
			err = multierror.Append(err, localErr)
		}
	}
	return err
}
