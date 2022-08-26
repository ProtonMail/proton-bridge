package tests

import (
	"fmt"
	"net/smtp"

	"github.com/ProtonMail/proton-bridge/v2/internal/constants"
)

func (t *testCtx) newSMTPClient(userID, clientID string) error {
	return t.newSMTPClientOnPort(userID, clientID, t.bridge.GetSMTPPort())
}

func (t *testCtx) newSMTPClientOnPort(userID, clientID string, imapPort int) error {
	client, err := smtp.Dial(fmt.Sprintf("%v:%d", constants.Host, imapPort))
	if err != nil {
		return err
	}

	t.smtpClients[clientID] = &smtpClient{
		userID: userID,
		client: client,
	}

	return nil
}

func (t *testCtx) getSMTPClient(clientID string) (string, *smtp.Client) {
	return t.smtpClients[clientID].userID, t.smtpClients[clientID].client
}
