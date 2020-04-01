package pmapi

import (
	"sync"

	"github.com/getsentry/raven-go"
	"github.com/sirupsen/logrus"
)

// ClientManager is a manager of clients.
type ClientManager struct {
	clients       map[string]*Client
	clientsLocker sync.Locker

	tokens       map[string]string
	tokensLocker sync.Locker

	config *ClientConfig

	bridgeAuths chan ClientAuth
	clientAuths chan ClientAuth
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
		clients:       make(map[string]*Client),
		clientsLocker: &sync.Mutex{},
		tokens:        make(map[string]string),
		tokensLocker:  &sync.Mutex{},
		config:        config,
		bridgeAuths:   make(chan ClientAuth),
		clientAuths:   make(chan ClientAuth),
	}

	go cm.forwardClientAuths()

	return
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
			// TODO: Try again!
			logrus.WithError(err).Error("Client logout failed, not trying again")
		}
		client.clearSensitiveData()
		cm.clearToken(userID)
	}()

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

	logrus.WithField("userID", userID).WithField("token", token).Info("Updating refresh token")

	cm.tokens[userID] = token
}

func (cm *ClientManager) clearToken(userID string) {
	cm.tokensLocker.Lock()
	defer cm.tokensLocker.Unlock()

	logrus.WithField("userID", userID).Info("Clearing refresh token")

	delete(cm.tokens, userID)
}

// handleClientAuth
func (cm *ClientManager) handleClientAuth(ca ClientAuth) {
	// TODO: Maybe want to logout the client in case of nil auth.
	if _, ok := cm.clients[ca.UserID]; !ok {
		return
	}

	if ca.Auth == nil {
		cm.clearToken(ca.UserID)
	} else {
		cm.setToken(ca.UserID, ca.Auth.GenToken())
	}
}
