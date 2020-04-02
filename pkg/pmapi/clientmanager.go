package pmapi

import (
	"net/http"
	"sync"
	"time"

	"github.com/getsentry/raven-go"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

var proxyUseDuration = 24 * time.Hour

// ClientManager is a manager of clients.
type ClientManager struct {
	config *ClientConfig

	clients       map[string]*Client
	clientsLocker sync.Locker

	tokens           map[string]string
	tokenExpirations map[string]*tokenExpiration
	tokensLocker     sync.Locker

	url       string
	urlLocker sync.Locker

	bridgeAuths chan ClientAuth
	clientAuths chan ClientAuth

	allowProxy    bool
	proxyProvider *proxyProvider
}

type ClientAuth struct {
	UserID string
	Auth   *Auth
}

type tokenExpiration struct {
	timer  *time.Timer
	cancel chan (struct{})
}

// NewClientManager creates a new ClientMan which manages clients configured with the given client config.
func NewClientManager(config *ClientConfig) (cm *ClientManager) {
	if err := raven.SetDSN(config.SentryDSN); err != nil {
		logrus.WithError(err).Error("Could not set up sentry DSN")
	}

	cm = &ClientManager{
		config: config,

		clients:       make(map[string]*Client),
		clientsLocker: &sync.Mutex{},

		tokens:           make(map[string]string),
		tokenExpirations: make(map[string]*tokenExpiration),
		tokensLocker:     &sync.Mutex{},

		url:       RootURL,
		urlLocker: &sync.Mutex{},

		bridgeAuths: make(chan ClientAuth),
		clientAuths: make(chan ClientAuth),

		proxyProvider: newProxyProvider(dohProviders, proxyQuery),
	}

	go cm.forwardClientAuths()

	return
}

// SetClientRoundTripper sets the roundtripper used by clients created by this client manager.
func (cm *ClientManager) SetClientRoundTripper(rt http.RoundTripper) {
	logrus.Info("Setting client roundtripper")
	cm.config.Transport = rt
}

// GetClient returns a client for the given userID.
// If the client does not exist already, it is created.
func (cm *ClientManager) GetClient(userID string) *Client {
	if client, ok := cm.clients[userID]; ok {
		return client
	}

	cm.clients[userID] = newClient(cm, userID)

	return cm.clients[userID]
}

// LogoutClient logs out the client with the given userID and ensures its sensitive data is successfully cleared.
func (cm *ClientManager) LogoutClient(userID string) {
	client, ok := cm.clients[userID]

	if !ok {
		return
	}

	delete(cm.clients, userID)

	go func() {
		if err := client.logout(); err != nil {
			// TODO: Try again! This should loop until it succeeds (might fail the first time due to internet).
			logrus.WithError(err).Error("Client logout failed, not trying again")
		}
		client.clearSensitiveData()
		cm.clearToken(userID)
	}()

	return
}

// GetRootURL returns the root URL to make requests to.
// It does not include the protocol i.e. no "https://".
func (cm *ClientManager) GetRootURL() string {
	cm.urlLocker.Lock()
	defer cm.urlLocker.Unlock()

	return cm.url
}

// IsProxyAllowed returns whether the user has allowed us to switch to a proxy if need be.
func (cm *ClientManager) IsProxyAllowed() bool {
	cm.urlLocker.Lock()
	defer cm.urlLocker.Unlock()

	return cm.allowProxy
}

// AllowProxy allows the client manager to switch clients over to a proxy if need be.
func (cm *ClientManager) AllowProxy() {
	cm.urlLocker.Lock()
	defer cm.urlLocker.Unlock()

	cm.allowProxy = true
}

// DisallowProxy prevents the client manager from switching clients over to a proxy if need be.
func (cm *ClientManager) DisallowProxy() {
	cm.urlLocker.Lock()
	defer cm.urlLocker.Unlock()

	cm.allowProxy = false
	cm.url = RootURL
}

// IsProxyEnabled returns whether we are currently proxying requests.
func (cm *ClientManager) IsProxyEnabled() bool {
	cm.urlLocker.Lock()
	defer cm.urlLocker.Unlock()

	return cm.url != RootURL
}

// FindProxy returns a usable proxy server.
func (cm *ClientManager) SwitchToProxy() (proxy string, err error) {
	cm.urlLocker.Lock()
	defer cm.urlLocker.Unlock()

	logrus.Info("Attempting to switch to a proxy")

	if proxy, err = cm.proxyProvider.findProxy(); err != nil {
		err = errors.Wrap(err, "failed to find a usable proxy")
		return
	}

	logrus.WithField("proxy", proxy).Info("Switching to a proxy")

	cm.url = proxy

	// TODO: Disable again after 24 hours.

	return
}

// GetConfig returns the config used to configure clients.
func (cm *ClientManager) GetConfig() *ClientConfig {
	return cm.config
}

// GetToken returns the token for the given userID.
func (cm *ClientManager) GetToken(userID string) string {
	return cm.tokens[userID]
}

// GetBridgeAuthChannel returns a channel on which client auths can be received.
func (cm *ClientManager) GetBridgeAuthChannel() chan ClientAuth {
	return cm.bridgeAuths
}

// getClientAuthChannel returns a channel on which clients should send auths.
func (cm *ClientManager) getClientAuthChannel() chan ClientAuth {
	return cm.clientAuths
}

// forwardClientAuths handles all incoming auths from clients before forwarding them on the bridge auth channel.
func (cm *ClientManager) forwardClientAuths() {
	for auth := range cm.clientAuths {
		logrus.Debug("ClientManager received auth from client")
		cm.handleClientAuth(auth)
		logrus.Debug("ClientManager is forwarding auth to bridge")
		cm.bridgeAuths <- auth
	}
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
	if exp, ok := cm.tokenExpirations[userID]; ok {
		exp.timer.Stop()
		close(exp.cancel)
	}

	cm.tokenExpirations[userID] = &tokenExpiration{
		timer:  time.NewTimer(expiration),
		cancel: make(chan struct{}),
	}

	go cm.watchTokenExpiration(userID)
}

func (cm *ClientManager) clearToken(userID string) {
	cm.tokensLocker.Lock()
	defer cm.tokensLocker.Unlock()

	logrus.WithField("userID", userID).Info("Clearing token")

	delete(cm.tokens, userID)
}

// handleClientAuth updates or clears client authorisation based on auths received.
func (cm *ClientManager) handleClientAuth(ca ClientAuth) {
	// If we aren't managing this client, there's nothing to do.
	if _, ok := cm.clients[ca.UserID]; !ok {
		logrus.WithField("userID", ca.UserID).Info("Handling auth for unmanaged client")
		return
	}

	// If the auth is nil, we should clear the token.
	// TODO: Maybe we should trigger a client logout here? Then we don't have to remember to log it out ourself.
	if ca.Auth == nil {
		cm.clearToken(ca.UserID)
		return
	}

	cm.setToken(ca.UserID, ca.Auth.GenToken(), time.Duration(ca.Auth.ExpiresIn)*time.Second)
}

func (cm *ClientManager) watchTokenExpiration(userID string) {
	expiration := cm.tokenExpirations[userID]

	select {
	case <-expiration.timer.C:
		logrus.WithField("userID", userID).Info("Auth token expired! Refreshing")
		cm.clients[userID].AuthRefresh(cm.tokens[userID])

	case <-expiration.cancel:
		logrus.WithField("userID", userID).Info("Auth was refreshed before it expired, cancelling this watcher")
	}
}
