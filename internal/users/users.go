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

// Package users provides core business logic providing API over credentials store and PM API.
package users

import (
	"context"
	"strings"
	"sync"
	"time"

	"github.com/ProtonMail/proton-bridge/v2/internal/events"
	"github.com/ProtonMail/proton-bridge/v2/internal/metrics"
	"github.com/ProtonMail/proton-bridge/v2/internal/users/credentials"
	"github.com/ProtonMail/proton-bridge/v2/pkg/keychain"
	"github.com/ProtonMail/proton-bridge/v2/pkg/listener"
	"github.com/ProtonMail/proton-bridge/v2/pkg/pmapi"
	"github.com/hashicorp/go-multierror"
	"github.com/pkg/errors"
	logrus "github.com/sirupsen/logrus"
)

var (
	log                   = logrus.WithField("pkg", "users") //nolint:gochecknoglobals
	isApplicationOutdated = false                            //nolint:gochecknoglobals

	// ErrWrongMailboxPassword is returned when login password is OK but
	// not the mailbox one.
	ErrWrongMailboxPassword = errors.New("wrong mailbox password")

	// ErrUserAlreadyConnected is returned when authentication was OK but
	// there is already active account for this user.
	ErrUserAlreadyConnected = errors.New("user is already connected")
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

	lock sync.RWMutex
}

func New(
	locations Locator,
	panicHandler PanicHandler,
	eventListener listener.Listener,
	clientManager pmapi.Manager,
	credStorer CredentialsStorer,
	storeFactory StoreMaker,
) *Users {
	log.Trace("Creating new users")

	u := &Users{
		locations:     locations,
		panicHandler:  panicHandler,
		events:        eventListener,
		clientManager: clientManager,
		credStorer:    credStorer,
		storeFactory:  storeFactory,
		lock:          sync.RWMutex{},
	}

	go func() {
		defer panicHandler.HandlePanic()
		u.watchEvents()
	}()

	if u.credStorer == nil {
		log.Error("No credentials store is available")
	} else if err := u.loadUsersFromCredentialsStore(); err != nil {
		log.WithError(err).Error("Could not load all users from credentials store")
	}

	return u
}

func (u *Users) watchEvents() {
	upgradeCh := u.events.ProvideChannel(events.UpgradeApplicationEvent)
	internetConnChangedCh := u.events.ProvideChannel(events.InternetConnChangedEvent)

	for {
		select {
		case <-upgradeCh:
			isApplicationOutdated = true
			u.closeAllConnections()
		case stat := <-internetConnChangedCh:
			if stat != events.InternetOn {
				continue
			}
			for _, user := range u.users {
				if user.store == nil {
					if err := user.loadStore(); err != nil {
						log.WithError(err).Error("Failed to load store after reconnecting")
					}
				}

				if user.totalBytes == 0 {
					user.UpdateSpace(nil)
				}
			}
		}
	}
}

func (u *Users) loadUsersFromCredentialsStore() error {
	u.lock.Lock()
	defer u.lock.Unlock()

	userIDs, err := u.credStorer.List()
	if err != nil {
		notifyKeychainRepair(u.events, err)
		return err
	}

	for _, userID := range userIDs {
		l := log.WithField("user", userID)
		user, creds, err := newUser(u.panicHandler, userID, u.events, u.credStorer, u.storeFactory)
		if err != nil {
			l.WithError(err).Warn("Could not create user, skipping")
			continue
		}

		u.users = append(u.users, user)

		if creds.IsConnected() {
			// If there is no connection, we don't want to retry. Load should
			// happen fast enough to not block GUI. When connection is back up,
			// watchEvents and unlockIfNecessary will finish user init later.
			if err := u.loadConnectedUser(pmapi.ContextWithoutRetry(context.Background()), user, creds); err != nil {
				l.WithError(err).Warn("Could not load connected user")
			}
		} else {
			l.Warn("User is disconnected and must be connected manually")
			if err := user.connect(u.clientManager.NewClient("", "", "", time.Time{}), creds); err != nil {
				l.WithError(err).Warn("Could not load disconnected user")
			}
		}
	}

	return err
}

func (u *Users) loadConnectedUser(ctx context.Context, user *User, creds *credentials.Credentials) error {
	uid, ref, err := creds.SplitAPIToken()
	if err != nil {
		return errors.Wrap(err, "could not get user's refresh token")
	}

	client, auth, err := u.clientManager.NewClientWithRefresh(ctx, uid, ref)
	if err != nil {
		// When client cannot be refreshed right away due to no connection,
		// we create client which will refresh automatically when possible.
		connectErr := user.connect(u.clientManager.NewClient(uid, "", ref, time.Time{}), creds)

		switch errors.Cause(err) {
		case pmapi.ErrNoConnection, pmapi.ErrUpgradeApplication:
			return connectErr
		}

		if pmapi.IsFailedAuth(connectErr) {
			if logoutErr := user.logout(); logoutErr != nil {
				logrus.WithError(logoutErr).Warn("Could not logout user")
			}
		}
		return errors.Wrap(err, "could not refresh token")
	}

	// Update the user's credentials with the latest auth used to connect this user.
	if creds, err = u.credStorer.UpdateToken(creds.UserID, auth.UID, auth.RefreshToken); err != nil {
		notifyKeychainRepair(u.events, err)
		return errors.Wrap(err, "could not create get user's refresh token")
	}

	return user.connect(client, creds)
}

func (u *Users) closeAllConnections() {
	for _, user := range u.users {
		user.CloseAllConnections()
	}
}

// Login authenticates a user by username/password, returning an authorised client and an auth object.
// The authorisation scope may not yet be full if the user has 2FA enabled.
func (u *Users) Login(username string, password []byte) (authClient pmapi.Client, auth *pmapi.Auth, err error) {
	u.crashBandicoot(username)

	return u.clientManager.NewClientWithLogin(context.Background(), username, password)
}

// FinishLogin finishes the login procedure and adds the user into the credentials store.
func (u *Users) FinishLogin(client pmapi.Client, auth *pmapi.Auth, password []byte) (user *User, err error) { //nolint:funlen
	apiUser, passphrase, err := getAPIUser(context.Background(), client, password)
	if err != nil {
		return nil, err
	}

	if user, ok := u.hasUser(apiUser.ID); ok {
		if user.IsConnected() {
			if err := client.AuthDelete(context.Background()); err != nil {
				logrus.WithError(err).Warn("Failed to delete new auth session")
			}

			return user, ErrUserAlreadyConnected
		}

		// Update the user's credentials with the latest auth used to connect this user.
		if _, err := u.credStorer.UpdateToken(auth.UserID, auth.UID, auth.RefreshToken); err != nil {
			notifyKeychainRepair(u.events, err)
			return nil, errors.Wrap(err, "failed to load user credentials")
		}

		// Update the password in case the user changed it.
		creds, err := u.credStorer.UpdatePassword(apiUser.ID, passphrase)
		if err != nil {
			notifyKeychainRepair(u.events, err)
			return nil, errors.Wrap(err, "failed to update password of user in credentials store")
		}

		// will go and unlock cache if not already done
		if err := user.connect(client, creds); err != nil {
			return nil, errors.Wrap(err, "failed to reconnect existing user")
		}

		u.events.Emit(events.UserRefreshEvent, apiUser.ID)

		return user, nil
	}

	if err := u.addNewUser(client, apiUser, auth, passphrase); err != nil {
		return nil, errors.Wrap(err, "failed to add new user")
	}

	u.events.Emit(events.UserRefreshEvent, apiUser.ID)

	return u.GetUser(apiUser.ID)
}

// addNewUser adds a new user.
func (u *Users) addNewUser(client pmapi.Client, apiUser *pmapi.User, auth *pmapi.Auth, passphrase []byte) error {
	u.lock.Lock()
	defer u.lock.Unlock()

	if _, err := u.credStorer.Add(apiUser.ID, apiUser.Name, auth.UID, auth.RefreshToken, passphrase, client.Addresses().ActiveEmails()); err != nil {
		notifyKeychainRepair(u.events, err)
		return errors.Wrap(err, "failed to add user credentials to credentials store")
	}

	user, creds, err := newUser(u.panicHandler, apiUser.ID, u.events, u.credStorer, u.storeFactory)
	if err != nil {
		return errors.Wrap(err, "failed to create new user")
	}

	if err := user.connect(client, creds); err != nil {
		return errors.Wrap(err, "failed to connect new user")
	}

	if err := u.SendMetric(metrics.New(metrics.Setup, metrics.NewUser, metrics.NoLabel)); err != nil {
		log.WithError(err).Error("Failed to send metric")
	}

	u.users = append(u.users, user)

	return nil
}

func getAPIUser(ctx context.Context, client pmapi.Client, password []byte) (*pmapi.User, []byte, error) {
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
		return nil, nil, ErrWrongMailboxPassword
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

	return result
}

// DeleteUser deletes user completely; it logs user out from the API, stops any
// active connection, deletes from credentials store and removes from the Bridge struct.
func (u *Users) DeleteUser(userID string, clearStore bool) error {
	u.lock.Lock()
	defer u.lock.Unlock()
	defer u.events.Emit(events.UserRefreshEvent, userID)

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
				notifyKeychainRepair(u.events, err)
				log.WithError(err).Error("Cannot remove user")
				return err
			}
			u.users = append(u.users[:idx], u.users[idx+1:]...)
			return nil
		}
	}

	return errors.New("user " + userID + " not found")
}

// ClearUsers deletes all users.
func (u *Users) ClearUsers() error {
	var result error

	for _, user := range u.GetUsers() {
		if err := u.DeleteUser(user.ID(), false); err != nil {
			result = multierror.Append(result, err)
		}
	}

	return result
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

func notifyKeychainRepair(l listener.Listener, err error) {
	if err == keychain.ErrMacKeychainRebuild {
		l.Emit(events.CredentialsErrorEvent, err.Error())
	}
}
