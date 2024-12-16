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
	"math"
	"strconv"
	"strings"
	"time"

	"github.com/ProtonMail/proton-bridge/v3/internal/plan"
	"github.com/ProtonMail/proton-bridge/v3/internal/updater"
	"github.com/sirupsen/logrus"
)

func NewHeartbeat(manager HeartbeatManager, imapPort, smtpPort int, cacheDir, keychain string) Heartbeat {
	heartbeat := Heartbeat{
		log:     logrus.WithField("pkg", "telemetry"),
		manager: manager,
		metrics: HeartbeatData{
			MeasurementGroup: "bridge.any.heartbeat",
			Event:            "bridge_heartbeat_new",
			Dimensions:       NewHeartbeatDimensions(),
		},
		defaultIMAPPort: imapPort,
		defaultSMTPPort: smtpPort,
		defaultCache:    cacheDir,
		defaultKeychain: keychain,
		defaultUserPlan: plan.Unknown,
	}
	return heartbeat
}

func (heartbeat *Heartbeat) SetRollout(val float64) {
	heartbeat.metrics.Values.Rollout = int(math.Floor(val * 10))
}

func (heartbeat *Heartbeat) GetRollout() int {
	return heartbeat.metrics.Values.Rollout
}

func (heartbeat *Heartbeat) SetNumberConnectedAccounts(val int) {
	heartbeat.metrics.Values.NumberConnectedAccounts = val
}

func (heartbeat *Heartbeat) SetAutoUpdate(val bool) {
	heartbeat.metrics.Dimensions.AutoUpdateEnabled = strconv.FormatBool(val)
}

func (heartbeat *Heartbeat) SetAutoStart(val bool) {
	heartbeat.metrics.Dimensions.AutoStartEnabled = strconv.FormatBool(val)
}

func (heartbeat *Heartbeat) SetBeta(val updater.Channel) {
	heartbeat.metrics.Dimensions.BetaEnabled = strconv.FormatBool(val == updater.EarlyChannel)
}

func (heartbeat *Heartbeat) SetDoh(val bool) {
	heartbeat.metrics.Dimensions.DohEnabled = strconv.FormatBool(val)
}

func (heartbeat *Heartbeat) SetSplitMode(val bool) {
	heartbeat.metrics.Dimensions.UseSplitMode = strconv.FormatBool(val)
}

func (heartbeat *Heartbeat) SetUserPlan(val string) {
	mappedUserPlan := plan.MapUserPlan(val)
	if plan.IsHigherPriority(heartbeat.metrics.Dimensions.UserPlanGroup, mappedUserPlan) {
		heartbeat.metrics.Dimensions.UserPlanGroup = val
	}
}

func (heartbeat *Heartbeat) SetContactedByAppleNotes(uaName string) {
	uaNameLowered := strings.ToLower(uaName)
	if strings.Contains(uaNameLowered, "mac") && strings.Contains(uaNameLowered, "notes") {
		heartbeat.metrics.Dimensions.ContactedByAppleNotes = strconv.FormatBool(true)
	}
}

func (heartbeat *Heartbeat) SetShowAllMail(val bool) {
	heartbeat.metrics.Dimensions.ShowAllMail = strconv.FormatBool(val)
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
	heartbeat.metrics.Dimensions.UseDefaultIMAPPort = strconv.FormatBool(val == heartbeat.defaultIMAPPort)
}

func (heartbeat *Heartbeat) SetSMTPPort(val int) {
	heartbeat.metrics.Dimensions.UseDefaultSMTPPort = strconv.FormatBool(val == heartbeat.defaultSMTPPort)
}

func (heartbeat *Heartbeat) SetCacheLocation(val string) {
	heartbeat.metrics.Dimensions.UseDefaultCacheLocation = strconv.FormatBool(val == heartbeat.defaultCache)
}

func (heartbeat *Heartbeat) SetKeyChainPref(val string) {
	heartbeat.metrics.Dimensions.UseDefaultKeychain = strconv.FormatBool(val == heartbeat.defaultKeychain)
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
