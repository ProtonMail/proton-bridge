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
	"strings"

	"github.com/ProtonMail/proton-bridge/internal/bridge"
	imapquota "github.com/emersion/go-imap-quota"
	goIMAPBackend "github.com/emersion/go-imap/backend"
)

var (
	errNoSuchMailbox = errors.New("no such mailbox") //nolint[gochecknoglobals]
)

type imapUser struct {
	panicHandler panicHandler
	backend      *imapBackend
	user         bridgeUser
	client       bridge.PMAPIProvider

	storeUser    storeUserProvider
	storeAddress storeAddressProvider

	currentAddressLowercase string
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

	client := user.GetTemporaryPMAPIClient()

	return &imapUser{
		panicHandler: panicHandler,
		backend:      backend,
		user:         user,
		client:       client,

		storeUser:    storeUser,
		storeAddress: storeAddress,

		currentAddressLowercase: strings.ToLower(address),
	}, err
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

// GetMailbox returns a mailbox. If it doesn't exist, it returns ErrNoSuchMailbox.
func (iu *imapUser) GetMailbox(name string) (mb goIMAPBackend.Mailbox, err error) {
	// Called from go-imap in goroutines - we need to handle panics for each function.
	defer iu.panicHandler.HandlePanic()

	storeMailbox, err := iu.storeAddress.GetMailbox(name)
	if err != nil {
		log.WithField("name", name).WithError(err).Error("Could not get mailbox")
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

	usedSpace, maxSpace, err := iu.storeUser.GetSpace()
	if err != nil {
		log.Error("Failed getting quota: ", err)
		return nil, err
	}

	resources := make(map[string][2]uint32)
	var list [2]uint32
	list[0] = uint32(usedSpace / 1000)
	list[1] = uint32(maxSpace / 1000)
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
