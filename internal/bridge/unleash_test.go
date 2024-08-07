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
// along with Proton Mail Bridge. If not, see <https://www.gnu.org/licenses/>.

package bridge_test

import (
	"context"
	"testing"
	"time"

	"github.com/ProtonMail/go-proton-api"
	"github.com/ProtonMail/go-proton-api/server"
	"github.com/ProtonMail/proton-bridge/v3/internal/bridge"
	"github.com/ProtonMail/proton-bridge/v3/internal/unleash"
	"github.com/stretchr/testify/require"
)

func Test_UnleashService(t *testing.T) {
	withEnv(t, func(ctx context.Context, s *server.Server, netCtl *proton.NetCtl, locator bridge.Locator, storeKey []byte) {
		unleash.ModifyPollPeriodAndJitter(500*time.Millisecond, 0)

		withBridge(ctx, t, s.GetHostURL(), netCtl, locator, storeKey, func(b *bridge.Bridge, _ *bridge.Mocks) {
			// Initial startup assumes there is no cached feature flags.
			require.Equal(t, b.GetFeatureFlagValue("test-1"), false)
			require.Equal(t, b.GetFeatureFlagValue("test-2"), false)
			require.Equal(t, b.GetFeatureFlagValue("test-3"), false)

			s.PushFeatureFlag("test-1")
			s.PushFeatureFlag("test-2")

			// Wait for poll.
			time.Sleep(time.Millisecond * 700)
			require.Equal(t, b.GetFeatureFlagValue("test-1"), true)
			require.Equal(t, b.GetFeatureFlagValue("test-2"), true)
			require.Equal(t, b.GetFeatureFlagValue("test-3"), false)

			s.PushFeatureFlag("test-3")
			time.Sleep(time.Millisecond * 700) // Wait for poll again
			require.Equal(t, b.GetFeatureFlagValue("test-1"), true)
			require.Equal(t, b.GetFeatureFlagValue("test-2"), true)
			require.Equal(t, b.GetFeatureFlagValue("test-3"), true)
		})

		// Wait for Bridge to close.
		time.Sleep(time.Millisecond * 500)

		// Second instance should have a feature flag cache file available. Therefore, all of the flags should evaluate to true on startup.
		withBridge(ctx, t, s.GetHostURL(), netCtl, locator, storeKey, func(b *bridge.Bridge, _ *bridge.Mocks) {
			require.Equal(t, b.GetFeatureFlagValue("test-1"), true)
			require.Equal(t, b.GetFeatureFlagValue("test-2"), true)
			require.Equal(t, b.GetFeatureFlagValue("test-3"), true)

			s.DeleteFeatureFlags()

			require.Equal(t, b.GetFeatureFlagValue("test-1"), true)
			require.Equal(t, b.GetFeatureFlagValue("test-2"), true)
			require.Equal(t, b.GetFeatureFlagValue("test-3"), true)

			time.Sleep(time.Millisecond * 700)

			require.Equal(t, b.GetFeatureFlagValue("test-1"), false)
			require.Equal(t, b.GetFeatureFlagValue("test-2"), false)
			require.Equal(t, b.GetFeatureFlagValue("test-3"), false)

			s.PushFeatureFlag("test-3")
			require.Equal(t, b.GetFeatureFlagValue("test-1"), false)
			require.Equal(t, b.GetFeatureFlagValue("test-2"), false)
			require.Equal(t, b.GetFeatureFlagValue("test-3"), false)

			time.Sleep(time.Millisecond * 700)
			require.Equal(t, b.GetFeatureFlagValue("test-1"), false)
			require.Equal(t, b.GetFeatureFlagValue("test-2"), false)
			require.Equal(t, b.GetFeatureFlagValue("test-3"), true)
		})
	})
}
