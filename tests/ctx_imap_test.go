// Copyright (c) 2023 Proton AG
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
	"time"

	"github.com/ProtonMail/proton-bridge/v3/internal/constants"
	"github.com/emersion/go-imap/client"
)

func (t *testCtx) newIMAPClient(userID, clientID string) error {
	return t.newIMAPClientOnPort(userID, clientID, t.bridge.GetIMAPPort())
}

func (t *testCtx) newIMAPClientOnPort(userID, clientID string, imapPort int) error {
	cli, err := eventuallyDial(fmt.Sprintf("%v:%d", constants.Host, imapPort))
	if err != nil {
		return err
	}

	t.imapClients[clientID] = &imapClient{
		userID: userID,
		client: cli,
	}

	return nil
}

func (t *testCtx) getIMAPClient(clientID string) (string, *client.Client) {
	return t.imapClients[clientID].userID, t.imapClients[clientID].client
}

func eventuallyDial(addr string) (cli *client.Client, err error) {
	var sleep = 1 * time.Second
	for i := 0; i < 5; i++ {
		cli, err := client.Dial(addr)
		if err == nil {
			return cli, nil
		}
		time.Sleep(sleep)
		sleep *= 2
	}
	return nil, fmt.Errorf("after 5 attempts, last error: %s", err)
}
