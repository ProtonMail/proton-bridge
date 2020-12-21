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

// Package settings provides access to persistent user settings.
package settings

import (
	"fmt"
	"math/rand"
	"path/filepath"
	"time"
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
	AutoUpdateKey          = "autoupdate"
	CookiesKey             = "cookies"
	ReportOutgoingNoEncKey = "report_outgoing_email_without_encryption"
	LastVersionKey         = "last_used_version"
	UpdateChannelKey       = "update_channel"
	RolloutKey             = "rollout"
	PreferredKeychainKey   = "preferred_keychain"
)

type Settings struct {
	*keyValueStore

	settingsPath string
}

func New(settingsPath string) *Settings {
	s := &Settings{
		keyValueStore: newKeyValueStore(filepath.Join(settingsPath, "prefs.json")),
		settingsPath:  settingsPath,
	}

	s.setDefaultValues()

	return s
}

const (
	DefaultIMAPPort = "1143"
	DefaultSMTPPort = "1025"
	DefaultAPIPort  = "1042"
)

func (s *Settings) setDefaultValues() {
	s.setDefault(FirstStartKey, "true")
	s.setDefault(FirstStartGUIKey, "true")
	s.setDefault(NextHeartbeatKey, fmt.Sprintf("%v", time.Now().Unix()))
	s.setDefault(AllowProxyKey, "true")
	s.setDefault(AutostartKey, "true")
	s.setDefault(AutoUpdateKey, "true")
	s.setDefault(ReportOutgoingNoEncKey, "false")
	s.setDefault(LastVersionKey, "")
	s.setDefault(UpdateChannelKey, "")
	s.setDefault(RolloutKey, fmt.Sprintf("%v", rand.Float64()))
	s.setDefault(PreferredKeychainKey, "")

	s.setDefault(APIPortKey, DefaultAPIPort)
	s.setDefault(IMAPPortKey, DefaultIMAPPort)
	s.setDefault(SMTPPortKey, DefaultSMTPPort)

	// By default, stick to STARTTLS. If the user uses catalina+applemail they'll have to change to SSL.
	s.setDefault(SMTPSSLKey, "false")
}
