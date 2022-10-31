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

package store

import (
	"github.com/ProtonMail/proton-bridge/v2/pkg/pmapi"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	bolt "go.etcd.io/bbolt"
)

// ErrAllMailOpNotAllowed is error user when user tries to do unsupported
// operation on All Mail folder.
var ErrAllMailOpNotAllowed = errors.New("operation not allowed for 'All Mail' folder")

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
	msg, err := storeMailbox.client().GetMessage(exposeContextForIMAP(), apiID)
	if err != nil {
		return nil, err
	}
	return newStoreMessage(storeMailbox, msg), nil
}

func (storeMailbox *Mailbox) ImportMessage(enc []byte, seen bool, labelIDs []string, flags, time int64) (string, error) {
	defer storeMailbox.pollNow()

	if storeMailbox.labelID != pmapi.AllMailLabel {
		labelIDs = append(labelIDs, storeMailbox.labelID)
	}

	importReqs := &pmapi.ImportMsgReq{
		Metadata: &pmapi.ImportMetadata{
			AddressID: storeMailbox.storeAddress.addressID,
			Unread:    pmapi.Boolean(!seen),
			Flags:     flags,
			Time:      time,
			LabelIDs:  labelIDs,
		},
		Message: append(enc, "\r\n"...),
	}

	res, err := storeMailbox.client().Import(exposeContextForIMAP(), pmapi.ImportMsgReqs{importReqs})
	if err != nil {
		return "", err
	}

	if len(res) == 0 {
		return "", errors.New("no import response")
	}

	return res[0].MessageID, res[0].Error
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
	// Edge case is want to untrash message by drag&drop to AllMail (to not
	// have it in trash but to not delete message forever). IMAP move would
	// work okay but some clients might use COPY&EXPUNGE or APPEND&EXPUNGE.
	// In this case COPY or APPEND is noop because the message is already
	// in All mail. The consequent EXPUNGE would delete message forever.
	if storeMailbox.labelID == pmapi.AllMailLabel {
		return ErrAllMailOpNotAllowed
	}
	defer storeMailbox.pollNow()
	return storeMailbox.client().LabelMessages(exposeContextForIMAP(), apiIDs, storeMailbox.labelID)
}

// UnlabelMessages removes the label by calling an API.
// It has to be propagated to all the same messages in all mailboxes.
// The propagation is processed by the event loop.
func (storeMailbox *Mailbox) UnlabelMessages(apiIDs []string) error {
	storeMailbox.log.WithField("messages", apiIDs).
		Trace("Unlabeling messages")
	if storeMailbox.labelID == pmapi.AllMailLabel {
		return ErrAllMailOpNotAllowed
	}
	defer storeMailbox.pollNow()
	return storeMailbox.client().UnlabelMessages(exposeContextForIMAP(), apiIDs, storeMailbox.labelID)
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
		if message, _ := storeMailbox.store.getMessageFromDB(apiID); message == nil || message.Unread {
			ids = append(ids, apiID)
		}
	}
	if len(ids) == 0 {
		return nil
	}
	return storeMailbox.client().MarkMessagesRead(exposeContextForIMAP(), ids)
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
	return storeMailbox.client().MarkMessagesUnread(exposeContextForIMAP(), apiIDs)
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
	return storeMailbox.client().LabelMessages(exposeContextForIMAP(), apiIDs, pmapi.StarredLabel)
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
	return storeMailbox.client().UnlabelMessages(exposeContextForIMAP(), apiIDs, pmapi.StarredLabel)
}

// MarkMessagesDeleted adds local flag \Deleted. This is not propagated to API
// until RemoveDeleted is called.
func (storeMailbox *Mailbox) MarkMessagesDeleted(apiIDs []string) error {
	log.WithFields(logrus.Fields{
		"messages": apiIDs,
		"label":    storeMailbox.labelID,
		"mailbox":  storeMailbox.Name,
	}).Trace("Marking messages as deleted")
	if storeMailbox.labelID == pmapi.AllMailLabel {
		return ErrAllMailOpNotAllowed
	}
	return storeMailbox.store.db.Update(func(tx *bolt.Tx) error {
		return storeMailbox.txMarkMessagesAsDeleted(tx, apiIDs, true)
	})
}

// MarkMessagesUndeleted removes local flag \Deleted. This is not propagated to
// API.
func (storeMailbox *Mailbox) MarkMessagesUndeleted(apiIDs []string) error {
	log.WithFields(logrus.Fields{
		"messages": apiIDs,
		"label":    storeMailbox.labelID,
		"mailbox":  storeMailbox.Name,
	}).Trace("Marking messages as undeleted")
	if storeMailbox.labelID == pmapi.AllMailLabel {
		return ErrAllMailOpNotAllowed
	}
	return storeMailbox.store.db.Update(func(tx *bolt.Tx) error {
		return storeMailbox.txMarkMessagesAsDeleted(tx, apiIDs, false)
	})
}

// RemoveDeleted sends request to API to remove message from mailbox.
// If the mailbox is All Mail or All Sent, it does nothing.
// If the mailbox is Trash or Spam and message is not in any other mailbox, messages is deleted.
// In all other cases the message is only removed from the mailbox.
// If nil is passed, all messages with \Deleted flag are removed.
// In other cases only messages with \Deleted flag and included in the passed list.
func (storeMailbox *Mailbox) RemoveDeleted(apiIDs []string) error {
	storeMailbox.log.Trace("Deleting messages")

	deletedAPIIDs, err := storeMailbox.GetDeletedAPIIDs()
	if err != nil {
		return err
	}

	if apiIDs == nil {
		apiIDs = deletedAPIIDs
	} else {
		filteredAPIIDs := []string{}
		for _, apiID := range apiIDs {
			found := false
			for _, deletedAPIID := range deletedAPIIDs {
				if apiID == deletedAPIID {
					found = true
					break
				}
			}
			if found {
				filteredAPIIDs = append(filteredAPIIDs, apiID)
			}
		}
		apiIDs = filteredAPIIDs
	}

	if len(apiIDs) == 0 {
		storeMailbox.log.Debug("List to expunge is empty")
		return nil
	}

	defer storeMailbox.pollNow()

	switch storeMailbox.labelID {
	case pmapi.AllMailLabel, pmapi.AllSentLabel:
		break
	case pmapi.TrashLabel, pmapi.SpamLabel:
		if err := storeMailbox.deleteFromTrashOrSpam(apiIDs); err != nil {
			return err
		}
	case pmapi.DraftLabel:
		storeMailbox.log.WithField("ids", apiIDs).Warn("Deleting drafts")
		if err := storeMailbox.client().DeleteMessages(exposeContextForIMAP(), apiIDs); err != nil {
			return err
		}
	default:
		if err := storeMailbox.client().UnlabelMessages(exposeContextForIMAP(), apiIDs, storeMailbox.labelID); err != nil {
			return err
		}
	}
	return nil
}

// deleteFromTrashOrSpam will remove messages from API forever. If messages
// still has some custom label the message will not be deleted. Instead it will
// be removed from Trash or Spam.
func (storeMailbox *Mailbox) deleteFromTrashOrSpam(apiIDs []string) error {
	l := storeMailbox.log.WithField("messages", apiIDs)
	l.Trace("Deleting messages from trash")

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
		if err := storeMailbox.client().UnlabelMessages(exposeContextForIMAP(), messageIDsToUnlabel, storeMailbox.labelID); err != nil {
			l.WithError(err).Warning("Cannot unlabel before deleting")
		}
	}
	if len(messageIDsToDelete) > 0 {
		storeMailbox.log.WithField("ids", messageIDsToDelete).Warn("Deleting messages")
		if err := storeMailbox.client().DeleteMessages(exposeContextForIMAP(), messageIDsToDelete); err != nil {
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
func (storeMailbox *Mailbox) txCreateOrUpdateMessages(tx *bolt.Tx, msgs []*pmapi.Message) error { //nolint:funlen
	shouldSendMailboxUpdate := false

	// Buckets are not initialized right away because it's a heavy operation.
	// The best option is to get the same bucket only once and only when needed.
	var apiBucket, imapBucket, deletedBucket *bolt.Bucket

	// Collect updates to send them later, after possibly sending the status/EXISTS update.
	updates := make([]func(), 0, len(msgs))

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
		if msg.IsDraft() {
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
				if deletedBucket == nil {
					deletedBucket = storeMailbox.txGetDeletedIDsBucket(tx)
				}
				isMarkedAsDeleted := deletedBucket.Get([]byte(msg.ID)) != nil
				if seqErr == nil {
					storeMailbox.store.notifyUpdateMessage(
						storeMailbox.storeAddress.address,
						storeMailbox.labelName,
						btoi(uidb),
						seqNum,
						msg,
						isMarkedAsDeleted,
					)
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

		updates = append(updates, func() {
			storeMailbox.store.notifyUpdateMessage(
				storeMailbox.storeAddress.address,
				storeMailbox.labelName,
				uid,
				seqNum,
				msg,
				false, // new message is never marked as deleted
			)
		})

		shouldSendMailboxUpdate = true
	}

	if shouldSendMailboxUpdate {
		if err := storeMailbox.txMailboxStatusUpdate(tx); err != nil {
			return err
		}
	}

	for _, update := range updates {
		update()
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
	deletedBucket := storeMailbox.txGetDeletedIDsBucket(tx)

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

	if err := deletedBucket.Delete(apiIDb); err != nil {
		return errors.Wrap(err, "cannot delete from mark-as-deleted bucket")
	}

	if seqNumErr == nil {
		storeMailbox.store.notifyDeleteMessage(
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
	total, unread, unreadSeqNum, err := storeMailbox.txGetCounts(tx)
	if err != nil {
		return errors.Wrap(err, "cannot get counts for mailbox status update")
	}
	storeMailbox.store.notifyMailboxStatus(
		storeMailbox.storeAddress.address,
		storeMailbox.labelName,
		total,
		unread,
		unreadSeqNum,
	)
	return nil
}

func (storeMailbox *Mailbox) txMarkMessagesAsDeleted(tx *bolt.Tx, apiIDs []string, markAsDeleted bool) error {
	// Load all buckets before looping over apiIDs
	metaBucket := tx.Bucket(metadataBucket)
	apiBucket := storeMailbox.txGetAPIIDsBucket(tx)
	uidBucket := storeMailbox.txGetIMAPIDsBucket(tx)
	deletedBucket := storeMailbox.txGetDeletedIDsBucket(tx)
	for _, apiID := range apiIDs {
		if markAsDeleted {
			if err := deletedBucket.Put([]byte(apiID), []byte{1}); err != nil {
				return err
			}
		} else {
			if err := deletedBucket.Delete([]byte(apiID)); err != nil {
				return err
			}
		}

		msg, err := storeMailbox.store.txGetMessageFromBucket(metaBucket, apiID)
		if err != nil {
			return err
		}

		uid, err := storeMailbox.txGetUIDFromBucket(apiBucket, apiID)
		if err != nil {
			return err
		}

		seqNum, err := storeMailbox.txGetSequenceNumberOfUID(uidBucket, itob(uid))
		if err != nil {
			return err
		}

		// In order to send flags in format
		// S: * 2 FETCH (FLAGS (\Deleted \Seen))
		storeMailbox.store.notifyUpdateMessage(
			storeMailbox.storeAddress.address,
			storeMailbox.labelName,
			uid,
			seqNum,
			msg,
			markAsDeleted,
		)
	}

	return nil
}
