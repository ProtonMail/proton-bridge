// Copyright (c) 2020 Proton Technologies AG
//
// This file is part of ProtonMail Bridge.
//
// ProtonMail Bridge is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// ProtonMail Bridge is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with ProtonMail Bridge.  If not, see <https://www.gnu.org/licenses/>.

package cmd

import (
	"os"
	"path/filepath"
	"runtime"
	"runtime/pprof"
)

// MakeMemoryProfile generates memory pprof.
func MakeMemoryProfile() {
	name := "./mem.pprof"
	f, err := os.Create(name)
	if err != nil {
		log.Error("Could not create memory profile: ", err)
	}
	if abs, err := filepath.Abs(name); err == nil {
		name = abs
	}
	log.Info("Writing memory profile to ", name)
	runtime.GC() // get up-to-date statistics
	if err := pprof.WriteHeapProfile(f); err != nil {
		log.Error("Could not write memory profile: ", err)
	}
	_ = f.Close()
}
