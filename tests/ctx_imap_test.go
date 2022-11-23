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
// along with Proton Mail Bridge.  If not, see <https://www.gnu.org/licenses/>.

package tests

import (
	"fmt"

	"github.com/ProtonMail/proton-bridge/v3/internal/constants"
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
