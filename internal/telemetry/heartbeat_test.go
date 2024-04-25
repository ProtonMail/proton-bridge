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

package telemetry_test

import (
	"context"
	"testing"
	"time"

	"github.com/ProtonMail/proton-bridge/v3/internal/telemetry"
	"github.com/ProtonMail/proton-bridge/v3/internal/telemetry/mocks"
	"github.com/golang/mock/gomock"
)

func TestHeartbeat_default_heartbeat(t *testing.T) {
	withHeartbeat(t, 1143, 1025, "/tmp", "defaultKeychain", func(hb *telemetry.Heartbeat, mock *mocks.MockHeartbeatManager) {
		data := telemetry.HeartbeatData{
			MeasurementGroup: "bridge.any.usage",
			Event:            "bridge_heartbeat",
			Values: telemetry.HeartbeatValues{
				NbAccount: 1,
			},
			Dimensions: telemetry.HeartbeatDimensions{
				AutoUpdate:         "on",
				AutoStart:          "on",
				Beta:               "off",
				Doh:                "off",
				SplitMode:          "off",
				ShowAllMail:        "off",
				IMAPConnectionMode: "ssl",
				SMTPConnectionMode: "ssl",
				IMAPPort:           "default",
				SMTPPort:           "default",
				CacheLocation:      "default",
				KeychainPref:       "default",
				PrevVersion:        "1.2.3",
				Rollout:            "10",
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
	heartbeat.SetNbAccount(1)
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
