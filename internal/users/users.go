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
	"context"
	"strings"
	"sync"
	"time"

	"github.com/ProtonMail/proton-bridge/internal/events"
	imapcache "github.com/ProtonMail/proton-bridge/internal/imap/cache"
	"github.com/ProtonMail/proton-bridge/internal/metrics"
	"github.com/ProtonMail/proton-bridge/internal/users/credentials"
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
	clientManager pmapi.Manager
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
}

func New(
	locations Locator,
	panicHandler PanicHandler,
	eventListener listener.Listener,
	clientManager pmapi.Manager,
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
	}

	// FIXME(conman): Handle force upgrade events.
	/*
		go func() {
			defer panicHandler.HandlePanic()
			u.watchAppOutdated()
		}()
	*/

	if u.credStorer == nil {
		log.Error("No credentials store is available")
	} else if err := u.loadUsersFromCredentialsStore(context.TODO()); err != nil {
		log.WithError(err).Error("Could not load all users from credentials store")
	}

	return u
}

func (u *Users) loadUsersFromCredentialsStore(ctx context.Context) error {
	u.lock.Lock()
	defer u.lock.Unlock()

	userIDs, err := u.credStorer.List()
	if err != nil {
		return err
	}

	for _, userID := range userIDs {
		user, creds, err := newUser(u.panicHandler, userID, u.events, u.credStorer, u.storeFactory, u.useOnlyActiveAddresses)
		if err != nil {
			logrus.WithError(err).Warn("Could not create user, skipping")
			continue
		}

		u.users = append(u.users, user)

		if creds.IsConnected() {
			if err := u.loadConnectedUser(ctx, user, creds); err != nil {
				logrus.WithError(err).Warn("Could not load connected user")
			}
		} else {
			logrus.Warn("User is disconnected and must be connected manually")

			if err := u.loadDisconnectedUser(ctx, user, creds); err != nil {
				logrus.WithError(err).Warn("Could not load disconnected user")
			}
		}
	}

	return err
}

func (u *Users) loadDisconnectedUser(ctx context.Context, user *User, creds *credentials.Credentials) error {
	// FIXME(conman): We shouldn't be creating unauthorized clients... this is hacky, just to avoid huge refactor!
	return user.connect(ctx, u.clientManager.NewClient("", "", "", time.Time{}), creds)
}

func (u *Users) loadConnectedUser(ctx context.Context, user *User, creds *credentials.Credentials) error {
	uid, ref, err := creds.SplitAPIToken()
	if err != nil {
		return errors.Wrap(err, "could not get user's refresh token")
	}

	client, auth, err := u.clientManager.NewClientWithRefresh(ctx, uid, ref)
	if err != nil {
		// FIXME(conman): This is a problem... if we weren't able to create a new client due to internet,
		// we need to be able to retry later, but I deleted all the hacky "retry auth if necessary" stuff...
		return user.connect(ctx, u.clientManager.NewClient(uid, "", ref, time.Time{}), creds)
	}

	// Update the user's credentials with the latest auth used to connect this user.
	if creds, err = u.credStorer.UpdateToken(auth.UserID, auth.UID, auth.RefreshToken); err != nil {
		return errors.Wrap(err, "could not create get user's refresh token")
	}

	return user.connect(ctx, client, creds)
}

func (u *Users) watchAppOutdated() {
	// FIXME(conman): handle force upgrade events.

	/*
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
	*/
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

	return u.clientManager.NewClientWithLogin(context.TODO(), username, password)
}

// FinishLogin finishes the login procedure and adds the user into the credentials store.
func (u *Users) FinishLogin(client pmapi.Client, auth *pmapi.Auth, password string) (user *User, err error) { //nolint[funlen]
	apiUser, passphrase, err := getAPIUser(context.TODO(), client, password)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get API user")
	}

	if user, ok := u.hasUser(apiUser.ID); ok {
		if user.IsConnected() {
			if err := client.AuthDelete(context.TODO()); err != nil {
				logrus.WithError(err).Warn("Failed to delete new auth session")
			}

			return nil, errors.New("user is already connected")
		}

		// Update the user's credentials with the latest auth used to connect this user.
		if _, err := u.credStorer.UpdateToken(auth.UserID, auth.UID, auth.RefreshToken); err != nil {
			return nil, errors.Wrap(err, "failed to load user credentials")
		}

		// Update the password in case the user changed it.
		creds, err := u.credStorer.UpdatePassword(apiUser.ID, string(passphrase))
		if err != nil {
			return nil, errors.Wrap(err, "failed to update password of user in credentials store")
		}

		if err := user.connect(context.TODO(), client, creds); err != nil {
			return nil, errors.Wrap(err, "failed to reconnect existing user")
		}

		return user, nil
	}

	if err := u.addNewUser(context.TODO(), client, apiUser, auth, passphrase); err != nil {
		return nil, errors.Wrap(err, "failed to add new user")
	}

	u.events.Emit(events.UserRefreshEvent, apiUser.ID)

	return u.GetUser(apiUser.ID)
}

// addNewUser adds a new user.
func (u *Users) addNewUser(ctx context.Context, client pmapi.Client, apiUser *pmapi.User, auth *pmapi.Auth, passphrase []byte) error {
	u.lock.Lock()
	defer u.lock.Unlock()

	var emails []string

	if u.useOnlyActiveAddresses {
		emails = client.Addresses().ActiveEmails()
	} else {
		emails = client.Addresses().AllEmails()
	}

	if _, err := u.credStorer.Add(apiUser.ID, apiUser.Name, auth.UID, auth.RefreshToken, string(passphrase), emails); err != nil {
		return errors.Wrap(err, "failed to add user credentials to credentials store")
	}

	user, creds, err := newUser(u.panicHandler, apiUser.ID, u.events, u.credStorer, u.storeFactory, u.useOnlyActiveAddresses)
	if err != nil {
		return errors.Wrap(err, "failed to create new user")
	}

	if err := user.connect(ctx, client, creds); err != nil {
		return errors.Wrap(err, "failed to connect new user")
	}

	if err := u.SendMetric(metrics.New(metrics.Setup, metrics.NewUser, metrics.NoLabel)); err != nil {
		log.WithError(err).Error("Failed to send metric")
	}

	u.users = append(u.users, user)

	return nil
}

func getAPIUser(ctx context.Context, client pmapi.Client, password string) (*pmapi.User, []byte, error) {
	salt, err := client.AuthSalt(ctx)
	if err != nil {
		return nil, nil, errors.Wrap(err, "failed to get salt")
	}

	passphrase, err := pmapi.HashMailboxPassword(password, salt)
	if err != nil {
		return nil, nil, errors.Wrap(err, "failed to hash password")
	}

	// We unlock the user's PGP key here to detect if the user's mailbox password is wrong.
	if err := client.Unlock(ctx, passphrase); err != nil {
		return nil, nil, errors.Wrap(err, "failed to unlock client")
	}

	user, err := client.CurrentUser(ctx)
	if err != nil {
		return nil, nil, errors.Wrap(err, "failed to load user data")
	}

	return user, passphrase, nil
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
	cat, act, lab := m.Get()

	if err := u.clientManager.SendSimpleMetric(context.Background(), string(cat), string(act), string(lab)); err != nil {
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
	// FIXME(conman): Support DoH.
	// u.apiManager.AllowProxy()
}

// DisallowProxy instructs the app to not use DoH to access an API proxy if necessary.
// It also needs to work before the app is initialised (because we may need to use the proxy at startup).
func (u *Users) DisallowProxy() {
	// FIXME(conman): Support DoH.
	// u.apiManager.DisallowProxy()
}

// CheckConnection returns whether there is an internet connection.
// This should use the connection manager when it is eventually implemented.
func (u *Users) CheckConnection() error {
	// FIXME(conman): Other parts of bridge that rely on this method should register as a connection observer.
	panic("TODO: register as a connection observer to get this information")
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
