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

package imap

import (
	"errors"
	"fmt"
	"net/mail"
	"strings"
	"sync"
	"time"

	"github.com/ProtonMail/proton-bridge/internal/imap/uidplus"
	"github.com/ProtonMail/proton-bridge/pkg/message"
	"github.com/ProtonMail/proton-bridge/pkg/parallel"
	"github.com/ProtonMail/proton-bridge/pkg/pmapi"
	"github.com/emersion/go-imap"
	"github.com/sirupsen/logrus"
)

// UpdateMessagesFlags alters flags for the specified message(s).
//
// If the Backend implements Updater, it must notify the client immediately
// via a message update.
func (im *imapMailbox) UpdateMessagesFlags(uid bool, seqSet *imap.SeqSet, operation imap.FlagsOp, flags []string) error {
	log.WithFields(logrus.Fields{
		"flags":     flags,
		"operation": operation,
	}).Debug("Updating message flags")

	// Called from go-imap in goroutines - we need to handle panics for each function.
	defer im.panicHandler.HandlePanic()

	messageIDs, err := im.apiIDsFromSeqSet(uid, seqSet)
	if err != nil || len(messageIDs) == 0 {
		return err
	}

	for _, f := range flags {
		switch f {
		case imap.SeenFlag:
			switch operation {
			case imap.SetFlags, imap.AddFlags:
				_ = im.storeMailbox.MarkMessagesRead(messageIDs)
			case imap.RemoveFlags:
				_ = im.storeMailbox.MarkMessagesUnread(messageIDs)
			}
		case imap.FlaggedFlag:
			switch operation {
			case imap.SetFlags, imap.AddFlags:
				_ = im.storeMailbox.MarkMessagesStarred(messageIDs)
			case imap.RemoveFlags:
				_ = im.storeMailbox.MarkMessagesUnstarred(messageIDs)
			}
		case imap.DeletedFlag:
			if operation == imap.RemoveFlags {
				break // Nothing to do, no message has the \Deleted flag.
			}
			_ = im.storeMailbox.DeleteMessages(messageIDs)
		case imap.AnsweredFlag, imap.DraftFlag, imap.RecentFlag:
			// Not supported.
		case message.AppleMailJunkFlag, message.ThunderbirdJunkFlag:
			storeMailbox, err := im.storeAddress.GetMailbox(pmapi.SpamLabel)
			if err != nil {
				return err
			}

			// Handle custom junk flags for Apple Mail and Thunderbird.
			switch operation {
			// No label removal is necessary because Spam and Inbox are both exclusive labels so the backend
			// will automatically take care of label removal.
			case imap.SetFlags, imap.AddFlags:
				_ = storeMailbox.LabelMessages(messageIDs)
			case imap.RemoveFlags:
				_ = storeMailbox.UnlabelMessages(messageIDs)
			}
		}
	}

	return nil
}

// CopyMessages copies the specified message(s) to the end of the specified
// destination mailbox. The flags and internal date of the message(s) SHOULD
// be preserved, and the Recent flag SHOULD be set, in the copy.
func (im *imapMailbox) CopyMessages(uid bool, seqSet *imap.SeqSet, targetLabel string) error {
	// Called from go-imap in goroutines - we need to handle panics for each function.
	defer im.panicHandler.HandlePanic()

	return im.labelMessages(uid, seqSet, targetLabel, false)
}

// MoveMessages adds dest's label and removes this mailbox' label from each message.
//
// This should not be used until MOVE extension has option to send UIDPLUS
// responses.
func (im *imapMailbox) MoveMessages(uid bool, seqSet *imap.SeqSet, targetLabel string) error {
	// Called from go-imap in goroutines - we need to handle panics for each function.
	defer im.panicHandler.HandlePanic()

	return im.labelMessages(uid, seqSet, targetLabel, true)
}

func (im *imapMailbox) labelMessages(uid bool, seqSet *imap.SeqSet, targetLabel string, move bool) error {
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

	// Label messages first to not lose them. If message is only in trash and we unlabel
	// it, it will be removed completely and we cannot label it back.
	if err := targetStoreMailbox.LabelMessages(messageIDs); err != nil {
		return err
	}
	if move {
		if err := im.storeMailbox.UnlabelMessages(messageIDs); err != nil {
			return err
		}
	}

	targetSeqSet := targetStoreMailbox.GetUIDList(messageIDs)
	return uidplus.CopyResponse(targetStoreMailbox.UIDValidity(), sourceSeqSet, targetSeqSet)
}

// SearchMessages searches messages. The returned list must contain UIDs if
// uid is set to true, or sequence numbers otherwise.
func (im *imapMailbox) SearchMessages(isUID bool, criteria *imap.SearchCriteria) (ids []uint32, err error) { //nolint[gocyclo]
	// Called from go-imap in goroutines - we need to handle panics for each function.
	defer im.panicHandler.HandlePanic()

	if criteria.Not != nil || criteria.Or != nil {
		return nil, errors.New("unsupported search query")
	}

	if criteria.Body != nil || criteria.Text != nil {
		log.Warn("Body and Text criteria not applied.")
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

	var apiIDsFromUID []string
	if criteria.Uid != nil {
		if apiIDs, err := im.apiIDsFromSeqSet(true, criteria.Uid); err == nil {
			apiIDsFromUID = append(apiIDsFromUID, apiIDs...)
		}
	}

	// Apply filters.
	for _, apiID := range apiIDs {
		// Filter on UIDs.
		if len(apiIDsFromUID) > 0 && !isStringInList(apiIDsFromUID, apiID) {
			continue
		}

		// Get message.
		storeMessage, err := im.storeMailbox.GetMessage(apiID)
		if err != nil {
			log.Warnf("search messages: cannot get message %q from db: %v", apiID, err)
			continue
		}
		m := storeMessage.Message()

		// Filter addresses.
		/*if criteria.From != "" && !addressMatch([]*mail.Address{m.Sender}, criteria.From) {
			continue
		}
		if criteria.To != "" && !addressMatch(m.ToList, criteria.To) {
			continue
		}
		if criteria.Cc != "" && !addressMatch(m.CCList, criteria.Cc) {
			continue
		}
		if criteria.Bcc != "" && !addressMatch(m.BCCList, criteria.Bcc) {
			continue
		}*/

		// Filter strings.
		/*if criteria.Subject != "" && !strings.Contains(strings.ToLower(m.Subject), strings.ToLower(criteria.Subject)) {
			continue
		}
		if criteria.Keyword != "" && !hasKeyword(m, criteria.Keyword) {
			continue
		}
		if criteria.Unkeyword != "" && hasKeyword(m, criteria.Unkeyword) {
			continue
		}
		if criteria.Header[0] != "" {
			h := message.GetHeader(m)
			if val := h.Get(criteria.Header[0]); val == "" {
				continue // Field is not in header.
			} else if criteria.Header[1] != "" && !strings.Contains(strings.ToLower(val), strings.ToLower(criteria.Header[1])) {
				continue // Field is in header, second criteria is non-zero and field value not matched (case insensitive).
			}
		}

		// Filter flags.
		if criteria.Flagged && !isStringInList(m.LabelIDs, pmapi.StarredLabel) {
			continue
		}
		if criteria.Unflagged && isStringInList(m.LabelIDs, pmapi.StarredLabel) {
			continue
		}
		if criteria.Seen && m.Unread == 1 {
			continue
		}
		if criteria.Unseen && m.Unread == 0 {
			continue
		}
		if criteria.Deleted {
			continue
		}
		// if criteria.Undeleted { // All messages matches this criteria }
		if criteria.Draft && (m.Has(pmapi.FlagSent) || m.Has(pmapi.FlagReceived)) {
			continue
		}
		if criteria.Undraft && !(m.Has(pmapi.FlagSent) || m.Has(pmapi.FlagReceived)) {
			continue
		}
		if criteria.Answered && !(m.Has(pmapi.FlagReplied) || m.Has(pmapi.FlagRepliedAll)) {
			continue
		}
		if criteria.Unanswered && (m.Has(pmapi.FlagReplied) || m.Has(pmapi.FlagRepliedAll)) {
			continue
		}
		if criteria.Recent && m.Has(pmapi.FlagOpened) { // opened means not recent
			continue
		}
		if criteria.Old && !m.Has(pmapi.FlagOpened) {
			continue
		}
		if criteria.New && !(!m.Has(pmapi.FlagOpened) && m.Unread == 1) {
			continue
		}*/

		// Filter internal date.
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
		/*if !criteria.On.IsZero() {
			truncated := criteria.On.Truncate(24 * time.Hour)
			if m.Time < truncated.Unix() || m.Time > truncated.Add(24*time.Hour).Unix() {
				continue
			}
		}*/
		if !(criteria.SentBefore.IsZero() && criteria.SentSince.IsZero() /*&& criteria.SentOn.IsZero()*/) {
			if t, err := m.Header.Date(); err == nil && !t.IsZero() {
				// Filter header date.
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
				/*if !criteria.SentOn.IsZero() {
					truncated := criteria.SentOn.Truncate(24 * time.Hour)
					if t.Unix() < truncated.Unix() || t.Unix() > truncated.Add(24*time.Hour).Unix() {
						continue
					}
				}*/
			}
		}

		// Filter size (only if size was already calculated).
		if m.Size > 0 {
			if criteria.Larger != 0 && m.Size <= int64(criteria.Larger) {
				continue
			}
			if criteria.Smaller != 0 && m.Size >= int64(criteria.Smaller) {
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
func (im *imapMailbox) ListMessages(isUID bool, seqSet *imap.SeqSet, items []imap.FetchItem, msgResponse chan<- *imap.Message) (err error) { //nolint[funlen]
	defer func() {
		close(msgResponse)
		if err != nil {
			log.Errorf("cannot list messages (%v, %v, %v): %v", isUID, seqSet, items, err)
		}
		// Called from go-imap in goroutines - we need to handle panics for each function.
		im.panicHandler.HandlePanic()
	}()

	var markAsReadIDs []string
	markAsReadMutex := &sync.Mutex{}

	l := log.WithField("cmd", "ListMessages")

	apiIDs, err := im.apiIDsFromSeqSet(isUID, seqSet)
	if err != nil {
		err = fmt.Errorf("list messages seq: %v", err)
		l.WithField("seq", seqSet).Error(err)
		return err
	}

	// From RFC: UID range of 559:* always includes the UID of the last message
	// in the mailbox, even if 559 is higher than any assigned UID value.
	// See: https://tools.ietf.org/html/rfc3501#page-61
	if isUID && seqSet.Dynamic() && len(apiIDs) == 0 {
		l.Debug("Requesting empty UID dynamic fetch, adding latest message")
		apiID, err := im.storeMailbox.GetLatestAPIID()
		if err != nil {
			return nil
		}
		apiIDs = []string{apiID}
	}

	input := make([]interface{}, len(apiIDs))
	for i, apiID := range apiIDs {
		input[i] = apiID
	}

	processCallback := func(value interface{}) (interface{}, error) {
		apiID := value.(string)

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

		if storeMessage.Message().Unread == 1 {
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
		msg := value.(*imap.Message)
		msgResponse <- msg
		return nil
	}

	err = parallel.RunParallel(fetchMessagesWorkers, input, processCallback, collectCallback)
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

func isAddressInList(addrs []*mail.Address, query string) bool { //nolint[deadcode]
	for _, addr := range addrs {
		if strings.Contains(addr.Address, query) || strings.Contains(addr.Name, query) {
			return true
		}
	}
	return false
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

func hasKeyword(m *pmapi.Message, keyword string) bool {
	for _, v := range message.GetHeader(m) {
		if strings.Contains(strings.ToLower(strings.Join(v, " ")), strings.ToLower(keyword)) {
			return true
		}
	}
	return false
}
