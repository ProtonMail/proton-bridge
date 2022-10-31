// Copyright (c) 2022 Proton AG
//
// This file is part of Proton Mail Bridge.Bridge.
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

package mocks

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/logrusorgru/aurora"
)

type debug struct {
	verbosity int
	reqTag    string
}

func newDebug(reqTag string) *debug {
	return &debug{
		verbosity: verbosityLevelFromEnv(),
		reqTag:    reqTag,
	}
}

func verbosityLevelFromEnv() int {
	verbosityName := os.Getenv("VERBOSITY")
	switch strings.ToLower(verbosityName) {
	case "error", "fatal", "panic":
		return 0
	case "warning":
		return 1
	case "info":
		return 2
	case "debug", "trace":
		return 3
	}
	return 2
}

func (d *debug) printReq(command string) {
	if d.verbosity > 0 {
		fmt.Println(aurora.Green(fmt.Sprintf("Req %s: %s", d.reqTag, command)))
	}
}

func (d *debug) printRes(line string) {
	if d.verbosity > 1 {
		line = strings.ReplaceAll(line, "\n", "")
		line = strings.ReplaceAll(line, "\r", "")
		fmt.Println(aurora.Cyan(fmt.Sprintf("Res %s: %s", d.reqTag, line)))
	}
}

func (d *debug) printErr(line string) {
	fmt.Print(aurora.Bold(aurora.Red(fmt.Sprintf("Res %s: %s", d.reqTag, line))))
}

func (d *debug) printTime(diff time.Duration) {
	if d.verbosity > 0 {
		fmt.Println(aurora.Green(fmt.Sprintf("Time %s:%v", d.reqTag, diff)))
	}
}
