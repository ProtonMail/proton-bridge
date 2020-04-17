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

package store

import (
	"github.com/ProtonMail/proton-bridge/pkg/pmapi"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	bolt "go.etcd.io/bbolt"
)

// GetMessage returns the `pmapi.Message` struct wrapped in `StoreMessage`
// tied to this mailbox.
func (storeMailbox *Mailbox) GetMessage(apiID string) (*Message, error) {
	msg, err := storeMailbox.store.getMessageFromDB(apiID)
	if err != nil {
		return nil, err
	}
	return newStoreMessage(storeMailbox, msg), nil
}

// FetchMessage fetches the message with the given `apiID`, stores it in the database, and returns a new store message
// wrapping it.
func (storeMailbox *Mailbox) FetchMessage(apiID string) (*Message, error) {
	msg, err := storeMailbox.api().GetMessage(apiID)
	if err != nil {
		return nil, err
	}
	return newStoreMessage(storeMailbox, msg), nil
}

// ImportMessage imports the message by calling an API.
// It has to be propagated to all mailboxes which is done by the event loop.
func (storeMailbox *Mailbox) ImportMessage(msg *pmapi.Message, body []byte, labelIDs []string) error {
	defer storeMailbox.pollNow()

	if storeMailbox.labelID != pmapi.AllMailLabel {
		labelIDs = append(labelIDs, storeMailbox.labelID)
	}

	importReqs := &pmapi.ImportMsgReq{
		AddressID: msg.AddressID,
		Body:      body,
		Unread:    msg.Unread,
		Flags:     msg.Flags,
		Time:      msg.Time,
		LabelIDs:  labelIDs,
	}

	res, err := storeMailbox.api().Import([]*pmapi.ImportMsgReq{importReqs})
	if err == nil && len(res) > 0 {
		msg.ID = res[0].MessageID
	}
	return err
}

// LabelMessages adds the label by calling an API.
// It has to be propagated to all the same messages in all mailboxes.
// The propagation is processed by the event loop.
func (storeMailbox *Mailbox) LabelMessages(apiIDs []string) error {
	log.WithFields(logrus.Fields{
		"messages": apiIDs,
		"label":    storeMailbox.labelID,
		"mailbox":  storeMailbox.Name,
	}).Trace("Labeling messages")
	defer storeMailbox.pollNow()
	return storeMailbox.api().LabelMessages(apiIDs, storeMailbox.labelID)
}

// UnlabelMessages removes the label by calling an API.
// It has to be propagated to all the same messages in all mailboxes.
// The propagation is processed by the event loop.
func (storeMailbox *Mailbox) UnlabelMessages(apiIDs []string) error {
	log.WithFields(logrus.Fields{
		"messages": apiIDs,
		"label":    storeMailbox.labelID,
		"mailbox":  storeMailbox.Name,
	}).Trace("Unlabeling messages")
	defer storeMailbox.pollNow()
	return storeMailbox.api().UnlabelMessages(apiIDs, storeMailbox.labelID)
}

// MarkMessagesRead marks the message read by calling an API.
// It has to be propagated to metadata mailbox which is done by the event loop.
func (storeMailbox *Mailbox) MarkMessagesRead(apiIDs []string) error {
	log.WithFields(logrus.Fields{
		"messages": apiIDs,
		"label":    storeMailbox.labelID,
		"mailbox":  storeMailbox.Name,
	}).Trace("Marking messages as read")
	defer storeMailbox.pollNow()

	// Before deleting a message, TB sets \Seen flag which causes an event update
	// and thus a refresh of the message by deleting and creating it again.
	// TB does not notice this and happily continues with next command to move
	// the message to the Trash but the message does not exist anymore.
	// Therefore we do not issue API update if the message is already read.
	ids := []string{}
	for _, apiID := range apiIDs {
		if message, _ := storeMailbox.store.getMessageFromDB(apiID); message == nil || message.Unread == 1 {
			ids = append(ids, apiID)
		}
	}
	return storeMailbox.api().MarkMessagesRead(ids)
}

// MarkMessagesUnread marks the message unread by calling an API.
// It has to be propagated to metadata mailbox which is done by the event loop.
func (storeMailbox *Mailbox) MarkMessagesUnread(apiIDs []string) error {
	log.WithFields(logrus.Fields{
		"messages": apiIDs,
		"label":    storeMailbox.labelID,
		"mailbox":  storeMailbox.Name,
	}).Trace("Marking messages as unread")
	defer storeMailbox.pollNow()
	return storeMailbox.api().MarkMessagesUnread(apiIDs)
}

// MarkMessagesStarred adds the Starred label by calling an API.
// It has to be propagated to all the same messages in all mailboxes.
// The propagation is processed by the event loop.
func (storeMailbox *Mailbox) MarkMessagesStarred(apiIDs []string) error {
	log.WithFields(logrus.Fields{
		"messages": apiIDs,
		"label":    storeMailbox.labelID,
		"mailbox":  storeMailbox.Name,
	}).Trace("Marking messages as starred")
	defer storeMailbox.pollNow()
	return storeMailbox.api().LabelMessages(apiIDs, pmapi.StarredLabel)
}

// MarkMessagesUnstarred removes the Starred label by calling an API.
// It has to be propagated to all the same messages in all mailboxes.
// The propagation is processed by the event loop.
func (storeMailbox *Mailbox) MarkMessagesUnstarred(apiIDs []string) error {
	log.WithFields(logrus.Fields{
		"messages": apiIDs,
		"label":    storeMailbox.labelID,
		"mailbox":  storeMailbox.Name,
	}).Trace("Marking messages as unstarred")
	defer storeMailbox.pollNow()
	return storeMailbox.api().UnlabelMessages(apiIDs, pmapi.StarredLabel)
}

// DeleteMessages deletes messages.
// If the mailbox is All Mail or All Sent, it does nothing.
// If the mailbox is Trash or Spam and message is not in any other mailbox, messages is deleted.
// In all other cases the message is only removed from the mailbox.
func (storeMailbox *Mailbox) DeleteMessages(apiIDs []string) error {
	log.WithFields(logrus.Fields{
		"messages": apiIDs,
		"label":    storeMailbox.labelID,
		"mailbox":  storeMailbox.Name,
	}).Trace("Deleting messages")
	defer storeMailbox.pollNow()

	switch storeMailbox.labelID {
	case pmapi.AllMailLabel, pmapi.AllSentLabel:
		break
	case pmapi.TrashLabel, pmapi.SpamLabel:
		messageIDsToDelete := []string{}
		messageIDsToUnlabel := []string{}
		for _, apiID := range apiIDs {
			msg, err := storeMailbox.store.getMessageFromDB(apiID)
			if err != nil {
				return err
			}

			otherLabels := false
			// If the message has any custom label, we don't want to delete it, only remove trash/spam label.
			for _, label := range msg.LabelIDs {
				if label != pmapi.SpamLabel && label != pmapi.TrashLabel && label != pmapi.AllMailLabel && label != pmapi.AllSentLabel && label != pmapi.DraftLabel && label != pmapi.AllDraftsLabel {
					otherLabels = true
					break
				}
			}

			if otherLabels {
				messageIDsToUnlabel = append(messageIDsToUnlabel, apiID)
			} else {
				messageIDsToDelete = append(messageIDsToDelete, apiID)
			}
		}
		if len(messageIDsToUnlabel) > 0 {
			if err := storeMailbox.api().UnlabelMessages(messageIDsToUnlabel, storeMailbox.labelID); err != nil {
				log.WithError(err).Warning("Cannot unlabel before deleting")
			}
		}
		if len(messageIDsToDelete) > 0 {
			if err := storeMailbox.api().DeleteMessages(messageIDsToDelete); err != nil {
				return err
			}
		}
	case pmapi.DraftLabel:
		if err := storeMailbox.api().DeleteMessages(apiIDs); err != nil {
			return err
		}
	default:
		if err := storeMailbox.api().UnlabelMessages(apiIDs, storeMailbox.labelID); err != nil {
			return err
		}
	}
	return nil
}

func (storeMailbox *Mailbox) txSkipAndRemoveFromMailbox(tx *bolt.Tx, msg *pmapi.Message) (skipAndRemove bool) {
	defer func() {
		if skipAndRemove {
			if err := storeMailbox.txDeleteMessage(tx, msg.ID); err != nil {
				storeMailbox.log.WithError(err).Error("Cannot remove message")
			}
		}
	}()

	mode, err := storeMailbox.store.getAddressMode()
	if err != nil {
		log.WithError(err).Error("Could not determine address mode")
		return
	}

	skipAndRemove = true

	// If it's split mode and it shouldn't be under this address, it should be skipped and removed.
	if mode == splitMode && storeMailbox.storeAddress.addressID != msg.AddressID {
		return
	}

	// If the message belongs in this mailbox, don't skip/remove it.
	for _, labelID := range msg.LabelIDs {
		if labelID == storeMailbox.labelID {
			skipAndRemove = false
			return
		}
	}

	return skipAndRemove
}

// txCreateOrUpdateMessages will delete, create or update message from mailbox.
func (storeMailbox *Mailbox) txCreateOrUpdateMessages(tx *bolt.Tx, msgs []*pmapi.Message) error { //nolint[funlen]
	shouldSendMailboxUpdate := false

	// Buckets are not initialized right away because it's a heavy operation.
	// The best option is to get the same bucket only once and only when needed.
	var apiBucket, imapBucket *bolt.Bucket
	for _, msg := range msgs {
		if storeMailbox.txSkipAndRemoveFromMailbox(tx, msg) {
			continue
		}

		// Update message.
		if apiBucket == nil {
			apiBucket = storeMailbox.txGetAPIIDsBucket(tx)
		}

		// Draft bodies can change and bodies are not re-fetched by IMAP clients.
		// Every change has to be a new message; we need to delete the old one and always recreate it.
		if msg.Type == pmapi.MessageTypeDraft {
			if err := storeMailbox.txDeleteMessage(tx, msg.ID); err != nil {
				return errors.Wrap(err, "cannot delete old draft")
			}
		} else {
			uidb := apiBucket.Get([]byte(msg.ID))

			if uidb != nil {
				if imapBucket == nil {
					imapBucket = storeMailbox.txGetIMAPIDsBucket(tx)
				}
				seqNum, seqErr := storeMailbox.txGetSequenceNumberOfUID(imapBucket, uidb)
				if seqErr == nil {
					storeMailbox.store.imapUpdateMessage(
						storeMailbox.storeAddress.address,
						storeMailbox.labelName,
						btoi(uidb),
						seqNum,
						msg,
					)
					shouldSendMailboxUpdate = true
				}
				continue
			}
		}

		// Create a new message.
		if imapBucket == nil {
			imapBucket = storeMailbox.txGetIMAPIDsBucket(tx)
		}
		uid, err := storeMailbox.txGetNextUID(imapBucket, true)
		if err != nil {
			return errors.Wrap(err, "cannot generate new UID")
		}
		uidb := itob(uid)

		if err = imapBucket.Put(uidb, []byte(msg.ID)); err != nil {
			return errors.Wrap(err, "cannot add to IMAP bucket")
		}
		if err = apiBucket.Put([]byte(msg.ID), uidb); err != nil {
			return errors.Wrap(err, "cannot add to API bucket")
		}

		seqNum, err := storeMailbox.txGetSequenceNumberOfUID(imapBucket, uidb)
		if err != nil {
			return errors.Wrap(err, "cannot get sequence number from UID")
		}
		storeMailbox.store.imapUpdateMessage(
			storeMailbox.storeAddress.address,
			storeMailbox.labelName,
			uid,
			seqNum,
			msg,
		)
		shouldSendMailboxUpdate = true
	}

	if shouldSendMailboxUpdate {
		if err := storeMailbox.txMailboxStatusUpdate(tx); err != nil {
			return err
		}
	}

	return nil
}

// txDeleteMessage deletes the message from the mailbox bucket.
// and issues message delete and mailbox update changes to updates channel.
func (storeMailbox *Mailbox) txDeleteMessage(tx *bolt.Tx, apiID string) error {
	apiBucket := storeMailbox.txGetAPIIDsBucket(tx)
	apiIDb := []byte(apiID)
	uidb := apiBucket.Get(apiIDb)
	if uidb == nil {
		return nil
	}

	imapBucket := storeMailbox.txGetIMAPIDsBucket(tx)

	seqNum, seqNumErr := storeMailbox.txGetSequenceNumberOfUID(imapBucket, uidb)
	if seqNumErr != nil {
		storeMailbox.log.WithField("apiID", apiID).WithError(seqNumErr).Warn("Cannot get seqNum of deleting message")
	}

	if err := imapBucket.Delete(uidb); err != nil {
		return errors.Wrap(err, "cannot delete from IMAP bucket")
	}

	if err := apiBucket.Delete(apiIDb); err != nil {
		return errors.Wrap(err, "cannot delete from API bucket")
	}

	if seqNumErr == nil {
		storeMailbox.store.imapDeleteMessage(
			storeMailbox.storeAddress.address,
			storeMailbox.labelName,
			seqNum,
		)
		// Outlook for Mac has problems with sending an EXISTS after deleting
		// messages, mostly after moving message to other folder. It causes
		// Outlook to rebuild the whole mailbox. [RFC-3501] says it's not
		// necessary to send an EXISTS response with the new value.
	}
	return nil
}

func (storeMailbox *Mailbox) txMailboxStatusUpdate(tx *bolt.Tx) error {
	total, unread, err := storeMailbox.txGetCounts(tx)
	if err != nil {
		return errors.Wrap(err, "cannot get counts for mailbox status update")
	}
	storeMailbox.store.imapMailboxStatus(
		storeMailbox.storeAddress.address,
		storeMailbox.labelName,
		total,
		unread,
	)
	return nil
}
