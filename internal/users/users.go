// Copyright (c) 2021 Proton Technologies AG
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

// Package users provides core business logic providing API over credentials store and PM API.
package users

import (
	"strings"
	"sync"

	"github.com/ProtonMail/proton-bridge/internal/events"
	imapcache "github.com/ProtonMail/proton-bridge/internal/imap/cache"
	"github.com/ProtonMail/proton-bridge/internal/metrics"
	"github.com/ProtonMail/proton-bridge/pkg/listener"
	"github.com/ProtonMail/proton-bridge/pkg/pmapi"
	"github.com/hashicorp/go-multierror"
	"github.com/pkg/errors"
	logrus "github.com/sirupsen/logrus"
)

var (
	log                   = logrus.WithField("pkg", "users") //nolint[gochecknoglobals]
	isApplicationOutdated = false                            //nolint[gochecknoglobals]

	// ErrWrongMailboxPassword is returned when login password is OK but not the mailbox one.
	ErrWrongMailboxPassword = errors.New("wrong mailbox password")
)

// Users is a struct handling users.
type Users struct {
	locations     Locator
	panicHandler  PanicHandler
	events        listener.Listener
	clientManager ClientManager
	credStorer    CredentialsStorer
	storeFactory  StoreMaker

	// users is a list of accounts that have been added to the app.
	// They are stored sorted in the credentials store in the order
	// that they were added to the app chronologically.
	// People are used to that and so we preserve that ordering here.
	users []*User

	// useOnlyActiveAddresses determines whether credentials keeps only active
	// addresses or all of them. Each usage has to be consisteng, e.g., once
	// user is added, it saves address list to credentials and next time loads
	// as is, without requesting server again.
	useOnlyActiveAddresses bool

	lock sync.RWMutex

	// stopAll can be closed to stop all goroutines from looping (watchAppOutdated, watchAPIAuths, heartbeat etc).
	stopAll chan struct{}
}

func New(
	locations Locator,
	panicHandler PanicHandler,
	eventListener listener.Listener,
	clientManager ClientManager,
	credStorer CredentialsStorer,
	storeFactory StoreMaker,
	useOnlyActiveAddresses bool,
) *Users {
	log.Trace("Creating new users")

	u := &Users{
		locations:              locations,
		panicHandler:           panicHandler,
		events:                 eventListener,
		clientManager:          clientManager,
		credStorer:             credStorer,
		storeFactory:           storeFactory,
		useOnlyActiveAddresses: useOnlyActiveAddresses,
		lock:                   sync.RWMutex{},
		stopAll:                make(chan struct{}),
	}

	go func() {
		defer panicHandler.HandlePanic()
		u.watchAppOutdated()
	}()

	go func() {
		defer panicHandler.HandlePanic()
		u.watchAPIAuths()
	}()

	if u.credStorer == nil {
		log.Error("No credentials store is available")
	} else if err := u.loadUsersFromCredentialsStore(); err != nil {
		log.WithError(err).Error("Could not load all users from credentials store")
	}

	return u
}

func (u *Users) loadUsersFromCredentialsStore() (err error) {
	u.lock.Lock()
	defer u.lock.Unlock()

	userIDs, err := u.credStorer.List()
	if err != nil {
		return
	}

	for _, userID := range userIDs {
		l := log.WithField("user", userID)

		user, newUserErr := newUser(u.panicHandler, userID, u.events, u.credStorer, u.clientManager, u.storeFactory)
		if newUserErr != nil {
			l.WithField("user", userID).WithError(newUserErr).Warn("Could not load user, skipping")
			continue
		}

		u.users = append(u.users, user)

		if initUserErr := user.init(); initUserErr != nil {
			l.WithField("user", userID).WithError(initUserErr).Warn("Could not initialise user")
		}
	}

	return err
}

func (u *Users) watchAppOutdated() {
	ch := make(chan string)

	u.events.Add(events.UpgradeApplicationEvent, ch)

	for {
		select {
		case <-ch:
			isApplicationOutdated = true
			u.closeAllConnections()

		case <-u.stopAll:
			return
		}
	}
}

// watchAPIAuths receives auths from the client manager and sends them to the appropriate user.
func (u *Users) watchAPIAuths() {
	for {
		select {
		case auth := <-u.clientManager.GetAuthUpdateChannel():
			log.Debug("Users received auth from ClientManager")

			user, ok := u.hasUser(auth.UserID)
			if !ok {
				log.WithField("userID", auth.UserID).Info("User not available for auth update")
				continue
			}

			if auth.Auth != nil {
				user.updateAuthToken(auth.Auth)
			} else if err := user.logout(); err != nil {
				log.WithError(err).
					WithField("userID", auth.UserID).
					Error("User logout failed while watching API auths")
			}

		case <-u.stopAll:
			return
		}
	}
}

func (u *Users) closeAllConnections() {
	for _, user := range u.users {
		user.CloseAllConnections()
	}
}

// Login authenticates a user by username/password, returning an authorised client and an auth object.
// The authorisation scope may not yet be full if the user has 2FA enabled.
func (u *Users) Login(username, password string) (authClient pmapi.Client, auth *pmapi.Auth, err error) {
	u.crashBandicoot(username)

	// We need to use anonymous client because we don't yet have userID and so can't save auth tokens yet.
	authClient = u.clientManager.GetAnonymousClient()

	authInfo, err := authClient.AuthInfo(username)
	if err != nil {
		log.WithField("username", username).WithError(err).Error("Could not get auth info for user")
		return
	}

	if auth, err = authClient.Auth(username, password, authInfo); err != nil {
		log.WithField("username", username).WithError(err).Error("Could not get auth for user")
		return
	}

	return
}

// FinishLogin finishes the login procedure and adds the user into the credentials store.
func (u *Users) FinishLogin(authClient pmapi.Client, auth *pmapi.Auth, mbPassphrase string) (user *User, err error) { //nolint[funlen]
	defer func() {
		if err != nil {
			log.WithError(err).Debug("Login not finished; removing auth session")
			if delAuthErr := authClient.DeleteAuth(); delAuthErr != nil {
				log.WithError(delAuthErr).Error("Failed to clear login session after unlock")
			}
		}
		// The anonymous client will be removed from list and authentication will not be deleted.
		authClient.Logout()
	}()

	apiUser, hashedPassphrase, err := getAPIUser(authClient, mbPassphrase)
	if err != nil {
		log.WithError(err).Error("Failed to get API user")
		return
	}

	log.Info("Got API user")

	var ok bool
	if user, ok = u.hasUser(apiUser.ID); ok {
		if err = u.connectExistingUser(user, auth, hashedPassphrase); err != nil {
			log.WithError(err).Error("Failed to connect existing user")
			return
		}
	} else {
		if err = u.addNewUser(apiUser, auth, hashedPassphrase); err != nil {
			log.WithError(err).Error("Failed to add new user")
			return
		}
	}

	// Old credentials use username as key (user ID) which needs to be removed
	// once user logs in again with proper ID fetched from API.
	if _, ok := u.hasUser(apiUser.Name); ok {
		if err := u.DeleteUser(apiUser.Name, true); err != nil {
			log.WithError(err).Error("Failed to delete old user")
		}
	}

	u.events.Emit(events.UserRefreshEvent, apiUser.ID)

	return u.GetUser(apiUser.ID)
}

// connectExistingUser connects an existing user.
func (u *Users) connectExistingUser(user *User, auth *pmapi.Auth, hashedPassphrase string) (err error) {
	if user.IsConnected() {
		return errors.New("user is already connected")
	}

	log.Info("Connecting existing user")

	// Update the user's password in the cred store in case they changed it.
	if err = u.credStorer.UpdatePassword(user.ID(), hashedPassphrase); err != nil {
		return errors.Wrap(err, "failed to update password of user in credentials store")
	}

	client := u.clientManager.GetClient(user.ID())

	if auth, err = client.AuthRefresh(auth.GenToken()); err != nil {
		return errors.Wrap(err, "failed to refresh auth token of new client")
	}

	if err = u.credStorer.UpdateToken(user.ID(), auth.GenToken()); err != nil {
		return errors.Wrap(err, "failed to update token of user in credentials store")
	}

	if err = user.init(); err != nil {
		return errors.Wrap(err, "failed to initialise user")
	}

	return
}

// addNewUser adds a new user.
func (u *Users) addNewUser(apiUser *pmapi.User, auth *pmapi.Auth, hashedPassphrase string) (err error) {
	u.lock.Lock()
	defer u.lock.Unlock()

	client := u.clientManager.GetClient(apiUser.ID)

	if auth, err = client.AuthRefresh(auth.GenToken()); err != nil {
		return errors.Wrap(err, "failed to refresh token in new client")
	}

	if apiUser, err = client.CurrentUser(); err != nil {
		return errors.Wrap(err, "failed to update API user")
	}

	var emails []string //nolint[prealloc]
	if u.useOnlyActiveAddresses {
		emails = client.Addresses().ActiveEmails()
	} else {
		emails = client.Addresses().AllEmails()
	}

	if _, err = u.credStorer.Add(apiUser.ID, apiUser.Name, auth.GenToken(), hashedPassphrase, emails); err != nil {
		return errors.Wrap(err, "failed to add user to credentials store")
	}

	user, err := newUser(u.panicHandler, apiUser.ID, u.events, u.credStorer, u.clientManager, u.storeFactory)
	if err != nil {
		return errors.Wrap(err, "failed to create user")
	}

	// The user needs to be part of the users list in order for it to receive an auth during initialisation.
	u.users = append(u.users, user)

	if err = user.init(); err != nil {
		u.users = u.users[:len(u.users)-1]
		return errors.Wrap(err, "failed to initialise user")
	}

	if err := u.SendMetric(metrics.New(metrics.Setup, metrics.NewUser, metrics.NoLabel)); err != nil {
		log.WithError(err).Error("Failed to send metric")
	}

	return err
}

func getAPIUser(client pmapi.Client, mbPassphrase string) (user *pmapi.User, hashedPassphrase string, err error) {
	salt, err := client.AuthSalt()
	if err != nil {
		log.WithError(err).Error("Could not get salt")
		return nil, "", err
	}

	hashedPassphrase, err = pmapi.HashMailboxPassword(mbPassphrase, salt)
	if err != nil {
		log.WithError(err).Error("Could not hash mailbox password")
		return nil, "", err
	}

	// We unlock the user's PGP key here to detect if the user's mailbox password is wrong.
	if err = client.Unlock([]byte(hashedPassphrase)); err != nil {
		log.WithError(err).Error("Wrong mailbox password")
		return nil, "", ErrWrongMailboxPassword
	}

	if user, err = client.CurrentUser(); err != nil {
		log.WithError(err).Error("Could not load user data")
		return nil, "", err
	}

	return user, hashedPassphrase, nil
}

// GetUsers returns all added users into keychain (even logged out users).
func (u *Users) GetUsers() []*User {
	u.lock.RLock()
	defer u.lock.RUnlock()

	return u.users
}

// GetUser returns a user by `query` which is compared to users' ID, username or any attached e-mail address.
func (u *Users) GetUser(query string) (*User, error) {
	u.crashBandicoot(query)

	u.lock.RLock()
	defer u.lock.RUnlock()

	for _, user := range u.users {
		if strings.EqualFold(user.ID(), query) || strings.EqualFold(user.Username(), query) {
			return user, nil
		}
		for _, address := range user.GetAddresses() {
			if strings.EqualFold(address, query) {
				return user, nil
			}
		}
	}

	return nil, errors.New("user " + query + " not found")
}

// ClearData closes all connections (to release db files and so on) and clears all data.
func (u *Users) ClearData() error {
	var result error

	for _, user := range u.users {
		if err := user.Logout(); err != nil {
			result = multierror.Append(result, err)
		}
		if err := user.closeStore(); err != nil {
			result = multierror.Append(result, err)
		}
	}

	if err := u.locations.Clear(); err != nil {
		result = multierror.Append(result, err)
	}

	// Need to clear imap cache otherwise fetch response will be remembered
	// from previous test
	imapcache.Clear()

	return result
}

// DeleteUser deletes user completely; it logs user out from the API, stops any
// active connection, deletes from credentials store and removes from the Bridge struct.
func (u *Users) DeleteUser(userID string, clearStore bool) error {
	u.lock.Lock()
	defer u.lock.Unlock()

	log := log.WithField("user", userID)

	for idx, user := range u.users {
		if user.ID() == userID {
			if err := user.Logout(); err != nil {
				log.WithError(err).Error("Cannot logout user")
				// We can try to continue to remove the user.
				// Token will still be valid, but will expire eventually.
			}

			if err := user.closeStore(); err != nil {
				log.WithError(err).Error("Failed to close user store")
			}
			if clearStore {
				// Clear cache after closing connections (done in logout).
				if err := user.clearStore(); err != nil {
					log.WithError(err).Error("Failed to clear user")
				}
			}

			if err := u.credStorer.Delete(userID); err != nil {
				log.WithError(err).Error("Cannot remove user")
				return err
			}
			u.users = append(u.users[:idx], u.users[idx+1:]...)
			return nil
		}
	}

	return errors.New("user " + userID + " not found")
}

// SendMetric sends a metric. We don't want to return any errors, only log them.
func (u *Users) SendMetric(m metrics.Metric) error {
	c := u.clientManager.GetAnonymousClient()
	defer c.Logout()

	cat, act, lab := m.Get()
	if err := c.SendSimpleMetric(string(cat), string(act), string(lab)); err != nil {
		return err
	}

	log.WithFields(logrus.Fields{
		"cat": cat,
		"act": act,
		"lab": lab,
	}).Debug("Metric successfully sent")

	return nil
}

// AllowProxy instructs the app to use DoH to access an API proxy if necessary.
// It also needs to work before the app is initialised (because we may need to use the proxy at startup).
func (u *Users) AllowProxy() {
	u.clientManager.AllowProxy()
}

// DisallowProxy instructs the app to not use DoH to access an API proxy if necessary.
// It also needs to work before the app is initialised (because we may need to use the proxy at startup).
func (u *Users) DisallowProxy() {
	u.clientManager.DisallowProxy()
}

// CheckConnection returns whether there is an internet connection.
// This should use the connection manager when it is eventually implemented.
func (u *Users) CheckConnection() error {
	return u.clientManager.CheckConnection()
}

// StopWatchers stops all goroutines.
func (u *Users) StopWatchers() {
	close(u.stopAll)
}

// hasUser returns whether the struct currently has a user with ID `id`.
func (u *Users) hasUser(id string) (user *User, ok bool) {
	for _, u := range u.users {
		if u.ID() == id {
			user, ok = u, true
			return
		}
	}

	return
}

// "Easter egg" for testing purposes.
func (u *Users) crashBandicoot(username string) {
	if username == "crash@bandicoot" {
		panic("Your wish is my command… I crash!")
	}
}
