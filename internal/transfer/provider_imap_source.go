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
	"fmt"
	"io/ioutil"

	"github.com/emersion/go-imap"
)

type imapMessageInfo struct {
	id   string
	uid  uint32
	size uint32
}

const (
	imapPageSize     = uint32(2000)             // Optimized on Gmail.
	imapMaxFetchSize = uint32(50 * 1000 * 1000) // Size in octets. If 0, it will use one fetch per message.
)

// TransferTo exports messages based on rules to channel.
func (p *IMAPProvider) TransferTo(rules transferRules, progress *Progress, ch chan<- Message) {
	log.Info("Started transfer from IMAP to channel")
	defer log.Info("Finished transfer from IMAP to channel")

	p.timeIt.clear()
	defer p.timeIt.logResults()

	imapMessageInfoMap := p.loadMessageInfoMap(rules, progress)

	for rule := range rules.iterateActiveRules() {
		log.WithField("rule", rule).Debug("Processing rule")
		messagesInfo := imapMessageInfoMap[rule.SourceMailbox.Name]
		p.transferTo(rule, messagesInfo, progress, ch)
	}
}

func (p *IMAPProvider) loadMessageInfoMap(rules transferRules, progress *Progress) map[string]map[string]imapMessageInfo {
	res := map[string]map[string]imapMessageInfo{}

	for rule := range rules.iterateActiveRules() {
		if progress.shouldStop() {
			break
		}

		mailboxName := rule.SourceMailbox.Name
		var mailbox *imap.MailboxStatus
		progress.callWrap(func() error {
			var err error
			mailbox, err = p.selectIn(mailboxName)
			return err
		})
		if mailbox.Messages == 0 {
			continue
		}

		messagesInfo := p.loadMessagesInfo(rule, progress, mailbox.UidValidity, mailbox.Messages)
		res[rule.SourceMailbox.Name] = messagesInfo
		progress.updateCount(rule.SourceMailbox.Name, uint(len(messagesInfo)))
	}
	progress.countsFinal()

	return res
}

func (p *IMAPProvider) loadMessagesInfo(rule *Rule, progress *Progress, uidValidity, count uint32) map[string]imapMessageInfo {
	p.timeIt.start("load", rule.SourceMailbox.Name)
	defer p.timeIt.stop("load", rule.SourceMailbox.Name)

	log := log.WithField("mailbox", rule.SourceMailbox.Name)
	messagesInfo := map[string]imapMessageInfo{}

	fetchItems := []imap.FetchItem{imap.FetchUid, imap.FetchRFC822Size}
	if rule.HasTimeLimit() {
		fetchItems = append(fetchItems, imap.FetchEnvelope)
	}

	processMessageCallback := func(imapMessage *imap.Message) {
		if rule.HasTimeLimit() {
			t := imapMessage.Envelope.Date.Unix()
			if t != 0 && !rule.isTimeInRange(t) {
				log.WithField("uid", imapMessage.Uid).Debug("Message skipped due to time")
				return
			}
		}
		id := getUniqueMessageID(rule.SourceMailbox.Name, uidValidity, imapMessage.Uid)
		// We use ID as key to ensure we have every unique message only once.
		// Some IMAP servers responded twice the same message...
		messagesInfo[id] = imapMessageInfo{
			id:   id,
			uid:  imapMessage.Uid,
			size: imapMessage.Size,
		}
		progress.addMessage(id, []string{rule.SourceMailbox.Name}, rule.TargetMailboxNames())
	}

	pageStart := uint32(1)
	pageEnd := imapPageSize
	for {
		if progress.shouldStop() || pageStart > count {
			break
		}

		// Some servers do not accept message sequence number higher than the total count.
		if pageEnd > count {
			pageEnd = count
		}

		seqSet := &imap.SeqSet{}
		seqSet.AddRange(pageStart, pageEnd)
		err := p.fetch(rule.SourceMailbox.Name, seqSet, fetchItems, processMessageCallback)
		if err != nil {
			log.WithError(err).WithField("idx", seqSet).Warning("Load batch fetch failed, trying one by one")
			for ; pageStart <= pageEnd; pageStart++ {
				seqSet := &imap.SeqSet{}
				seqSet.AddNum(pageStart)
				if err := p.fetch(rule.SourceMailbox.Name, seqSet, fetchItems, processMessageCallback); err != nil {
					log.WithError(err).WithField("idx", seqSet).Warning("Load fetch failed, skipping the message")
				}
			}
		}

		pageStart = pageEnd + 1
		pageEnd += imapPageSize
	}
	return messagesInfo
}

func (p *IMAPProvider) transferTo(rule *Rule, messagesInfo map[string]imapMessageInfo, progress *Progress, ch chan<- Message) {
	progress.callWrap(func() error {
		_, err := p.selectIn(rule.SourceMailbox.Name)
		return err
	})

	seqSet := &imap.SeqSet{}
	seqSetSize := uint32(0)
	uidToID := map[uint32]string{}

	for _, messageInfo := range messagesInfo {
		if progress.shouldStop() {
			break
		}

		if seqSetSize != 0 && (seqSetSize+messageInfo.size) > imapMaxFetchSize {
			log.WithField("mailbox", rule.SourceMailbox.Name).WithField("seq", seqSet).WithField("size", seqSetSize).Debug("Fetching messages")
			p.exportMessages(rule, progress, ch, seqSet, uidToID)

			seqSet = &imap.SeqSet{}
			seqSetSize = 0
			uidToID = map[uint32]string{}
		}

		seqSet.AddNum(messageInfo.uid)
		seqSetSize += messageInfo.size
		uidToID[messageInfo.uid] = messageInfo.id
	}

	if len(uidToID) != 0 {
		log.WithField("mailbox", rule.SourceMailbox.Name).WithField("seq", seqSet).WithField("size", seqSetSize).Debug("Fetching messages")
		p.exportMessages(rule, progress, ch, seqSet, uidToID)
	}
}

func (p *IMAPProvider) exportMessages(rule *Rule, progress *Progress, ch chan<- Message, seqSet *imap.SeqSet, uidToID map[uint32]string) {
	section := &imap.BodySectionName{}
	items := []imap.FetchItem{imap.FetchUid, imap.FetchFlags, section.FetchItem()}

	processMessageCallback := func(imapMessage *imap.Message) {
		if progress.shouldStop() {
			return
		}

		id, ok := uidToID[imapMessage.Uid]

		// Sometimes, server sends not requested messages.
		if !ok {
			log.WithField("uid", imapMessage.Uid).Warning("Message skipped: not requested")
			return
		}

		// Sometimes, server sends message twice, once with body and once without it.
		bodyReader := imapMessage.GetBody(section)
		if bodyReader == nil {
			log.WithField("uid", imapMessage.Uid).Warning("Message skipped: no body")
			return
		}

		body, err := ioutil.ReadAll(bodyReader)
		progress.messageExported(id, body, err)
		if err == nil {
			msg := p.exportMessage(rule, id, imapMessage, body)

			p.timeIt.stop("fetch", rule.SourceMailbox.Name)
			ch <- msg
			p.timeIt.start("fetch", rule.SourceMailbox.Name)
		}
	}

	p.timeIt.start("fetch", rule.SourceMailbox.Name)
	progress.callWrap(func() error {
		return p.uidFetch(rule.SourceMailbox.Name, seqSet, items, processMessageCallback)
	})
	p.timeIt.stop("fetch", rule.SourceMailbox.Name)
}

func (p *IMAPProvider) exportMessage(rule *Rule, id string, imapMessage *imap.Message, body []byte) Message {
	unread := true
	for _, flag := range imapMessage.Flags {
		if flag == imap.SeenFlag {
			unread = false
		}
	}

	return Message{
		ID:      id,
		Unread:  unread,
		Body:    body,
		Sources: []Mailbox{rule.SourceMailbox},
		Targets: rule.TargetMailboxes,
	}
}

func getUniqueMessageID(mailboxName string, uidValidity, uid uint32) string {
	return fmt.Sprintf("%s_%d:%d", mailboxName, uidValidity, uid)
}
