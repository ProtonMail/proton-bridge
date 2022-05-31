// Copyright (c) 2022 Proton AG
//
// This file is part of Proton Mail Bridge.Bridge.
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

package context

import (
	"time"

	"github.com/ProtonMail/go-autostart"
	"github.com/ProtonMail/proton-bridge/v2/internal/bridge"
	"github.com/ProtonMail/proton-bridge/v2/internal/config/settings"
	"github.com/ProtonMail/proton-bridge/v2/internal/config/useragent"
	"github.com/ProtonMail/proton-bridge/v2/internal/constants"
	"github.com/ProtonMail/proton-bridge/v2/internal/sentry"
	"github.com/ProtonMail/proton-bridge/v2/internal/store/cache"
	"github.com/ProtonMail/proton-bridge/v2/internal/users"
	"github.com/ProtonMail/proton-bridge/v2/pkg/listener"
	"github.com/ProtonMail/proton-bridge/v2/pkg/message"
	"github.com/ProtonMail/proton-bridge/v2/pkg/pmapi"
)

// GetBridge returns bridge instance.
func (ctx *TestContext) GetBridge() *bridge.Bridge {
	return ctx.bridge
}

// withBridgeInstance creates a bridge instance for use in the test.
// TestContext has this by default once called with env variable TEST_APP=bridge.
func (ctx *TestContext) withBridgeInstance() {
	ctx.bridge = newBridgeInstance(ctx.t, ctx.locations, ctx.cache, ctx.settings, ctx.credStore, ctx.listener, ctx.clientManager)
	ctx.users = ctx.bridge.Users
	ctx.addCleanupChecked(ctx.bridge.ClearData, "Cleaning bridge data")
}

// RestartBridge closes store for each user and recreates a bridge instance the same way as `withBridgeInstance`.
// NOTE: This is a very problematic method. It doesn't stop the goroutines doing the event loop and the sync.
//       These goroutines can continue to run and can cause problems or unexpected behaviour (especially
//       regarding authorization, because if an auth fails, it will log out the user).
//       To truly emulate bridge restart, we need a way to immediately stop those goroutines.
//       I have added a channel that waits up to one second for the event loop to stop, but that isn't great.
func (ctx *TestContext) RestartBridge() error {
	for _, user := range ctx.bridge.GetUsers() {
		_ = user.GetStore().Close()
	}

	time.Sleep(50 * time.Millisecond)

	ctx.withBridgeInstance()

	return nil
}

// newBridgeInstance creates a new bridge instance configured to use the given config/credstore.
// NOTE(GODT-1158): Need some tests with on-disk cache as well! Configurable in feature file or envvar?
func newBridgeInstance(
	t *bddT,
	locations bridge.Locator,
	cacheProvider bridge.CacheProvider,
	fakeSettings *fakeSettings,
	credStore users.CredentialsStorer,
	eventListener listener.Listener,
	clientManager pmapi.Manager,
) *bridge.Bridge {
	return bridge.New(
		locations,
		cacheProvider,
		fakeSettings,
		sentry.NewReporter("bridge", constants.Version, useragent.New()),
		&panicHandler{t: t},
		eventListener,
		cache.NewInMemoryCache(100*(1<<20)),
		message.NewBuilder(fakeSettings.GetInt(settings.FetchWorkers), fakeSettings.GetInt(settings.AttachmentWorkers)),
		clientManager,
		credStore,
		newFakeUpdater(),
		newFakeVersioner(),
		&autostart.App{
			Name: "bridge",
			Exec: []string{"bridge"},
		},
	)
}
