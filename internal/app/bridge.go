package app

import (
	"encoding/base64"
	"fmt"
	"os"
	"runtime"

	"github.com/Masterminds/semver/v3"
	"github.com/ProtonMail/go-autostart"
	"github.com/ProtonMail/gopenpgp/v2/crypto"
	"github.com/ProtonMail/proton-bridge/v2/internal/bridge"
	"github.com/ProtonMail/proton-bridge/v2/internal/certs"
	"github.com/ProtonMail/proton-bridge/v2/internal/constants"
	"github.com/ProtonMail/proton-bridge/v2/internal/dialer"
	"github.com/ProtonMail/proton-bridge/v2/internal/locations"
	"github.com/ProtonMail/proton-bridge/v2/internal/updater"
	"github.com/ProtonMail/proton-bridge/v2/internal/useragent"
	"github.com/ProtonMail/proton-bridge/v2/internal/vault"
	"github.com/ProtonMail/proton-bridge/v2/internal/versioner"
	"github.com/ProtonMail/proton-bridge/v2/pkg/keychain"
	"github.com/sirupsen/logrus"
	"golang.org/x/exp/slices"
)

const vaultSecretName = "bridge-vault-key"

func newBridge(locations *locations.Locations, identifier *useragent.UserAgent) (*bridge.Bridge, error) {
	// Create the underlying dialer used by the bridge.
	// It only connects to trusted servers and reports any untrusted servers it finds.
	pinningDialer := dialer.NewPinningTLSDialer(
		dialer.NewBasicTLSDialer(constants.APIHost),
		dialer.NewTLSReporter(constants.APIHost, constants.AppVersion, identifier, dialer.TrustedAPIPins),
		dialer.NewTLSPinChecker(dialer.TrustedAPIPins),
	)

	// Create a proxy dialer which switches to a proxy if the request fails.
	proxyDialer := dialer.NewProxyTLSDialer(pinningDialer, constants.APIHost)

	// Create the autostarter.
	autostarter, err := newAutostarter()
	if err != nil {
		return nil, fmt.Errorf("could not create autostarter: %w", err)
	}

	// Create the update installer.
	updater, err := newUpdater(locations)
	if err != nil {
		return nil, fmt.Errorf("could not create updater: %w", err)
	}

	// Get the current bridge version.
	version, err := semver.NewVersion(constants.Version)
	if err != nil {
		return nil, fmt.Errorf("could not create version: %w", err)
	}

	// Create the encVault.
	encVault, insecure, corrupt, err := newVault(locations)
	if err != nil {
		return nil, fmt.Errorf("could not create vault: %w", err)
	} else if insecure {
		logrus.Warn("The vault key could not be retrieved; the vault will not be encrypted")
	} else if corrupt {
		logrus.Warn("The vault is corrupt and has been wiped")
	}

	// Install the certificates if needed.
	if installed := encVault.GetCertsInstalled(); !installed {
		if err := certs.NewInstaller().InstallCert(encVault.GetBridgeTLSCert()); err != nil {
			return nil, fmt.Errorf("failed to install certs: %w", err)
		}

		if err := encVault.SetCertsInstalled(true); err != nil {
			return nil, fmt.Errorf("failed to set certs installed: %w", err)
		}

		if err := encVault.SetCertsInstalled(true); err != nil {
			return nil, fmt.Errorf("could not set certs installed: %w", err)
		}
	}

	// Create a new bridge.
	bridge, err := bridge.New(
		constants.APIHost,
		locations,
		encVault,
		identifier,
		pinningDialer,
		dialer.CreateTransportWithDialer(proxyDialer),
		proxyDialer,
		autostarter,
		updater,
		version,
	)
	if err != nil {
		return nil, fmt.Errorf("could not create bridge: %w", err)
	}

	// If the vault could not be loaded properly, push errors to the bridge.
	switch {
	case insecure:
		bridge.PushError(vault.ErrInsecure)

	case corrupt:
		bridge.PushError(vault.ErrCorrupt)
	}

	return bridge, nil
}

func newVault(locations *locations.Locations) (*vault.Vault, bool, bool, error) {
	var insecure bool

	vaultDir, err := locations.ProvideSettingsPath()
	if err != nil {
		return nil, false, false, fmt.Errorf("could not get vault dir: %w", err)
	}

	var vaultKey []byte

	if key, err := getVaultKey(vaultDir); err != nil {
		insecure = true
	} else {
		vaultKey = key
	}

	gluonDir, err := locations.ProvideGluonPath()
	if err != nil {
		return nil, false, false, fmt.Errorf("could not provide gluon path: %w", err)
	}

	vault, corrupt, err := vault.New(vaultDir, gluonDir, vaultKey)
	if err != nil {
		return nil, false, false, fmt.Errorf("could not create vault: %w", err)
	}

	return vault, insecure, corrupt, nil
}

func getVaultKey(vaultDir string) ([]byte, error) {
	helper, err := vault.GetHelper(vaultDir)
	if err != nil {
		return nil, fmt.Errorf("could not get keychain helper: %w", err)
	}

	keychain, err := keychain.NewKeychain(helper, constants.KeyChainName)
	if err != nil {
		return nil, fmt.Errorf("could not create keychain: %w", err)
	}

	secrets, err := keychain.List()
	if err != nil {
		return nil, fmt.Errorf("could not list keychain: %w", err)
	}

	if !slices.Contains(secrets, vaultSecretName) {
		tok, err := crypto.RandomToken(32)
		if err != nil {
			return nil, fmt.Errorf("could not generate random token: %w", err)
		}

		if err := keychain.Put(vaultSecretName, base64.StdEncoding.EncodeToString(tok)); err != nil {
			return nil, fmt.Errorf("could not put keychain item: %w", err)
		}
	}

	_, keyEnc, err := keychain.Get(vaultSecretName)
	if err != nil {
		return nil, fmt.Errorf("could not get keychain item: %w", err)
	}

	keyDec, err := base64.StdEncoding.DecodeString(keyEnc)
	if err != nil {
		return nil, fmt.Errorf("could not decode keychain item: %w", err)
	}

	return keyDec, nil
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
