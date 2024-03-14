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

package bridge_test

import (
	"bytes"
	"context"
	"crypto/tls"
	"fmt"
	"io"
	"net"
	"strings"
	"testing"
	"time"

	"github.com/ProtonMail/gluon/rfc822"
	"github.com/ProtonMail/go-proton-api"
	"github.com/ProtonMail/go-proton-api/server"
	"github.com/ProtonMail/proton-bridge/v3/internal/bridge"
	"github.com/ProtonMail/proton-bridge/v3/internal/constants"
	"github.com/ProtonMail/proton-bridge/v3/internal/events"
	go_imap "github.com/emersion/go-imap"
	"github.com/emersion/go-sasl"
	"github.com/emersion/go-smtp"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/require"
)

func TestBridge_HandleDraftsSendFromOtherClient(t *testing.T) {
	getGluonHeaderID := func(literal []byte) (string, string) {
		h, err := rfc822.NewHeader(literal)
		require.NoError(t, err)

		gluonID, ok := h.GetChecked("X-Pm-Gluon-Id")
		require.True(t, ok)

		externalID, ok := h.GetChecked("Message-Id")
		require.True(t, ok)

		return gluonID, externalID
	}

	withEnv(t, func(ctx context.Context, s *server.Server, netCtl *proton.NetCtl, locator bridge.Locator, storeKey []byte) {
		_, _, err := s.CreateUser("imap", password)
		require.NoError(t, err)

		_, _, err = s.CreateUser("bar", password)
		require.NoError(t, err)

		// The initial user should be fully synced.
		withBridge(ctx, t, s.GetHostURL(), netCtl, locator, storeKey, func(b *bridge.Bridge, _ *bridge.Mocks) {
			syncCh, done := chToType[events.Event, events.SyncFinished](b.GetEvents(events.SyncFinished{}))
			defer done()

			userID, err := b.LoginFull(ctx, "imap", password, nil, nil)
			require.NoError(t, err)

			require.Equal(t, userID, (<-syncCh).UserID)

			info, err := b.GetUserInfo(userID)
			require.NoError(t, err)
			require.True(t, info.State == bridge.Connected)

			client, err := eventuallyDial(fmt.Sprintf("%v:%v", constants.Host, b.GetIMAPPort()))
			require.NoError(t, err)
			require.NoError(t, client.Login(info.Addresses[0], string(info.BridgePass)))
			defer func() { _ = client.Logout() }()

			// Create first draft in client.
			literal := fmt.Sprintf(`From: %v
To: %v
Date: Fri, 3 Feb 2023 01:04:32 +0100
Subject: Foo

Hello
`, info.Addresses[0], "bar@proton.local")

			require.NoError(t, client.Append("Drafts", nil, time.Now(), strings.NewReader(literal)))
			// Verify the draft is available in client.
			require.Eventually(t, func() bool {
				status, err := client.Status("Drafts", []go_imap.StatusItem{go_imap.StatusMessages})
				require.NoError(t, err)
				return status.Messages == 1
			}, 2*time.Second, time.Second)

			// Retrieve the new literal so we can have the Proton Message ID.
			messages, err := clientFetch(client, "Drafts")
			require.NoError(t, err)
			require.Equal(t, 1, len(messages))

			newLiteral, err := io.ReadAll(messages[0].GetBody(must(go_imap.ParseBodySectionName("BODY[]"))))
			require.NoError(t, err)
			logrus.Info(string(newLiteral))

			newLiteralID, newLiteralExternID := getGluonHeaderID(newLiteral)

			// Modify new literal.
			newLiteralModified := append(newLiteral, []byte(" world from client2")...) //nolint:gocritic

			func() {
				smtpClient, err := smtp.Dial(net.JoinHostPort(constants.Host, fmt.Sprint(b.GetSMTPPort())))
				require.NoError(t, err)
				defer func() { _ = smtpClient.Close() }()

				// Upgrade to TLS.
				require.NoError(t, smtpClient.StartTLS(&tls.Config{InsecureSkipVerify: true}))

				// Authorize with SASL PLAIN.
				require.NoError(t, smtpClient.Auth(sasl.NewPlainClient(
					info.Addresses[0],
					info.Addresses[0],
					string(info.BridgePass)),
				))

				// Send the message.
				require.NoError(t, smtpClient.SendMail(
					info.Addresses[0],
					[]string{"bar@proton.local"},
					bytes.NewReader(newLiteralModified),
				))
			}()

			// Append message to Sent as the imap client would.
			require.NoError(t, client.Append("Sent", nil, time.Now(), strings.NewReader(literal)))

			// Verify the sent message gets updated with the new literal.
			require.Eventually(t, func() bool {
				// Check if sent message matches the latest draft.
				messagesClient1, err := clientFetch(client, "Sent", "BODY[TEXT]", "BODY[]")
				require.NoError(t, err)

				if len(messagesClient1) != 1 {
					return false
				}

				sentLiteral, err := io.ReadAll(messagesClient1[0].GetBody(must(go_imap.ParseBodySectionName("BODY[]"))))
				require.NoError(t, err)

				sentLiteralID, sentLiteralExternID := getGluonHeaderID(sentLiteral)

				sentLiteralText, err := io.ReadAll(messagesClient1[0].GetBody(must(go_imap.ParseBodySectionName("BODY[TEXT]"))))
				require.NoError(t, err)

				sentLiteralStr := string(sentLiteralText)

				literalMatches := sentLiteralStr == "Hello\r\n world from client2\r\n"

				idIsDifferent := sentLiteralID != newLiteralID

				externIDMatches := sentLiteralExternID == newLiteralExternID

				return literalMatches && idIsDifferent && externIDMatches
			}, 2*time.Second, time.Second)
		})
	}, server.WithMessageDedup())
}
