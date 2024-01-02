// Copyright (c) 2024 Proton AG
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
	"context"
	"strconv"
	"time"

	"github.com/ProtonMail/proton-bridge/v3/internal/updater"
	"github.com/sirupsen/logrus"
)

func NewHeartbeat(manager HeartbeatManager, imapPort, smtpPort int, cacheDir, keychain string) Heartbeat {
	heartbeat := Heartbeat{
		log:     logrus.WithField("pkg", "telemetry"),
		manager: manager,
		metrics: HeartbeatData{
			MeasurementGroup: "bridge.any.usage",
			Event:            "bridge_heartbeat",
		},
		defaultIMAPPort: imapPort,
		defaultSMTPPort: smtpPort,
		defaultCache:    cacheDir,
		defaultKeychain: keychain,
	}
	return heartbeat
}

func (heartbeat *Heartbeat) SetRollout(val float64) {
	heartbeat.metrics.Dimensions.Rollout = strconv.Itoa(int(val * 100))
}

func (heartbeat *Heartbeat) SetNbAccount(val int) {
	heartbeat.metrics.Values.NbAccount = val
}

func (heartbeat *Heartbeat) SetAutoUpdate(val bool) {
	if val {
		heartbeat.metrics.Dimensions.AutoUpdate = dimensionON
	} else {
		heartbeat.metrics.Dimensions.AutoUpdate = dimensionOFF
	}
}

func (heartbeat *Heartbeat) SetAutoStart(val bool) {
	if val {
		heartbeat.metrics.Dimensions.AutoStart = dimensionON
	} else {
		heartbeat.metrics.Dimensions.AutoStart = dimensionOFF
	}
}

func (heartbeat *Heartbeat) SetBeta(val updater.Channel) {
	if val == updater.EarlyChannel {
		heartbeat.metrics.Dimensions.Beta = dimensionON
	} else {
		heartbeat.metrics.Dimensions.Beta = dimensionOFF
	}
}

func (heartbeat *Heartbeat) SetDoh(val bool) {
	if val {
		heartbeat.metrics.Dimensions.Doh = dimensionON
	} else {
		heartbeat.metrics.Dimensions.Doh = dimensionOFF
	}
}

func (heartbeat *Heartbeat) SetSplitMode(val bool) {
	if val {
		heartbeat.metrics.Dimensions.SplitMode = dimensionON
	} else {
		heartbeat.metrics.Dimensions.SplitMode = dimensionOFF
	}
}

func (heartbeat *Heartbeat) SetShowAllMail(val bool) {
	if val {
		heartbeat.metrics.Dimensions.ShowAllMail = dimensionON
	} else {
		heartbeat.metrics.Dimensions.ShowAllMail = dimensionOFF
	}
}

func (heartbeat *Heartbeat) SetIMAPConnectionMode(val bool) {
	if val {
		heartbeat.metrics.Dimensions.IMAPConnectionMode = dimensionSSL
	} else {
		heartbeat.metrics.Dimensions.IMAPConnectionMode = dimensionStartTLS
	}
}

func (heartbeat *Heartbeat) SetSMTPConnectionMode(val bool) {
	if val {
		heartbeat.metrics.Dimensions.SMTPConnectionMode = dimensionSSL
	} else {
		heartbeat.metrics.Dimensions.SMTPConnectionMode = dimensionStartTLS
	}
}

func (heartbeat *Heartbeat) SetIMAPPort(val int) {
	if val == heartbeat.defaultIMAPPort {
		heartbeat.metrics.Dimensions.IMAPPort = dimensionDefault
	} else {
		heartbeat.metrics.Dimensions.IMAPPort = dimensionCustom
	}
}

func (heartbeat *Heartbeat) SetSMTPPort(val int) {
	if val == heartbeat.defaultSMTPPort {
		heartbeat.metrics.Dimensions.SMTPPort = dimensionDefault
	} else {
		heartbeat.metrics.Dimensions.SMTPPort = dimensionCustom
	}
}

func (heartbeat *Heartbeat) SetCacheLocation(val string) {
	if val == heartbeat.defaultCache {
		heartbeat.metrics.Dimensions.CacheLocation = dimensionDefault
	} else {
		heartbeat.metrics.Dimensions.CacheLocation = dimensionCustom
	}
}

func (heartbeat *Heartbeat) SetKeyChainPref(val string) {
	if val == heartbeat.defaultKeychain {
		heartbeat.metrics.Dimensions.KeychainPref = dimensionDefault
	} else {
		heartbeat.metrics.Dimensions.KeychainPref = dimensionCustom
	}
}

func (heartbeat *Heartbeat) SetPrevVersion(val string) {
	heartbeat.metrics.Dimensions.PrevVersion = val
}

func (heartbeat *Heartbeat) TrySending(ctx context.Context) {
	if heartbeat.manager.IsTelemetryAvailable(ctx) {
		lastSent := heartbeat.manager.GetLastHeartbeatSent()
		now := time.Now()
		if now.Year() > lastSent.Year() || (now.Year() == lastSent.Year() && now.YearDay() > lastSent.YearDay()) {
			if !heartbeat.manager.SendHeartbeat(ctx, &heartbeat.metrics) {
				heartbeat.log.WithFields(logrus.Fields{
					"metrics": heartbeat.metrics,
				}).Error("Failed to send heartbeat")
				return
			}
			heartbeat.log.WithFields(logrus.Fields{
				"metrics": heartbeat.metrics,
			}).Info("Heartbeat sent")

			if err := heartbeat.manager.SetLastHeartbeatSent(now); err != nil {
				heartbeat.log.WithError(err).Warn("Cannot save last heartbeat sent to the vault.")
			}
		}
	}
}
