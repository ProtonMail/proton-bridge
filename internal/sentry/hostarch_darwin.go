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
// along with Proton Mail Bridge. If not, see <https://www.gnu.org/licenses/>.

//go:build darwin
// +build darwin

package sentry

import (
	"github.com/elastic/go-sysinfo/types"
	"golang.org/x/sys/unix"
)

const translatedProcDarwin = "sysctl.proc_translated"

func getHostArch(host types.Host) string {
	// It is not possible to retrieve real hardware architecture once using
	// rosetta. But it is possible to detect the process translation if
	// rosetta is used.
	res, err := unix.SysctlRaw(translatedProcDarwin)
	if err != nil || len(res) > 4 {
		return host.Info().Architecture + "_err"
	}

	if res[0] == 1 {
		return host.Info().Architecture + "_rosetta"
	}

	return host.Info().Architecture
}
