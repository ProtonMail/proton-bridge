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

//go:build !windows
// +build !windows

package store

import (
	"runtime"
	"syscall"
)

func getCurrentFDLimit() (int, error) {
	var limits syscall.Rlimit
	err := syscall.Getrlimit(syscall.RLIMIT_NOFILE, &limits)
	if err != nil {
		return 0, err
	}
	return int(limits.Cur), nil
}

func countOpenedFDs(limit int) int {
	openedFDs := 0

	for i := 0; i < limit; i++ {
		_, _, err := syscall.Syscall(syscall.SYS_FCNTL, uintptr(i), uintptr(syscall.F_GETFL), 0)
		if err == 0 {
			openedFDs++
		}
	}

	return openedFDs
}

func isFdCloseToULimit() bool {
	if runtime.GOOS != "darwin" && runtime.GOOS != "linux" {
		return false
	}

	limit, err := getCurrentFDLimit()
	if err != nil {
		log.WithError(err).Error("Cannot get current FD limit")
		return false
	}

	openedFDs := countOpenedFDs(limit)

	log.
		WithField("noGoroutines", runtime.NumCgoCall()).
		WithField("noFDs", openedFDs).
		WithField("limitFD", limit).
		Info("File descriptor check")
	return openedFDs >= int(0.95*float64(limit))
}
