package pmapi

import (
	"net/http"
	"sync"

	"github.com/getsentry/raven-go"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

// ClientManager is a manager of clients.
type ClientManager struct {
	config *ClientConfig

	clients       map[string]*Client
	clientsLocker sync.Locker

	tokens       map[string]string
	tokensLocker sync.Locker

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

// NewClientManager creates a new ClientMan which manages clients configured with the given client config.
func NewClientManager(config *ClientConfig) (cm *ClientManager) {
	if err := raven.SetDSN(config.SentryDSN); err != nil {
		logrus.WithError(err).Error("Could not set up sentry DSN")
	}

	cm = &ClientManager{
		config: config,

		clients:       make(map[string]*Client),
		clientsLocker: &sync.Mutex{},

		tokens:       make(map[string]string),
		tokensLocker: &sync.Mutex{},

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

// SetRootURL sets the root URL to make requests to.
func (cm *ClientManager) SetRootURL(url string) {
	cm.urlLocker.Lock()
	defer cm.urlLocker.Unlock()

	logrus.WithField("url", url).Info("Changing to a new root URL")

	cm.url = url
}

// IsProxyAllowed returns whether the user has allowed us to switch to a proxy if need be.
func (cm *ClientManager) IsProxyAllowed() bool {
	return cm.allowProxy
}

// AllowProxy allows the client manager to switch clients over to a proxy if need be.
func (cm *ClientManager) AllowProxy() {
	cm.allowProxy = true
}

// DisallowProxy prevents the client manager from switching clients over to a proxy if need be.
func (cm *ClientManager) DisallowProxy() {
	cm.allowProxy = false
}

// FindProxy returns a usable proxy server.
func (cm *ClientManager) SwitchToProxy() (proxy string, err error) {
	logrus.Info("Attempting gto switch to a proxy")

	if proxy, err = cm.proxyProvider.findProxy(); err != nil {
		err = errors.Wrap(err, "failed to find usable proxy")
		return
	}

	cm.SetRootURL(proxy)

	// TODO: Disable after 24 hours.

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
	return cm.clientAuths
}

// getClientAuthChannel returns a channel on which clients should send auths.
func (cm *ClientManager) getClientAuthChannel() chan ClientAuth {
	return cm.clientAuths
}

// forwardClientAuths handles all incoming auths from clients before forwarding them on the bridge auth channel.
func (cm *ClientManager) forwardClientAuths() {
	for auth := range cm.clientAuths {
		cm.handleClientAuth(auth)
		cm.bridgeAuths <- auth
	}
}

func (cm *ClientManager) setToken(userID, token string) {
	cm.tokensLocker.Lock()
	defer cm.tokensLocker.Unlock()

	logrus.WithField("userID", userID).Info("Updating refresh token")

	cm.tokens[userID] = token
}

func (cm *ClientManager) clearToken(userID string) {
	cm.tokensLocker.Lock()
	defer cm.tokensLocker.Unlock()

	logrus.WithField("userID", userID).Info("Clearing refresh token")

	delete(cm.tokens, userID)
}

// handleClientAuth updates or clears client authorisation based on auths received.
func (cm *ClientManager) handleClientAuth(ca ClientAuth) {
	// If we aren't managing this client, there's nothing to do.
	if _, ok := cm.clients[ca.UserID]; !ok {
		return
	}

	// If the auth is nil, we should clear the token.
	// TODO: Maybe we should trigger a client logout here? Then we don't have to remember to log it out ourself.
	if ca.Auth == nil {
		cm.clearToken(ca.UserID)
		return
	}

	cm.setToken(ca.UserID, ca.Auth.GenToken())
}
