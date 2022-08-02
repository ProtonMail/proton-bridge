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

// Package imap provides IMAP server of the Bridge.
//
// Methods are called by the go-imap library in parallel.
// Additional parallelism is achieved while handling each IMAP request.
//
// For example, ListMessages internally uses `fetchWorkers` workers to resolve each requested item.
// When IMAP clients request message literals (or parts thereof), we sometimes need to build RFC822 message literals.
// To do this, we pass build jobs to the message builder, which internally manages its own parallelism.
// Summary:
//  - each IMAP fetch request is handled in parallel,
//  - within each IMAP fetch request, individual items are handled by a pool of `fetchWorkers` workers,
//  - within each worker, build jobs are posted to the message builder,
//  - the message builder handles build jobs using its own, independent worker pool,
// The builder will handle jobs in parallel up to its own internal limit. This prevents it from overwhelming API.
package imap

import (
	"strings"
	"sync"
	"time"

	"github.com/ProtonMail/proton-bridge/v2/internal/bridge"
	"github.com/ProtonMail/proton-bridge/v2/internal/config/settings"
	"github.com/ProtonMail/proton-bridge/v2/internal/events"
	"github.com/ProtonMail/proton-bridge/v2/internal/users"
	"github.com/ProtonMail/proton-bridge/v2/pkg/listener"
	"github.com/emersion/go-imap"
	goIMAPBackend "github.com/emersion/go-imap/backend"
)

type panicHandler interface {
	HandlePanic()
}

type imapBackend struct {
	panicHandler  panicHandler
	bridge        bridger
	updates       *imapUpdates
	eventListener listener.Listener
	listWorkers   int

	users       map[string]*imapUser
	usersLocker sync.Locker

	imapCache     map[string]map[string]string
	imapCachePath string
	imapCacheLock *sync.RWMutex
}

type settingsProvider interface {
	GetInt(string) int
}

// NewIMAPBackend returns struct implementing go-imap/backend interface.
func NewIMAPBackend(
	panicHandler panicHandler,
	eventListener listener.Listener,
	cache cacheProvider,
	setting settingsProvider,
	bridge *bridge.Bridge,
) *imapBackend { //nolint:revive
	bridgeWrap := newBridgeWrap(bridge)

	imapWorkers := setting.GetInt(settings.IMAPWorkers)
	backend := newIMAPBackend(panicHandler, cache, bridgeWrap, eventListener, imapWorkers)

	go backend.monitorDisconnectedUsers()

	return backend
}

func newIMAPBackend(
	panicHandler panicHandler,
	cache cacheProvider,
	bridge bridger,
	eventListener listener.Listener,
	listWorkers int,
) *imapBackend {
	ib := &imapBackend{
		panicHandler:  panicHandler,
		bridge:        bridge,
		eventListener: eventListener,

		users:       map[string]*imapUser{},
		usersLocker: &sync.Mutex{},

		imapCachePath: cache.GetIMAPCachePath(),
		imapCacheLock: &sync.RWMutex{},
		listWorkers:   listWorkers,
	}
	ib.updates = newIMAPUpdates(ib)
	return ib
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

	if ib.bridge.HasError(bridge.ErrLocalCacheUnavailable) {
		return nil, users.ErrLoggedOutUser
	}

	imapUser, err := ib.getUser(username)
	if err != nil {
		log.WithError(err).Warn("Cannot get user")
		return nil, err
	}

	if err := imapUser.user.CheckBridgeLogin(password); err != nil {
		log.WithError(err).Error("Could not check bridge password")
		if err := imapUser.Logout(); err != nil {
			log.WithError(err).Warn("Could not logout user after unsuccessful login check")
		}
		// Apple Mail sometimes generates a lot of requests very quickly.
		// It's therefore good to have a timeout after a bad login so that we can slow
		// those requests down a little bit.
		time.Sleep(10 * time.Second)
		return nil, err
	}

	// The update channel should be nil until we try to login to IMAP for the first time
	// so that it doesn't make bridge slow for users who are only using bridge for SMTP
	// (otherwise the store will be locked for 1 sec per email during synchronization).
	if store := imapUser.user.GetStore(); store != nil {
		store.SetChangeNotifier(ib.updates)
	}

	return imapUser, nil
}

// Updates returns a channel of updates for IMAP IDLE extension.
func (ib *imapBackend) Updates() <-chan goIMAPBackend.Update {
	// Called from go-imap in goroutines - we need to handle panics for each function.
	defer ib.panicHandler.HandlePanic()

	return ib.updates.ch
}

func (ib *imapBackend) CreateMessageLimit() *uint32 {
	return nil
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
