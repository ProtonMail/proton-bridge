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
// along with Proton Mail Bridge. If not, see <https://www.gnu.org/licenses/>.

package store

import (
	"os"
	"runtime"

	"golang.org/x/sys/unix"
)

func isFdCloseToULimit() bool {
	if runtime.GOOS != "darwin" && runtime.GOOS != "linux" {
		return false
	}

	var fdPath string
	switch runtime.GOOS {
	case "darwin":
		fdPath = "/dev/fd"
	case "linux":
		fdPath = "/proc/self/fd"
	}
	f, err := os.Open(fdPath)
	if err != nil {
		log.Warn("isFdCloseToULimit: ", err)
		return false
	}
	d, err := f.ReadDir(-1)
	if err != nil {
		log.Warn("isFdCloseToULimit: ", err)
		return false
	}
	fd := len(d) - 1

	var lim unix.Rlimit
	err = unix.Getrlimit(unix.RLIMIT_NOFILE, &lim)
	if err != nil {
		log.Print(err)
	}
	ulimit := lim.Max

	log.Info("File descriptor check: num goroutines ", runtime.NumGoroutine(), " fd ", fd, " ulimit ", ulimit)
	return fd >= int(0.95*float64(ulimit))
}
