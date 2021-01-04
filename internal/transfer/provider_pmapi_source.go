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
	"fmt"
	"sync"

	pkgMessage "github.com/ProtonMail/proton-bridge/pkg/message"
	"github.com/ProtonMail/proton-bridge/pkg/pmapi"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

const pmapiListPageSize = 150

// TransferTo exports messages based on rules to channel.
func (p *PMAPIProvider) TransferTo(rules transferRules, progress *Progress, ch chan<- Message) {
	log.Info("Started transfer from PMAPI to channel")
	defer log.Info("Finished transfer from PMAPI to channel")

	p.timeIt.clear()
	defer p.timeIt.logResults()

	// TransferTo cannot end sooner than loadCounts goroutine because
	// loadCounts writes to channel in progress which would be closed.
	// That can happen for really small accounts.
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		p.loadCounts(rules, progress)
	}()

	for rule := range rules.iterateActiveRules() {
		p.transferTo(rule, progress, ch, rules.skipEncryptedMessages)
	}

	wg.Wait()
}

func (p *PMAPIProvider) loadCounts(rules transferRules, progress *Progress) {
	for rule := range rules.iterateActiveRules() {
		if progress.shouldStop() {
			break
		}

		rule := rule
		progress.callWrap(func() error {
			_, total, err := p.listMessages(&pmapi.MessagesFilter{
				AddressID: p.addressID,
				LabelID:   rule.SourceMailbox.ID,
				Begin:     rule.FromTime,
				End:       rule.ToTime,
				Limit:     0,
			})
			if err != nil {
				log.WithError(err).Warning("Problem to load counts")
				return err
			}
			progress.updateCount(rule.SourceMailbox.Name, uint(total))
			return nil
		})
	}
	progress.countsFinal()
}

func (p *PMAPIProvider) transferTo(rule *Rule, progress *Progress, ch chan<- Message, skipEncryptedMessages bool) {
	page := 0
	for {
		if progress.shouldStop() {
			break
		}

		isLastPage := true

		progress.callWrap(func() error {
			// Would be better to filter by Begin and BeginID to be sure
			// in case user deletes messages during the process, no message
			// is skipped (paging is off then), but API does not support
			// filtering by both mentioned fields at the same time.
			desc := false
			pmapiMessages, total, err := p.listMessages(&pmapi.MessagesFilter{
				AddressID: p.addressID,
				LabelID:   rule.SourceMailbox.ID,
				Begin:     rule.FromTime,
				End:       rule.ToTime,
				PageSize:  pmapiListPageSize,
				Page:      page,
				Sort:      "ID",
				Desc:      &desc,
			})
			if err != nil {
				return err
			}
			log.WithFields(logrus.Fields{
				"label": rule.SourceMailbox.ID,
				"page":  page,
				"total": total,
				"count": len(pmapiMessages),
			}).Debug("Listing messages")

			isLastPage = len(pmapiMessages) < pmapiListPageSize

			for _, pmapiMessage := range pmapiMessages {
				if progress.shouldStop() {
					break
				}

				msgID := fmt.Sprintf("%s_%s", rule.SourceMailbox.ID, pmapiMessage.ID)
				progress.addMessage(msgID, []string{rule.SourceMailbox.Name}, rule.TargetMailboxNames())
				msg, err := p.exportMessage(rule, progress, pmapiMessage.ID, msgID, skipEncryptedMessages)
				progress.messageExported(msgID, msg.Body, err)
				if err == nil {
					ch <- msg
				}
			}

			page++

			return nil
		})

		if isLastPage {
			break
		}
	}
}

func (p *PMAPIProvider) exportMessage(rule *Rule, progress *Progress, pmapiMsgID, msgID string, skipEncryptedMessages bool) (Message, error) {
	var msg *pmapi.Message
	progress.callWrap(func() error {
		var err error
		msg, err = p.getMessage(pmapiMsgID)
		return err
	})

	p.timeIt.start("build", msgID)
	defer p.timeIt.stop("build", msgID)

	msgBuilder := pkgMessage.NewBuilder(p.client(), msg)
	msgBuilder.EncryptedToHTML = false
	_, body, err := msgBuilder.BuildMessage()
	if err != nil {
		return Message{
			Body: body, // Keep body to show details about the message to user.
		}, errors.Wrap(err, "failed to build message")
	}

	if !msgBuilder.SuccessfullyDecrypted() && skipEncryptedMessages {
		return Message{
			Body: body, // Keep body to show details about the message to user.
		}, errors.New("skipping encrypted message")
	}

	unread := false
	if msg.Unread == 1 {
		unread = true
	}

	return Message{
		ID:      msgID,
		Unread:  unread,
		Body:    body,
		Sources: []Mailbox{rule.SourceMailbox},
		Targets: rule.TargetMailboxes,
	}, nil
}
