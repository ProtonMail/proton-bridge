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
	"strconv"
	"time"

	"github.com/ProtonMail/proton-bridge/internal/config/settings"
	"github.com/ProtonMail/proton-bridge/internal/constants"
	"github.com/ProtonMail/proton-bridge/internal/metrics"
	"github.com/ProtonMail/proton-bridge/internal/users"
	"github.com/ProtonMail/proton-bridge/pkg/pmapi"

	"github.com/ProtonMail/proton-bridge/pkg/listener"
	logrus "github.com/sirupsen/logrus"
)

var (
	log = logrus.WithField("pkg", "bridge") //nolint[gochecknoglobals]
)

type Bridge struct {
	*users.Users

	settings      SettingsProvider
	clientManager users.ClientManager

	userAgentClientName    string
	userAgentClientVersion string
	userAgentOS            string
}

func New(
	locations Locator,
	cache Cacher,
	s SettingsProvider,
	panicHandler users.PanicHandler,
	eventListener listener.Listener,
	clientManager users.ClientManager,
	credStorer users.CredentialsStorer,
) *Bridge {
	// Allow DoH before starting the app if the user has previously set this setting.
	// This allows us to start even if protonmail is blocked.
	if s.GetBool(settings.AllowProxyKey) {
		clientManager.AllowProxy()
	}

	storeFactory := newStoreFactory(cache, panicHandler, clientManager, eventListener)
	u := users.New(locations, panicHandler, eventListener, clientManager, credStorer, storeFactory, true)
	b := &Bridge{
		Users: u,

		settings:      s,
		clientManager: clientManager,
	}

	if s.GetBool(settings.FirstStartKey) {
		b.SendMetric(metrics.New(metrics.Setup, metrics.FirstStart, metrics.Label(constants.Version)))
		s.SetBool(settings.FirstStartKey, false)
	}

	go b.heartbeat()

	return b
}

// heartbeat sends a heartbeat signal once a day.
func (b *Bridge) heartbeat() {
	ticker := time.NewTicker(1 * time.Minute)

	for range ticker.C {
		next, err := strconv.ParseInt(b.settings.Get(settings.NextHeartbeatKey), 10, 64)
		if err != nil {
			continue
		}
		nextTime := time.Unix(next, 0)
		if time.Now().After(nextTime) {
			b.SendMetric(metrics.New(metrics.Heartbeat, metrics.Daily, metrics.NoLabel))
			nextTime = nextTime.Add(24 * time.Hour)
			b.settings.Set(settings.NextHeartbeatKey, strconv.FormatInt(nextTime.Unix(), 10))
		}
	}
}

// GetCurrentClient returns currently connected client (e.g. Thunderbird).
func (b *Bridge) GetCurrentClient() string {
	res := b.userAgentClientName
	if b.userAgentClientVersion != "" {
		res = res + " " + b.userAgentClientVersion
	}
	return res
}

// SetCurrentClient updates client info (e.g. Thunderbird) and sets the user agent
// on pmapi. By default no client is used, IMAP has to detect it on first login.
func (b *Bridge) SetCurrentClient(clientName, clientVersion string) {
	b.userAgentClientName = clientName
	b.userAgentClientVersion = clientVersion
	b.updateUserAgent()
}

// SetCurrentOS updates OS and sets the user agent on pmapi. By default we use
// `runtime.GOOS`, but this can be overridden in case of better detection.
func (b *Bridge) SetCurrentOS(os string) {
	b.userAgentOS = os
	b.updateUserAgent()
}

func (b *Bridge) updateUserAgent() {
	b.clientManager.SetUserAgent(b.userAgentClientName, b.userAgentClientVersion, b.userAgentOS)
}

// ReportBug reports a new bug from the user.
func (b *Bridge) ReportBug(osType, osVersion, description, accountName, address, emailClient string) error {
	c := b.clientManager.GetAnonymousClient()
	defer c.Logout()

	title := "[Bridge] Bug"
	report := pmapi.ReportReq{
		OS:          osType,
		OSVersion:   osVersion,
		Browser:     emailClient,
		Title:       title,
		Description: description,
		Username:    accountName,
		Email:       address,
	}

	if err := c.Report(report); err != nil {
		log.Error("Reporting bug failed: ", err)
		return err
	}

	log.Info("Bug successfully reported")

	return nil
}
