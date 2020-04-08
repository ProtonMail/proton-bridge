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

// Package bridge provides core business logic providing API over credentials store and PM API.
package bridge

import (
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/ProtonMail/proton-bridge/internal/events"
	"github.com/ProtonMail/proton-bridge/internal/metrics"
	"github.com/ProtonMail/proton-bridge/internal/preferences"
	"github.com/ProtonMail/proton-bridge/internal/store"
	"github.com/ProtonMail/proton-bridge/pkg/config"
	"github.com/ProtonMail/proton-bridge/pkg/listener"
	"github.com/ProtonMail/proton-bridge/pkg/pmapi"
	"github.com/hashicorp/go-multierror"
	"github.com/pkg/errors"
	logrus "github.com/sirupsen/logrus"
)

var (
	log                   = config.GetLogEntry("bridge") //nolint[gochecknoglobals]
	isApplicationOutdated = false                        //nolint[gochecknoglobals]
)

// Bridge is a struct handling users.
type Bridge struct {
	config        Configer
	pref          PreferenceProvider
	panicHandler  PanicHandler
	events        listener.Listener
	version       string
	clientManager ClientManager
	credStorer    CredentialsStorer
	storeCache    *store.Cache

	// users is a list of accounts that have been added to bridge.
	// They are stored sorted in the credentials store in the order
	// that they were added to bridge chronologically.
	// People are used to that and so we preserve that ordering here.
	users []*User

	// idleUpdates is a channel which the imap backend listens to and which it uses
	// to send idle updates to the mail client (eg thunderbird).
	// The user stores should send idle updates on this channel.
	idleUpdates chan interface{}

	lock sync.RWMutex

	userAgentClientName    string
	userAgentClientVersion string
	userAgentOS            string
}

func New(
	config Configer,
	pref PreferenceProvider,
	panicHandler PanicHandler,
	eventListener listener.Listener,
	version string,
	clientManager ClientManager,
	credStorer CredentialsStorer,
) *Bridge {
	log.Trace("Creating new bridge")

	b := &Bridge{
		config:        config,
		pref:          pref,
		panicHandler:  panicHandler,
		events:        eventListener,
		version:       version,
		clientManager: clientManager,
		credStorer:    credStorer,
		storeCache:    store.NewCache(config.GetIMAPCachePath()),
		idleUpdates:   make(chan interface{}),
		lock:          sync.RWMutex{},
	}

	// Allow DoH before starting bridge if the user has previously set this setting.
	// This allows us to start even if protonmail is blocked.
	if pref.GetBool(preferences.AllowProxyKey) {
		b.AllowProxy()
	}

	go func() {
		defer panicHandler.HandlePanic()
		b.watchBridgeOutdated()
	}()

	go func() {
		defer panicHandler.HandlePanic()
		b.watchAPIAuths()
	}()

	if b.credStorer == nil {
		log.Error("Bridge has no credentials store")
	} else if err := b.loadUsersFromCredentialsStore(); err != nil {
		log.WithError(err).Error("Could not load all users from credentials store")
	}

	if pref.GetBool(preferences.FirstStartKey) {
		b.SendMetric(metrics.New(metrics.Setup, metrics.FirstStart, metrics.Label(version)))
	}

	go b.heartbeat()

	return b
}

// heartbeat sends a heartbeat signal once a day.
func (b *Bridge) heartbeat() {
	for range time.NewTicker(1 * time.Hour).C {
		next, err := strconv.ParseInt(b.pref.Get(preferences.NextHeartbeatKey), 10, 64)
		if err != nil {
			continue
		}
		nextTime := time.Unix(next, 0)
		if time.Now().After(nextTime) {
			b.SendMetric(metrics.New(metrics.Heartbeat, metrics.Daily, metrics.NoLabel))
			nextTime = nextTime.Add(24 * time.Hour)
			b.pref.Set(preferences.NextHeartbeatKey, strconv.FormatInt(nextTime.Unix(), 10))
		}
	}
}

func (b *Bridge) loadUsersFromCredentialsStore() (err error) {
	b.lock.Lock()
	defer b.lock.Unlock()

	userIDs, err := b.credStorer.List()
	if err != nil {
		return
	}

	for _, userID := range userIDs {
		l := log.WithField("user", userID)

		user, newUserErr := newUser(b.panicHandler, userID, b.events, b.credStorer, b.clientManager, b.storeCache, b.config.GetDBDir())
		if newUserErr != nil {
			l.WithField("user", userID).WithError(newUserErr).Warn("Could not load user, skipping")
			continue
		}

		b.users = append(b.users, user)

		if initUserErr := user.init(b.idleUpdates); initUserErr != nil {
			l.WithField("user", userID).WithError(initUserErr).Warn("Could not initialise user")
		}
	}

	return err
}

func (b *Bridge) watchBridgeOutdated() {
	ch := make(chan string)
	b.events.Add(events.UpgradeApplicationEvent, ch)
	for range ch {
		isApplicationOutdated = true
		b.closeAllConnections()
	}
}

// watchAPIAuths receives auths from the client manager and sends them to the appropriate user.
func (b *Bridge) watchAPIAuths() {
	for auth := range b.clientManager.GetAuthUpdateChannel() {
		logrus.Debug("Bridge received auth from ClientManager")

		user, ok := b.hasUser(auth.UserID)
		if !ok {
			logrus.WithField("userID", auth.UserID).Info("User not available for auth update")
			continue
		}

		if auth.Auth != nil {
			user.updateAuthToken(auth.Auth)
		} else {
			user.logout()
		}
	}
}

func (b *Bridge) closeAllConnections() {
	for _, user := range b.users {
		user.closeAllConnections()
	}
}

// Login authenticates a user.
// The login flow:
//  * Authenticate user:
//      client, auth, err := bridge.Authenticate(username, password)
//
//  * In case user `auth.HasTwoFactor()`, ask for it and fully authenticate the user.
// 	    auth2FA, err := client.Auth2FA(twoFactorCode)
//
//  * In case user `auth.HasMailboxPassword()`, ask for it, otherwise use `password`
//    and then finish the login procedure.
//		user, err := bridge.FinishLogin(client, auth, mailboxPassword)
func (b *Bridge) Login(username, password string) (authClient pmapi.Client, auth *pmapi.Auth, err error) {
	b.crashBandicoot(username)

	// We need to use anonymous client because we don't yet have userID and so can't save auth tokens yet.
	authClient = b.clientManager.GetAnonymousClient()

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
// See `Login` for more details of the login flow.
func (b *Bridge) FinishLogin(authClient pmapi.Client, auth *pmapi.Auth, mbPassword string) (user *User, err error) { //nolint[funlen]
	defer func() {
		if err == pmapi.ErrUpgradeApplication {
			b.events.Emit(events.UpgradeApplicationEvent, "")
		}
	}()

	apiUser, hashedPassword, err := getAPIUser(authClient, auth, mbPassword)
	if err != nil {
		log.WithError(err).Error("Failed to get API user")
		return
	}

	if user, err = b.GetUser(apiUser.ID); err == nil {
		if err = b.connectExistingUser(user, auth, hashedPassword); err != nil {
			log.WithError(err).Error("Failed to connect existing user")
			return
		}
	} else {
		if err = b.addNewUser(apiUser, auth, hashedPassword); err != nil {
			log.WithError(err).Error("Failed to add new user")
			return
		}
	}

	b.events.Emit(events.UserRefreshEvent, apiUser.ID)

	return b.GetUser(apiUser.ID)
}

// connectExistingUser connects an existing bridge user to the bridge.
func (b *Bridge) connectExistingUser(user *User, auth *pmapi.Auth, hashedPassword string) (err error) {
	if user.IsConnected() {
		return errors.New("user is already connected")
	}

	// Update the user's password in the cred store in case they changed it.
	if err = b.credStorer.UpdatePassword(user.ID(), hashedPassword); err != nil {
		return errors.Wrap(err, "failed to update password of user in credentials store")
	}

	client := b.clientManager.GetClient(user.ID())

	if auth, err = client.AuthRefresh(auth.GenToken()); err != nil {
		return errors.Wrap(err, "failed to refresh auth token of new client")
	}

	if err = b.credStorer.UpdateToken(user.ID(), auth.GenToken()); err != nil {
		return errors.Wrap(err, "failed to update token of user in credentials store")
	}

	if err = user.init(b.idleUpdates); err != nil {
		return errors.Wrap(err, "failed to initialise user")
	}

	return
}

// addNewUser adds a new bridge user to the bridge.
func (b *Bridge) addNewUser(user *pmapi.User, auth *pmapi.Auth, hashedPassword string) (err error) {
	b.lock.Lock()
	defer b.lock.Unlock()

	client := b.clientManager.GetClient(user.ID)

	if auth, err = client.AuthRefresh(auth.GenToken()); err != nil {
		return errors.Wrap(err, "failed to refresh token in new client")
	}

	if user, err = client.UpdateUser(); err != nil {
		return errors.Wrap(err, "failed to update API user")
	}

	activeEmails := client.Addresses().ActiveEmails()

	if _, err = b.credStorer.Add(user.ID, user.Name, auth.GenToken(), hashedPassword, activeEmails); err != nil {
		return errors.Wrap(err, "failed to add user to credentials store")
	}

	bridgeUser, err := newUser(b.panicHandler, user.ID, b.events, b.credStorer, b.clientManager, b.storeCache, b.config.GetDBDir())
	if err != nil {
		return errors.Wrap(err, "failed to create user")
	}

	// The user needs to be part of the users list in order for it to receive an auth during initialisation.
	// TODO: If adding the user fails, we don't want to leave it there.
	b.users = append(b.users, bridgeUser)

	if err = bridgeUser.init(b.idleUpdates); err != nil {
		return errors.Wrap(err, "failed to initialise user")
	}

	b.SendMetric(metrics.New(metrics.Setup, metrics.NewUser, metrics.NoLabel))

	return
}

func getAPIUser(client pmapi.Client, auth *pmapi.Auth, mbPassword string) (user *pmapi.User, hashedPassword string, err error) {
	hashedPassword, err = pmapi.HashMailboxPassword(mbPassword, auth.KeySalt)
	if err != nil {
		log.WithError(err).Error("Could not hash mailbox password")
		return
	}

	if _, err = client.Unlock(hashedPassword); err != nil {
		log.WithError(err).Error("Could not decrypt keyring")
		return
	}

	if user, err = client.UpdateUser(); err != nil {
		log.WithError(err).Error("Could not update API user")
		return
	}

	return
}

// GetUsers returns all added users into keychain (even logged out users).
func (b *Bridge) GetUsers() []*User {
	b.lock.RLock()
	defer b.lock.RUnlock()

	return b.users
}

// GetUser returns a user by `query` which is compared to users' ID, username or any attached e-mail address.
func (b *Bridge) GetUser(query string) (*User, error) {
	b.crashBandicoot(query)

	b.lock.RLock()
	defer b.lock.RUnlock()

	for _, user := range b.users {
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
func (b *Bridge) ClearData() error {
	var result *multierror.Error
	for _, user := range b.users {
		if err := user.Logout(); err != nil {
			result = multierror.Append(result, err)
		}
		if err := user.closeStore(); err != nil {
			result = multierror.Append(result, err)
		}
	}
	if err := b.config.ClearData(); err != nil {
		result = multierror.Append(result, err)
	}
	return result.ErrorOrNil()
}

// DeleteUser deletes user completely; it logs user out from the API, stops any
// active connection, deletes from credentials store and removes from the Bridge struct.
func (b *Bridge) DeleteUser(userID string, clearStore bool) error {
	b.lock.Lock()
	defer b.lock.Unlock()

	log := log.WithField("user", userID)

	for idx, user := range b.users {
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

			if err := b.credStorer.Delete(userID); err != nil {
				log.WithError(err).Error("Cannot remove user")
				return err
			}
			b.users = append(b.users[:idx], b.users[idx+1:]...)
			return nil
		}
	}

	return errors.New("user " + userID + " not found")
}

// ReportBug reports a new bug from the user.
func (b *Bridge) ReportBug(osType, osVersion, description, accountName, address, emailClient string) error {
	c := b.clientManager.GetClient("bug_reporter")
	defer c.Logout()

	title := "[Bridge] Bug"
	if err := c.ReportBugWithEmailClient(
		osType,
		osVersion,
		title,
		description,
		accountName,
		address,
		emailClient,
	); err != nil {
		log.Error("Reporting bug failed: ", err)
		return err
	}

	log.Info("Bug successfully reported")

	return nil
}

// SendMetric sends a metric. We don't want to return any errors, only log them.
func (b *Bridge) SendMetric(m metrics.Metric) {
	c := b.clientManager.GetClient("metric_reporter")
	defer c.Logout()

	cat, act, lab := m.Get()
	if err := c.SendSimpleMetric(string(cat), string(act), string(lab)); err != nil {
		log.Error("Sending metric failed: ", err)
	}

	log.WithFields(logrus.Fields{
		"cat": cat,
		"act": act,
		"lab": lab,
	}).Debug("Metric successfully sent")
}

// GetCurrentClient returns currently connected client (e.g. Thunderbird).
func (b *Bridge) GetCurrentClient() string {
	res := b.userAgentClientName
	if b.userAgentClientVersion != "" {
		res = res + " " + b.userAgentClientVersion
	}
	return res
}

// SetCurrentClient updates client info (e.g. Thunderbird) and sets the user agent
// on pmapi. By default no client is used, IMAP has to detect it on first login.
func (b *Bridge) SetCurrentClient(clientName, clientVersion string) {
	b.userAgentClientName = clientName
	b.userAgentClientVersion = clientVersion
	b.updateCurrentUserAgent()
}

// SetCurrentOS updates OS and sets the user agent on pmapi. By default we use
// `runtime.GOOS`, but this can be overridden in case of better detection.
func (b *Bridge) SetCurrentOS(os string) {
	b.userAgentOS = os
	b.updateCurrentUserAgent()
}

// GetIMAPUpdatesChannel sets the channel on which idle events should be sent.
func (b *Bridge) GetIMAPUpdatesChannel() chan interface{} {
	if b.idleUpdates == nil {
		log.Warn("Bridge updates channel is nil")
	}

	return b.idleUpdates
}

// AllowProxy instructs bridge to use DoH to access an API proxy if necessary.
// It also needs to work before bridge is initialised (because we may need to use the proxy at startup).
func (b *Bridge) AllowProxy() {
	b.clientManager.AllowProxy()
}

// DisallowProxy instructs bridge to not use DoH to access an API proxy if necessary.
// It also needs to work before bridge is initialised (because we may need to use the proxy at startup).
func (b *Bridge) DisallowProxy() {
	b.clientManager.DisallowProxy()
}

func (b *Bridge) updateCurrentUserAgent() {
	UpdateCurrentUserAgent(b.version, b.userAgentOS, b.userAgentClientName, b.userAgentClientVersion)
}

// hasUser returns whether the bridge currently has a user with ID `id`.
func (b *Bridge) hasUser(id string) (user *User, ok bool) {
	for _, u := range b.users {
		if u.ID() == id {
			user, ok = u, true
			return
		}
	}

	return
}

// "Easter egg" for testing purposes.
func (b *Bridge) crashBandicoot(username string) {
	if username == "crash@bandicoot" {
		panic("Your wish is my command… I crash!")
	}
}
