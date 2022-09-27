package tests

import (
	"context"
	"fmt"

	"github.com/ProtonMail/proton-bridge/v2/internal/bridge"
	"github.com/ProtonMail/proton-bridge/v2/internal/events"
	"github.com/ProtonMail/proton-bridge/v2/internal/useragent"
	"github.com/ProtonMail/proton-bridge/v2/internal/vault"
)

func (t *testCtx) startBridge() error {
	// Bridge will enable the proxy by default at startup.
	t.mocks.ProxyDialer.EXPECT().AllowProxy()

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

	// Create the bridge.
	bridge, err := bridge.New(
		t.api.GetHostURL(),
		t.locator,
		vault,
		useragent.New(),
		t.mocks.TLSReporter,
		t.mocks.ProxyDialer,
		t.mocks.Autostarter,
		t.mocks.Updater,
		t.version,
	)
	if err != nil {
		return err
	}

	// Save the bridge t.
	t.bridge = bridge

	// Connect the event channels.
	t.userLoginCh = chToType[events.Event, events.UserLoggedIn](bridge.GetEvents(events.UserLoggedIn{}))
	t.userLogoutCh = chToType[events.Event, events.UserLoggedOut](bridge.GetEvents(events.UserLoggedOut{}))
	t.userDeletedCh = chToType[events.Event, events.UserDeleted](bridge.GetEvents(events.UserDeleted{}))
	t.userDeauthCh = chToType[events.Event, events.UserDeauth](bridge.GetEvents(events.UserDeauth{}))
	t.syncStartedCh = chToType[events.Event, events.SyncStarted](bridge.GetEvents(events.SyncStarted{}))
	t.syncFinishedCh = chToType[events.Event, events.SyncFinished](bridge.GetEvents(events.SyncFinished{}))
	t.forcedUpdateCh = chToType[events.Event, events.UpdateForced](bridge.GetEvents(events.UpdateForced{}))
	t.connStatusCh, _ = bridge.GetEvents(events.ConnStatusUp{}, events.ConnStatusDown{})
	t.updateCh, _ = bridge.GetEvents(events.UpdateAvailable{}, events.UpdateNotAvailable{}, events.UpdateInstalled{}, events.UpdateForced{})

	return nil
}

func (t *testCtx) stopBridge() error {
	if err := t.bridge.Close(context.Background()); err != nil {
		return err
	}

	t.bridge = nil

	return nil
}
