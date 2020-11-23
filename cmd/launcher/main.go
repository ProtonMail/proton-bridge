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

package main

import (
	"os"
	"os/exec"
	"path/filepath"
	"runtime"

	"github.com/ProtonMail/go-appdir"
	"github.com/ProtonMail/gopenpgp/v2/crypto"
	"github.com/ProtonMail/proton-bridge/internal/constants"
	"github.com/ProtonMail/proton-bridge/internal/crash"
	"github.com/ProtonMail/proton-bridge/internal/locations"
	"github.com/ProtonMail/proton-bridge/internal/logging"
	"github.com/ProtonMail/proton-bridge/internal/updater"
	"github.com/ProtonMail/proton-bridge/internal/versioner"
	"github.com/ProtonMail/proton-bridge/pkg/sentry"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

const appName = "ProtonMail Launcher"

var (
	ConfigName = "" // nolint[gochecknoglobals]
	ExeName    = "" // nolint[gochecknoglobals]
)

func main() { // nolint[funlen]
	sentryReporter := sentry.NewReporter(appName, constants.Version)

	crashHandler := crash.NewHandler(sentryReporter.Report)
	defer crashHandler.HandlePanic()

	locations := locations.New(
		appdir.New(filepath.Join(constants.VendorName, ConfigName)),
		ConfigName,
	)

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

	exe, err := getPathToExecutable(ExeName, versioner, kr)
	if err != nil {
		if exe, err = getFallbackExecutable(ExeName, versioner); err != nil {
			logrus.WithError(err).Fatal("Failed to find any launchable executable")
		}
	}

	launcher, err := os.Executable()
	if err != nil {
		logrus.WithError(err).Fatal("Failed to determine path to launcher")
	}

	cmd := exec.Command(exe, appendLauncherPath(launcher, os.Args[1:])...) // nolint[gosec]

	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	// On windows, if you use Run(), a terminal stays open; we don't want that.
	if runtime.GOOS == "windows" {
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
		if v != "--launcher" {
			continue
		}

		hasFlag = true

		if k+1 >= len(res) {
			continue
		}

		res[k+1] = path
	}

	if !hasFlag {
		res = append(res, "--launcher", path)
	}

	return res
}

func getPathToExecutable(name string, versioner *versioner.Versioner, kr *crypto.KeyRing) (string, error) {
	versions, err := versioner.ListVersions()
	if err != nil {
		return "", errors.Wrap(err, "failed to list available versions")
	}

	for _, version := range versions {
		vlog := logrus.WithField("version", version)

		if err := version.VerifyFiles(kr); err != nil {
			vlog.WithError(err).Error("Failed to verify files")
			continue
		}

		exe, err := version.GetExecutable(name)
		if err != nil {
			vlog.WithError(err).Error("Failed to get executable")
			continue
		}

		return exe, nil
	}

	return "", errors.New("no available versions")
}

func getFallbackExecutable(name string, versioner *versioner.Versioner) (string, error) {
	logrus.Info("Searching for fallback executable")

	launcher, err := os.Executable()
	if err != nil {
		return "", errors.Wrap(err, "failed to determine path to launcher")
	}

	return versioner.GetExecutableInDirectory(name, filepath.Dir(launcher))
}
