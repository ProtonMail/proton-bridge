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

package bridge

import (
	"fmt"
	"path/filepath"
	"runtime"
	"strings"
	"sync"

	"github.com/ProtonMail/proton-bridge/internal/bridge/credentials"
	"github.com/ProtonMail/proton-bridge/internal/events"
	"github.com/ProtonMail/proton-bridge/internal/store"
	"github.com/ProtonMail/proton-bridge/pkg/listener"
	"github.com/ProtonMail/proton-bridge/pkg/pmapi"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

// ErrLoggedOutUser is sent to IMAP and SMTP if user exists, password is OK but user is logged out from bridge.
var ErrLoggedOutUser = errors.New("bridge account is logged out, use bridge to login again")

// User is a struct on top of API client and credentials store.
type User struct {
	log           *logrus.Entry
	panicHandler  PanicHandler
	listener      listener.Listener
	clientManager ClientManager
	credStorer    CredentialsStorer

	imapUpdatesChannel chan interface{}

	store      *store.Store
	storeCache *store.Cache
	storePath  string

	userID string
	creds  *credentials.Credentials

	lock         sync.RWMutex
	isAuthorized bool

	unlockingKeyringLock sync.Mutex
	wasKeyringUnlocked   bool
}

// newUser creates a new bridge user.
func newUser(
	panicHandler PanicHandler,
	userID string,
	eventListener listener.Listener,
	credStorer CredentialsStorer,
	clientManager ClientManager,
	storeCache *store.Cache,
	storeDir string,
) (u *User, err error) {
	log := log.WithField("user", userID)
	log.Debug("Creating or loading user")

	creds, err := credStorer.Get(userID)
	if err != nil {
		return nil, errors.Wrap(err, "failed to load user credentials")
	}

	u = &User{
		log:           log,
		panicHandler:  panicHandler,
		listener:      eventListener,
		credStorer:    credStorer,
		clientManager: clientManager,
		storeCache:    storeCache,
		storePath:     getUserStorePath(storeDir, userID),
		userID:        userID,
		creds:         creds,
	}

	return
}

func (u *User) client() pmapi.Client {
	return u.clientManager.GetClient(u.userID)
}

// init initialises a bridge user. This includes reloading its credentials from the credentials store
// (such as when logging out and back in, you need to reload the credentials because the new credentials will
// have the apitoken and password), authorising the user against the api, loading the user store (creating a new one
// if necessary), and setting the imap idle updates channel (used to send imap idle updates to the imap backend if
// something in the store changed).
func (u *User) init(idleUpdates chan interface{}) (err error) {
	u.unlockingKeyringLock.Lock()
	u.wasKeyringUnlocked = false
	u.unlockingKeyringLock.Unlock()

	u.log.Info("Initialising user")

	// Reload the user's credentials (if they log out and back in we need the new
	// version with the apitoken and mailbox password).
	creds, err := u.credStorer.Get(u.userID)
	if err != nil {
		return errors.Wrap(err, "failed to load user credentials")
	}
	u.creds = creds

	// Try to authorise the user if they aren't already authorised.
	// Note: we still allow users to set up bridge if the internet is off.
	if authErr := u.authorizeIfNecessary(false); authErr != nil {
		switch errors.Cause(authErr) {
		case pmapi.ErrAPINotReachable, pmapi.ErrUpgradeApplication, ErrLoggedOutUser:
			u.log.WithError(authErr).Warn("Could not authorize user")
		default:
			if logoutErr := u.logout(); logoutErr != nil {
				u.log.WithError(logoutErr).Warn("Could not logout user")
			}
			return errors.Wrap(authErr, "failed to authorize user")
		}
	}

	// Logged-out user keeps store running to access offline data.
	// Therefore it is necessary to close it before re-init.
	if u.store != nil {
		if err := u.store.Close(); err != nil {
			log.WithError(err).Error("Not able to close store")
		}
		u.store = nil
	}
	store, err := store.New(u.panicHandler, u, u.clientManager, u.listener, u.storePath, u.storeCache)
	if err != nil {
		return errors.Wrap(err, "failed to create store")
	}
	u.store = store

	// Save the imap updates channel here so it can be set later when imap connects.
	u.imapUpdatesChannel = idleUpdates

	return err
}

func (u *User) SetIMAPIdleUpdateChannel() {
	if u.store == nil {
		return
	}

	u.store.SetIMAPUpdateChannel(u.imapUpdatesChannel)
}

// authorizeIfNecessary checks whether user is logged in and is connected to api auth channel.
// If user is not already connected to the api auth channel (for example there was no internet during start),
// it tries to connect it.
func (u *User) authorizeIfNecessary(emitEvent bool) (err error) {
	// If user is connected and has an auth channel, then perfect, nothing to do here.
	if u.creds.IsConnected() && u.HasAPIAuth() {
		// The keyring  unlock is triggered here to resolve state where apiClient
		// is authenticated (we have auth token) but it was not possible to download
		// and unlock the keys (internet not reachable).
		return u.unlockIfNecessary()
	}

	if !u.creds.IsConnected() {
		err = ErrLoggedOutUser
	} else if err = u.authorizeAndUnlock(); err != nil {
		u.log.WithError(err).Error("Could not authorize and unlock user")

		switch errors.Cause(err) {
		case pmapi.ErrUpgradeApplication:
			u.listener.Emit(events.UpgradeApplicationEvent, "")

		case pmapi.ErrAPINotReachable:
			u.listener.Emit(events.InternetOffEvent, "")

		default:
			if errLogout := u.credStorer.Logout(u.userID); errLogout != nil {
				u.log.WithField("err", errLogout).Error("Could not log user out from credentials store")
			}
		}
	}

	if emitEvent && err != nil &&
		errors.Cause(err) != pmapi.ErrUpgradeApplication &&
		errors.Cause(err) != pmapi.ErrAPINotReachable {
		u.listener.Emit(events.LogoutEvent, u.userID)
	}

	return err
}

// unlockIfNecessary will not trigger keyring unlocking if it was already successfully unlocked.
func (u *User) unlockIfNecessary() error {
	u.unlockingKeyringLock.Lock()
	defer u.unlockingKeyringLock.Unlock()

	if u.wasKeyringUnlocked {
		return nil
	}

	if _, err := u.client().Unlock(u.creds.MailboxPassword); err != nil {
		return errors.Wrap(err, "failed to unlock user")
	}

	if err := u.client().UnlockAddresses([]byte(u.creds.MailboxPassword)); err != nil {
		return errors.Wrap(err, "failed to unlock user addresses")
	}

	u.wasKeyringUnlocked = true
	return nil
}

// authorizeAndUnlock tries to authorize the user with the API using the the user's APIToken.
// If that succeeds, it tries to unlock the user's keys and addresses.
func (u *User) authorizeAndUnlock() (err error) {
	if u.creds.APIToken == "" {
		u.log.Warn("Could not connect to API auth channel, have no API token")
		return nil
	}

	if _, err := u.client().AuthRefresh(u.creds.APIToken); err != nil {
		return errors.Wrap(err, "failed to refresh API auth")
	}

	if _, err = u.client().Unlock(u.creds.MailboxPassword); err != nil {
		return errors.Wrap(err, "failed to unlock user")
	}

	if err = u.client().UnlockAddresses([]byte(u.creds.MailboxPassword)); err != nil {
		return errors.Wrap(err, "failed to unlock user addresses")
	}

	return nil
}

func (u *User) updateAuthToken(auth *pmapi.Auth) {
	u.log.Debug("User received auth from bridge")

	if err := u.credStorer.UpdateToken(u.userID, auth.GenToken()); err != nil {
		u.log.WithError(err).Error("Failed to update refresh token in credentials store")
		return
	}

	u.refreshFromCredentials()

	u.isAuthorized = true
}

// clearStore removes the database.
func (u *User) clearStore() error {
	u.log.Trace("Clearing user store")

	if u.store != nil {
		if err := u.store.Remove(); err != nil {
			return errors.Wrap(err, "failed to remove store")
		}
	} else {
		u.log.Warn("Store is not initialized: cleaning up store files manually")
		if err := store.RemoveStore(u.storeCache, u.storePath, u.userID); err != nil {
			return errors.Wrap(err, "failed to remove store manually")
		}
	}
	return nil
}

// closeStore just closes the store without deleting it.
func (u *User) closeStore() error {
	u.log.Trace("Closing user store")

	if u.store != nil {
		if err := u.store.Close(); err != nil {
			return errors.Wrap(err, "failed to close store")
		}
	}

	return nil
}

// getUserStorePath returns the file path of the store database for the given userID.
func getUserStorePath(storeDir string, userID string) (path string) {
	fileName := fmt.Sprintf("mailbox-%v.db", userID)
	return filepath.Join(storeDir, fileName)
}

// GetTemporaryPMAPIClient returns an authorised PMAPI client.
// Do not use! It's only for backward compatibility of old SMTP and IMAP implementations.
// After proper refactor of SMTP and IMAP remove this method.
func (u *User) GetTemporaryPMAPIClient() pmapi.Client {
	return u.client()
}

// ID returns the user's userID.
func (u *User) ID() string {
	return u.userID
}

// Username returns the user's username as found in the user's credentials.
func (u *User) Username() string {
	u.lock.RLock()
	defer u.lock.RUnlock()

	return u.creds.Name
}

// IsConnected returns whether user is logged in.
func (u *User) IsConnected() bool {
	u.lock.RLock()
	defer u.lock.RUnlock()

	return u.creds.IsConnected()
}

// IsCombinedAddressMode returns whether user is set in combined or split mode.
// Combined mode is the default mode and is what users typically need.
// Split mode is mostly for outlook as it cannot handle sending e-mails from an
// address other than the primary one.
func (u *User) IsCombinedAddressMode() bool {
	if u.store != nil {
		return u.store.IsCombinedMode()
	}

	return u.creds.IsCombinedAddressMode
}

// GetPrimaryAddress returns the user's original address (which is
// not necessarily the same as the primary address, because a primary address
// might be an alias and be in position one).
func (u *User) GetPrimaryAddress() string {
	u.lock.RLock()
	defer u.lock.RUnlock()

	return u.creds.EmailList()[0]
}

// GetStoreAddresses returns all addresses used by the store (so in combined mode,
// that's just the original address, but in split mode, that's all active addresses).
func (u *User) GetStoreAddresses() []string {
	u.lock.RLock()
	defer u.lock.RUnlock()

	if u.IsCombinedAddressMode() {
		return u.creds.EmailList()[:1]
	}

	return u.creds.EmailList()
}

// getStoreAddresses returns a user's used addresses (with the original address in first place).
func (u *User) getStoreAddresses() []string { // nolint[unused]
	addrInfo, err := u.store.GetAddressInfo()
	if err != nil {
		u.log.WithError(err).Error("Failed getting address info from store")
		return nil
	}

	addresses := []string{}
	for _, addr := range addrInfo {
		addresses = append(addresses, addr.Address)
	}

	if u.IsCombinedAddressMode() {
		return addresses[:1]
	}

	return addresses
}

// GetAddresses returns list of all addresses.
func (u *User) GetAddresses() []string {
	u.lock.RLock()
	defer u.lock.RUnlock()

	return u.creds.EmailList()
}

// GetAddressID returns the API ID of the given address.
func (u *User) GetAddressID(address string) (id string, err error) {
	u.lock.RLock()
	defer u.lock.RUnlock()

	address = strings.ToLower(address)

	if u.store == nil {
		err = errors.New("store is not initialised")
		return
	}

	return u.store.GetAddressID(address)
}

// GetBridgePassword returns bridge password. This is not a password of the PM
// account, but generated password for local purposes to not use a PM account
// in the clients (such as Thunderbird).
func (u *User) GetBridgePassword() string {
	u.lock.RLock()
	defer u.lock.RUnlock()

	return u.creds.BridgePassword
}

// CheckBridgeLogin checks whether the user is logged in and the bridge
// IMAP/SMTP password is correct.
func (u *User) CheckBridgeLogin(password string) error {
	if isApplicationOutdated {
		u.listener.Emit(events.UpgradeApplicationEvent, "")
		return pmapi.ErrUpgradeApplication
	}

	u.lock.RLock()
	defer u.lock.RUnlock()

	// True here because users should be notified by popup of auth failure.
	if err := u.authorizeIfNecessary(true); err != nil {
		u.log.WithError(err).Error("Failed to authorize user")
		return err
	}

	return u.creds.CheckPassword(password)
}

// UpdateUser updates user details from API and saves to the credentials.
func (u *User) UpdateUser() error {
	u.lock.Lock()
	defer u.lock.Unlock()

	if err := u.authorizeIfNecessary(true); err != nil {
		return errors.Wrap(err, "cannot update user")
	}

	_, err := u.client().UpdateUser()
	if err != nil {
		return err
	}

	if _, err = u.client().Unlock(u.creds.MailboxPassword); err != nil {
		return err
	}

	if err := u.client().UnlockAddresses([]byte(u.creds.MailboxPassword)); err != nil {
		return err
	}

	emails := u.client().Addresses().ActiveEmails()
	if err := u.credStorer.UpdateEmails(u.userID, emails); err != nil {
		return err
	}

	u.refreshFromCredentials()

	return nil
}

// SwitchAddressMode changes mode from combined to split and vice versa. The mode to switch to is determined by the
// state of the user's credentials in the credentials store. See `IsCombinedAddressMode` for more details.
func (u *User) SwitchAddressMode() (err error) {
	u.log.Trace("Switching user address mode")

	u.lock.Lock()
	defer u.lock.Unlock()
	u.closeAllConnections()

	if u.store == nil {
		err = errors.New("store is not initialised")
		return
	}

	newAddressModeState := !u.IsCombinedAddressMode()

	if err = u.store.UseCombinedMode(newAddressModeState); err != nil {
		u.log.WithError(err).Error("Could not switch store address mode")
		return
	}

	if u.creds.IsCombinedAddressMode != newAddressModeState {
		if err = u.credStorer.SwitchAddressMode(u.userID); err != nil {
			u.log.WithError(err).Error("Could not switch credentials store address mode")
			return
		}
	}

	u.refreshFromCredentials()

	return err
}

// logout is the same as Logout, but for internal purposes (logged out from
// the server) which emits LogoutEvent to notify other parts of the Bridge.
func (u *User) logout() error {
	u.lock.Lock()
	wasConnected := u.creds.IsConnected()
	u.lock.Unlock()

	err := u.Logout()

	if wasConnected {
		u.listener.Emit(events.LogoutEvent, u.userID)
		u.listener.Emit(events.UserRefreshEvent, u.userID)
	}

	u.isAuthorized = false

	return err
}

// Logout logs out the user from pmapi, the credentials store, the mail store, and tries to remove as much
// sensitive data as possible.
func (u *User) Logout() (err error) {
	u.lock.Lock()
	defer u.lock.Unlock()

	u.log.Debug("Logging out user")

	if !u.creds.IsConnected() {
		return
	}

	u.unlockingKeyringLock.Lock()
	u.wasKeyringUnlocked = false
	u.unlockingKeyringLock.Unlock()

	u.client().Logout()

	if err = u.credStorer.Logout(u.userID); err != nil {
		u.log.WithError(err).Warn("Could not log user out from credentials store")

		if err = u.credStorer.Delete(u.userID); err != nil {
			u.log.WithError(err).Error("Could not delete user from credentials store")
		}
	}

	u.refreshFromCredentials()

	// Do not close whole store, just event loop. Some information might be needed offline (e.g. addressID)
	u.closeEventLoop()

	u.closeAllConnections()

	runtime.GC()

	return err
}

func (u *User) refreshFromCredentials() {
	if credentials, err := u.credStorer.Get(u.userID); err != nil {
		log.WithError(err).Error("Cannot refresh user credentials")
	} else {
		u.creds = credentials
	}
}

func (u *User) closeEventLoop() {
	if u.store == nil {
		return
	}

	u.store.CloseEventLoop()
}

// closeAllConnections calls CloseConnection for all users addresses.
func (u *User) closeAllConnections() {
	for _, address := range u.creds.EmailList() {
		u.CloseConnection(address)
	}

	if u.store != nil {
		u.store.SetIMAPUpdateChannel(nil)
	}
}

// CloseConnection emits closeConnection event on `address` which should close all active connection.
func (u *User) CloseConnection(address string) {
	u.listener.Emit(events.CloseConnectionEvent, address)
}

func (u *User) GetStore() *store.Store {
	return u.store
}

func (u *User) HasAPIAuth() bool {
	return u.isAuthorized
}
