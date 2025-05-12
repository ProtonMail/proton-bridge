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

package observability

import (
	"strings"

	"github.com/ProtonMail/proton-bridge/v3/internal/updater"
)

// settingsGetter - interface that maps to bridge object methods such that we
// can pass the whole object instead of individual function callbacks.
type settingsGetter interface {
	GetCurrentUserAgent() string
	GetProxyAllowed() bool
	GetUpdateChannel() updater.Channel
}

// User agent mapping.
const (
	emailAgentAppleMail   = "apple_mail"
	emailAgentOutlook     = "outlook"
	emailAgentThunderbird = "thunderbird"
	emailAgentOther       = "other"
	emailAgentUnknown     = "unknown"
)

func matchUserAgent(userAgent string) string {
	if userAgent == "" {
		return emailAgentUnknown
	}

	userAgent = strings.ToLower(userAgent)
	switch {
	case strings.Contains(userAgent, "outlook"):
		return emailAgentOutlook
	case strings.Contains(userAgent, "thunderbird"):
		return emailAgentThunderbird
	case strings.Contains(userAgent, "mac") && strings.Contains(userAgent, "mail"):
		return emailAgentAppleMail
	case strings.Contains(userAgent, "mac") && strings.Contains(userAgent, "notes"):
		return emailAgentUnknown
	default:
		return emailAgentOther
	}
}

func getEnabled(value bool) string {
	if !value {
		return "disabled"
	}
	return "enabled"
}

func BucketIMAPConnections(val int) string {
	switch {
	case val < 10:
		return "<10"
	case val < 25:
		return "10-24"
	case val < 50:
		return "25-49"
	case val < 100:
		return "50-99"
	case val < 200:
		return "100-199"
	case val < 300:
		return "200-299"
	case val < 500:
		return "300-499"
	case val < 1000:
		return "500-999"
	case val < 2000:
		return "1000-1999"
	case val < 3000:
		return "2000-2999"
	default:
		return "3000+"
	}
}
