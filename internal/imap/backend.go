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

// Package imap provides IMAP server of the Bridge.
package imap

import (
	"strings"
	"sync"
	"time"

	imapid "github.com/ProtonMail/go-imap-id"
	"github.com/ProtonMail/proton-bridge/internal/bridge"
	"github.com/ProtonMail/proton-bridge/internal/events"
	"github.com/ProtonMail/proton-bridge/pkg/listener"
	"github.com/emersion/go-imap"
	goIMAPBackend "github.com/emersion/go-imap/backend"
	"github.com/sirupsen/logrus"
)

type panicHandler interface {
	HandlePanic()
}

type imapBackend struct {
	panicHandler  panicHandler
	bridge        bridger
	updates       chan goIMAPBackend.Update
	eventListener listener.Listener

	users       map[string]*imapUser
	usersLocker sync.Locker

	lastMailClient       imapid.ID
	lastMailClientLocker sync.Locker

	imapCache     map[string]map[string]string
	imapCachePath string
	imapCacheLock *sync.RWMutex
}

// NewIMAPBackend returns struct implementing go-imap/backend interface.
func NewIMAPBackend(
	panicHandler panicHandler,
	eventListener listener.Listener,
	cfg configProvider,
	bridge *bridge.Bridge,
) *imapBackend { //nolint[golint]
	bridgeWrap := newBridgeWrap(bridge)
	backend := newIMAPBackend(panicHandler, cfg, bridgeWrap, eventListener)

	// We want idle updates coming from bridge's updates channel (which in turn come
	// from the bridge users' stores) to be sent to the imap backend's update channel.
	backend.updates = bridge.GetIMAPUpdatesChannel()

	go backend.monitorDisconnectedUsers()

	return backend
}

func newIMAPBackend(
	panicHandler panicHandler,
	cfg configProvider,
	bridge bridger,
	eventListener listener.Listener,
) *imapBackend {
	return &imapBackend{
		panicHandler:  panicHandler,
		bridge:        bridge,
		updates:       make(chan goIMAPBackend.Update),
		eventListener: eventListener,

		users:       map[string]*imapUser{},
		usersLocker: &sync.Mutex{},

		lastMailClient:       imapid.ID{imapid.FieldName: clientNone},
		lastMailClientLocker: &sync.Mutex{},

		imapCachePath: cfg.GetIMAPCachePath(),
		imapCacheLock: &sync.RWMutex{},
	}
}

func (ib *imapBackend) getUser(address string) (*imapUser, error) {
	ib.usersLocker.Lock()
	defer ib.usersLocker.Unlock()

	address = strings.ToLower(address)
	imapUser, ok := ib.users[address]
	if ok {
		return imapUser, nil
	}
	return ib.createUser(address)
}

// createUser require that address MUST be in lowercase.
func (ib *imapBackend) createUser(address string) (*imapUser, error) {
	log.WithField("address", address).Debug("Creating new IMAP user")

	user, err := ib.bridge.GetUser(address)
	if err != nil {
		return nil, err
	}

	// Make sure you return the same user for all valid addresses when in combined mode.
	if user.IsCombinedAddressMode() {
		address = strings.ToLower(user.GetPrimaryAddress())
		if combinedUser, ok := ib.users[address]; ok {
			return combinedUser, nil
		}
	}

	// Client can log in only using address so we can properly close all IMAP connections.
	var addressID string
	if addressID, err = user.GetAddressID(address); err != nil {
		return nil, err
	}

	newUser, err := newIMAPUser(ib.panicHandler, ib, user, addressID, address)
	if err != nil {
		return nil, err
	}

	ib.users[address] = newUser

	return newUser, nil
}

// deleteUser removes a user from the users map.
// This is a safe operation even if the user doesn't exist so it is no problem if it is done twice.
func (ib *imapBackend) deleteUser(address string) {
	log.WithField("address", address).Debug("Deleting IMAP user")

	ib.usersLocker.Lock()
	defer ib.usersLocker.Unlock()

	delete(ib.users, strings.ToLower(address))
}

// Login authenticates a user.
func (ib *imapBackend) Login(_ *imap.ConnInfo, username, password string) (goIMAPBackend.User, error) {
	// Called from go-imap in goroutines - we need to handle panics for each function.
	defer ib.panicHandler.HandlePanic()

	imapUser, err := ib.getUser(username)
	if err != nil {
		log.WithError(err).Warn("Cannot get user")
		return nil, err
	}

	if err := imapUser.user.CheckBridgeLogin(password); err != nil {
		log.WithError(err).Error("Could not check bridge password")
		_ = imapUser.Logout()
		// Apple Mail sometimes generates a lot of requests very quickly.
		// It's therefore good to have a timeout after a bad login so that we can slow
		// those requests down a little bit.
		time.Sleep(10 * time.Second)
		return nil, err
	}

	// The update channel should be nil until we try to login to IMAP for the first time
	// so that it doesn't make bridge slow for users who are only using bridge for SMTP
	// (otherwise the store will be locked for 1 sec per email during synchronization).
	imapUser.user.SetIMAPIdleUpdateChannel()

	return imapUser, nil
}

// Updates returns a channel of updates for IMAP IDLE extension.
func (ib *imapBackend) Updates() <-chan goIMAPBackend.Update {
	// Called from go-imap in goroutines - we need to handle panics for each function.
	defer ib.panicHandler.HandlePanic()

	return ib.updates
}

func (ib *imapBackend) CreateMessageLimit() *uint32 {
	return nil
}

func (ib *imapBackend) setLastMailClient(id imapid.ID) {
	ib.lastMailClientLocker.Lock()
	defer ib.lastMailClientLocker.Unlock()

	if name, ok := id[imapid.FieldName]; ok && ib.lastMailClient[imapid.FieldName] != name {
		ib.lastMailClient = imapid.ID{}
		for k, v := range id {
			ib.lastMailClient[k] = v
		}
		log.Warn("Mail Client ID changed to ", ib.lastMailClient)
		ib.bridge.SetCurrentClient(
			ib.lastMailClient[imapid.FieldName],
			ib.lastMailClient[imapid.FieldVersion],
		)
	}
}

// monitorDisconnectedUsers removes users when it receives a close connection event for them.
func (ib *imapBackend) monitorDisconnectedUsers() {
	ch := make(chan string)
	ib.eventListener.Add(events.CloseConnectionEvent, ch)

	for address := range ch {
		// delete the user to ensure future imap login attempts use the latest bridge user
		// (bridge user might be removed-readded so we want to use the new bridge user object).
		ib.deleteUser(address)
	}
}

func (ib *imapBackend) upgradeError(err error) {
	logrus.WithError(err).Error("IMAP connection couldn't be upgraded to TLS during STARTTLS")

	if strings.Contains(err.Error(), "remote error: tls: bad certificate") {
		logrus.Info("TODO: Show troubleshooting popup")
	}
}
