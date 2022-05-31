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
	"fmt"
	"net/mail"
	"strings"
	"sync"
	"time"

	"github.com/ProtonMail/proton-bridge/v2/internal/imap/uidplus"
	"github.com/ProtonMail/proton-bridge/v2/pkg/message"
	"github.com/ProtonMail/proton-bridge/v2/pkg/parallel"
	"github.com/ProtonMail/proton-bridge/v2/pkg/pmapi"
	"github.com/emersion/go-imap"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

// UpdateMessagesFlags alters flags for the specified message(s).
//
// If the Backend implements Updater, it must notify the client immediately
// via a message update.
func (im *imapMailbox) UpdateMessagesFlags(uid bool, seqSet *imap.SeqSet, operation imap.FlagsOp, flags []string) error {
	return im.logCommand(func() error {
		return im.updateMessagesFlags(uid, seqSet, operation, flags)
	}, "STORE", uid, seqSet, operation, flags)
}

func (im *imapMailbox) updateMessagesFlags(uid bool, seqSet *imap.SeqSet, operation imap.FlagsOp, flags []string) error {
	log.WithFields(logrus.Fields{
		"flags":     flags,
		"operation": operation,
	}).Debug("Updating message flags")

	// Called from go-imap in goroutines - we need to handle panics for each function.
	defer im.panicHandler.HandlePanic()

	im.user.backend.updates.block(im.user.currentAddressLowercase, im.name, operationUpdateMessage)
	defer im.user.backend.updates.unblock(im.user.currentAddressLowercase, im.name, operationUpdateMessage)

	messageIDs, err := im.apiIDsFromSeqSet(uid, seqSet)
	if err != nil || len(messageIDs) == 0 {
		return err
	}

	if operation == imap.SetFlags {
		return im.setFlags(messageIDs, flags)
	}
	return im.addOrRemoveFlags(operation, messageIDs, flags)
}

// setFlags is used for FLAGS command (not +FLAGS or -FLAGS), which means
// to set flags passed as an argument and unset the rest. For example,
// if message is not read, is flagged and is not deleted, call FLAGS \Seen
// should flag message as read, unflagged and keep undeleted.
func (im *imapMailbox) setFlags(messageIDs, flags []string) error { //nolint:funlen
	seen := false
	flagged := false
	deleted := false
	spam := false

	for _, f := range flags {
		switch f {
		case imap.SeenFlag:
			seen = true
		case imap.FlaggedFlag:
			flagged = true
		case imap.DeletedFlag:
			deleted = true
		case message.AppleMailJunkFlag, message.ThunderbirdJunkFlag:
			spam = true
		}
	}

	if seen {
		if err := im.storeMailbox.MarkMessagesRead(messageIDs); err != nil {
			return err
		}
	} else {
		if err := im.storeMailbox.MarkMessagesUnread(messageIDs); err != nil {
			return err
		}
	}

	if flagged {
		if err := im.storeMailbox.MarkMessagesStarred(messageIDs); err != nil {
			return err
		}
	} else {
		if err := im.storeMailbox.MarkMessagesUnstarred(messageIDs); err != nil {
			return err
		}
	}

	if deleted {
		if err := im.storeMailbox.MarkMessagesDeleted(messageIDs); err != nil {
			return err
		}
	} else {
		if err := im.storeMailbox.MarkMessagesUndeleted(messageIDs); err != nil {
			return err
		}
	}

	// Spam should not be taken into action here as Outlook is using FLAGS
	// without preserving junk flag. Probably it's because junk is not standard
	// in the rfc3501 and thus Outlook expects calling FLAGS \Seen will not
	// change the state of junk or other non-standard flags.
	// Still, its safe to label as spam once any client sends the request.
	if spam {
		spamMailbox, err := im.storeAddress.GetMailbox("Spam")
		if err != nil {
			return err
		}
		if err := spamMailbox.LabelMessages(messageIDs); err != nil {
			return err
		}
	}

	return nil
}

func (im *imapMailbox) addOrRemoveFlags(operation imap.FlagsOp, messageIDs, flags []string) error { //nolint:funlen
	for _, f := range flags {
		// Adding flag 'nojunk' is equivalent to removing flag 'junk'
		if (operation == imap.AddFlags) && (f == "nojunk") {
			operation = imap.RemoveFlags
			f = "junk"
		}

		switch f {
		case imap.SeenFlag:
			switch operation { //nolint:exhaustive // imap.SetFlags is processed by im.setFlags
			case imap.AddFlags:
				if err := im.storeMailbox.MarkMessagesRead(messageIDs); err != nil {
					return err
				}
			case imap.RemoveFlags:
				if err := im.storeMailbox.MarkMessagesUnread(messageIDs); err != nil {
					return err
				}
			}
		case imap.FlaggedFlag:
			switch operation { //nolint:exhaustive // imap.SetFlag is processed by im.setFlags
			case imap.AddFlags:
				if err := im.storeMailbox.MarkMessagesStarred(messageIDs); err != nil {
					return err
				}
			case imap.RemoveFlags:
				if err := im.storeMailbox.MarkMessagesUnstarred(messageIDs); err != nil {
					return err
				}
			}
		case imap.DeletedFlag:
			switch operation { //nolint:exhaustive // imap.SetFlag is processed by im.setFlags
			case imap.AddFlags:
				if err := im.storeMailbox.MarkMessagesDeleted(messageIDs); err != nil {
					return err
				}
			case imap.RemoveFlags:
				if err := im.storeMailbox.MarkMessagesUndeleted(messageIDs); err != nil {
					return err
				}
			}
		case imap.AnsweredFlag, imap.DraftFlag, imap.RecentFlag:
			// Not supported.
		case strings.ToLower(message.AppleMailJunkFlag), strings.ToLower(message.ThunderbirdJunkFlag):
			spamMailbox, err := im.storeAddress.GetMailbox("Spam")
			if err != nil {
				return err
			}
			// Handle custom junk flags for Apple Mail and Thunderbird.
			switch operation { //nolint:exhaustive // imap.SetFlag is processed by im.setFlags
			// No label removal is necessary because Spam and Inbox are both exclusive labels so the backend
			// will automatically take care of label removal.
			case imap.AddFlags:
				if err := spamMailbox.LabelMessages(messageIDs); err != nil {
					return err
				}
			case imap.RemoveFlags:
				// During spam flag removal only messages which
				// are in Spam folder should be moved to Inbox.
				// For other messages it is NOOP.
				messagesInSpam := []string{}
				for _, mID := range messageIDs {
					if uid := spamMailbox.GetUIDList([]string{mID}); len(*uid) != 0 {
						messagesInSpam = append(messagesInSpam, mID)
					}
				}
				if len(messagesInSpam) != 0 {
					inboxMailbox, err := im.storeAddress.GetMailbox("INBOX")
					if err != nil {
						return err
					}
					if err := inboxMailbox.LabelMessages(messagesInSpam); err != nil {
						return err
					}
				}
			}
		}
	}

	return nil
}

// CopyMessages copies the specified message(s) to the end of the specified
// destination mailbox. The flags and internal date of the message(s) SHOULD
// be preserved, and the Recent flag SHOULD be set, in the copy.
func (im *imapMailbox) CopyMessages(uid bool, seqSet *imap.SeqSet, targetLabel string) error {
	return im.logCommand(func() error {
		return im.copyMessages(uid, seqSet, targetLabel)
	}, "COPY", uid, seqSet, targetLabel)
}

func (im *imapMailbox) copyMessages(uid bool, seqSet *imap.SeqSet, targetLabel string) error {
	// Called from go-imap in goroutines - we need to handle panics for each function.
	defer im.panicHandler.HandlePanic()

	return im.labelMessages(uid, seqSet, targetLabel, false)
}

// MoveMessages adds dest's label and removes this mailbox' label from each message.
//
// This should not be used until MOVE extension has option to send UIDPLUS
// responses.
func (im *imapMailbox) MoveMessages(uid bool, seqSet *imap.SeqSet, targetLabel string) error {
	return im.logCommand(func() error {
		return im.moveMessages(uid, seqSet, targetLabel)
	}, "MOVE", uid, seqSet, targetLabel)
}

func (im *imapMailbox) moveMessages(uid bool, seqSet *imap.SeqSet, targetLabel string) error {
	// Called from go-imap in goroutines - we need to handle panics for each function.
	defer im.panicHandler.HandlePanic()

	// Moving from All Mail is not allowed.
	if im.storeMailbox.LabelID() == pmapi.AllMailLabel {
		return errors.New("move from All Mail is not allowed")
	}
	return im.labelMessages(uid, seqSet, targetLabel, true)
}

func (im *imapMailbox) labelMessages(uid bool, seqSet *imap.SeqSet, targetLabel string, move bool) error { //nolint:funlen
	messageIDs, err := im.apiIDsFromSeqSet(uid, seqSet)
	if err != nil || len(messageIDs) == 0 {
		return err
	}

	// It is needed to get UID list before LabelingMessages because
	// messages can be removed from source during labeling (e.g. folder1 -> folder2).
	sourceSeqSet := im.storeMailbox.GetUIDList(messageIDs)

	targetStoreMailbox, err := im.storeAddress.GetMailbox(targetLabel)
	if err != nil {
		return err
	}

	// Moving or copying from Inbox to Sent or from Sent to Inbox is no-op.
	// Inbox and Sent is the same mailbox and message is showen in one or
	// the other based on message flags.
	// COPY operation has to be forbidden otherwise move by COPY+EXPUNGE
	// would lead to message found only in All Mail, because COPY is no-op
	// and EXPUNGE is translated as unlabel from the source.
	// MOVE operation could be allowed, just it will do no change. It's better
	// to refuse it as well so client is kept in proper state and no sync
	// is needed.
	isInboxOrSent := func(labelID string) bool {
		return labelID == pmapi.InboxLabel || labelID == pmapi.SentLabel
	}
	if isInboxOrSent(im.storeMailbox.LabelID()) && isInboxOrSent(targetStoreMailbox.LabelID()) {
		if im.storeMailbox.LabelID() == pmapi.InboxLabel {
			return errors.New("move from Inbox to Sent is not allowed")
		}
		return errors.New("move from Sent to Inbox is not allowed")
	}

	deletedIDs := []string{}
	allDeletedIDs, err := im.storeMailbox.GetDeletedAPIIDs()
	if err != nil {
		log.WithError(err).Warn("Problem to get deleted API IDs")
	} else {
		for _, messageID := range messageIDs {
			for _, deletedID := range allDeletedIDs {
				if messageID == deletedID {
					deletedIDs = append(deletedIDs, deletedID)
				}
			}
		}
	}

	// Label messages first to not lose them. If message is only in trash and we unlabel
	// it, it will be removed completely and we cannot label it back.
	if err := targetStoreMailbox.LabelMessages(messageIDs); err != nil {
		return err
	}
	// Folder cannot be unlabeled. Every message has to belong to exactly one folder.
	// In case of labeling message to folder, the original one is implicitly unlabeled.
	// Therefore, we have to unlabel explicitly only if the source mailbox is label.
	if im.storeMailbox.IsLabel() && move {
		if err := im.storeMailbox.UnlabelMessages(messageIDs); err != nil {
			return err
		}
	}

	// Preserve \Deleted flag at target location.
	if len(deletedIDs) > 0 {
		if err := targetStoreMailbox.MarkMessagesDeleted(deletedIDs); err != nil {
			log.WithError(err).Warn("Problem to preserve deleted flag for copied messages")
		}
	}

	targetSeqSet := targetStoreMailbox.GetUIDList(messageIDs)
	return uidplus.CopyResponse(targetStoreMailbox.UIDValidity(), sourceSeqSet, targetSeqSet)
}

// SearchMessages searches messages. The returned list must contain UIDs if
// uid is set to true, or sequence numbers otherwise.
func (im *imapMailbox) SearchMessages(isUID bool, criteria *imap.SearchCriteria) (ids []uint32, err error) { //nolint:gocyclo,funlen
	// Called from go-imap in goroutines - we need to handle panics for each function.
	defer im.panicHandler.HandlePanic()

	if criteria.Not != nil || criteria.Or != nil {
		return nil, errors.New("unsupported search query")
	}

	if criteria.Body != nil || criteria.Text != nil {
		log.Warn("Body and Text criteria not applied")
	}

	var apiIDs []string
	if criteria.SeqNum != nil {
		apiIDs, err = im.apiIDsFromSeqSet(false, criteria.SeqNum)
	} else {
		apiIDs, err = im.storeMailbox.GetAPIIDsFromSequenceRange(1, 0)
	}
	if err != nil {
		return nil, err
	}

	if criteria.Uid != nil {
		apiIDsByUID, err := im.apiIDsFromSeqSet(true, criteria.Uid)
		if err != nil {
			return nil, err
		}
		apiIDs = arrayIntersection(apiIDs, apiIDsByUID)
	}

	for _, apiID := range apiIDs {
		// Get message.
		storeMessage, err := im.storeMailbox.GetMessage(apiID)
		if err != nil {
			log.Warnf("search messages: cannot get message %q from db: %v", apiID, err)
			continue
		}
		m := storeMessage.Message()

		// Filter by time.
		if !criteria.Before.IsZero() {
			if truncated := criteria.Before.Truncate(24 * time.Hour); m.Time > truncated.Unix() {
				continue
			}
		}
		if !criteria.Since.IsZero() {
			if truncated := criteria.Since.Truncate(24 * time.Hour); m.Time < truncated.Unix() {
				continue
			}
		}

		// In order to speed up search it is not needed to always
		// retrieve the fully cached header.
		header := storeMessage.GetMIMEHeaderFast()

		if !criteria.SentBefore.IsZero() || !criteria.SentSince.IsZero() {
			t, err := mail.Header(header).Date()
			if err != nil || t.IsZero() {
				t = time.Unix(m.Time, 0)
			}
			if !criteria.SentBefore.IsZero() {
				if truncated := criteria.SentBefore.Truncate(24 * time.Hour); t.Unix() > truncated.Unix() {
					continue
				}
			}
			if !criteria.SentSince.IsZero() {
				if truncated := criteria.SentSince.Truncate(24 * time.Hour); t.Unix() < truncated.Unix() {
					continue
				}
			}
		}

		// Filter by headers.
		headerMatch := true
		for criteriaKey, criteriaValues := range criteria.Header {
			for _, criteriaValue := range criteriaValues {
				if criteriaValue == "" {
					continue
				}
				switch criteriaKey {
				case "Subject":
					headerMatch = strings.Contains(strings.ToLower(m.Subject), strings.ToLower(criteriaValue))
				case "From":
					headerMatch = addressMatch([]*mail.Address{m.Sender}, criteriaValue)
				case "To":
					headerMatch = addressMatch(m.ToList, criteriaValue)
				case "Cc":
					headerMatch = addressMatch(m.CCList, criteriaValue)
				case "Bcc":
					headerMatch = addressMatch(m.BCCList, criteriaValue)
				default:
					if messageValue := header.Get(criteriaKey); messageValue == "" {
						headerMatch = false // Field is not in header.
					} else if !strings.Contains(strings.ToLower(messageValue), strings.ToLower(criteriaValue)) {
						headerMatch = false // Field is in header but value not matched (case insensitive).
					}
				}
				if !headerMatch {
					break
				}
			}
			if !headerMatch {
				break
			}
		}
		if !headerMatch {
			continue
		}

		// Filter by flags.
		messageFlagsMap := make(map[string]bool)
		if isStringInList(m.LabelIDs, pmapi.StarredLabel) {
			messageFlagsMap[imap.FlaggedFlag] = true
		}
		if !m.Unread {
			messageFlagsMap[imap.SeenFlag] = true
		}
		if m.Has(pmapi.FlagReplied) || m.Has(pmapi.FlagRepliedAll) {
			messageFlagsMap[imap.AnsweredFlag] = true
		}
		if m.Has(pmapi.FlagSent) || m.Has(pmapi.FlagReceived) {
			messageFlagsMap[imap.DraftFlag] = true
		}
		if !m.Has(pmapi.FlagOpened) {
			messageFlagsMap[imap.RecentFlag] = true
		}
		if storeMessage.IsMarkedDeleted() {
			messageFlagsMap[imap.DeletedFlag] = true
		}

		flagMatch := true
		for _, flag := range criteria.WithFlags {
			if !messageFlagsMap[flag] {
				flagMatch = false
				break
			}
		}
		for _, flag := range criteria.WithoutFlags {
			if messageFlagsMap[flag] {
				flagMatch = false
				break
			}
		}
		if !flagMatch {
			continue
		}

		// Filter by size (only if size was already calculated).
		size, err := storeMessage.GetRFC822Size()
		if err != nil {
			return nil, err
		}

		if size > 0 {
			if criteria.Larger != 0 && int64(size) <= int64(criteria.Larger) {
				continue
			}
			if criteria.Smaller != 0 && int64(size) >= int64(criteria.Smaller) {
				continue
			}
		}

		// Add the ID to response.
		var id uint32
		if isUID {
			id, err = storeMessage.UID()
			if err != nil {
				return nil, err
			}
		} else {
			id, err = storeMessage.SequenceNumber()
			if err != nil {
				return nil, err
			}
		}
		ids = append(ids, id)
	}

	return ids, nil
}

// ListMessages returns a list of messages. seqset must be interpreted as UIDs
// if uid is set to true and as message sequence numbers otherwise. See RFC
// 3501 section 6.4.5 for a list of items that can be requested.
//
// Messages must be sent to msgResponse. When the function returns, msgResponse must be closed.
func (im *imapMailbox) ListMessages(isUID bool, seqSet *imap.SeqSet, items []imap.FetchItem, msgResponse chan<- *imap.Message) error {
	return im.logCommand(func() error {
		return im.listMessages(isUID, seqSet, items, msgResponse)
	}, "FETCH", isUID, seqSet, items)
}

func (im *imapMailbox) listMessages(isUID bool, seqSet *imap.SeqSet, items []imap.FetchItem, msgResponse chan<- *imap.Message) (err error) { //nolint:funlen
	defer func() {
		close(msgResponse)
		if err != nil {
			log.Errorf("cannot list messages (%v, %v, %v): %v", isUID, seqSet, items, err)
		}
		// Called from go-imap in goroutines - we need to handle panics for each function.
		im.panicHandler.HandlePanic()
	}()

	if !isUID {
		// EXPUNGE cannot be sent during listing and can come only from
		// the event loop, so we prevent any server side update to avoid
		// the problem.
		im.user.backend.updates.forbidExpunge(im.storeMailbox.LabelID())
		defer im.user.backend.updates.allowExpunge(im.storeMailbox.LabelID())
	}

	var markAsReadIDs []string
	markAsReadMutex := &sync.Mutex{}

	l := log.WithField("cmd", "ListMessages")

	apiIDs, err := im.apiIDsFromSeqSet(isUID, seqSet)
	if err != nil {
		err = fmt.Errorf("list messages seq: %v", err)
		l.WithField("seq", seqSet).Error(err)
		return err
	}

	input := make([]interface{}, len(apiIDs))
	for i, apiID := range apiIDs {
		input[i] = apiID
	}

	processCallback := func(value interface{}) (interface{}, error) {
		apiID := value.(string) //nolint:forcetypeassert // we want to panic here

		storeMessage, err := im.storeMailbox.GetMessage(apiID)
		if err != nil {
			err = fmt.Errorf("list message from db: %v", err)
			l.WithField("apiID", apiID).Error(err)
			return nil, err
		}

		msg, err := im.getMessage(storeMessage, items)
		if err != nil {
			err = fmt.Errorf("list message build: %v", err)
			l.WithField("metaID", storeMessage.ID()).Error(err)
			return nil, err
		}

		if storeMessage.Message().Unread {
			for section := range msg.Body {
				// Peek means get messages without marking them as read.
				// If client does not only ask for peek, we have to mark them as read.
				if !section.Peek {
					markAsReadMutex.Lock()
					markAsReadIDs = append(markAsReadIDs, storeMessage.ID())
					markAsReadMutex.Unlock()
					msg.Flags = append(msg.Flags, imap.SeenFlag)
					break
				}
			}
		}

		return msg, nil
	}

	collectCallback := func(idx int, value interface{}) error {
		msg := value.(*imap.Message) //nolint:forcetypeassert // we want to panic here
		msgResponse <- msg
		return nil
	}

	err = parallel.RunParallel(im.user.backend.listWorkers, input, processCallback, collectCallback)
	if err != nil {
		return err
	}

	if len(markAsReadIDs) > 0 {
		if err := im.storeMailbox.MarkMessagesRead(markAsReadIDs); err != nil {
			l.Warnf("Cannot mark messages as read: %v", err)
		}
	}
	return nil
}

// apiIDsFromSeqSet takes an IMAP sequence set (which can contain either
// sequence numbers or UIDs) and returns all known API IDs in this range.
func (im *imapMailbox) apiIDsFromSeqSet(uid bool, seqSet *imap.SeqSet) ([]string, error) {
	apiIDs := []string{}
	for _, seq := range seqSet.Set {
		var newAPIIDs []string
		var err error
		if uid {
			newAPIIDs, err = im.storeMailbox.GetAPIIDsFromUIDRange(seq.Start, seq.Stop)
		} else {
			newAPIIDs, err = im.storeMailbox.GetAPIIDsFromSequenceRange(seq.Start, seq.Stop)
		}
		if err != nil {
			return []string{}, err
		}
		apiIDs = append(apiIDs, newAPIIDs...)
	}
	if len(apiIDs) == 0 {
		log.Debugf("Requested empty message list: %v %v", uid, seqSet)
	}
	return apiIDs, nil
}

func arrayIntersection(a, b []string) (c []string) {
	m := make(map[string]bool)
	for _, item := range a {
		m[item] = true
	}
	for _, item := range b {
		if _, ok := m[item]; ok {
			c = append(c, item)
		}
	}
	return
}

func isStringInList(list []string, s string) bool {
	for _, v := range list {
		if v == s {
			return true
		}
	}
	return false
}

func addressMatch(addresses []*mail.Address, criteria string) bool {
	for _, addr := range addresses {
		if strings.Contains(strings.ToLower(addr.String()), strings.ToLower(criteria)) {
			return true
		}
	}
	return false
}
