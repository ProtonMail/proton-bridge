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

import (
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/emersion/go-mbox"
	"github.com/hashicorp/go-multierror"
)

// DefaultMailboxes returns the default mailboxes for default rules if no other is found.
func (p *MBOXProvider) DefaultMailboxes(sourceMailbox Mailbox) []Mailbox {
	return []Mailbox{{
		Name: sourceMailbox.Name,
	}}
}

// CreateMailbox does nothing. Files are created dynamically during the import.
func (p *MBOXProvider) CreateMailbox(mailbox Mailbox) (Mailbox, error) {
	return mailbox, nil
}

// TransferFrom imports messages from channel.
func (p *MBOXProvider) TransferFrom(rules transferRules, progress *Progress, ch <-chan Message) {
	log.Info("Started transfer from channel to MBOX")
	defer log.Info("Finished transfer from channel to MBOX")

	for msg := range ch {
		if progress.shouldStop() {
			break
		}

		err := p.writeMessage(msg)
		progress.messageImported(msg.ID, "", err)
	}
}

func (p *MBOXProvider) writeMessage(msg Message) error {
	var multiErr error
	for _, mailbox := range msg.Targets {
		mboxName := sanitizeFileName(mailbox.Name)
		if !strings.HasSuffix(mboxName, ".mbox") {
			mboxName += ".mbox"
		}

		mboxPath := filepath.Join(p.root, mboxName)
		mboxFile, err := os.OpenFile(mboxPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0600)
		if err != nil {
			multiErr = multierror.Append(multiErr, err)
			continue
		}

		msgFrom := ""
		msgTime := time.Now()
		if header, err := getMessageHeader(msg.Body); err == nil {
			if date, err := header.Date(); err == nil {
				msgTime = date
			}
			if addresses, err := header.AddressList("from"); err == nil && len(addresses) > 0 {
				msgFrom = addresses[0].Address
			}
		}

		mboxWriter := mbox.NewWriter(mboxFile)
		messageWriter, err := mboxWriter.CreateMessage(msgFrom, msgTime)
		if err != nil {
			multiErr = multierror.Append(multiErr, err)
			continue
		}

		_, err = messageWriter.Write(msg.Body)
		if err != nil {
			multiErr = multierror.Append(multiErr, err)
			continue
		}
	}
	return multiErr
}
