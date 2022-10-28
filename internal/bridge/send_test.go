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

package bridge_test

import (
	"context"
	"fmt"
	"net/smtp"
	"testing"
	"time"

	"github.com/ProtonMail/proton-bridge/v2/internal/bridge"
	"github.com/ProtonMail/proton-bridge/v2/internal/constants"
	"github.com/emersion/go-imap"
	"github.com/emersion/go-imap/client"
	"github.com/stretchr/testify/require"
	"gitlab.protontech.ch/go/liteapi"
	"gitlab.protontech.ch/go/liteapi/server"
)

func TestBridge_Send(t *testing.T) {
	withEnv(t, func(ctx context.Context, s *server.Server, netCtl *liteapi.NetCtl, locator bridge.Locator, storeKey []byte) {
		_, _, err := s.CreateUser("recipient", "recipient@pm.me", password)
		require.NoError(t, err)

		withBridge(ctx, t, s.GetHostURL(), netCtl, locator, storeKey, func(bridge *bridge.Bridge, mocks *bridge.Mocks) {
			senderUserID, err := bridge.LoginFull(ctx, username, password, nil, nil)
			require.NoError(t, err)

			recipientUserID, err := bridge.LoginFull(ctx, "recipient", password, nil, nil)
			require.NoError(t, err)

			senderInfo, err := bridge.GetUserInfo(senderUserID)
			require.NoError(t, err)

			recipientInfo, err := bridge.GetUserInfo(recipientUserID)
			require.NoError(t, err)

			for i := 0; i < 10; i++ {
				// Send an email from sender to recipient.
				smtpClient, err := smtp.Dial(fmt.Sprintf("%v:%v", constants.Host, bridge.GetSMTPPort()))
				require.NoError(t, err)
				defer smtpClient.Close() //nolint:errcheck

				require.NoError(t, smtpClient.Auth(smtp.PlainAuth("", senderInfo.Addresses[0], string(senderInfo.BridgePass), constants.Host)))
				require.NoError(t, smtpClient.Mail(senderInfo.Addresses[0]))
				require.NoError(t, smtpClient.Rcpt("recipient@pm.me"))

				wc, err := smtpClient.Data()
				require.NoError(t, err)

				n, err := fmt.Fprintf(wc, "Subject: Test %v\r\n\r\nHello world!", i)
				require.NoError(t, err)
				require.Greater(t, n, 0)
				require.NoError(t, wc.Close())

				// Sender should see the message in the Sent folder.
				senderIMAPClient, err := client.Dial(fmt.Sprintf("%v:%v", constants.Host, bridge.GetIMAPPort()))
				require.NoError(t, err)
				require.NoError(t, senderIMAPClient.Login(senderInfo.Addresses[0], string(senderInfo.BridgePass)))
				defer senderIMAPClient.Logout() //nolint:errcheck

				require.Eventually(t, func() bool {
					status, err := senderIMAPClient.Status(`Sent`, []imap.StatusItem{imap.StatusMessages})
					require.NoError(t, err)
					return status.Messages == uint32(i+1)
				}, 10*time.Second, 100*time.Millisecond)

				// Recipient should see the message in the Inbox.
				recipientIMAPClient, err := client.Dial(fmt.Sprintf("%v:%v", constants.Host, bridge.GetIMAPPort()))
				require.NoError(t, err)
				require.NoError(t, recipientIMAPClient.Login(recipientInfo.Addresses[0], string(recipientInfo.BridgePass)))
				defer recipientIMAPClient.Logout() //nolint:errcheck

				require.Eventually(t, func() bool {
					status, err := recipientIMAPClient.Status(`Inbox`, []imap.StatusItem{imap.StatusMessages})
					require.NoError(t, err)
					return status.Messages == uint32(i+1)
				}, 10*time.Second, 100*time.Millisecond)
			}
		})
	})
}
