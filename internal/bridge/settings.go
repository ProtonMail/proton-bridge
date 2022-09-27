package bridge

import (
	"context"

	"github.com/Masterminds/semver/v3"
	"github.com/ProtonMail/proton-bridge/v2/internal/updater"
	"github.com/ProtonMail/proton-bridge/v2/internal/vault"
)

func (bridge *Bridge) GetKeychainApp() (string, error) {
	vaultDir, err := bridge.locator.ProvideSettingsPath()
	if err != nil {
		return "", err
	}

	return vault.GetHelper(vaultDir)
}

func (bridge *Bridge) SetKeychainApp(helper string) error {
	vaultDir, err := bridge.locator.ProvideSettingsPath()
	if err != nil {
		return err
	}

	return vault.SetHelper(vaultDir, helper)
}

func (bridge *Bridge) GetIMAPPort() int {
	return bridge.vault.GetIMAPPort()
}

func (bridge *Bridge) SetIMAPPort(newPort int) error {
	if newPort == bridge.vault.GetIMAPPort() {
		return nil
	}

	if err := bridge.vault.SetIMAPPort(newPort); err != nil {
		return err
	}

	return bridge.restartIMAP(context.Background())
}

func (bridge *Bridge) GetIMAPSSL() bool {
	return bridge.vault.GetIMAPSSL()
}

func (bridge *Bridge) SetIMAPSSL(newSSL bool) error {
	if newSSL == bridge.vault.GetIMAPSSL() {
		return nil
	}

	if err := bridge.vault.SetIMAPSSL(newSSL); err != nil {
		return err
	}

	return bridge.restartIMAP(context.Background())
}

func (bridge *Bridge) GetSMTPPort() int {
	return bridge.vault.GetSMTPPort()
}

func (bridge *Bridge) SetSMTPPort(newPort int) error {
	if newPort == bridge.vault.GetSMTPPort() {
		return nil
	}

	if err := bridge.vault.SetSMTPPort(newPort); err != nil {
		return err
	}

	return bridge.restartSMTP()
}

func (bridge *Bridge) GetSMTPSSL() bool {
	return bridge.vault.GetSMTPSSL()
}

func (bridge *Bridge) SetSMTPSSL(newSSL bool) error {
	if newSSL == bridge.vault.GetSMTPSSL() {
		return nil
	}

	if err := bridge.vault.SetSMTPSSL(newSSL); err != nil {
		return err
	}

	return bridge.restartSMTP()
}

func (bridge *Bridge) GetGluonDir() string {
	return bridge.vault.GetGluonDir()
}

func (bridge *Bridge) SetGluonDir(ctx context.Context, newGluonDir string) error {
	if newGluonDir == bridge.GetGluonDir() {
		return nil
	}

	if err := bridge.closeIMAP(context.Background()); err != nil {
		return err
	}

	if err := moveDir(bridge.GetGluonDir(), newGluonDir); err != nil {
		return err
	}

	if err := bridge.vault.SetGluonDir(newGluonDir); err != nil {
		return err
	}

	imapServer, err := newIMAPServer(bridge.vault.GetGluonDir(), bridge.curVersion, bridge.tlsConfig)
	if err != nil {
		return err
	}

	for _, user := range bridge.users {
		imapConn, err := user.NewGluonConnector(ctx)
		if err != nil {
			return err
		}

		if err := imapServer.LoadUser(context.Background(), imapConn, user.GluonID(), user.GluonKey()); err != nil {
			return err
		}
	}

	bridge.imapServer = imapServer

	return bridge.serveIMAP()
}

func (bridge *Bridge) GetProxyAllowed() bool {
	return bridge.vault.GetProxyAllowed()
}

func (bridge *Bridge) SetProxyAllowed(allowed bool) error {
	if allowed {
		bridge.proxyDialer.AllowProxy()
	} else {
		bridge.proxyDialer.DisallowProxy()
	}

	return bridge.vault.SetProxyAllowed(allowed)
}

func (bridge *Bridge) GetShowAllMail() bool {
	return bridge.vault.GetShowAllMail()
}

func (bridge *Bridge) SetShowAllMail(show bool) error {
	panic("TODO")
}

func (bridge *Bridge) GetAutostart() bool {
	return bridge.vault.GetAutostart()
}

func (bridge *Bridge) SetAutostart(autostart bool) error {
	if err := bridge.vault.SetAutostart(autostart); err != nil {
		return err
	}

	var err error

	if autostart {
		err = bridge.autostarter.Enable()
	} else {
		err = bridge.autostarter.Disable()
	}

	return err
}

func (bridge *Bridge) GetAutoUpdate() bool {
	return bridge.vault.GetAutoUpdate()
}

func (bridge *Bridge) SetAutoUpdate(autoUpdate bool) error {
	if bridge.vault.GetAutoUpdate() == autoUpdate {
		return nil
	}

	if err := bridge.vault.SetAutoUpdate(autoUpdate); err != nil {
		return err
	}

	bridge.updateCheckCh <- struct{}{}

	return nil
}

func (bridge *Bridge) GetUpdateChannel() updater.Channel {
	return updater.Channel(bridge.vault.GetUpdateChannel())
}

func (bridge *Bridge) SetUpdateChannel(channel updater.Channel) error {
	if bridge.vault.GetUpdateChannel() == channel {
		return nil
	}

	if err := bridge.vault.SetUpdateChannel(channel); err != nil {
		return err
	}

	bridge.updateCheckCh <- struct{}{}

	return nil
}

func (bridge *Bridge) GetLastVersion() *semver.Version {
	return bridge.vault.GetLastVersion()
}

func (bridge *Bridge) GetFirstStart() bool {
	return bridge.vault.GetFirstStart()
}

func (bridge *Bridge) SetFirstStart(firstStart bool) error {
	return bridge.vault.SetFirstStart(firstStart)
}

func (bridge *Bridge) GetFirstStartGUI() bool {
	return bridge.vault.GetFirstStartGUI()
}

func (bridge *Bridge) SetFirstStartGUI(firstStart bool) error {
	return bridge.vault.SetFirstStartGUI(firstStart)
}

func (bridge *Bridge) GetColorScheme() string {
	return bridge.vault.GetColorScheme()
}

func (bridge *Bridge) SetColorScheme(colorScheme string) error {
	return bridge.vault.SetColorScheme(colorScheme)
}
