// Copyright (c) 2023 Proton AG
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

package telemetry

import (
	"github.com/ProtonMail/proton-bridge/v3/internal/updater"
)

func NewHeartbeat(imapPort, smtpPort int, cacheDir, keychain string) Heartbeat {
	heartbeat := Heartbeat{
		Metrics: HeartbeatData{
			MeasurementGroup: "bridge.amy.usage",
			Event:            "bridge_heartbeat",
		},
		DefaultIMAPPort: imapPort,
		DefaultSMTPPort: smtpPort,
		DefaultCache:    cacheDir,
		DefaultKeychain: keychain,
	}
	return heartbeat
}

func (heartbeat *Heartbeat) SetRollout(val float64) {
	heartbeat.Metrics.Values.Rollout = int(val * 100)
}

func (heartbeat *Heartbeat) SetNbAccount(val int) {
	heartbeat.Metrics.Values.NbAccount = val
}

func (heartbeat *Heartbeat) SetAutoUpdate(val bool) {
	if val {
		heartbeat.Metrics.Dimensions.AutoUpdate = dimensionON
	} else {
		heartbeat.Metrics.Dimensions.AutoUpdate = dimensionOFF
	}
}

func (heartbeat *Heartbeat) SetAutoStart(val bool) {
	if val {
		heartbeat.Metrics.Dimensions.AutoStart = dimensionON
	} else {
		heartbeat.Metrics.Dimensions.AutoStart = dimensionOFF
	}
}

func (heartbeat *Heartbeat) SetBeta(val updater.Channel) {
	if val == updater.EarlyChannel {
		heartbeat.Metrics.Dimensions.Beta = dimensionON
	} else {
		heartbeat.Metrics.Dimensions.Beta = dimensionOFF
	}
}

func (heartbeat *Heartbeat) SetDoh(val bool) {
	if val {
		heartbeat.Metrics.Dimensions.Doh = dimensionON
	} else {
		heartbeat.Metrics.Dimensions.Doh = dimensionOFF
	}
}

func (heartbeat *Heartbeat) SetSplitMode(val bool) {
	if val {
		heartbeat.Metrics.Dimensions.SplitMode = dimensionON
	} else {
		heartbeat.Metrics.Dimensions.SplitMode = dimensionOFF
	}
}

func (heartbeat *Heartbeat) SetShowAllMail(val bool) {
	if val {
		heartbeat.Metrics.Dimensions.ShowAllMail = dimensionON
	} else {
		heartbeat.Metrics.Dimensions.ShowAllMail = dimensionOFF
	}
}

func (heartbeat *Heartbeat) SetIMAPConnectionMode(val bool) {
	if val {
		heartbeat.Metrics.Dimensions.IMAPConnectionMode = dimensionSSL
	} else {
		heartbeat.Metrics.Dimensions.IMAPConnectionMode = dimensionStartTLS
	}
}

func (heartbeat *Heartbeat) SetSMTPConnectionMode(val bool) {
	if val {
		heartbeat.Metrics.Dimensions.SMTPConnectionMode = dimensionSSL
	} else {
		heartbeat.Metrics.Dimensions.SMTPConnectionMode = dimensionStartTLS
	}
}

func (heartbeat *Heartbeat) SetIMAPPort(val int) {
	if val == heartbeat.DefaultIMAPPort {
		heartbeat.Metrics.Dimensions.IMAPPort = dimensionDefault
	} else {
		heartbeat.Metrics.Dimensions.IMAPPort = dimensionCustom
	}
}

func (heartbeat *Heartbeat) SetSMTPPort(val int) {
	if val == heartbeat.DefaultSMTPPort {
		heartbeat.Metrics.Dimensions.SMTPPort = dimensionDefault
	} else {
		heartbeat.Metrics.Dimensions.SMTPPort = dimensionCustom
	}
}

func (heartbeat *Heartbeat) SetCacheLocation(val string) {
	if val != heartbeat.DefaultCache {
		heartbeat.Metrics.Dimensions.CacheLocation = dimensionDefault
	} else {
		heartbeat.Metrics.Dimensions.CacheLocation = dimensionCustom
	}
}

func (heartbeat *Heartbeat) SetKeyChainPref(val string) {
	if val != heartbeat.DefaultKeychain {
		heartbeat.Metrics.Dimensions.KeychainPref = dimensionDefault
	} else {
		heartbeat.Metrics.Dimensions.KeychainPref = dimensionCustom
	}
}

func (heartbeat *Heartbeat) SetPrevVersion(val string) {
	heartbeat.Metrics.Dimensions.PrevVersion = val
}
