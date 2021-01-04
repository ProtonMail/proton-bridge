// Copyright (c) 2021 Proton Technologies AG
//
// This file is part of ProtonMail Bridge.Bridge.
//
// ProtonMail Bridge is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// ProtonMail Bridge is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with ProtonMail Bridge.  If not, see <https://www.gnu.org/licenses/>.

package context

import (
	"time"

	"github.com/ProtonMail/proton-bridge/internal/bridge"
	"github.com/ProtonMail/proton-bridge/internal/preferences"
	"github.com/ProtonMail/proton-bridge/internal/users"
	"github.com/ProtonMail/proton-bridge/pkg/listener"
)

// GetBridge returns bridge instance.
func (ctx *TestContext) GetBridge() *bridge.Bridge {
	return ctx.bridge
}

// withBridgeInstance creates a bridge instance for use in the test.
// TestContext has this by default once called with env variable TEST_APP=bridge.
func (ctx *TestContext) withBridgeInstance() {
	ctx.bridge = newBridgeInstance(ctx.t, ctx.cfg, ctx.credStore, ctx.listener, ctx.clientManager)
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

	ctx.bridge.StopWatchers()
	time.Sleep(50 * time.Millisecond)

	ctx.withBridgeInstance()

	return nil
}

// newBridgeInstance creates a new bridge instance configured to use the given config/credstore.
func newBridgeInstance(
	t *bddT,
	cfg *fakeConfig,
	credStore users.CredentialsStorer,
	eventListener listener.Listener,
	clientManager users.ClientManager,
) *bridge.Bridge {
	panicHandler := &panicHandler{t: t}
	pref := preferences.New(cfg)
	return bridge.New(cfg, pref, panicHandler, eventListener, clientManager, credStore)
}
