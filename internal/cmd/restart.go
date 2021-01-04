// Copyright (c) 2021 Proton Technologies AG
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
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"

	"github.com/ProtonMail/proton-bridge/internal/frontend"
	"github.com/ProtonMail/proton-bridge/pkg/config"
	"github.com/ProtonMail/proton-bridge/pkg/sentry"
	"github.com/urfave/cli"
)

const (
	// After how many crashes app gives up starting.
	maxAllowedCrashes = 10
)

var (
	// How many crashes happened so far in a row.
	// It will be filled from args by `filterRestartNumberFromArgs`.
	// Every call of `HandlePanic` will increase this number.
	// Then it will be passed as argument to the next try by `RestartApp`.
	numberOfCrashes = 0 //nolint[gochecknoglobals]
)

// filterRestartNumberFromArgs removes flag with a number how many restart we already did.
// See restartApp how that number is used.
func filterRestartNumberFromArgs() {
	tmp := os.Args[:0]
	for i, arg := range os.Args {
		if !strings.HasPrefix(arg, "--restart_") {
			tmp = append(tmp, arg)
			continue
		}
		var err error
		numberOfCrashes, err = strconv.Atoi(os.Args[i][10:])
		if err != nil {
			numberOfCrashes = maxAllowedCrashes
		}
	}
	os.Args = tmp
}

// DisableRestart disables restart once `RestartApp` is called.
func DisableRestart() {
	numberOfCrashes = maxAllowedCrashes
}

// RestartApp starts a new instance in background.
func RestartApp() {
	if numberOfCrashes >= maxAllowedCrashes {
		log.Error("Too many crashes")
		return
	}
	if exeFile, err := os.Executable(); err == nil {
		arguments := append(os.Args[1:], fmt.Sprintf("--restart_%d", numberOfCrashes))
		cmd := exec.Command(exeFile, arguments...) //nolint[gosec]
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		cmd.Stdin = os.Stdin
		if err := cmd.Start(); err != nil {
			log.Error("Restart failed: ", err)
		}
	}
}

// PanicHandler defines HandlePanic which can be used anywhere in defer.
type PanicHandler struct {
	AppName string
	Config  *config.Config
	Err     *error // Pointer to error of cli action.
}

// HandlePanic should be called in defer to ensure restart of app after error.
func (ph *PanicHandler) HandlePanic() {
	sentry.SkipDuringUnwind()

	r := recover()
	if r == nil {
		return
	}

	config.HandlePanic(ph.Config, fmt.Sprintf("Recover: %v", r))
	frontend.HandlePanic(ph.AppName)

	*ph.Err = cli.NewExitError("Panic and restart", 255)
	numberOfCrashes++
	log.Error("Restarting after panic")
	RestartApp()
	os.Exit(255)
}
