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

package pmapi

import (
	"fmt"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

const maxLogoutRetries = 5

// ClientManager is a manager of clients.
type ClientManager struct {
	// newClient is used to create new Clients. By default this creates pmapi clients but it can be overridden to
	// create other types of clients (e.g. for integration tests).
	newClient func(userID string) Client

	config       *ClientConfig
	roundTripper http.RoundTripper

	clients       map[string]Client
	clientsLocker sync.Locker

	tokens       map[string]string
	tokensLocker sync.Locker

	expirations       map[string]*tokenExpiration
	expiredTokens     chan string
	expirationsLocker sync.Locker

	authUpdates chan ClientAuth

	host, scheme string
	hostLocker   sync.RWMutex

	allowProxy       bool
	proxyProvider    *proxyProvider
	proxyUseDuration time.Duration

	idGen idGen

	log *logrus.Entry
}

type idGen int

func (i *idGen) next() int {
	(*i)++
	return int(*i)
}

// ClientAuth holds an API auth produced by a Client for a specific user.
type ClientAuth struct {
	UserID string
	Auth   *Auth
}

// tokenExpiration manages the expiration of an access token.
type tokenExpiration struct {
	timer  *time.Timer
	cancel chan (struct{})
}

// NewClientManager creates a new ClientMan which manages clients configured with the given client config.
func NewClientManager(config *ClientConfig) (cm *ClientManager) {
	cm = &ClientManager{
		config:       config,
		roundTripper: http.DefaultTransport,

		clients:       make(map[string]Client),
		clientsLocker: &sync.Mutex{},

		tokens:       make(map[string]string),
		tokensLocker: &sync.Mutex{},

		expirations:       make(map[string]*tokenExpiration),
		expiredTokens:     make(chan string),
		expirationsLocker: &sync.Mutex{},

		host:       rootURL,
		scheme:     rootScheme,
		hostLocker: sync.RWMutex{},

		authUpdates: make(chan ClientAuth),

		proxyProvider:    newProxyProvider(dohProviders, proxyQuery),
		proxyUseDuration: proxyUseDuration,

		log: logrus.WithField("pkg", "pmapi-manager"),
	}

	cm.newClient = func(userID string) Client {
		return newClient(cm, userID)
	}
	cm.SetUserAgent("", "", "") // Set default user agent.

	go cm.watchTokenExpirations()

	return cm
}

// SetClientConstructor sets the method used to construct clients.
// By default this is `pmapi.newClient` but can be overridden with this method.
func (cm *ClientManager) SetClientConstructor(f func(userID string) Client) {
	cm.newClient = f
}

// SetRoundTripper sets the roundtripper used by clients created by this client manager.
func (cm *ClientManager) SetRoundTripper(rt http.RoundTripper) {
	cm.roundTripper = rt
}

func (cm *ClientManager) SetUserAgent(clientName, clientVersion, os string) {
	cm.config.UserAgent = formatUserAgent(clientName, clientVersion, os)
}

// GetClient returns a client for the given userID.
// If the client does not exist already, it is created.
func (cm *ClientManager) GetClient(userID string) Client {
	cm.clientsLocker.Lock()
	defer cm.clientsLocker.Unlock()

	if client, ok := cm.clients[userID]; ok {
		return client
	}

	cm.clients[userID] = cm.newClient(userID)

	return cm.clients[userID]
}

// GetAnonymousClient returns an anonymous client.
func (cm *ClientManager) GetAnonymousClient() Client {
	return cm.GetClient(fmt.Sprintf("anonymous-%v", cm.idGen.next()))
}

// LogoutClient logs out the client with the given userID and ensures its sensitive data is successfully cleared.
func (cm *ClientManager) LogoutClient(userID string) {
	cm.clientsLocker.Lock()
	defer cm.clientsLocker.Unlock()

	client, ok := cm.clients[userID]
	if !ok {
		return
	}

	delete(cm.clients, userID)

	go func() {
		defer client.ClearData()
		defer cm.clearToken(userID)

		if strings.HasPrefix(userID, "anonymous-") {
			return
		}

		var retries int

		for client.DeleteAuth() == ErrAPINotReachable {
			retries++

			if retries > maxLogoutRetries {
				cm.log.Error("Failed to delete client auth (retried too many times)")
				break
			}

			cm.log.Warn("Failed to delete client auth because API was not reachable, retrying...")
		}
	}()
}

// GetRootURL returns the full root URL (scheme+host).
func (cm *ClientManager) GetRootURL() string {
	cm.hostLocker.RLock()
	defer cm.hostLocker.RUnlock()

	return fmt.Sprintf("%v://%v", cm.scheme, cm.host)
}

// getHost returns the host to make requests to.
// It does not include the protocol i.e. no "https://" (use getScheme for that).
func (cm *ClientManager) getHost() string {
	cm.hostLocker.RLock()
	defer cm.hostLocker.RUnlock()

	return cm.host
}

// IsProxyAllowed returns whether the user has allowed us to switch to a proxy if need be.
func (cm *ClientManager) IsProxyAllowed() bool {
	cm.hostLocker.RLock()
	defer cm.hostLocker.RUnlock()

	return cm.allowProxy
}

// AllowProxy allows the client manager to switch clients over to a proxy if need be.
func (cm *ClientManager) AllowProxy() {
	cm.hostLocker.Lock()
	defer cm.hostLocker.Unlock()

	cm.allowProxy = true
}

// DisallowProxy prevents the client manager from switching clients over to a proxy if need be.
func (cm *ClientManager) DisallowProxy() {
	cm.hostLocker.Lock()
	defer cm.hostLocker.Unlock()

	cm.allowProxy = false
	cm.host = rootURL
}

// IsProxyEnabled returns whether we are currently proxying requests.
func (cm *ClientManager) IsProxyEnabled() bool {
	cm.hostLocker.RLock()
	defer cm.hostLocker.RUnlock()

	return cm.host != rootURL
}

// switchToReachableServer switches to using a reachable server (either proxy or standard API).
func (cm *ClientManager) switchToReachableServer() (proxy string, err error) {
	cm.hostLocker.Lock()
	defer cm.hostLocker.Unlock()

	logrus.Info("Attempting to switch to a proxy")

	if proxy, err = cm.proxyProvider.findReachableServer(); err != nil {
		err = errors.Wrap(err, "failed to find a usable proxy")
		return
	}

	// If the chosen proxy is the standard API, we want to use it but still show the troubleshooting screen.
	if proxy == rootURL {
		logrus.Info("The standard API is reachable again; connection drop was only intermittent")
		err = ErrAPINotReachable
		cm.host = proxy
		return
	}

	logrus.WithField("proxy", proxy).Info("Switching to a proxy")

	// If the host is currently the rootURL, it's the first time we are enabling a proxy.
	// This means we want to disable it again in 24 hours.
	if cm.host == rootURL {
		go func() {
			<-time.After(cm.proxyUseDuration)

			cm.hostLocker.Lock()
			defer cm.hostLocker.Unlock()

			cm.host = rootURL
		}()
	}

	cm.host = proxy

	return proxy, err
}

// GetToken returns the token for the given userID.
func (cm *ClientManager) GetToken(userID string) string {
	cm.tokensLocker.Lock()
	defer cm.tokensLocker.Unlock()

	return cm.tokens[userID]
}

// GetAuthUpdateChannel returns a channel on which client auths can be received.
func (cm *ClientManager) GetAuthUpdateChannel() chan ClientAuth {
	return cm.authUpdates
}

// ErrNoInternetConnection indicates that both protonstatus and the API are unreachable.
var ErrNoInternetConnection = errors.New("no internet connection")

// CheckConnection returns an error if there is no internet connection.
// This should be moved to the ConnectionManager when it is implemented.
func (cm *ClientManager) CheckConnection() error {
	client := getHTTPClient(cm.config, cm.roundTripper)

	// Do not cumulate timeouts, use goroutines.
	retStatus := make(chan error)
	retAPI := make(chan error)

	// Check protonstatus.com without SSL for performance reasons. vpn_status endpoint is fast and
	// returns only OK; this endpoint is not known by the public. We check the connection only.
	go checkConnection(client, "http://protonstatus.com/vpn_status", retStatus)

	// Check of API reachability also uses a fast endpoint.
	go checkConnection(client, cm.GetRootURL()+"/tests/ping", retAPI)

	errStatus := <-retStatus
	errAPI := <-retAPI

	switch {
	case errStatus == nil && errAPI == nil:
		return nil

	case errStatus == nil && errAPI != nil:
		cm.log.Error("ProtonStatus is reachable but API is not")
		return ErrAPINotReachable

	case errStatus != nil && errAPI == nil:
		cm.log.Warn("API is reachable but protonstatus is not")
		return nil

	case errStatus != nil && errAPI != nil:
		cm.log.Error("Both ProtonStatus and API are unreachable")
		return ErrNoInternetConnection
	}

	return nil
}

func checkConnection(client *http.Client, url string, errorChannel chan error) {
	resp, err := client.Get(url)
	if err != nil {
		errorChannel <- err
		return
	}

	_ = resp.Body.Close()

	if resp.StatusCode != 200 {
		errorChannel <- fmt.Errorf("HTTP status code %d", resp.StatusCode)
		return
	}

	errorChannel <- nil
}

// setTokenIfUnset sets the token for the given userID if it wasn't already set.
// The set token does not expire.
func (cm *ClientManager) setTokenIfUnset(userID, token string) {
	cm.tokensLocker.Lock()
	defer cm.tokensLocker.Unlock()

	if _, ok := cm.tokens[userID]; ok {
		return
	}

	logrus.WithField("userID", userID).Info("Setting token because it is currently unset")

	cm.tokens[userID] = token
}

// setToken sets the token for the given userID with the given expiration time.
func (cm *ClientManager) setToken(userID, token string, expiration time.Duration) {
	cm.tokensLocker.Lock()
	defer cm.tokensLocker.Unlock()

	logrus.WithField("userID", userID).Info("Updating token")

	cm.tokens[userID] = token

	cm.setTokenExpiration(userID, expiration)
}

// setTokenExpiration will ensure the token is refreshed if it expires.
// If the token already has an expiration time set, it is replaced.
func (cm *ClientManager) setTokenExpiration(userID string, expiration time.Duration) {
	cm.expirationsLocker.Lock()
	defer cm.expirationsLocker.Unlock()

	// Reduce the expiration by one minute so we can do the refresh with enough time to spare.
	expiration -= time.Minute

	if exp, ok := cm.expirations[userID]; ok {
		exp.timer.Stop()
		close(exp.cancel)
	}

	cm.expirations[userID] = &tokenExpiration{
		timer:  time.NewTimer(expiration),
		cancel: make(chan struct{}),
	}

	go func(expiration *tokenExpiration) {
		select {
		case <-expiration.timer.C:
			cm.expiredTokens <- userID

		case <-expiration.cancel:
			logrus.WithField("userID", userID).Debug("Auth was refreshed before it expired")
		}
	}(cm.expirations[userID])
}

func (cm *ClientManager) clearToken(userID string) {
	cm.tokensLocker.Lock()
	defer cm.tokensLocker.Unlock()

	logrus.WithField("userID", userID).Info("Clearing token")

	delete(cm.tokens, userID)
}

// HandleAuth updates or clears client authorisation based on auths received and then forwards the auth onwards.
func (cm *ClientManager) HandleAuth(ca ClientAuth) {
	cm.clientsLocker.Lock()
	defer cm.clientsLocker.Unlock()

	// If we aren't managing this client, there's nothing to do.
	if _, ok := cm.clients[ca.UserID]; !ok {
		logrus.WithField("userID", ca.UserID).Info("Not handling auth for unmanaged client")
		return
	}

	// If the auth is nil, we should clear the token.
	if ca.Auth == nil {
		cm.clearToken(ca.UserID)
		go cm.LogoutClient(ca.UserID)
		return
	}

	cm.setToken(ca.UserID, ca.Auth.GenToken(), time.Duration(ca.Auth.ExpiresIn)*time.Second)

	logrus.Debug("ClientManager is forwarding auth update...")
	cm.authUpdates <- ca
	logrus.Debug("Auth update was forwarded")
}

// watchTokenExpirations refreshes any tokens which are about to expire.
func (cm *ClientManager) watchTokenExpirations() {
	for userID := range cm.expiredTokens {
		log := cm.log.WithField("userID", userID)

		log.Info("Auth token expired! Refreshing")

		client, ok := cm.clients[userID]
		if !ok {
			log.Warn("Can't refresh expired token because there is no such client")
			continue
		}

		token, ok := cm.tokens[userID]
		if !ok {
			log.Warn("Can't refresh expired token because there is no such token")
			continue
		}

		if _, err := client.AuthRefresh(token); err != nil {
			log.WithError(err).Error("Failed to refresh expired token")
		}
	}
}
