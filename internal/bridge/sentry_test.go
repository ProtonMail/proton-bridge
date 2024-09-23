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
	"context"
	"fmt"
	"net"
	"testing"

	"github.com/ProtonMail/gluon/liner"
	"github.com/ProtonMail/go-proton-api"
	"github.com/ProtonMail/go-proton-api/server"
	"github.com/ProtonMail/proton-bridge/v3/internal/bridge"
	"github.com/ProtonMail/proton-bridge/v3/internal/constants"
	"github.com/ProtonMail/proton-bridge/v3/internal/events"
	"github.com/stretchr/testify/require"
)

func TestBridge_Report(t *testing.T) {
	withEnv(t, func(ctx context.Context, s *server.Server, netCtl *proton.NetCtl, locator bridge.Locator, storeKey []byte) {
		withBridge(ctx, t, s.GetHostURL(), netCtl, locator, storeKey, func(b *bridge.Bridge, _ *bridge.Mocks) {
			syncCh, done := chToType[events.Event, events.SyncFinished](b.GetEvents(events.SyncFinished{}))
			defer done()

			// Log in the user.
			userID, err := b.LoginFull(ctx, username, password, nil, nil)
			require.NoError(t, err)

			// Wait until the sync has finished.
			require.Equal(t, userID, (<-syncCh).UserID)

			// Get the IMAP info.
			info, err := b.GetUserInfo(userID)
			require.NoError(t, err)
			require.True(t, info.State == bridge.Connected)

			// Dial the IMAP port.
			conn, err := net.Dial("tcp", fmt.Sprintf("%v:%v", constants.Host, b.GetIMAPPort()))
			require.NoError(t, err)
			defer func() { require.NoError(t, conn.Close()) }()

			// Read lines from the IMAP port.
			lineCh := liner.New(conn).Lines(func() error { return nil })

			// On connection, we should get the greeting.
			require.Contains(t, string((<-lineCh).Line), "* OK")

			// Send garbage data.
			must(conn.Write([]byte("tag garbage\r\n")))

			// Bridge will reply with BAD.
			require.Contains(t, string((<-lineCh).Line), "tag BAD")
		})
	})
}
