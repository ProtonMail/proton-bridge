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
// along with Proton Mail Bridge. If not, see <https://www.gnu.org/licenses/>.

// Package bridge provides core functionality of Bridge app.
package bridge

import (
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/Masterminds/semver/v3"
	"github.com/ProtonMail/go-autostart"
	"github.com/ProtonMail/proton-bridge/v2/internal/config/settings"
	"github.com/ProtonMail/proton-bridge/v2/internal/constants"
	"github.com/ProtonMail/proton-bridge/v2/internal/metrics"
	"github.com/ProtonMail/proton-bridge/v2/internal/sentry"
	"github.com/ProtonMail/proton-bridge/v2/internal/store/cache"
	"github.com/ProtonMail/proton-bridge/v2/internal/updater"
	"github.com/ProtonMail/proton-bridge/v2/internal/users"
	"github.com/ProtonMail/proton-bridge/v2/pkg/message"
	"github.com/ProtonMail/proton-bridge/v2/pkg/pmapi"

	"github.com/ProtonMail/proton-bridge/v2/pkg/listener"
	logrus "github.com/sirupsen/logrus"
)

var log = logrus.WithField("pkg", "bridge") //nolint:gochecknoglobals

var ErrLocalCacheUnavailable = errors.New("local cache is unavailable")

type Bridge struct {
	*users.Users

	locations     Locator
	settings      SettingsProvider
	clientManager pmapi.Manager
	updater       Updater
	versioner     Versioner
	cacheProvider CacheProvider
	autostart     *autostart.App
	// Bridge's global errors list.
	errors []error

	isFirstStart bool
	lastVersion  string
}

func New(
	locations Locator,
	cacheProvider CacheProvider,
	setting SettingsProvider,
	sentryReporter *sentry.Reporter,
	panicHandler users.PanicHandler,
	eventListener listener.Listener,
	cache cache.Cache,
	builder *message.Builder,
	clientManager pmapi.Manager,
	credStorer users.CredentialsStorer,
	updater Updater,
	versioner Versioner,
	autostart *autostart.App,
) *Bridge {
	// Allow DoH before starting the app if the user has previously set this setting.
	// This allows us to start even if protonmail is blocked.
	if setting.GetBool(settings.AllowProxyKey) {
		clientManager.AllowProxy()
	}

	u := users.New(
		locations,
		panicHandler,
		eventListener,
		clientManager,
		credStorer,
		newStoreFactory(cacheProvider, sentryReporter, panicHandler, eventListener, cache, builder),
	)

	b := &Bridge{
		Users:         u,
		locations:     locations,
		settings:      setting,
		clientManager: clientManager,
		updater:       updater,
		versioner:     versioner,
		cacheProvider: cacheProvider,
		autostart:     autostart,
		isFirstStart:  false,
	}

	if setting.GetBool(settings.FirstStartKey) {
		b.isFirstStart = true
		if err := b.SendMetric(metrics.New(metrics.Setup, metrics.FirstStart, metrics.Label(constants.Version))); err != nil {
			logrus.WithError(err).Error("Failed to send metric")
		}
		setting.SetBool(settings.FirstStartKey, false)
	}

	// Keep in bridge and update in settings the last used version.
	b.lastVersion = b.settings.Get(settings.LastVersionKey)
	b.settings.Set(settings.LastVersionKey, constants.Version)

	go b.heartbeat()

	return b
}

// heartbeat sends a heartbeat signal once a day.
func (b *Bridge) heartbeat() {
	for range time.Tick(time.Minute) {
		lastHeartbeatDay, err := strconv.ParseInt(b.settings.Get(settings.LastHeartbeatKey), 10, 64)
		if err != nil {
			continue
		}

		// If we're still on the same day, don't send a heartbeat.
		if time.Now().YearDay() == int(lastHeartbeatDay) {
			continue
		}

		// We're on the next (or a different) day, so send a heartbeat.
		if err := b.SendMetric(metrics.New(metrics.Heartbeat, metrics.Daily, metrics.NoLabel)); err != nil {
			logrus.WithError(err).Error("Failed to send heartbeat")
			continue
		}

		// Heartbeat was sent successfully so update the last heartbeat day.
		b.settings.Set(settings.LastHeartbeatKey, fmt.Sprintf("%v", time.Now().YearDay()))
	}
}

// GetUpdateChannel returns currently set update channel.
func (b *Bridge) GetUpdateChannel() updater.UpdateChannel {
	return updater.UpdateChannel(b.settings.Get(settings.UpdateChannelKey))
}

// SetUpdateChannel switches update channel.
func (b *Bridge) SetUpdateChannel(channel updater.UpdateChannel) {
	b.settings.Set(settings.UpdateChannelKey, string(channel))
}

func (b *Bridge) resetToLatestStable() error {
	version, err := b.updater.Check()
	if err != nil {
		// If we can not check for updates - just remove all local updates and reset to base installer version.
		// Not using `b.locations.ClearUpdates()` because `versioner.RemoveOtherVersions` can also handle
		// case when it is needed to remove currently running verion.
		if err := b.versioner.RemoveOtherVersions(semver.MustParse("0.0.0")); err != nil {
			log.WithError(err).Error("Failed to clear updates while downgrading channel")
		}
		return nil
	}

	// If current version is same as upstream stable version - do nothing.
	if version.Version.Equal(semver.MustParse(constants.Version)) {
		return nil
	}

	if err := b.updater.InstallUpdate(version); err != nil {
		return err
	}

	return b.versioner.RemoveOtherVersions(version.Version)
}

// FactoryReset will remove all local cache and settings.
// It will also downgrade to latest stable version if user is on early version.
func (b *Bridge) FactoryReset() {
	wasEarly := b.GetUpdateChannel() == updater.EarlyChannel

	b.settings.Set(settings.UpdateChannelKey, string(updater.StableChannel))

	if wasEarly {
		if err := b.resetToLatestStable(); err != nil {
			log.WithError(err).Error("Failed to reset to latest stable version")
		}
	}

	if err := b.Users.ClearData(); err != nil {
		log.WithError(err).Error("Failed to remove bridge data")
	}

	if err := b.Users.ClearUsers(); err != nil {
		log.WithError(err).Error("Failed to remove bridge users")
	}
}

// GetKeychainApp returns current keychain helper.
func (b *Bridge) GetKeychainApp() string {
	return b.settings.Get(settings.PreferredKeychainKey)
}

// SetKeychainApp sets current keychain helper.
func (b *Bridge) SetKeychainApp(helper string) {
	b.settings.Set(settings.PreferredKeychainKey, helper)
}

func (b *Bridge) EnableCache() error {
	if err := b.Users.EnableCache(); err != nil {
		return err
	}

	b.settings.SetBool(settings.CacheEnabledKey, true)

	return nil
}

func (b *Bridge) DisableCache() error {
	if err := b.Users.DisableCache(); err != nil {
		return err
	}

	b.settings.SetBool(settings.CacheEnabledKey, false)
	// Reset back to the default location when disabling.
	b.settings.Set(settings.CacheLocationKey, b.cacheProvider.GetDefaultMessageCacheDir())

	return nil
}

func (b *Bridge) MigrateCache(from, to string) error {
	if err := b.Users.MigrateCache(from, to); err != nil {
		return err
	}

	b.settings.Set(settings.CacheLocationKey, to)

	return nil
}

// SetProxyAllowed instructs the app whether to use DoH to access an API proxy if necessary.
// It also needs to work before the app is initialised (because we may need to use the proxy at startup).
func (b *Bridge) SetProxyAllowed(proxyAllowed bool) {
	b.settings.SetBool(settings.AllowProxyKey, proxyAllowed)
	if proxyAllowed {
		b.clientManager.AllowProxy()
	} else {
		b.clientManager.DisallowProxy()
	}
}

// GetProxyAllowed returns whether use of DoH is enabled to access an API proxy if necessary.
func (b *Bridge) GetProxyAllowed() bool {
	return b.settings.GetBool(settings.AllowProxyKey)
}

// AddError add an error to a global error list if it does not contain it yet. Adding nil is noop.
func (b *Bridge) AddError(err error) {
	if err == nil {
		return
	}
	if b.HasError(err) {
		return
	}

	b.errors = append(b.errors, err)
}

// DelError removes an error from global error list.
func (b *Bridge) DelError(err error) {
	for idx, val := range b.errors {
		if val == err {
			b.errors = append(b.errors[:idx], b.errors[idx+1:]...)
			return
		}
	}
}

// HasError returnes true if global error list contains an err.
func (b *Bridge) HasError(err error) bool {
	for _, val := range b.errors {
		if val == err {
			return true
		}
	}

	return false
}

// GetLastVersion returns the version which was used in previous execution of
// Bridge.
func (b *Bridge) GetLastVersion() string {
	return b.lastVersion
}

// IsFirstStart returns true when Bridge is running for first time or after
// factory reset.
func (b *Bridge) IsFirstStart() bool {
	return b.isFirstStart
}
