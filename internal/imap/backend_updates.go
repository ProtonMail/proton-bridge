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

package imap

import (
	"strings"
	"time"

	"github.com/ProtonMail/proton-bridge/internal/store"
	"github.com/ProtonMail/proton-bridge/pkg/message"
	"github.com/ProtonMail/proton-bridge/pkg/pmapi"
	imap "github.com/emersion/go-imap"
	goIMAPBackend "github.com/emersion/go-imap/backend"
	"github.com/sirupsen/logrus"
)

type operation string

const (
	operationUpdateMessage operation = "store"
	operationDeleteMessage operation = "expunge"
)

func (ib *imapBackend) setUpdatesBeBlocking(address, mailboxName string, op operation) {
	ib.changeUpdatesBlocking(address, mailboxName, op, true)
}

func (ib *imapBackend) unsetUpdatesBeBlocking(address, mailboxName string, op operation) {
	ib.changeUpdatesBlocking(address, mailboxName, op, false)
}

func (ib *imapBackend) changeUpdatesBlocking(address, mailboxName string, op operation, block bool) {
	ib.updatesBlockingLocker.Lock()
	defer ib.updatesBlockingLocker.Unlock()

	key := strings.ToLower(address + "_" + mailboxName + "_" + string(op))
	if block {
		ib.updatesBlocking[key] = true
	} else {
		delete(ib.updatesBlocking, key)
	}
}

func (ib *imapBackend) isBlocking(address, mailboxName string, op operation) bool {
	key := strings.ToLower(address + "_" + mailboxName + "_" + string(op))
	return ib.updatesBlocking[key]
}

func (ib *imapBackend) Notice(address, notice string) {
	update := new(goIMAPBackend.StatusUpdate)
	update.Update = goIMAPBackend.NewUpdate(address, "")
	update.StatusResp = &imap.StatusResp{
		Type: imap.StatusRespOk,
		Code: imap.CodeAlert,
		Info: notice,
	}
	ib.sendIMAPUpdate(update, false)
}

func (ib *imapBackend) UpdateMessage(
	address, mailboxName string,
	uid, sequenceNumber uint32,
	msg *pmapi.Message, hasDeletedFlag bool,
) {
	log.WithFields(logrus.Fields{
		"address": address,
		"mailbox": mailboxName,
		"seqNum":  sequenceNumber,
		"uid":     uid,
		"flags":   message.GetFlags(msg),
		"deleted": hasDeletedFlag,
	}).Trace("IDLE update")
	update := new(goIMAPBackend.MessageUpdate)
	update.Update = goIMAPBackend.NewUpdate(address, mailboxName)
	update.Message = imap.NewMessage(sequenceNumber, []imap.FetchItem{imap.FetchFlags, imap.FetchUid})
	update.Message.Flags = message.GetFlags(msg)
	if hasDeletedFlag {
		update.Message.Flags = append(update.Message.Flags, imap.DeletedFlag)
	}
	update.Message.Uid = uid
	ib.sendIMAPUpdate(update, ib.isBlocking(address, mailboxName, operationUpdateMessage))
}

func (ib *imapBackend) DeleteMessage(address, mailboxName string, sequenceNumber uint32) {
	log.WithFields(logrus.Fields{
		"address": address,
		"mailbox": mailboxName,
		"seqNum":  sequenceNumber,
	}).Trace("IDLE delete")
	update := new(goIMAPBackend.ExpungeUpdate)
	update.Update = goIMAPBackend.NewUpdate(address, mailboxName)
	update.SeqNum = sequenceNumber
	ib.sendIMAPUpdate(update, ib.isBlocking(address, mailboxName, operationDeleteMessage))
}

func (ib *imapBackend) MailboxCreated(address, mailboxName string) {
	log.WithFields(logrus.Fields{
		"address": address,
		"mailbox": mailboxName,
	}).Trace("IDLE mailbox info")
	update := new(goIMAPBackend.MailboxInfoUpdate)
	update.Update = goIMAPBackend.NewUpdate(address, "")
	update.MailboxInfo = &imap.MailboxInfo{
		Attributes: []string{imap.NoInferiorsAttr},
		Delimiter:  store.PathDelimiter,
		Name:       mailboxName,
	}
	ib.sendIMAPUpdate(update, false)
}

func (ib *imapBackend) MailboxStatus(address, mailboxName string, total, unread, unreadSeqNum uint32) {
	log.WithFields(logrus.Fields{
		"address":      address,
		"mailbox":      mailboxName,
		"total":        total,
		"unread":       unread,
		"unreadSeqNum": unreadSeqNum,
	}).Trace("IDLE status")
	update := new(goIMAPBackend.MailboxUpdate)
	update.Update = goIMAPBackend.NewUpdate(address, mailboxName)
	update.MailboxStatus = imap.NewMailboxStatus(mailboxName, []imap.StatusItem{imap.StatusMessages, imap.StatusUnseen})
	update.MailboxStatus.Messages = total
	update.MailboxStatus.Unseen = unread
	update.MailboxStatus.UnseenSeqNum = unreadSeqNum
	ib.sendIMAPUpdate(update, false)
}

func (ib *imapBackend) sendIMAPUpdate(update goIMAPBackend.Update, block bool) {
	if ib.updates == nil {
		log.Trace("IMAP IDLE unavailable")
		return
	}

	done := update.Done()
	go func() {
		select {
		case <-time.After(1 * time.Second):
			log.Warn("IMAP update could not be sent (timeout)")
			return
		case ib.updates <- update:
		}
	}()

	if !block {
		return
	}

	select {
	case <-done:
	case <-time.After(1 * time.Second):
		log.Warn("IMAP update could not be delivered (timeout).")
		return
	}
}
