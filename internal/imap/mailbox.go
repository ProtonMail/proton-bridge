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
	"strings"

	"github.com/ProtonMail/proton-bridge/pkg/message"
	"github.com/ProtonMail/proton-bridge/pkg/pmapi"
	"github.com/emersion/go-imap"
	specialuse "github.com/emersion/go-imap-specialuse"
	"github.com/sirupsen/logrus"
)

type imapMailbox struct {
	panicHandler panicHandler
	user         *imapUser
	name         string

	log *logrus.Entry

	storeUser    storeUserProvider
	storeAddress storeAddressProvider
	storeMailbox storeMailboxProvider
}

// newIMAPMailbox returns struct implementing go-imap/mailbox interface.
func newIMAPMailbox(panicHandler panicHandler, user *imapUser, storeMailbox storeMailboxProvider) *imapMailbox {
	return &imapMailbox{
		panicHandler: panicHandler,
		user:         user,
		name:         storeMailbox.Name(),

		log: log.
			WithField("addressID", user.storeAddress.AddressID()).
			WithField("userID", user.storeUser.UserID()).
			WithField("labelID", storeMailbox.LabelID()),

		storeUser:    user.storeUser,
		storeAddress: user.storeAddress,
		storeMailbox: storeMailbox,
	}
}

// Name returns this mailbox name.
func (im *imapMailbox) Name() string {
	// Called from go-imap in goroutines - we need to handle panics for each function.
	defer im.panicHandler.HandlePanic()

	return im.name
}

// Info returns this mailbox info.
func (im *imapMailbox) Info() (*imap.MailboxInfo, error) {
	// Called from go-imap in goroutines - we need to handle panics for each function.
	defer im.panicHandler.HandlePanic()

	info := &imap.MailboxInfo{
		Attributes: im.getFlags(),
		Delimiter:  im.storeMailbox.GetDelimiter(),
		Name:       im.name,
	}

	return info, nil
}

func (im *imapMailbox) getFlags() []string {
	flags := []string{}
	if !im.storeMailbox.IsFolder() || im.storeMailbox.IsSystem() {
		flags = append(flags, imap.NoInferiorsAttr) // Subfolders are not supported for System or Label
	}
	switch im.storeMailbox.LabelID() {
	case pmapi.SentLabel:
		flags = append(flags, specialuse.Sent)
	case pmapi.TrashLabel:
		flags = append(flags, specialuse.Trash)
	case pmapi.SpamLabel:
		flags = append(flags, specialuse.Junk)
	case pmapi.ArchiveLabel:
		flags = append(flags, specialuse.Archive)
	case pmapi.AllMailLabel:
		flags = append(flags, specialuse.All)
	case pmapi.DraftLabel:
		flags = append(flags, specialuse.Drafts)
	}

	return flags
}

// Status returns this mailbox status. The fields Name, Flags and
// PermanentFlags in the returned MailboxStatus must be always populated. This
// function does not affect the state of any messages in the mailbox. See RFC
// 3501 section 6.3.10 for a list of items that can be requested.
//
// It always returns the state of DB (which could be different to server status).
// Additionally it checks that all stored numbers are same as in DB and polls events if needed.
func (im *imapMailbox) Status(items []imap.StatusItem) (*imap.MailboxStatus, error) {
	// Called from go-imap in goroutines - we need to handle panics for each function.
	defer im.panicHandler.HandlePanic()

	l := log.WithField("status-label", im.storeMailbox.LabelID())
	l.Data["user"] = im.storeUser.UserID()
	l.Data["address"] = im.storeAddress.AddressID()
	status := imap.NewMailboxStatus(im.name, items)
	status.UidValidity = im.storeMailbox.UIDValidity()
	status.Flags = []string{
		imap.SeenFlag, strings.ToUpper(imap.SeenFlag),
		imap.FlaggedFlag, strings.ToUpper(imap.FlaggedFlag),
		imap.DeletedFlag, strings.ToUpper(imap.DeletedFlag),
		imap.DraftFlag, strings.ToUpper(imap.DraftFlag),
		message.AppleMailJunkFlag,
		message.ThunderbirdJunkFlag,
		message.ThunderbirdNonJunkFlag,
	}
	status.PermanentFlags = append([]string{}, status.Flags...)

	dbTotal, dbUnread, dbUnreadSeqNum, err := im.storeMailbox.GetCounts()
	l.WithFields(logrus.Fields{
		"total":        dbTotal,
		"unread":       dbUnread,
		"unreadSeqNum": dbUnreadSeqNum,
		"err":          err,
	}).Debug("DB counts")
	if err == nil {
		status.Messages = uint32(dbTotal)
		status.Unseen = uint32(dbUnread)
		status.UnseenSeqNum = uint32(dbUnreadSeqNum)
	}

	if status.UidNext, err = im.storeMailbox.GetNextUID(); err != nil {
		return nil, err
	}

	return status, nil
}

// SetSubscribed adds or removes the mailbox to the server's set of "active"
// or "subscribed" mailboxes.
func (im *imapMailbox) SetSubscribed(subscribed bool) error {
	// Called from go-imap in goroutines - we need to handle panics for each function.
	defer im.panicHandler.HandlePanic()

	label := im.storeMailbox.LabelID()
	if subscribed && !im.user.isSubscribed(label) {
		im.user.removeFromCache(SubscriptionException, label)
	}
	if !subscribed && im.user.isSubscribed(label) {
		im.user.addToCache(SubscriptionException, label)
	}
	return nil
}

// Check requests a checkpoint of the currently selected mailbox. A checkpoint
// refers to any implementation-dependent housekeeping associated with the
// mailbox (e.g., resolving the server's in-memory state of the mailbox with
// the state on its disk). A checkpoint MAY take a non-instantaneous amount of
// real time to complete. If a server implementation has no such housekeeping
// considerations, CHECK is equivalent to NOOP.
func (im *imapMailbox) Check() error {
	return nil
}

// Expunge permanently removes all messages that have the \Deleted flag set
// from the currently selected mailbox.
func (im *imapMailbox) Expunge() error {
	im.user.backend.setUpdatesBeBlocking(im.user.currentAddressLowercase, im.name, operationDeleteMessage)
	defer im.user.backend.unsetUpdatesBeBlocking(im.user.currentAddressLowercase, im.name, operationDeleteMessage)

	return im.storeMailbox.RemoveDeleted()
}

func (im *imapMailbox) ListQuotas() ([]string, error) {
	return []string{""}, nil
}
