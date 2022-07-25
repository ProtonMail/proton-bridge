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

package pmapi

import (
	"runtime"

	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

type Config struct {
	// HostURL is the base URL of API.
	HostURL string

	// AppVersion sets version to headers of each request.
	AppVersion string

	// UserAgent sets user agent to headers of each request.
	// Used only if GetUserAgent is not set.
	UserAgent string

	// GetUserAgent is dynamic version of UserAgent.
	// Overrides UserAgent.
	GetUserAgent func() string

	// UpgradeApplicationHandler is used to notify when there is a force upgrade.
	UpgradeApplicationHandler func()

	// TLSIssueHandler is used to notify when there is a TLS issue.
	TLSIssueHandler func()
}

func NewConfig(appVersionName, appVersion string) Config {
	return Config{
		HostURL:    getRootURL(),
		AppVersion: getAPIOS() + cases.Title(language.Und).String(appVersionName) + "_" + appVersion,
	}
}

func (c *Config) getUserAgent() string {
	if c.GetUserAgent == nil {
		return c.UserAgent
	}
	return c.GetUserAgent()
}

// getAPIOS returns actual operating system.
func getAPIOS() string {
	switch os := runtime.GOOS; os {
	case "darwin": // nolint: const
		return "macOS"
	case "linux":
		return "Linux"
	case "windows":
		return "Windows"
	}
	return "Linux"
}
