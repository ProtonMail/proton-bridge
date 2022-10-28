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
	"context"
	"crypto/tls"
	"fmt"
	"net/http/cookiejar"
	"os"
	"time"

	"github.com/sirupsen/logrus"

	"github.com/ProtonMail/proton-bridge/v2/internal/bridge"
	"github.com/ProtonMail/proton-bridge/v2/internal/cookies"
	"github.com/ProtonMail/proton-bridge/v2/internal/events"
	"github.com/ProtonMail/proton-bridge/v2/internal/useragent"
	"github.com/ProtonMail/proton-bridge/v2/internal/vault"
	"gitlab.protontech.ch/go/liteapi"
)

func (t *testCtx) startBridge() error {
	// Bridge will enable the proxy by default at startup.
	t.mocks.ProxyCtl.EXPECT().AllowProxy()

	// Get the path to the vault.
	vaultDir, err := t.locator.ProvideSettingsPath()
	if err != nil {
		return err
	}

	// Get the default gluon path.
	gluonDir, err := t.locator.ProvideGluonPath()
	if err != nil {
		return err
	}

	// Create the vault.
	vault, corrupt, err := vault.New(vaultDir, gluonDir, t.storeKey)
	if err != nil {
		return err
	} else if corrupt {
		return fmt.Errorf("vault is corrupt")
	}

	// Create the underlying cookie jar.
	jar, err := cookiejar.New(nil)
	if err != nil {
		return err
	}

	// Create the persisting cookie jar.
	persister, err := cookies.NewCookieJar(jar, vault)
	if err != nil {
		return err
	}

	var logIMAP bool

	if len(os.Getenv("FEATURE_TEST_LOG_IMAP")) != 0 {
		logrus.SetLevel(logrus.TraceLevel)
		logIMAP = true
	}

	// Create the bridge.
	bridge, eventCh, err := bridge.New(
		// App stuff
		t.locator,
		vault,
		t.mocks.Autostarter,
		t.mocks.Updater,
		t.version,

		// API stuff
		t.api.GetHostURL(),
		persister,
		useragent.New(),
		t.mocks.TLSReporter,
		liteapi.NewDialer(t.netCtl, &tls.Config{InsecureSkipVerify: true}).GetRoundTripper(),
		t.mocks.ProxyCtl,

		// Logging stuff
		logIMAP,
		logIMAP,
		false,
	)
	if err != nil {
		return err
	}

	t.events.collectFrom(eventCh)

	// Wait for the users to be loaded.
	t.events.await(events.AllUsersLoaded{}, 30*time.Second)

	// Save the bridge to the context.
	t.bridge = bridge

	return nil
}

func (t *testCtx) stopBridge() error {
	if t.bridge == nil {
		return fmt.Errorf("bridge is not running")
	}

	t.bridge.Close(context.Background())

	t.bridge = nil

	return nil
}
