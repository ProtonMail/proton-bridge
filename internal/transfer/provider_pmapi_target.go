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
	"bytes"
	"context"
	"fmt"
	"io"
	"io/ioutil"
	"net/mail"
	"sync"

	pkgMsg "github.com/ProtonMail/proton-bridge/pkg/message"
	"github.com/ProtonMail/proton-bridge/pkg/pmapi"
	"github.com/pkg/errors"
)

const (
	pmapiImportBatchMaxItems = 10
	pmapiImportBatchMaxSize  = 25 * 1000 * 1000 // 25 MB
	pmapiImportWorkers       = 4                // To keep memory under 1 GB.
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

	label, err := p.client.CreateLabel(context.TODO(), &pmapi.Label{
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

	p.timeIt.clear()
	defer p.timeIt.logResults()

	// Cache has to be cleared before each transfer to not contain
	// old stuff from previous cancelled run.
	p.nextImportRequests = map[string]*pmapi.ImportMsgReq{}
	p.nextImportRequestsSize = 0

	preparedImportRequestsCh := make(chan map[string]*pmapi.ImportMsgReq)
	wg := p.startImportWorkers(progress, preparedImportRequestsCh)

	for msg := range ch {
		if progress.shouldStop() {
			break
		}

		if p.isMessageDraft(msg) {
			p.transferDraft(rules, progress, msg)
		} else {
			p.transferMessage(rules, progress, msg, preparedImportRequestsCh)
		}
	}

	if len(p.nextImportRequests) > 0 {
		preparedImportRequestsCh <- p.nextImportRequests
	}
	close(preparedImportRequestsCh)
	wg.Wait()
}

func (p *PMAPIProvider) isMessageDraft(msg Message) bool {
	for _, target := range msg.Targets {
		if target.ID == pmapi.DraftLabel {
			return true
		}
	}
	return false
}

func (p *PMAPIProvider) transferDraft(rules transferRules, progress *Progress, msg Message) {
	importedID, err := p.importDraft(msg, rules.globalMailbox)
	progress.messageImported(msg.ID, importedID, err)
}

func (p *PMAPIProvider) importDraft(msg Message, globalMailbox *Mailbox) (string, error) { //nolint[funlen]
	message, attachmentReaders, err := p.parseMessage(msg)
	if err != nil {
		return "", errors.Wrap(err, "failed to parse message")
	}

	if message.Sender == nil {
		mainAddress := p.client().Addresses().Main()
		message.Sender = &mail.Address{
			Name:    mainAddress.DisplayName,
			Address: mainAddress.Email,
		}
	}

	// Trying to encrypt an encrypted draft will return an error;
	// users are forbidden to import messages encrypted with foreign keys to drafts.
	if message.IsEncrypted() {
		return "", errors.New("refusing to import draft encrypted by foreign key")
	}

	p.timeIt.start("encrypt", msg.ID)
	err = message.Encrypt(p.keyRing, nil)
	p.timeIt.stop("encrypt", msg.ID)
	if err != nil {
		return "", errors.Wrap(err, "failed to encrypt draft")
	}

	if globalMailbox != nil {
		message.LabelIDs = append(message.LabelIDs, globalMailbox.ID)
	}

	attachments := message.Attachments
	message.Attachments = nil

	draft, err := p.createDraft(msg.ID, message, "", pmapi.DraftActionReply)
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

		p.timeIt.start("encrypt", msg.ID)
		r = bytes.NewReader(attachmentBody)
		encReader, err := attachment.Encrypt(p.keyRing, r)
		p.timeIt.stop("encrypt", msg.ID)
		if err != nil {
			return "", errors.Wrap(err, "failed to encrypt attachment")
		}

		_, err = p.createAttachment(msg.ID, attachment, encReader, sigReader)
		if err != nil {
			return "", errors.Wrap(err, "failed to create attachment")
		}
	}

	return draft.ID, nil
}

func (p *PMAPIProvider) transferMessage(rules transferRules, progress *Progress, msg Message, preparedImportRequestsCh chan map[string]*pmapi.ImportMsgReq) {
	importMsgReq, err := p.generateImportMsgReq(rules, progress, msg)
	if err != nil {
		progress.messageImported(msg.ID, "", err)
		return
	}
	if importMsgReq == nil || progress.shouldStop() {
		return
	}

	importMsgReqSize := len(importMsgReq.Message)
	if p.nextImportRequestsSize+importMsgReqSize > pmapiImportBatchMaxSize || len(p.nextImportRequests) == pmapiImportBatchMaxItems {
		preparedImportRequestsCh <- p.nextImportRequests
		p.nextImportRequests = map[string]*pmapi.ImportMsgReq{}
		p.nextImportRequestsSize = 0
	}
	p.nextImportRequests[msg.ID] = importMsgReq
	p.nextImportRequestsSize += importMsgReqSize
}

func (p *PMAPIProvider) generateImportMsgReq(rules transferRules, progress *Progress, msg Message) (*pmapi.ImportMsgReq, error) {
	message, attachmentReaders, err := p.parseMessage(msg)
	if err != nil {
		return nil, errors.Wrap(err, "failed to parse message")
	}

	var body []byte
	if message.IsEncrypted() {
		if rules.skipEncryptedMessages {
			progress.messageSkipped(msg.ID)
			return nil, nil
		}
		body = msg.Body
	} else {
		p.timeIt.start("encrypt", msg.ID)
		body, err = p.encryptMessage(message, attachmentReaders)
		p.timeIt.stop("encrypt", msg.ID)
		if err != nil {
			return nil, errors.Wrap(err, "failed to encrypt message")
		}
	}

	var unread pmapi.Boolean

	if msg.Unread {
		unread = pmapi.True
	} else {
		unread = pmapi.False
	}

	labelIDs := []string{}
	for _, target := range msg.Targets {
		// Frontend should not set All Mail to Rules, but to be sure...
		if target.ID != pmapi.AllMailLabel {
			labelIDs = append(labelIDs, target.ID)
		}
	}
	if rules.globalMailbox != nil {
		labelIDs = append(labelIDs, rules.globalMailbox.ID)
	}

	return &pmapi.ImportMsgReq{
		Metadata: &pmapi.ImportMetadata{
			AddressID: p.addressID,
			Unread:    unread,
			Time:      message.Time,
			Flags:     computeMessageFlags(message.Header),
			LabelIDs:  labelIDs,
		},
		Message: body,
	}, nil
}

func (p *PMAPIProvider) parseMessage(msg Message) (m *pmapi.Message, r []io.Reader, err error) {
	p.timeIt.start("parse", msg.ID)
	defer p.timeIt.stop("parse", msg.ID)
	message, _, _, attachmentReaders, err := pkgMsg.Parse(bytes.NewBuffer(msg.Body))
	return message, attachmentReaders, err
}

func (p *PMAPIProvider) encryptMessage(msg *pmapi.Message, attachmentReaders []io.Reader) ([]byte, error) {
	if msg.MIMEType == pmapi.ContentTypeMultipartEncrypted {
		return []byte(msg.Body), nil
	}
	return pkgMsg.BuildEncrypted(msg, attachmentReaders, p.keyRing)
}

func computeMessageFlags(header mail.Header) (flag int64) {
	if header.Get("received") == "" {
		return pmapi.FlagSent
	}
	return pmapi.FlagReceived
}

func (p *PMAPIProvider) startImportWorkers(progress *Progress, preparedImportRequestsCh chan map[string]*pmapi.ImportMsgReq) *sync.WaitGroup {
	var wg sync.WaitGroup
	wg.Add(pmapiImportWorkers)
	for i := 0; i < pmapiImportWorkers; i++ {
		go func() {
			for importRequests := range preparedImportRequestsCh {
				p.importMessages(progress, importRequests)
			}
			wg.Done()
		}()
	}
	return &wg
}

func (p *PMAPIProvider) importMessages(progress *Progress, importRequests map[string]*pmapi.ImportMsgReq) {
	if progress.shouldStop() {
		return
	}

	importMsgIDs := []string{}
	importMsgRequests := pmapi.ImportMsgReqs{}
	for msgID, req := range importRequests {
		importMsgIDs = append(importMsgIDs, msgID)
		importMsgRequests = append(importMsgRequests, req)
	}
	log.WithField("msgIDs", importMsgIDs).Trace("Importing messages")
	results, err := p.importRequest(importMsgIDs[0], importMsgRequests)

	// In case the whole request failed, try to import every message one by one.
	if err != nil || len(results) == 0 {
		log.WithError(err).Warning("Importing messages failed, trying one by one")
		for msgID, req := range importRequests {
			importedID, err := p.importMessage(msgID, progress, req)
			progress.messageImported(msgID, importedID, err)
		}
		return
	}

	// In case request passed but some messages failed, try to import the failed ones alone.
	for index, result := range results {
		msgID := importMsgIDs[index]
		if result.Error != nil {
			log.WithError(result.Error).WithField("msg", msgID).Warning("Importing message failed, trying alone")
			req := importMsgRequests[index]
			importedID, err := p.importMessage(msgID, progress, req)
			progress.messageImported(msgID, importedID, err)
		} else {
			progress.messageImported(msgID, result.MessageID, nil)
		}
	}
}

func (p *PMAPIProvider) importMessage(msgSourceID string, progress *Progress, req *pmapi.ImportMsgReq) (importedID string, importedErr error) {
	progress.callWrap(func() error {
		results, err := p.importRequest(msgSourceID, pmapi.ImportMsgReqs{req})
		if err != nil {
			return errors.Wrap(err, "failed to import messages")
		}
		if len(results) == 0 {
			importedErr = errors.New("import ended with no result")
			return nil // This should not happen, only when there is bug which means we should skip this one.
		}
		if results[0].Error != nil {
			importedErr = errors.Wrap(results[0].Error, "failed to import message")
			return nil //nolint[nilerr] Call passed but API refused this message, skip this one.
		}
		importedID = results[0].MessageID
		return nil
	})
	return importedID, importedErr
}
