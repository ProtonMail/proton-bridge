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

package telemetry_test

import (
	"context"
	"testing"
	"time"

	"github.com/ProtonMail/proton-bridge/v3/internal/plan"
	"github.com/ProtonMail/proton-bridge/v3/internal/telemetry"
	"github.com/ProtonMail/proton-bridge/v3/internal/telemetry/mocks"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/require"
)

func TestHeartbeat_default_heartbeat(t *testing.T) {
	withHeartbeat(t, 1143, 1025, "/tmp", "defaultKeychain", func(hb *telemetry.Heartbeat, mock *mocks.MockHeartbeatManager) {
		data := telemetry.HeartbeatData{
			MeasurementGroup: "bridge.any.heartbeat",
			Event:            "bridge_heartbeat_new",
			Values: telemetry.HeartbeatValues{
				NumberConnectedAccounts: 1,
				Rollout:                 1,
			},
			Dimensions: telemetry.HeartbeatDimensions{
				AutoUpdateEnabled:       "true",
				AutoStartEnabled:        "true",
				BetaEnabled:             "false",
				DohEnabled:              "false",
				UseSplitMode:            "false",
				ShowAllMail:             "false",
				UseDefaultIMAPPort:      "true",
				UseDefaultSMTPPort:      "true",
				UseDefaultCacheLocation: "true",
				UseDefaultKeychain:      "true",
				ContactedByAppleNotes:   "false",
				PrevVersion:             "1.2.3",
				IMAPConnectionMode:      "ssl",
				SMTPConnectionMode:      "ssl",
				UserPlanGroup:           plan.Unknown,
			},
		}

		mock.EXPECT().IsTelemetryAvailable(context.Background()).Return(true)
		mock.EXPECT().GetLastHeartbeatSent().Return(time.Date(2022, 6, 4, 0, 0, 0, 0, time.UTC))
		mock.EXPECT().SendHeartbeat(context.Background(), &data).Return(true)
		mock.EXPECT().SetLastHeartbeatSent(gomock.Any()).Return(nil)

		hb.TrySending(context.Background())
	})
}

func TestHeartbeat_already_sent_heartbeat(t *testing.T) {
	withHeartbeat(t, 1143, 1025, "/tmp", "defaultKeychain", func(hb *telemetry.Heartbeat, mock *mocks.MockHeartbeatManager) {
		mock.EXPECT().IsTelemetryAvailable(context.Background()).Return(true)
		mock.EXPECT().GetLastHeartbeatSent().DoAndReturn(func() time.Time {
			curTime := time.Now()
			return time.Date(curTime.Year(), curTime.Month(), curTime.Day(), 0, 0, 0, 0, curTime.Location())
		})
		hb.TrySending(context.Background())
	})
}

func withHeartbeat(t *testing.T, imap, smtp int, cache, keychain string, tests func(hb *telemetry.Heartbeat, mock *mocks.MockHeartbeatManager)) {
	ctl := gomock.NewController(t)
	defer ctl.Finish()

	manager := mocks.NewMockHeartbeatManager(ctl)
	heartbeat := telemetry.NewHeartbeat(manager, imap, smtp, cache, keychain)

	heartbeat.SetRollout(0.1)
	heartbeat.SetNumberConnectedAccounts(1)
	heartbeat.SetSplitMode(false)
	heartbeat.SetAutoStart(true)
	heartbeat.SetAutoUpdate(true)
	heartbeat.SetBeta("stable")
	heartbeat.SetDoh(false)
	heartbeat.SetShowAllMail(false)
	heartbeat.SetIMAPConnectionMode(true)
	heartbeat.SetSMTPConnectionMode(true)
	heartbeat.SetIMAPPort(1143)
	heartbeat.SetSMTPPort(1025)
	heartbeat.SetCacheLocation("/tmp")
	heartbeat.SetKeyChainPref("defaultKeychain")
	heartbeat.SetPrevVersion("1.2.3")

	tests(&heartbeat, manager)
}

func Test_setRollout(t *testing.T) {
	hb := telemetry.Heartbeat{}
	type testStruct struct {
		val float64
		res int
	}

	tests := []testStruct{
		{0.02, 0},
		{0.04, 0},
		{0.09999, 0},
		{0.1, 1},
		{0.132323, 1},
		{0.2, 2},
		{0.25, 2},
		{0.7111, 7},
		{0.93, 9},
		{0.999, 9},
	}

	for _, test := range tests {
		hb.SetRollout(test.val)
		require.Equal(t, test.res, hb.GetRollout())
	}
}
