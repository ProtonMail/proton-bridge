// Copyright (c) 2021 Proton Technologies AG
//
// This file is part of ProtonMail Bridge.
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

// Package bridge provides core functionality of Bridge app.
package bridge

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/ProtonMail/proton-bridge/internal/config/settings"
	"github.com/ProtonMail/proton-bridge/internal/constants"
	"github.com/ProtonMail/proton-bridge/internal/metrics"
	"github.com/ProtonMail/proton-bridge/internal/sentry"
	"github.com/ProtonMail/proton-bridge/internal/store/cache"
	"github.com/ProtonMail/proton-bridge/internal/updater"
	"github.com/ProtonMail/proton-bridge/internal/users"
	"github.com/ProtonMail/proton-bridge/pkg/message"
	"github.com/ProtonMail/proton-bridge/pkg/pmapi"

	"github.com/ProtonMail/proton-bridge/pkg/listener"
	logrus "github.com/sirupsen/logrus"
)

var log = logrus.WithField("pkg", "bridge") //nolint[gochecknoglobals]

type Bridge struct {
	*users.Users

	locations     Locator
	settings      SettingsProvider
	clientManager pmapi.Manager
	updater       Updater
	versioner     Versioner
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
		Users: u,

		locations:     locations,
		settings:      setting,
		clientManager: clientManager,
		updater:       updater,
		versioner:     versioner,
	}

	if setting.GetBool(settings.FirstStartKey) {
		if err := b.SendMetric(metrics.New(metrics.Setup, metrics.FirstStart, metrics.Label(constants.Version))); err != nil {
			logrus.WithError(err).Error("Failed to send metric")
		}

		setting.SetBool(settings.FirstStartKey, false)
	}

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

// ReportBug reports a new bug from the user.
func (b *Bridge) ReportBug(osType, osVersion, description, accountName, address, emailClient string) error {
	return b.clientManager.ReportBug(context.Background(), pmapi.ReportBugReq{
		OS:          osType,
		OSVersion:   osVersion,
		Browser:     emailClient,
		Title:       "[Bridge] Bug",
		Description: description,
		Username:    accountName,
		Email:       address,
	})
}

// GetUpdateChannel returns currently set update channel.
func (b *Bridge) GetUpdateChannel() updater.UpdateChannel {
	return updater.UpdateChannel(b.settings.Get(settings.UpdateChannelKey))
}

// SetUpdateChannel switches update channel.
// Downgrading to previous version (by switching from early to stable, for example)
// requires clearing all data including update files due to possibility of
// inconsistency between versions and absence of backwards migration scripts.
func (b *Bridge) SetUpdateChannel(channel updater.UpdateChannel) (needRestart bool, err error) {
	b.settings.Set(settings.UpdateChannelKey, string(channel))

	version, err := b.updater.Check()
	if err != nil {
		return false, err
	}

	// We have to deal right away only with downgrade - that action needs to
	// clear data and updates, and install bridge right away. But regular
	// upgrade can be left out for periodic check.
	if !b.updater.IsDowngrade(version) {
		return false, nil
	}

	if err := b.Users.ClearData(); err != nil {
		log.WithError(err).Error("Failed to clear data while downgrading channel")
	}

	if err := b.locations.ClearUpdates(); err != nil {
		log.WithError(err).Error("Failed to clear updates while downgrading channel")
	}

	if err := b.updater.InstallUpdate(version); err != nil {
		return false, err
	}

	return true, b.versioner.RemoveOtherVersions(version.Version)
}

// FactoryReset will remove all local cache and settings.
// We want to downgrade to latest stable version if user is early higher than stable.
// Setting the channel back to stable will do this for us.
func (b *Bridge) FactoryReset() {
	if _, err := b.SetUpdateChannel(updater.StableChannel); err != nil {
		log.WithError(err).Error("Failed to revert to stable update channel")
	}

	if err := b.Users.ClearUsers(); err != nil {
		log.WithError(err).Error("Failed to remove bridge users")
	}

	if err := b.Users.ClearData(); err != nil {
		log.WithError(err).Error("Failed to remove bridge data")
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
	// Set this back to the default location before enabling.
	b.settings.Set(settings.CacheLocationKey, "")

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

	return nil
}

func (b *Bridge) MigrateCache(from, to string) error {
	if err := b.Users.MigrateCache(from, to); err != nil {
		return err
	}

	b.settings.Set(settings.CacheLocationKey, to)

	return nil
}
