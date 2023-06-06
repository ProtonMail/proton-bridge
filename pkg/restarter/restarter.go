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

package restarter

import (
	"os"
	"strconv"
	"strings"

	"github.com/bradenaw/juniper/xslices"
	"github.com/sirupsen/logrus"
	"golang.org/x/sys/execabs"
)

const (
	BridgeCrashCount = "BRIDGE_CRASH_COUNT"
	MaxCrashRestarts = 10
)

type Restarter struct {
	restart bool
	crash   bool

	exe   string
	flags []string
}

func New(exe string) *Restarter {
	return &Restarter{exe: exe}
}

func (restarter *Restarter) Set(restart, crash bool) {
	restarter.restart = restart
	restarter.crash = crash
}

func (restarter *Restarter) Override(exe string) {
	restarter.exe = exe
}

func (restarter *Restarter) AddFlags(flags ...string) {
	restarter.flags = append(restarter.flags, flags...)
}

func (restarter *Restarter) Restart() {
	if !restarter.restart {
		return
	}

	if restarter.exe == "" {
		return
	}

	env := getEnvMap()

	if restarter.crash {
		env[BridgeCrashCount] = increment(env[BridgeCrashCount])
	} else {
		env[BridgeCrashCount] = "0"
	}

	//nolint:gosec
	cmd := execabs.Command(
		restarter.exe,
		xslices.Join(removeFlagsWithValue(removeFlag(os.Args[1:], "no-window"), "parent-pid", "session-id"), restarter.flags)...,
	)

	l := logrus.WithFields(logrus.Fields{
		"exe":        restarter.exe,
		"crashCount": env[BridgeCrashCount],
		"args":       cmd.Args,
	})

	if nCrash, err := strconv.Atoi(env[BridgeCrashCount]); err != nil {
		l.WithError(err).Error("Crash count is not integer, ignoring")
		return
	} else if nCrash >= MaxCrashRestarts {
		l.Error("Crash count is too high, ignoring")
		return
	}

	cmd.Env = getEnvList(env)

	if err := run(cmd); err != nil {
		l.WithError(err).Error("Failed to restart")
	}
}

func getEnvMap() map[string]string {
	env := make(map[string]string)

	for _, entry := range os.Environ() {
		if split := strings.SplitN(entry, "=", 2); len(split) == 2 {
			env[split[0]] = split[1]
		}
	}

	return env
}

func getEnvList(envMap map[string]string) []string {
	env := make([]string, 0, len(envMap))

	for key, value := range envMap {
		env = append(env, key+"="+value)
	}

	return env
}

func increment(value string) string {
	var valueInt int

	if parsed, err := strconv.Atoi(value); err == nil {
		valueInt = parsed
	}

	return strconv.Itoa(valueInt + 1)
}

// removeFlagWithValue removes a flag that requires a value from a list of command line parameters.
// The flag can be of the following form `-flag value`, `--flag value`, `-flag=value` or `--flags=value`.
func removeFlagWithValue(argList []string, flag string) []string {
	var result []string

	for i := 0; i < len(argList); i++ {
		arg := argList[i]
		// "detect the parameter form "-flag value" or "--paramName value"
		if (arg == "-"+flag) || (arg == "--"+flag) {
			i++
			continue
		}

		//  "detect the form "--flag=value" or "--flag=value"
		if strings.HasPrefix(arg, "-"+flag+"=") || (strings.HasPrefix(arg, "--"+flag+"=")) {
			continue
		}

		result = append(result, arg)
	}

	return result
}

// removeFlagWithValue removes flags that require a value from a list of command line parameters.
// The flags can be of the following form `-flag value`, `--flag value`, `-flag=value` or `--flags=value`.
func removeFlagsWithValue(argList []string, flags ...string) []string {
	for _, flag := range flags {
		argList = removeFlagWithValue(argList, flag)
	}

	return argList
}

func removeFlag(argList []string, flag string) []string {
	return xslices.Filter(argList, func(arg string) bool { return (arg != "-"+flag) && (arg != "--"+flag) })
}
