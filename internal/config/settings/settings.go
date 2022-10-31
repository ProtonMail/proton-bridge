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

// Package settings provides access to persistent user settings.
package settings

import (
	"fmt"
	"math/rand"
	"path/filepath"
	"time"
)

type Key string

// Keys of preferences in JSON file.
const (
	FirstStartKey          Key = "first_time_start"
	FirstStartGUIKey       Key = "first_time_start_gui"
	LastHeartbeatKey       Key = "last_heartbeat"
	APIPortKey             Key = "user_port_api"
	IMAPPortKey            Key = "user_port_imap"
	SMTPPortKey            Key = "user_port_smtp"
	SMTPSSLKey             Key = "user_ssl_smtp"
	AllowProxyKey          Key = "allow_proxy"
	AutostartKey           Key = "autostart"
	AutoUpdateKey          Key = "autoupdate"
	CookiesKey             Key = "cookies"
	LastVersionKey         Key = "last_used_version"
	UpdateChannelKey       Key = "update_channel"
	RolloutKey             Key = "rollout"
	PreferredKeychainKey   Key = "preferred_keychain"
	CacheEnabledKey        Key = "cache_enabled"
	CacheCompressionKey    Key = "cache_compression"
	CacheLocationKey       Key = "cache_location"
	CacheMinFreeAbsKey     Key = "cache_min_free_abs"
	CacheMinFreeRatKey     Key = "cache_min_free_rat"
	CacheConcurrencyRead   Key = "cache_concurrent_read"
	CacheConcurrencyWrite  Key = "cache_concurrent_write"
	IMAPWorkers            Key = "imap_workers"
	FetchWorkers           Key = "fetch_workers"
	AttachmentWorkers      Key = "attachment_workers"
	ColorScheme            Key = "color_scheme"
	RebrandingMigrationKey Key = "rebranding_migrated"
	IsAllMailVisible       Key = "is_all_mail_visible"
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
	s.setDefault(LastHeartbeatKey, fmt.Sprintf("%v", time.Now().YearDay()))
	s.setDefault(AllowProxyKey, "true")
	s.setDefault(AutostartKey, "true")
	s.setDefault(AutoUpdateKey, "true")
	s.setDefault(LastVersionKey, "")
	s.setDefault(UpdateChannelKey, "")
	s.setDefault(RolloutKey, fmt.Sprintf("%v", rand.Float64())) //nolint:gosec // G404 It is OK to use weak random number generator here
	s.setDefault(PreferredKeychainKey, "")
	s.setDefault(CacheEnabledKey, "true")
	s.setDefault(CacheCompressionKey, "true")
	s.setDefault(CacheLocationKey, "")
	s.setDefault(CacheMinFreeAbsKey, "250000000")
	s.setDefault(CacheMinFreeRatKey, "")
	s.setDefault(CacheConcurrencyRead, "16")
	s.setDefault(CacheConcurrencyWrite, "16")
	s.setDefault(IMAPWorkers, "16")
	s.setDefault(FetchWorkers, "16")
	s.setDefault(AttachmentWorkers, "16")
	s.setDefault(ColorScheme, "")

	s.setDefault(APIPortKey, DefaultAPIPort)
	s.setDefault(IMAPPortKey, DefaultIMAPPort)
	s.setDefault(SMTPPortKey, DefaultSMTPPort)

	// By default, stick to STARTTLS. If the user uses catalina+applemail they'll have to change to SSL.
	s.setDefault(SMTPSSLKey, "false")

	s.setDefault(IsAllMailVisible, "true")
}
