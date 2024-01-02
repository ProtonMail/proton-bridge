// Copyright (c) 2024 Proton AG
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
// along with Proton Mail Bridge.  If not, see <https://www.gnu.org/licenses/>.

package tests

import (
	"fmt"
	"net/smtp"

	"github.com/ProtonMail/proton-bridge/v3/internal/constants"
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
