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
	"errors"
	"strings"
	"sync"

	"github.com/ProtonMail/proton-bridge/v2/pkg/pmapi"
	imapquota "github.com/emersion/go-imap-quota"
	goIMAPBackend "github.com/emersion/go-imap/backend"
)

type imapUser struct {
	panicHandler panicHandler
	backend      *imapBackend
	user         bridgeUser

	storeUser    storeUserProvider
	storeAddress storeAddressProvider

	currentAddressLowercase string

	// Some clients, for example Outlook, do MOVE by STORE \Deleted, APPEND,
	// EXPUNGE where APPEN and EXPUNGE can go in parallel. Usual IMAP servers
	// do not deduplicate messages and this it's not an issue, but for APPEND
	// for PM means just assigning label. That would cause to assign label and
	// then delete the message, or in other words cause data loss.
	// go-imap does not call CreateMessage till it gets the whole message from
	// IMAP client, therefore with big message, simple wait for APPEND before
	// performing EXPUNGE is not enough. There has to be two-way lock. Only
	// that way even if EXPUNGE is called few ms before APPEND and message
	// is deleted, APPEND will not just assing label but creates the message
	// again.
	// The issue is only when moving message from folder which is causing
	// real removal, so Trash and Spam. Those only need to use the lock to
	// not cause huge slow down as EXPUNGE is implicitly called also after
	// UNSELECT, CLOSE, or LOGOUT.
	appendExpungeLock sync.Mutex

	addressID  string           // cached value for logs to avoid lock
	mailboxIDs safeMapOfStrings // cached values for logs to avoid lock
}

// newIMAPUser returns struct implementing go-imap/user interface.
func newIMAPUser(
	panicHandler panicHandler,
	backend *imapBackend,
	user bridgeUser,
	addressID, address string,
) (*imapUser, error) {
	log.WithField("address", addressID).Debug("Creating new IMAP user")

	storeUser := user.GetStore()
	if storeUser == nil {
		return nil, errors.New("user database is not initialized")
	}

	storeAddress, err := storeUser.GetAddress(addressID)
	if err != nil {
		log.WithField("address", addressID).Debug("Could not get store user address")
		return nil, err
	}

	return &imapUser{
		panicHandler: panicHandler,
		backend:      backend,
		user:         user,

		storeUser:    storeUser,
		storeAddress: storeAddress,

		currentAddressLowercase: strings.ToLower(address),
		addressID:               addressID,
		mailboxIDs:              newSafeMapOfString(),
	}, err
}

// This method should eventually no longer be necessary. Everything should go via store.
func (iu *imapUser) client() pmapi.Client {
	return iu.user.GetClient()
}

func (iu *imapUser) isSubscribed(labelID string) bool {
	subscriptionExceptions := iu.backend.getCacheList(iu.storeUser.UserID(), SubscriptionException)
	exceptions := strings.Split(subscriptionExceptions, ";")

	for _, exception := range exceptions {
		if exception == labelID {
			return false
		}
	}
	return true
}

func (iu *imapUser) removeFromCache(label, value string) {
	iu.backend.removeFromCache(iu.storeUser.UserID(), label, value)
}

func (iu *imapUser) addToCache(label, value string) {
	iu.backend.addToCache(iu.storeUser.UserID(), label, value)
}

// Username returns this user's username.
func (iu *imapUser) Username() string {
	// Called from go-imap in goroutines - we need to handle panics for each function.
	defer iu.panicHandler.HandlePanic()

	return iu.storeAddress.AddressString()
}

// ListMailboxes returns a list of mailboxes belonging to this user.
// If subscribed is set to true, returns only subscribed mailboxes.
func (iu *imapUser) ListMailboxes(showOnlySubcribed bool) ([]goIMAPBackend.Mailbox, error) {
	// Called from go-imap in goroutines - we need to handle panics for each function.
	defer iu.panicHandler.HandlePanic()

	mailboxes := []goIMAPBackend.Mailbox{}
	for _, storeMailbox := range iu.storeAddress.ListMailboxes() {
		iu.mailboxIDs.set(storeMailbox.Name(), storeMailbox.LabelID())

		if storeMailbox.LabelID() == pmapi.AllMailLabel && !iu.backend.bridge.IsAllMailVisible() {
			continue
		}

		if showOnlySubcribed && !iu.isSubscribed(storeMailbox.LabelID()) {
			continue
		}
		mailbox := newIMAPMailbox(iu.panicHandler, iu, storeMailbox)
		mailboxes = append(mailboxes, mailbox)
	}

	mailboxes = append(mailboxes, newLabelsRootMailbox())
	mailboxes = append(mailboxes, newFoldersRootMailbox())

	log.WithField("mailboxes", mailboxes).Trace("Listing mailboxes")

	return mailboxes, nil
}

// GetMailbox returns a mailbox.
func (iu *imapUser) GetMailbox(name string) (mb goIMAPBackend.Mailbox, err error) {
	// Called from go-imap in goroutines - we need to handle panics for each function.
	defer iu.panicHandler.HandlePanic()

	storeMailbox, err := iu.storeAddress.GetMailbox(name)
	if err != nil {
		logMsg := log.WithField("name", name).WithError(err)

		// GODT-97: some clients perform SELECT "" in order to unselect.
		// We don't want to fill the logs with errors in this case.
		if name != "" {
			logMsg.Error("Could not get mailbox")
		} else {
			logMsg.Debug("Failed attempt to get mailbox with empty name")
		}

		return
	}

	return newIMAPMailbox(iu.panicHandler, iu, storeMailbox), nil
}

// CreateMailbox creates a new mailbox.
func (iu *imapUser) CreateMailbox(name string) error {
	// Called from go-imap in goroutines - we need to handle panics for each function.
	defer iu.panicHandler.HandlePanic()

	return iu.storeAddress.CreateMailbox(name)
}

// DeleteMailbox permanently removes the mailbox with the given name.
func (iu *imapUser) DeleteMailbox(name string) (err error) {
	// Called from go-imap in goroutines - we need to handle panics for each function.
	defer iu.panicHandler.HandlePanic()

	storeMailbox, err := iu.storeAddress.GetMailbox(name)
	if err != nil {
		log.WithField("name", name).WithError(err).Error("Could not get mailbox")
		return
	}

	return storeMailbox.Delete()
}

// RenameMailbox changes the name of a mailbox. It is an error to attempt to
// rename a mailbox that does not exist or to rename a mailbox to a name that
// already exists.
func (iu *imapUser) RenameMailbox(oldName, newName string) (err error) {
	// Called from go-imap in goroutines - we need to handle panics for each function.
	defer iu.panicHandler.HandlePanic()

	storeMailbox, err := iu.storeAddress.GetMailbox(oldName)
	if err != nil {
		log.WithField("name", oldName).WithError(err).Error("Could not get mailbox")
		return
	}

	return storeMailbox.Rename(newName)
}

// Logout is called when this User will no longer be used, likely because the
// client closed the connection.
func (iu *imapUser) Logout() (err error) {
	// Called from go-imap in goroutines - we need to handle panics for each function.
	defer iu.panicHandler.HandlePanic()

	log.Debug("IMAP client logged out address ", iu.storeAddress.AddressID())

	iu.backend.deleteUser(iu.currentAddressLowercase)

	return nil
}

func (iu *imapUser) GetQuota(name string) (*imapquota.Status, error) {
	// Called from go-imap in goroutines - we need to handle panics for each function.
	defer iu.panicHandler.HandlePanic()

	usedSpace, maxSpace, err := iu.storeUser.GetSpaceKB()
	if err != nil {
		log.Error("Failed getting quota: ", err)
		return nil, err
	}

	resources := make(map[string][2]uint32)
	var list [2]uint32
	list[0] = usedSpace
	list[1] = maxSpace
	resources[imapquota.ResourceStorage] = list
	status := &imapquota.Status{
		Name:      "",
		Resources: resources,
	}

	return status, nil
}

func (iu *imapUser) SetQuota(name string, resources map[string]uint32) error {
	// Called from go-imap in goroutines - we need to handle panics for each function.
	defer iu.panicHandler.HandlePanic()

	return errors.New("quota cannot be set")
}

func (iu *imapUser) CreateMessageLimit() *uint32 {
	// Called from go-imap in goroutines - we need to handle panics for each function.
	defer iu.panicHandler.HandlePanic()

	maxUpload, err := iu.storeUser.GetMaxUpload()
	if err != nil {
		log.Error("Failed getting current user for message limit: ", err)
		zero := uint32(0)
		return &zero
	}

	upload := uint32(maxUpload)
	return &upload
}
