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
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strconv"
	"strings"
)

func uLimit() int {
	if runtime.GOOS != "darwin" && runtime.GOOS != "linux" {
		return 0
	}
	out, err := exec.Command("bash", "-c", "ulimit -n").Output()
	if err != nil {
		log.Print(err)
		return 0
	}
	outStr := strings.Trim(string(out), " \n")
	num, err := strconv.Atoi(outStr)
	if err != nil {
		log.Print(err)
		return 0
	}
	return num
}

func isFdCloseToULimit() bool {
	if runtime.GOOS != "darwin" && runtime.GOOS != "linux" {
		return false
	}

	pid := fmt.Sprint(os.Getpid())
	out, err := exec.Command("lsof", "-p", pid).Output() //nolint:gosec
	if err != nil {
		log.Warn("isFdCloseToULimit: ", err)
		return false
	}
	lines := strings.Split(string(out), "\n")

	fd := len(lines) - 1
	ulimit := uLimit()
	log.Info("File descriptor check: num goroutines ", runtime.NumGoroutine(), " fd ", fd, " ulimit ", ulimit)
	return fd >= int(0.95*float64(ulimit))
}
