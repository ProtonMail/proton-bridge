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

package app

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/Masterminds/semver/v3"
	"github.com/ProtonMail/proton-bridge/v2/internal/updater"
	"github.com/ProtonMail/proton-bridge/v2/internal/vault"
	"github.com/allan-simon/go-singleinstance"
	"github.com/hashicorp/go-multierror"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

// nolint:gosec
func migrateOldSettings(vault *vault.Vault) error {
	configDir, err := os.UserConfigDir()
	if err != nil {
		return fmt.Errorf("failed to get user config dir: %w", err)
	}

	b, err := os.ReadFile(filepath.Join(configDir, "protonmail", "bridge", "prefs.json"))
	if err != nil {
		return fmt.Errorf("failed to read old prefs file: %w", err)
	}

	return migratePrefsToVault(vault, b)
}

// nolint:funlen
func migratePrefsToVault(vault *vault.Vault, b []byte) error {
	var prefs struct {
		IMAPPort int  `json:"user_port_imap,,string"`
		SMTPPort int  `json:"user_port_smtp,,string"`
		SMTPSSL  bool `json:"user_ssl_smtp,,string"`

		AutoUpdate    bool            `json:"autoupdate,,string"`
		UpdateChannel updater.Channel `json:"update_channel"`
		UpdateRollout float64         `json:"rollout,,string"`

		FirstStart    bool            `json:"first_time_start,,string"`
		FirstStartGUI bool            `json:"first_time_start_gui,,string"`
		ColorScheme   string          `json:"color_scheme"`
		LastVersion   *semver.Version `json:"last_used_version"`
		Autostart     bool            `json:"autostart,,string"`

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

	if err := vault.SetFirstStartGUI(prefs.FirstStartGUI); err != nil {
		errs = multierror.Append(errs, fmt.Errorf("failed to migrate first start GUI: %w", err))
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

	if err := vault.SetSyncWorkers(prefs.FetchWorkers); err != nil {
		errs = multierror.Append(errs, fmt.Errorf("failed to migrate sync workers: %w", err))
	}

	if err := vault.SetSyncBuffer(prefs.FetchWorkers); err != nil {
		errs = multierror.Append(errs, fmt.Errorf("failed to migrate sync buffer: %w", err))
	}

	if err := vault.SetSyncAttPool(prefs.AttachmentWorkers); err != nil {
		errs = multierror.Append(errs, fmt.Errorf("failed to migrate sync attachment pool: %w", err))
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
