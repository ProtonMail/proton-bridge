// Copyright (c) 2025 Proton AG
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

package app

import (
	"encoding/json"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/Masterminds/semver/v3"
	"github.com/ProtonMail/proton-bridge/v3/internal/legacy/credentials"
	"github.com/ProtonMail/proton-bridge/v3/internal/locations"
	"github.com/ProtonMail/proton-bridge/v3/internal/logging"
	"github.com/ProtonMail/proton-bridge/v3/internal/updater"
	"github.com/ProtonMail/proton-bridge/v3/internal/vault"
	"github.com/ProtonMail/proton-bridge/v3/pkg/algo"
	"github.com/ProtonMail/proton-bridge/v3/pkg/keychain"
	"github.com/allan-simon/go-singleinstance"
	"github.com/hashicorp/go-multierror"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

// nolint:gosec
func migrateKeychainHelper(locations *locations.Locations) error {
	logrus.Trace("Checking if keychain helper needs to be migrated")

	settings, err := locations.ProvideSettingsPath()
	if err != nil {
		return fmt.Errorf("failed to get settings path: %w", err)
	}

	// If keychain helper file is already there do not migrate again.
	if keychainName, _ := vault.GetHelper(settings); keychainName != "" {
		return nil
	}

	configDir, err := os.UserConfigDir()
	if err != nil {
		return fmt.Errorf("failed to get user config dir: %w", err)
	}

	b, err := os.ReadFile(filepath.Join(configDir, "protonmail", "bridge", "prefs.json"))
	if errors.Is(err, fs.ErrNotExist) {
		return nil
	} else if err != nil {
		return fmt.Errorf("failed to read old prefs file: %w", err)
	}

	var prefs struct {
		Helper string `json:"preferred_keychain"`
	}

	if err := json.Unmarshal(b, &prefs); err != nil {
		return fmt.Errorf("failed to unmarshal old prefs file: %w", err)
	}

	err = vault.SetHelper(settings, prefs.Helper)
	if err == nil {
		logrus.Info("Keychain helper has been migrated")
	}
	return err
}

// nolint:gosec
func migrateOldSettings(v *vault.Vault) error {
	logrus.Info("Migrating settings")

	configDir, err := os.UserConfigDir()
	if err != nil {
		return fmt.Errorf("failed to get user config dir: %w", err)
	}

	return migrateOldSettingsWithDir(configDir, v)
}

// nolint:gosec
func migrateOldSettingsWithDir(configDir string, v *vault.Vault) error {
	b, err := os.ReadFile(filepath.Join(configDir, "protonmail", "bridge", "prefs.json"))
	if errors.Is(err, fs.ErrNotExist) {
		return nil
	} else if err != nil {
		return fmt.Errorf("failed to read old prefs file: %w", err)
	}

	if err := migratePrefsToVault(v, b); err != nil {
		return fmt.Errorf("failed to migrate prefs to vault: %w", err)
	}

	logrus.Info("Migrating TLS certificate")

	certPEM, err := os.ReadFile(filepath.Join(configDir, "protonmail", "bridge", "cert.pem"))
	if errors.Is(err, fs.ErrNotExist) {
		return nil
	} else if err != nil {
		return fmt.Errorf("failed to read old cert file: %w", err)
	}

	keyPEM, err := os.ReadFile(filepath.Join(configDir, "protonmail", "bridge", "key.pem"))
	if errors.Is(err, fs.ErrNotExist) {
		return nil
	} else if err != nil {
		return fmt.Errorf("failed to read old key file: %w", err)
	}

	return v.SetBridgeTLSCertKey(certPEM, keyPEM)
}

func migrateOldAccounts(locations *locations.Locations, keychains *keychain.List, v *vault.Vault) error {
	logrus.Info("Migrating accounts")

	settings, err := locations.ProvideSettingsPath()
	if err != nil {
		return fmt.Errorf("failed to get settings path: %w", err)
	}

	helper, err := vault.GetHelper(settings)
	if err != nil {
		return fmt.Errorf("failed to get helper: %w", err)
	}
	keychain, _, err := keychain.NewKeychain(helper, "bridge", keychains.GetHelpers(), keychains.GetDefaultHelper())
	if err != nil {
		return fmt.Errorf("failed to create keychain: %w", err)
	}

	store := credentials.NewStore(keychain)

	users, err := store.List()
	if err != nil {
		return fmt.Errorf("failed to create credentials store: %w", err)
	}

	var migrationErrors error

	for _, userID := range users {
		if err := migrateOldAccount(userID, store, v); err != nil {
			migrationErrors = multierror.Append(migrationErrors, err)
		}
	}

	return migrationErrors
}

func migrateOldAccount(userID string, store *credentials.Store, v *vault.Vault) error {
	l := logrus.WithField("userID", userID)
	l.Info("Migrating account")

	creds, err := store.Get(userID)
	if err != nil {
		return fmt.Errorf("failed to get user %q: %w", userID, err)
	}

	authUID, authRef, err := creds.SplitAPIToken()
	if err != nil {
		return fmt.Errorf("failed to split api token for user %q: %w", userID, err)
	}

	var primaryEmail string
	if len(creds.EmailList()) > 0 {
		primaryEmail = creds.EmailList()[0]
	}

	user, err := v.AddUser(creds.UserID, creds.Name, primaryEmail, authUID, authRef, creds.MailboxPassword)
	if err != nil {
		return fmt.Errorf("failed to add user %q: %w", userID, err)
	}

	l = l.WithField("username", logging.Sensitive(user.Username()))
	l.Info("Migrated account with random bridge password")

	defer func() {
		if err := user.Close(); err != nil {
			logrus.WithField("userID", userID).WithError(err).Error("Failed to close vault user after migration")
		}
	}()

	dec, err := algo.B64RawDecode([]byte(creds.BridgePassword))
	if err != nil {
		return fmt.Errorf("failed to decode bridge password for user %q: %w", userID, err)
	}

	if err := user.SetBridgePass(dec); err != nil {
		return fmt.Errorf("failed to set bridge password for user %q: %w", userID, err)
	}

	l = l.WithField("password", logging.Sensitive(string(algo.B64RawEncode(dec))))
	l.Info("Migrated existing bridge password")

	if !creds.IsCombinedAddressMode {
		if err := user.SetAddressMode(vault.SplitMode); err != nil {
			return fmt.Errorf("failed to set split address mode to user %q: %w", userID, err)
		}
	}

	return nil
}

func migratePrefsToVault(vault *vault.Vault, b []byte) error {
	var prefs struct {
		IMAPPort int  `json:"user_port_imap,,string"`
		SMTPPort int  `json:"user_port_smtp,,string"`
		SMTPSSL  bool `json:"user_ssl_smtp,,string"`

		AutoUpdate    bool            `json:"autoupdate,,string"`
		UpdateChannel updater.Channel `json:"update_channel"`
		UpdateRollout float64         `json:"rollout,,string"`

		FirstStart  bool            `json:"first_time_start,,string"`
		ColorScheme string          `json:"color_scheme"`
		LastVersion *semver.Version `json:"last_used_version"`
		Autostart   bool            `json:"autostart,,string"`

		AllowProxy        bool `json:"allow_proxy,,string"`
		FetchWorkers      int  `json:"fetch_workers,,string"`
		AttachmentWorkers int  `json:"attachment_workers,,string"`
		ShowAllMail       bool `json:"is_all_mail_visible,,string"`

		Cookies string `json:"cookies"`
	}

	if err := json.Unmarshal(b, &prefs); err != nil {
		return fmt.Errorf("failed to unmarshal old prefs file: %w", err)
	}

	var errs error

	if err := vault.SetIMAPPort(prefs.IMAPPort); err != nil {
		errs = multierror.Append(errs, fmt.Errorf("failed to migrate IMAP port: %w", err))
	}

	if err := vault.SetSMTPPort(prefs.SMTPPort); err != nil {
		errs = multierror.Append(errs, fmt.Errorf("failed to migrate SMTP port: %w", err))
	}

	if err := vault.SetSMTPSSL(prefs.SMTPSSL); err != nil {
		errs = multierror.Append(errs, fmt.Errorf("failed to migrate SMTP SSL: %w", err))
	}

	if err := vault.SetAutoUpdate(prefs.AutoUpdate); err != nil {
		errs = multierror.Append(errs, fmt.Errorf("failed to migrate auto update: %w", err))
	}

	if err := vault.SetUpdateChannel(prefs.UpdateChannel); err != nil {
		errs = multierror.Append(errs, fmt.Errorf("failed to migrate update channel: %w", err))
	}

	if err := vault.SetUpdateRollout(prefs.UpdateRollout); err != nil {
		errs = multierror.Append(errs, fmt.Errorf("failed to migrate rollout: %w", err))
	}

	if err := vault.SetFirstStart(prefs.FirstStart); err != nil {
		errs = multierror.Append(errs, fmt.Errorf("failed to migrate first start: %w", err))
	}

	if err := vault.SetColorScheme(prefs.ColorScheme); err != nil {
		errs = multierror.Append(errs, fmt.Errorf("failed to migrate color scheme: %w", err))
	}

	if err := vault.SetLastVersion(prefs.LastVersion); err != nil {
		errs = multierror.Append(errs, fmt.Errorf("failed to migrate last version: %w", err))
	}

	if err := vault.SetAutostart(prefs.Autostart); err != nil {
		errs = multierror.Append(errs, fmt.Errorf("failed to migrate autostart: %w", err))
	}

	if err := vault.SetProxyAllowed(prefs.AllowProxy); err != nil {
		errs = multierror.Append(errs, fmt.Errorf("failed to migrate allow proxy: %w", err))
	}

	if err := vault.SetShowAllMail(prefs.ShowAllMail); err != nil {
		errs = multierror.Append(errs, fmt.Errorf("failed to migrate show all mail: %w", err))
	}

	if err := vault.SetCookies([]byte(prefs.Cookies)); err != nil {
		errs = multierror.Append(errs, fmt.Errorf("failed to migrate cookies: %w", err))
	}

	return errs
}

func migrateOldVersions() (allErrors error) {
	cacheDir, cacheError := os.UserCacheDir()
	if cacheError != nil {
		allErrors = multierror.Append(allErrors, errors.Wrap(cacheError, "cannot get os cache"))
		return // not need to continue for now (with more migrations might be still ok to continue)
	}

	if err := killV2AppAndRemoveV2LockFiles(filepath.Join(cacheDir, "protonmail", "bridge", "bridge.lock")); err != nil {
		allErrors = multierror.Append(allErrors, errors.Wrap(err, "cannot migrate lockfiles"))
	}

	return
}

func killV2AppAndRemoveV2LockFiles(lockFilePathV2 string) error {
	l := logrus.WithField("path", lockFilePathV2)

	if _, err := os.Stat(lockFilePathV2); os.IsNotExist(err) {
		l.Debug("no v2 lockfile")
		return nil
	}

	lock, err := singleinstance.CreateLockFile(lockFilePathV2)

	if err == nil {
		l.Debug("no other v2 instance is running")

		if errClose := lock.Close(); errClose != nil {
			l.WithError(errClose).Error("Cannot close lock file")
		}

		return os.Remove(lockFilePathV2)
	}

	// The other instance is an older version, so we should kill it.
	pid, err := getPID(lockFilePathV2)
	if err != nil {
		return errors.Wrap(err, "cannot get v2 pid")
	}

	if err := killPID(pid); err != nil {
		return errors.Wrapf(err, "cannot kill v2 app (PID %d)", pid)
	}

	// Need to wait some time to release file lock
	time.Sleep(time.Second)

	return nil
}

func getPID(lockFilePath string) (int, error) {
	file, err := os.Open(filepath.Clean(lockFilePath))
	if err != nil {
		return 0, err
	}
	defer func() { _ = file.Close() }()

	rawPID := make([]byte, 10) // PID is probably up to 7 digits long, 10 should be enough
	n, err := file.Read(rawPID)
	if err != nil {
		return 0, err
	}

	return strconv.Atoi(strings.TrimSpace(string(rawPID[:n])))
}

func killPID(pid int) error {
	p, err := os.FindProcess(pid)
	if err != nil {
		return err
	}

	return p.Kill()
}
