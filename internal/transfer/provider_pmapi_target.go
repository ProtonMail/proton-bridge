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
	"bytes"
	"fmt"
	"io"
	"io/ioutil"

	pkgMessage "github.com/ProtonMail/proton-bridge/pkg/message"
	"github.com/ProtonMail/proton-bridge/pkg/pmapi"
	"github.com/pkg/errors"
)

// DefaultMailboxes returns the default mailboxes for default rules if no other is found.
func (p *PMAPIProvider) DefaultMailboxes(_ Mailbox) []Mailbox {
	return []Mailbox{{
		ID:          pmapi.ArchiveLabel,
		Name:        "Archive",
		IsExclusive: true,
	}}
}

// CreateMailbox creates label in ProtonMail account.
func (p *PMAPIProvider) CreateMailbox(mailbox Mailbox) (Mailbox, error) {
	if mailbox.ID != "" {
		return Mailbox{}, errors.New("mailbox is already created")
	}

	exclusive := 0
	if mailbox.IsExclusive {
		exclusive = 1
	}

	label, err := p.client().CreateLabel(&pmapi.Label{
		Name:      mailbox.Name,
		Color:     mailbox.Color,
		Exclusive: exclusive,
		Type:      pmapi.LabelTypeMailbox,
	})
	if err != nil {
		return Mailbox{}, errors.Wrap(err, fmt.Sprintf("failed to create mailbox %s", mailbox.Name))
	}
	mailbox.ID = label.ID
	return mailbox, nil
}

// TransferFrom imports messages from channel.
func (p *PMAPIProvider) TransferFrom(rules transferRules, progress *Progress, ch <-chan Message) {
	log.Info("Started transfer from channel to PMAPI")
	defer log.Info("Finished transfer from channel to PMAPI")

	for msg := range ch {
		for progress.shouldStop() {
			break
		}

		var importedID string
		var err error
		if p.isMessageDraft(msg) {
			importedID, err = p.importDraft(msg, rules.globalMailbox)
		} else {
			importedID, err = p.importMessage(msg, rules.globalMailbox)
		}
		progress.messageImported(msg.ID, importedID, err)
	}
}

func (p *PMAPIProvider) isMessageDraft(msg Message) bool {
	for _, target := range msg.Targets {
		if target.ID == pmapi.DraftLabel {
			return true
		}
	}
	return false
}

func (p *PMAPIProvider) importDraft(msg Message, globalMailbox *Mailbox) (string, error) {
	message, attachmentReaders, err := p.parseMessage(msg)
	if err != nil {
		return "", errors.Wrap(err, "failed to parse message")
	}

	if err := message.Encrypt(p.keyRing, nil); err != nil {
		return "", errors.Wrap(err, "failed to encrypt draft")
	}

	if globalMailbox != nil {
		message.LabelIDs = append(message.LabelIDs, globalMailbox.ID)
	}

	attachments := message.Attachments
	message.Attachments = nil

	draft, err := p.createDraft(message, "", pmapi.DraftActionReply)
	if err != nil {
		return "", errors.Wrap(err, "failed to create draft")
	}

	for idx, attachment := range attachments {
		attachment.MessageID = draft.ID
		attachmentBody, _ := ioutil.ReadAll(attachmentReaders[idx])

		r := bytes.NewReader(attachmentBody)
		sigReader, err := attachment.DetachedSign(p.keyRing, r)
		if err != nil {
			return "", errors.Wrap(err, "failed to sign attachment")
		}

		r = bytes.NewReader(attachmentBody)
		encReader, err := attachment.Encrypt(p.keyRing, r)
		if err != nil {
			return "", errors.Wrap(err, "failed to encrypt attachment")
		}

		_, err = p.createAttachment(attachment, encReader, sigReader)
		if err != nil {
			return "", errors.Wrap(err, "failed to create attachment")
		}
	}

	return draft.ID, nil
}

func (p *PMAPIProvider) importMessage(msg Message, globalMailbox *Mailbox) (string, error) {
	message, attachmentReaders, err := p.parseMessage(msg)
	if err != nil {
		return "", errors.Wrap(err, "failed to parse message")
	}

	body, err := p.encryptMessage(message, attachmentReaders)
	if err != nil {
		return "", errors.Wrap(err, "failed to encrypt message")
	}

	unread := 0
	if msg.Unread {
		unread = 1
	}

	labelIDs := []string{}
	for _, target := range msg.Targets {
		// Frontend should not set All Mail to Rules, but to be sure...
		if target.ID != pmapi.AllMailLabel {
			labelIDs = append(labelIDs, target.ID)
		}
	}
	if globalMailbox != nil {
		labelIDs = append(labelIDs, globalMailbox.ID)
	}

	importMsgReq := &pmapi.ImportMsgReq{
		AddressID: p.addressID,
		Body:      body,
		Unread:    unread,
		Time:      message.Time,
		Flags:     computeMessageFlags(labelIDs),
		LabelIDs:  labelIDs,
	}

	results, err := p.importRequest([]*pmapi.ImportMsgReq{importMsgReq})
	if err != nil {
		return "", errors.Wrap(err, "failed to import messages")
	}
	if len(results) == 0 {
		return "", errors.New("import ended with no result")
	}
	if results[0].Error != nil {
		return "", errors.Wrap(results[0].Error, "failed to import message")
	}
	return results[0].MessageID, nil
}

func (p *PMAPIProvider) parseMessage(msg Message) (*pmapi.Message, []io.Reader, error) {
	message, _, _, attachmentReaders, err := pkgMessage.Parse(bytes.NewBuffer(msg.Body), "", "")
	return message, attachmentReaders, err
}

func (p *PMAPIProvider) encryptMessage(msg *pmapi.Message, attachmentReaders []io.Reader) ([]byte, error) {
	if msg.MIMEType == pmapi.ContentTypeMultipartEncrypted {
		return []byte(msg.Body), nil
	}
	return pkgMessage.BuildEncrypted(msg, attachmentReaders, p.keyRing)
}

func computeMessageFlags(labels []string) (flag int64) {
	for _, labelID := range labels {
		switch labelID {
		case pmapi.SentLabel:
			flag = (flag | pmapi.FlagSent)
		case pmapi.ArchiveLabel, pmapi.InboxLabel:
			flag = (flag | pmapi.FlagReceived)
		case pmapi.DraftLabel:
			log.Error("Found draft target in non-draft import")
		}
	}

	// NOTE: if the labels are custom only
	if flag == 0 {
		flag = pmapi.FlagReceived
	}

	return flag
}
