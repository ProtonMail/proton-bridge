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
// along with Proton Mail Bridge.  If not, see <https://www.gnu.org/licenses/>.

package vault

import (
	"math/rand"
	"runtime"

	"github.com/ProtonMail/proton-bridge/v2/internal/updater"
)

type Settings struct {
	GluonDir string

	IMAPPort int
	SMTPPort int
	IMAPSSL  bool
	SMTPSSL  bool

	UpdateChannel updater.Channel
	UpdateRollout float64

	ColorScheme  string
	ProxyAllowed bool
	ShowAllMail  bool
	Autostart    bool
	AutoUpdate   bool

	LastVersion   string
	FirstStart    bool
	FirstStartGUI bool

	SyncWorkers int
	SyncBuffer  int
}

func newDefaultSettings(gluonDir string) Settings {
	return Settings{
		GluonDir: gluonDir,

		IMAPPort: 1143,
		SMTPPort: 1025,
		IMAPSSL:  false,
		SMTPSSL:  false,

		UpdateChannel: updater.DefaultUpdateChannel,
		UpdateRollout: rand.Float64(), //nolint:gosec

		ColorScheme:  "",
		ProxyAllowed: true,
		ShowAllMail:  true,
		Autostart:    false,
		AutoUpdate:   true,

		LastVersion:   "0.0.0",
		FirstStart:    true,
		FirstStartGUI: true,

		SyncWorkers: runtime.NumCPU(),
		SyncBuffer:  runtime.NumCPU(),
	}
}
