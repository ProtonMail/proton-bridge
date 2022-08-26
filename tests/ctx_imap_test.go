package tests

import (
	"fmt"

	"github.com/ProtonMail/proton-bridge/v2/internal/constants"
	"github.com/emersion/go-imap/client"
)

func (t *testCtx) newIMAPClient(userID, clientID string) error {
	return t.newIMAPClientOnPort(userID, clientID, t.bridge.GetIMAPPort())
}

func (t *testCtx) newIMAPClientOnPort(userID, clientID string, imapPort int) error {
	client, err := client.Dial(fmt.Sprintf("%v:%d", constants.Host, imapPort))
	if err != nil {
		return err
	}

	t.imapClients[clientID] = &imapClient{
		userID: userID,
		client: client,
	}

	return nil
}

func (t *testCtx) getIMAPClient(clientID string) (string, *client.Client) {
	return t.imapClients[clientID].userID, t.imapClients[clientID].client
}
