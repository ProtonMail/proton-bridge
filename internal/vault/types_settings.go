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

package vault

import (
	"math/rand"
	"runtime"
	"time"

	"github.com/ProtonMail/proton-bridge/v3/internal/updater"
	"github.com/ProtonMail/proton-bridge/v3/internal/useragent"
	"github.com/ProtonMail/proton-bridge/v3/pkg/ports"
)

type Settings struct {
	GluonDir string

	IMAPPort int
	SMTPPort int
	IMAPSSL  bool
	SMTPSSL  bool

	UpdateChannel updater.Channel
	UpdateRollout float64

	ColorScheme       string
	ProxyAllowed      bool
	ShowAllMail       bool
	Autostart         bool
	AutoUpdate        bool
	TelemetryDisabled bool

	LastVersion string
	FirstStart  bool

	MaxSyncMemory uint64

	LastUserAgent string

	LastHeartbeatSent time.Time

	PasswordArchive PasswordArchive

	// **WARNING**: These entry can't be removed until they vault has proper migration support.
	SyncWorkers int
	SyncAttPool int
}

const DefaultMaxSyncMemory = 2 * 1024 * uint64(1024*1024)

func GetDefaultSyncWorkerCount() int {
	const minSyncWorkers = 16

	syncWorkers := runtime.NumCPU() * 4

	if syncWorkers < minSyncWorkers {
		syncWorkers = minSyncWorkers
	}

	return syncWorkers
}

func newDefaultSettings(gluonDir string) Settings {
	syncWorkers := GetDefaultSyncWorkerCount()
	imapPort := ports.FindFreePortFrom(1143)
	smtpPort := ports.FindFreePortFrom(1025, imapPort)

	return Settings{
		GluonDir: gluonDir,

		IMAPPort: imapPort,
		SMTPPort: smtpPort,
		IMAPSSL:  false,
		SMTPSSL:  false,

		UpdateChannel: updater.DefaultUpdateChannel,
		UpdateRollout: rand.Float64(), //nolint:gosec

		ColorScheme:       "",
		ProxyAllowed:      false,
		ShowAllMail:       true,
		Autostart:         true,
		AutoUpdate:        true,
		TelemetryDisabled: false,

		LastVersion: "0.0.0",
		FirstStart:  true,

		MaxSyncMemory: DefaultMaxSyncMemory,
		SyncWorkers:   syncWorkers,
		SyncAttPool:   syncWorkers,

		LastUserAgent:     useragent.DefaultUserAgent,
		LastHeartbeatSent: time.Time{},

		PasswordArchive: PasswordArchive{},
	}
}
