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
	"time"

	"github.com/ProtonMail/proton-bridge/v3/internal/plan"
	"github.com/sirupsen/logrus"
)

const (
	dimensionSSL      = "ssl"
	dimensionStartTLS = "starttls"
)

type Availability interface {
	IsTelemetryAvailable(ctx context.Context) bool
}

type HeartbeatManager interface {
	Availability
	SendHeartbeat(ctx context.Context, heartbeat *HeartbeatData) bool
	GetLastHeartbeatSent() time.Time
	SetLastHeartbeatSent(time.Time) error
	GetHeartbeatPeriodicInterval() time.Duration
}

type HeartbeatValues struct {
	NumberConnectedAccounts int `json:"numberConnectedAccounts"`
	Rollout                 int `json:"rolloutPercentage"`
}

type HeartbeatDimensions struct {
	// Fields below correspond to bool
	AutoUpdateEnabled       string `json:"isAutoUpdateEnabled"`
	AutoStartEnabled        string `json:"isAutoStartEnabled"`
	BetaEnabled             string `json:"isBetaEnabled"`
	DohEnabled              string `json:"isDohEnabled"`
	UseSplitMode            string `json:"usesSplitMode"`
	ShowAllMail             string `json:"useAllMail"`
	UseDefaultIMAPPort      string `json:"useDefaultImapPort"`
	UseDefaultSMTPPort      string `json:"useDefaultSmtpPort"`
	UseDefaultCacheLocation string `json:"useDefaultCacheLocation"`
	UseDefaultKeychain      string `json:"useDefaultKeychain"`
	ContactedByAppleNotes   string `json:"isContactedByAppleNotes"`

	// Fields below are enums.
	PrevVersion        string `json:"prevVersion"` // Free text (exception)
	IMAPConnectionMode string `json:"imapConnectionMode"`
	SMTPConnectionMode string `json:"smtpConnectionMode"`
	UserPlanGroup      string `json:"bridgePlanGroup"`
}

type HeartbeatData struct {
	MeasurementGroup string
	Event            string
	Values           HeartbeatValues
	Dimensions       HeartbeatDimensions
}

type Heartbeat struct {
	log     *logrus.Entry
	manager HeartbeatManager
	metrics HeartbeatData

	defaultIMAPPort int
	defaultSMTPPort int
	defaultCache    string
	defaultKeychain string
	defaultUserPlan string
}

func NewHeartbeatDimensions() HeartbeatDimensions {
	return HeartbeatDimensions{
		AutoUpdateEnabled:       "false",
		AutoStartEnabled:        "false",
		BetaEnabled:             "false",
		DohEnabled:              "false",
		UseSplitMode:            "false",
		ShowAllMail:             "false",
		UseDefaultIMAPPort:      "false",
		UseDefaultSMTPPort:      "false",
		UseDefaultCacheLocation: "false",
		UseDefaultKeychain:      "false",
		ContactedByAppleNotes:   "false",

		PrevVersion:        "unknown",
		IMAPConnectionMode: dimensionSSL,
		SMTPConnectionMode: dimensionSSL,
		UserPlanGroup:      plan.Unknown,
	}
}
