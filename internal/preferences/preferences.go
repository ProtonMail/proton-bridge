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

// Package preferences provides key names and defaults for preferences used in Bridge.
package preferences

import (
	"strconv"
	"time"

	"github.com/ProtonMail/proton-bridge/pkg/config"
	"github.com/sirupsen/logrus"
)

// Keys of preferences in JSON file.
const (
	FirstStartKey          = "first_time_start"
	FirstStartGUIKey       = "first_time_start_gui"
	NextHeartbeatKey       = "next_heartbeat"
	APIPortKey             = "user_port_api"
	IMAPPortKey            = "user_port_imap"
	SMTPPortKey            = "user_port_smtp"
	SMTPSSLKey             = "user_ssl_smtp"
	AllowProxyKey          = "allow_proxy"
	AutostartKey           = "autostart"
	CookiesKey             = "cookies"
	ReportOutgoingNoEncKey = "report_outgoing_email_without_encryption"
	LastVersionKey         = "last_used_version"
)

type configProvider interface {
	GetPreferencesPath() string
	GetDefaultAPIPort() int
	GetDefaultIMAPPort() int
	GetDefaultSMTPPort() int
}

var log = logrus.WithField("pkg", "store") //nolint[gochecknoglobals]

// New returns loaded preferences with Bridge defaults when values are not set yet.
func New(cfg configProvider) (pref *config.Preferences) {
	path := cfg.GetPreferencesPath()
	pref = config.NewPreferences(path)
	setDefaults(pref, cfg)

	log.WithField("path", path).Trace("Opened preferences")

	return
}

func setDefaults(preferences *config.Preferences, cfg configProvider) {
	preferences.SetDefault(FirstStartKey, "true")
	preferences.SetDefault(FirstStartGUIKey, "true")
	preferences.SetDefault(NextHeartbeatKey, strconv.FormatInt(time.Now().Unix(), 10))
	preferences.SetDefault(APIPortKey, strconv.Itoa(cfg.GetDefaultAPIPort()))
	preferences.SetDefault(IMAPPortKey, strconv.Itoa(cfg.GetDefaultIMAPPort()))
	preferences.SetDefault(SMTPPortKey, strconv.Itoa(cfg.GetDefaultSMTPPort()))
	preferences.SetDefault(AllowProxyKey, "true")
	preferences.SetDefault(AutostartKey, "true")
	preferences.SetDefault(ReportOutgoingNoEncKey, "false")
	preferences.SetDefault(LastVersionKey, "")

	// By default, stick to STARTTLS. If the user uses catalina+applemail they'll have to change to SSL.
	preferences.SetDefault(SMTPSSLKey, "false")
}
