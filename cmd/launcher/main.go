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

package main

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"time"

	"github.com/Masterminds/semver/v3"
	"github.com/ProtonMail/gopenpgp/v2/crypto"
	"github.com/ProtonMail/proton-bridge/v2/internal/config/useragent"
	"github.com/ProtonMail/proton-bridge/v2/internal/constants"
	"github.com/ProtonMail/proton-bridge/v2/internal/crash"
	"github.com/ProtonMail/proton-bridge/v2/internal/locations"
	"github.com/ProtonMail/proton-bridge/v2/internal/logging"
	"github.com/ProtonMail/proton-bridge/v2/internal/sentry"
	"github.com/ProtonMail/proton-bridge/v2/internal/updater"
	"github.com/ProtonMail/proton-bridge/v2/internal/versioner"
	"github.com/bradenaw/juniper/xslices"
	"github.com/elastic/go-sysinfo"
	"github.com/elastic/go-sysinfo/types"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"golang.org/x/sys/execabs"
)

const (
	appName    = "Proton Mail Launcher"
	configName = "bridge"
	exeName    = "bridge"
	guiName    = "proton-bridge"

	FlagCLI      = "--cli"
	FlagCLIShort = "-c"
	FlagLauncher = "--launcher"
	FlagWait     = "--wait"
)

func main() { //nolint:funlen
	reporter := sentry.NewReporter(appName, constants.Version, useragent.New())

	crashHandler := crash.NewHandler(reporter.ReportException)
	defer crashHandler.HandlePanic()

	locationsProvider, err := locations.NewDefaultProvider(filepath.Join(constants.VendorName, configName))
	if err != nil {
		logrus.WithError(err).Fatal("Failed to get locations provider")
	}

	locations := locations.New(locationsProvider, configName)

	logsPath, err := locations.ProvideLogsPath()
	if err != nil {
		logrus.WithError(err).Fatal("Failed to get logs path")
	}
	crashHandler.AddRecoveryAction(logging.DumpStackTrace(logsPath))

	if err := logging.Init(logsPath); err != nil {
		logrus.WithError(err).Fatal("Failed to setup logging")
	}

	logging.SetLevel(os.Getenv("VERBOSITY"))

	updatesPath, err := locations.ProvideUpdatesPath()
	if err != nil {
		logrus.WithError(err).Fatal("Failed to get updates path")
	}

	key, err := crypto.NewKeyFromArmored(updater.DefaultPublicKey)
	if err != nil {
		logrus.WithError(err).Fatal("Failed to create new verification key")
	}

	kr, err := crypto.NewKeyRing(key)
	if err != nil {
		logrus.WithError(err).Fatal("Failed to create new verification keyring")
	}

	versioner := versioner.New(updatesPath)

	exeToLaunch := guiName
	args := os.Args[1:]
	if inCLIMode(args) {
		exeToLaunch = exeName
	}

	exe, err := getPathToUpdatedExecutable(exeToLaunch, versioner, kr, reporter)
	if err != nil {
		if exe, err = getFallbackExecutable(exeToLaunch, versioner); err != nil {
			logrus.WithError(err).Fatal("Failed to find any launchable executable")
		}
	}

	launcher, err := os.Executable()
	if err != nil {
		logrus.WithError(err).Fatal("Failed to determine path to launcher")
	}

	args, wait, mainExe := findAndStripWait(args)
	if wait {
		waitForProcessToFinish(mainExe)
	}

	cmd := execabs.Command(exe, appendLauncherPath(launcher, args)...) //nolint:gosec

	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	// On windows, if you use Run(), a terminal stays open; we don't want that.
	if //goland:noinspection GoBoolExpressions
	runtime.GOOS == "windows" {
		err = cmd.Start()
	} else {
		err = cmd.Run()
	}

	if err != nil {
		logrus.WithError(err).Fatal("Failed to launch")
	}
}

func appendLauncherPath(path string, args []string) []string {
	res := append([]string{}, args...)

	hasFlag := false

	for k, v := range res {
		if v != FlagLauncher {
			continue
		}

		hasFlag = true

		if k+1 >= len(res) {
			continue
		}

		res[k+1] = path
	}

	if !hasFlag {
		res = append(res, FlagLauncher, path)
	}

	return res
}

func inCLIMode(args []string) bool {
	return sliceContains(args, FlagCLI) || sliceContains(args, FlagCLIShort)
}

// sliceContains checks if a value is present in a list.
func sliceContains[T comparable](list []T, s T) bool {
	return xslices.Any(list, func(arg T) bool { return arg == s })
}

// findAndStrip check if a value is present in s list and remove all occurrences of the value from this list.
func findAndStrip[T comparable](slice []T, v T) (strippedList []T, found bool) {
	strippedList = xslices.Filter(slice, func(value T) bool {
		return value != v
	})
	return strippedList, len(strippedList) != len(slice)
}

// findAndStripWait Check for waiter flag get its value and clean them both.
func findAndStripWait(args []string) ([]string, bool, string) {
	res := append([]string{}, args...)

	hasFlag := false
	var value string

	for k, v := range res {
		if v != FlagWait {
			continue
		}
		if k+1 >= len(res) {
			continue
		}
		hasFlag = true
		value = res[k+1]
	}

	if hasFlag {
		res, _ = findAndStrip(res, FlagWait)
		res, _ = findAndStrip(res, value)
	}
	return res, hasFlag, value
}

func getPathToUpdatedExecutable(
	name string,
	versioner *versioner.Versioner,
	kr *crypto.KeyRing,
	reporter *sentry.Reporter,
) (string, error) {
	versions, err := versioner.ListVersions()
	if err != nil {
		return "", errors.Wrap(err, "failed to list available versions")
	}

	currentVersion, err := semver.StrictNewVersion(constants.Version)
	if err != nil {
		logrus.WithField("version", constants.Version).WithError(err).Error("Failed to parse current version")
	}

	for _, version := range versions {
		vlog := logrus.WithField("version", version)

		if err := version.VerifyFiles(kr); err != nil {
			vlog.WithError(err).Error("Files failed verification and will be removed")

			if err := reporter.ReportMessage(fmt.Sprintf("version %v failed verification: %v", version, err)); err != nil {
				vlog.WithError(err).Error("Failed to report corrupt update files")
			}

			if err := version.Remove(); err != nil {
				vlog.WithError(err).Error("Failed to remove files")
			}

			continue
		}

		// Skip versions that are less or equal to launcher version.
		if currentVersion != nil && !version.SemVer().GreaterThan(currentVersion) {
			continue
		}

		exe, err := version.GetExecutable(name)
		if err != nil {
			vlog.WithError(err).Error("Failed to get executable")
			continue
		}

		return exe, nil
	}

	return "", errors.New("no available newer versions")
}

func getFallbackExecutable(name string, versioner *versioner.Versioner) (string, error) {
	logrus.Info("Searching for fallback executable")

	launcher, err := os.Executable()
	if err != nil {
		return "", errors.Wrap(err, "failed to determine path to launcher")
	}

	return versioner.GetExecutableInDirectory(name, filepath.Dir(launcher))
}

// waitForProcessToFinish waits until the process with the given path is finished.
func waitForProcessToFinish(exePath string) {
	for {
		processes, err := sysinfo.Processes()
		if err != nil {
			logrus.WithError(err).Error("Could not determine running processes")
			return
		}

		exeInfo, err := os.Stat(exePath)
		if err != nil {
			logrus.WithError(err).WithField("file", exeInfo).Error("Could not retrieve file info")
			return
		}

		if xslices.Any(processes, func(process types.Process) bool {
			info, err := process.Info()
			if err != nil {
				logrus.WithError(err).Error("Could not retrieve process info")
			}

			return sameFile(exeInfo, info.Exe)
		}) {
			logrus.Infof("Waiting for %v to finish.", exeInfo.Name())
			time.Sleep(1 * time.Second)
			continue
		}

		return
	}
}

func sameFile(info os.FileInfo, path string) bool {
	pathInfo, err := os.Stat(path)
	if err != nil {
		logrus.WithError(err).WithField("file", path).Error("Could not retrieve file info")
	}

	return os.SameFile(pathInfo, info)
}
