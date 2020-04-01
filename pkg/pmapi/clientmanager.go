package pmapi

import (
	"github.com/getsentry/raven-go"
	"github.com/sirupsen/logrus"
)

// ClientManager is a manager of clients.
type ClientManager struct {
	// TODO: Lockers.

	clients map[string]*Client
	tokens  map[string]string
	config  *ClientConfig
}

// NewClientManager creates a new ClientMan which manages clients configured with the given client config.
func NewClientManager(config *ClientConfig) *ClientManager {
	if err := raven.SetDSN(config.SentryDSN); err != nil {
		logrus.WithError(err).Error("Could not set up sentry DSN")
	}

	return &ClientManager{
		clients: make(map[string]*Client),
		tokens:  make(map[string]string),
		config:  config,
	}
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

// GetConfig returns the config used to configure clients.
func (cm *ClientManager) GetConfig() *ClientConfig {
	return cm.config
}

// GetToken returns the token for the given userID.
func (cm *ClientManager) GetToken(userID string) string {
	return cm.tokens[userID]
}

// SetToken sets the token for the given userID.
func (cm *ClientManager) SetToken(userID, token string) {
	cm.tokens[userID] = token
}

// SetTokenIfUnset sets the token for the given userID if it does not yet have a token.
func (cm *ClientManager) SetTokenIfUnset(userID, token string) {
	if _, ok := cm.tokens[userID]; ok {
		return
	}

	cm.tokens[userID] = token
}

// ClearToken clears the token of the given userID.
func (cm *ClientManager) ClearToken(userID string) {
	delete(cm.tokens, userID)
}
