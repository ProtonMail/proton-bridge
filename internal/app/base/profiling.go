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

package base

import (
	"os"
	"path/filepath"
	"runtime"
	"runtime/pprof"

	"github.com/sirupsen/logrus"
)

// startCPUProfile starts CPU pprof.
func startCPUProfile() {
	f, err := os.Create("./cpu.pprof")
	if err != nil {
		logrus.Fatal("Could not create CPU profile: ", err)
	}
	if err := pprof.StartCPUProfile(f); err != nil {
		logrus.Fatal("Could not start CPU profile: ", err)
	}
}

// makeMemoryProfile generates memory pprof.
func makeMemoryProfile() {
	name := "./mem.pprof"
	f, err := os.Create(name)
	if err != nil {
		logrus.Fatal("Could not create memory profile: ", err)
	}
	if abs, err := filepath.Abs(name); err == nil {
		name = abs
	}
	logrus.Info("Writing memory profile to ", name)
	runtime.GC() // get up-to-date statistics
	if err := pprof.WriteHeapProfile(f); err != nil {
		logrus.Fatal("Could not write memory profile: ", err)
	}
	_ = f.Close()
}
