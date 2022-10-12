package app

import (
	"fmt"
	"net/http"
	"os"
	"runtime"

	"github.com/Masterminds/semver/v3"
	"github.com/ProtonMail/go-autostart"
	"github.com/ProtonMail/gopenpgp/v2/crypto"
	"github.com/ProtonMail/proton-bridge/v2/internal/bridge"
	"github.com/ProtonMail/proton-bridge/v2/internal/constants"
	"github.com/ProtonMail/proton-bridge/v2/internal/dialer"
	"github.com/ProtonMail/proton-bridge/v2/internal/locations"
	"github.com/ProtonMail/proton-bridge/v2/internal/sentry"
	"github.com/ProtonMail/proton-bridge/v2/internal/updater"
	"github.com/ProtonMail/proton-bridge/v2/internal/useragent"
	"github.com/ProtonMail/proton-bridge/v2/internal/vault"
	"github.com/ProtonMail/proton-bridge/v2/internal/versioner"
	"github.com/sirupsen/logrus"
	"github.com/urfave/cli/v2"
)

const vaultSecretName = "bridge-vault-key"

// withBridge creates creates and tears down the bridge.
func withBridge(
	c *cli.Context,
	locations *locations.Locations,
	identifier *useragent.UserAgent,
	reporter *sentry.Reporter,
	vault *vault.Vault,
	cookieJar http.CookieJar,
	fn func(*bridge.Bridge) error,
) error {
	// Get the current bridge version.
	version, err := semver.NewVersion(constants.Version)
	if err != nil {
		return fmt.Errorf("could not create version: %w", err)
	}

	// Create the underlying dialer used by the bridge.
	// It only connects to trusted servers and reports any untrusted servers it finds.
	pinningDialer := dialer.NewPinningTLSDialer(
		dialer.NewBasicTLSDialer(constants.APIHost),
		dialer.NewTLSReporter(constants.APIHost, constants.AppVersion(version.Original()), identifier, dialer.TrustedAPIPins),
		dialer.NewTLSPinChecker(dialer.TrustedAPIPins),
	)

	// Create a proxy dialer which switches to a proxy if the request fails.
	proxyDialer := dialer.NewProxyTLSDialer(pinningDialer, constants.APIHost)

	// Create the autostarter.
	autostarter, err := newAutostarter()
	if err != nil {
		return fmt.Errorf("could not create autostarter: %w", err)
	}

	// Create the update installer.
	updater, err := newUpdater(locations)
	if err != nil {
		return fmt.Errorf("could not create updater: %w", err)
	}

	// Create a new bridge.
	bridge, err := bridge.New(
		// The app stuff.
		locations,
		vault,
		autostarter,
		updater,
		version,

		// The API stuff.
		constants.APIHost,
		cookieJar,
		identifier,
		pinningDialer,
		dialer.CreateTransportWithDialer(proxyDialer),
		proxyDialer,

		// The logging stuff.
		c.String(flagLogIMAP) == "client" || c.String(flagLogIMAP) == "all",
		c.String(flagLogIMAP) == "server" || c.String(flagLogIMAP) == "all",
		c.Bool(flagLogSMTP),
	)
	if err != nil {
		return fmt.Errorf("could not create bridge: %w", err)
	}

	// Close the bridge when we exit.
	defer func() {
		if err := bridge.Close(c.Context); err != nil {
			logrus.WithError(err).Error("Failed to close bridge")
		}
	}()

	return fn(bridge)
}

func newAutostarter() (*autostart.App, error) {
	exe, err := os.Executable()
	if err != nil {
		return nil, err
	}

	return &autostart.App{
		Name:        constants.FullAppName,
		DisplayName: constants.FullAppName,
		Exec:        []string{exe, "--" + flagNoWindow},
	}, nil
}

func newUpdater(locations *locations.Locations) (*updater.Updater, error) {
	updatesDir, err := locations.ProvideUpdatesPath()
	if err != nil {
		return nil, fmt.Errorf("could not provide updates path: %w", err)
	}

	key, err := crypto.NewKeyFromArmored(updater.DefaultPublicKey)
	if err != nil {
		return nil, fmt.Errorf("could not create key from armored: %w", err)
	}

	verifier, err := crypto.NewKeyRing(key)
	if err != nil {
		return nil, fmt.Errorf("could not create key ring: %w", err)
	}

	return updater.NewUpdater(
		updater.NewInstaller(versioner.New(updatesDir)),
		verifier,
		constants.UpdateName,
		runtime.GOOS,
	), nil
}
