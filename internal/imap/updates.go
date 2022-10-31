// Copyright (c) 2022 Proton AG
//
// This file is part of Proton Mail Bridge.
//
// Proton Mail Bridge is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// Proton Mail Bridge is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with Proton Mail Bridge. If not, see <https://www.gnu.org/licenses/>.

package imap

import (
	"strings"
	"sync"
	"time"

	"github.com/ProtonMail/proton-bridge/v2/internal/store"
	"github.com/ProtonMail/proton-bridge/v2/pkg/algo"
	"github.com/ProtonMail/proton-bridge/v2/pkg/message"
	"github.com/ProtonMail/proton-bridge/v2/pkg/pmapi"
	imap "github.com/emersion/go-imap"
	goIMAPBackend "github.com/emersion/go-imap/backend"
	"github.com/sirupsen/logrus"
)

type operation string

const (
	operationUpdateMessage operation = "store"
	operationDeleteMessage operation = "expunge"
)

type imapUpdates struct {
	lock            sync.Locker
	blocking        map[string]bool
	delayedExpunges map[string][]chan struct{}
	ch              chan goIMAPBackend.Update
	ib              *imapBackend
}

func newIMAPUpdates(ib *imapBackend) *imapUpdates {
	return &imapUpdates{
		lock:            &sync.Mutex{},
		blocking:        map[string]bool{},
		delayedExpunges: map[string][]chan struct{}{},
		ch:              make(chan goIMAPBackend.Update),
		ib:              ib,
	}
}

func (iu *imapUpdates) block(address, mailboxName string, op operation) {
	iu.lock.Lock()
	defer iu.lock.Unlock()

	iu.blocking[getBlockingKey(address, mailboxName, op)] = true
}

func (iu *imapUpdates) unblock(address, mailboxName string, op operation) {
	iu.lock.Lock()
	defer iu.lock.Unlock()

	delete(iu.blocking, getBlockingKey(address, mailboxName, op))
}

func (iu *imapUpdates) isBlocking(address, mailboxName string, op operation) bool {
	iu.lock.Lock()
	defer iu.lock.Unlock()

	return iu.blocking[getBlockingKey(address, mailboxName, op)]
}

func getBlockingKey(address, mailboxName string, op operation) string {
	return strings.ToLower(address + "_" + mailboxName + "_" + string(op))
}

func (iu *imapUpdates) forbidExpunge(mailboxID string) {
	iu.lock.Lock()
	defer iu.lock.Unlock()

	iu.delayedExpunges[mailboxID] = []chan struct{}{}
}

func (iu *imapUpdates) allowExpunge(mailboxID string) {
	iu.lock.Lock()
	defer iu.lock.Unlock()

	for _, ch := range iu.delayedExpunges[mailboxID] {
		close(ch)
	}
	delete(iu.delayedExpunges, mailboxID)
}

func (iu *imapUpdates) CanDelete(mailboxID string) (bool, func()) {
	iu.lock.Lock()
	defer iu.lock.Unlock()

	if iu.delayedExpunges[mailboxID] == nil {
		return true, nil
	}

	ch := make(chan struct{})
	iu.delayedExpunges[mailboxID] = append(iu.delayedExpunges[mailboxID], ch)
	return false, func() {
		log.WithField("mailbox", mailboxID).Debug("Expunge operations paused")
		<-ch
		log.WithField("mailbox", mailboxID).Debug("Expunge operations unpaused")
	}
}

func (iu *imapUpdates) Notice(address, notice string) {
	l := iu.updateLog(address, "")
	l.Info("Notice")
	update := new(goIMAPBackend.StatusUpdate)
	update.Update = goIMAPBackend.NewUpdate(address, "")
	update.StatusResp = &imap.StatusResp{
		Type: imap.StatusRespOk,
		Code: imap.CodeAlert,
		Info: notice,
	}
	iu.sendIMAPUpdate(l, update, false)
}

func (iu *imapUpdates) UpdateMessage(
	address, mailboxName string,
	uid, sequenceNumber uint32,
	msg *pmapi.Message, hasDeletedFlag bool,
) {
	l := iu.updateLog(address, mailboxName).
		WithFields(logrus.Fields{
			"seqNum":  sequenceNumber,
			"uid":     uid,
			"flags":   message.GetFlags(msg),
			"deleted": hasDeletedFlag,
		})
	l.Info("IDLE update")
	update := new(goIMAPBackend.MessageUpdate)
	update.Update = goIMAPBackend.NewUpdate(address, mailboxName)
	update.Message = imap.NewMessage(sequenceNumber, []imap.FetchItem{imap.FetchFlags, imap.FetchUid})
	update.Message.Flags = message.GetFlags(msg)
	if hasDeletedFlag {
		update.Message.Flags = append(update.Message.Flags, imap.DeletedFlag)
	}
	update.Message.Uid = uid
	iu.sendIMAPUpdate(l, update, iu.isBlocking(address, mailboxName, operationUpdateMessage))
}

func (iu *imapUpdates) DeleteMessage(address, mailboxName string, sequenceNumber uint32) {
	l := iu.updateLog(address, mailboxName).
		WithField("seqNum", sequenceNumber)
	l.Info("IDLE delete")
	update := new(goIMAPBackend.ExpungeUpdate)
	update.Update = goIMAPBackend.NewUpdate(address, mailboxName)
	update.SeqNum = sequenceNumber
	iu.sendIMAPUpdate(l, update, iu.isBlocking(address, mailboxName, operationDeleteMessage))
}

func (iu *imapUpdates) MailboxCreated(address, mailboxName string) {
	l := iu.updateLog(address, mailboxName)
	l.Info("IDLE mailbox info")
	update := new(goIMAPBackend.MailboxInfoUpdate)
	update.Update = goIMAPBackend.NewUpdate(address, "")
	update.MailboxInfo = &imap.MailboxInfo{
		Attributes: []string{imap.NoInferiorsAttr},
		Delimiter:  store.PathDelimiter,
		Name:       mailboxName,
	}
	iu.sendIMAPUpdate(l, update, false)
}

func (iu *imapUpdates) MailboxStatus(address, mailboxName string, total, unread, unreadSeqNum uint32) {
	l := iu.updateLog(address, mailboxName).
		WithFields(logrus.Fields{
			"total":        total,
			"unread":       unread,
			"unreadSeqNum": unreadSeqNum,
		})
	l.Info("IDLE status")
	update := new(goIMAPBackend.MailboxUpdate)
	update.Update = goIMAPBackend.NewUpdate(address, mailboxName)
	update.MailboxStatus = imap.NewMailboxStatus(mailboxName, []imap.StatusItem{imap.StatusMessages, imap.StatusUnseen})
	update.MailboxStatus.Messages = total
	update.MailboxStatus.Unseen = unread
	update.MailboxStatus.UnseenSeqNum = unreadSeqNum
	iu.sendIMAPUpdate(l, update, true)
}

func (iu *imapUpdates) sendIMAPUpdate(updateLog *logrus.Entry, update goIMAPBackend.Update, isBlocking bool) {
	l := updateLog.WithField("blocking", isBlocking)
	if iu.ch == nil {
		l.Info("IMAP IDLE unavailable")
		return
	}

	done := update.Done()
	go func() {
		select {
		case <-time.After(1 * time.Second):
			l.Warn("IMAP update could not be sent (timeout)")
			return
		case iu.ch <- update:
		}
	}()

	if !isBlocking {
		return
	}

	select {
	case <-done:
	case <-time.After(1 * time.Second):
		l.Warn("IMAP update could not be delivered (timeout)")
		return
	}
}

func (iu *imapUpdates) getIDs(address, mailboxName string) (addressID, mailboxID string) {
	addressID = "unknown-" + algo.HashBase64SHA256(address)
	mailboxID = "unknown-" + algo.HashBase64SHA256(mailboxName)

	if iu == nil || iu.ib == nil {
		return
	}

	user, err := iu.ib.getUser(address)
	if err != nil || user == nil || user.storeAddress == nil {
		return
	}
	addressID = user.addressID

	if v := user.mailboxIDs.get(mailboxName); v != "" {
		mailboxID = v
	}

	return
}

func (iu *imapUpdates) updateLog(address, mailboxName string) *logrus.Entry {
	addressID, mailboxID := iu.getIDs(address, mailboxName)
	return log.
		WithField("address", addressID).
		WithField("mailbox", mailboxID)
}
