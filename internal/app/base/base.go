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

// Package base implements a common application base currently shared by bridge and IE.
// The base includes the following:
//  - access to standard filesystem locations like config, cache, logging dirs
//  - an extensible crash handler
//  - versioned cache directory
//  - persistent settings
//  - event listener
//  - credentials store
//  - pmapi ClientManager
// In addition, the base initialises logging and reacts to command line arguments
// which control the log verbosity and enable cpu/memory profiling.
package base

import (
	"os"
	"path/filepath"
	"runtime"
	"runtime/pprof"

	"github.com/Masterminds/semver/v3"
	"github.com/ProtonMail/go-appdir"
	"github.com/ProtonMail/gopenpgp/v2/crypto"
	"github.com/ProtonMail/proton-bridge/internal/api"
	"github.com/ProtonMail/proton-bridge/internal/config/cache"
	"github.com/ProtonMail/proton-bridge/internal/config/settings"
	"github.com/ProtonMail/proton-bridge/internal/config/tls"
	"github.com/ProtonMail/proton-bridge/internal/constants"
	"github.com/ProtonMail/proton-bridge/internal/cookies"
	"github.com/ProtonMail/proton-bridge/internal/crash"
	"github.com/ProtonMail/proton-bridge/internal/events"
	"github.com/ProtonMail/proton-bridge/internal/locations"
	"github.com/ProtonMail/proton-bridge/internal/logging"
	"github.com/ProtonMail/proton-bridge/internal/updater"
	"github.com/ProtonMail/proton-bridge/internal/users/credentials"
	"github.com/ProtonMail/proton-bridge/internal/versioner"
	"github.com/ProtonMail/proton-bridge/pkg/listener"
	"github.com/ProtonMail/proton-bridge/pkg/pmapi"
	"github.com/ProtonMail/proton-bridge/pkg/sentry"
	"github.com/allan-simon/go-singleinstance"
	"github.com/sirupsen/logrus"
	"github.com/urfave/cli/v2"
)

type Base struct {
	CrashHandler *crash.Handler
	Locations    *locations.Locations
	Settings     *settings.Settings
	Lock         *os.File
	Cache        *cache.Cache
	Listener     listener.Listener
	Creds        *credentials.Store
	CM           *pmapi.ClientManager
	Updater      *updater.Updater
	Versioner    *versioner.Versioner
	TLS          *tls.TLS

	name  string
	usage string

	restart bool
}

func New( // nolint[funlen]
	appName,
	appUsage,
	configName,
	updateURLName,
	keychainName,
	cacheVersion string,
) (*Base, error) {
	sentryReporter := sentry.NewReporter(appName, constants.Version)

	crashHandler := crash.NewHandler(
		sentryReporter.Report,
		crash.ShowErrorNotification(appName),
	)
	defer crashHandler.HandlePanic()

	locations := locations.New(
		appdir.New(filepath.Join(constants.VendorName, configName)),
		configName,
	)
	if err := locations.Clean(); err != nil {
		return nil, err
	}

	logsPath, err := locations.ProvideLogsPath()
	if err != nil {
		return nil, err
	}
	if err := logging.Init(logsPath); err != nil {
		return nil, err
	}
	crashHandler.AddRecoveryAction(logging.DumpStackTrace(logsPath))

	settingsPath, err := locations.ProvideSettingsPath()
	if err != nil {
		return nil, err
	}
	settingsObj := settings.New(settingsPath)

	lock, err := singleinstance.CreateLockFile(locations.GetLockFile())
	if err != nil {
		logrus.Warnf("%v is already running", appName)
		return nil, api.CheckOtherInstanceAndFocus(settingsObj.GetInt(settings.APIPortKey))
	}

	cachePath, err := locations.ProvideCachePath()
	if err != nil {
		return nil, err
	}
	cache, err := cache.New(cachePath, cacheVersion)
	if err != nil {
		return nil, err
	}
	if err := cache.RemoveOldVersions(); err != nil {
		return nil, err
	}

	listener := listener.New()
	events.SetupEvents(listener)

	// NOTE: If we can't load the credentials for whatever reason,
	// do we really want to error out? Need to signal to frontend.
	creds, err := credentials.NewStore(keychainName)
	if err != nil {
		logrus.WithError(err).Error("Could not get credentials store")
		listener.Emit(events.CredentialsErrorEvent, err.Error())
	}

	jar, err := cookies.NewCookieJar(settingsObj)
	if err != nil {
		return nil, err
	}

	cm := pmapi.NewClientManager(pmapi.GetAPIConfig(configName, constants.Version))
	cm.SetRoundTripper(pmapi.GetRoundTripper(cm, listener))
	cm.SetCookieJar(jar)

	sentryReporter.SetUserAgentProvider(cm)

	tls := tls.New(settingsPath)

	key, err := crypto.NewKeyFromArmored(updater.DefaultPublicKey)
	if err != nil {
		return nil, err
	}

	kr, err := crypto.NewKeyRing(key)
	if err != nil {
		return nil, err
	}

	updatesDir, err := locations.ProvideUpdatesPath()
	if err != nil {
		return nil, err
	}

	versioner := versioner.New(updatesDir)

	installer := updater.NewInstaller(versioner)

	updater := updater.New(
		cm,
		installer,
		settingsObj,
		kr,
		semver.MustParse(constants.Version),
		updateURLName,
		runtime.GOOS,
	)

	return &Base{
		CrashHandler: crashHandler,
		Locations:    locations,
		Settings:     settingsObj,
		Lock:         lock,
		Cache:        cache,
		Listener:     listener,
		Creds:        creds,
		CM:           cm,
		Updater:      updater,
		Versioner:    versioner,
		TLS:          tls,

		name:  appName,
		usage: appUsage,
	}, nil
}

func (b *Base) NewApp(action func(*Base, *cli.Context) error) *cli.App {
	app := cli.NewApp()

	app.Name = b.name
	app.Usage = b.usage
	app.Version = constants.Version
	app.Action = b.run(action)
	app.Flags = []cli.Flag{
		&cli.BoolFlag{
			Name:    "cpu-prof",
			Aliases: []string{"p"},
			Usage:   "Generate CPU profile",
		},
		&cli.BoolFlag{
			Name:    "mem-prof",
			Aliases: []string{"m"},
			Usage:   "Generate memory profile",
		},
		&cli.StringFlag{
			Name:    "log-level",
			Aliases: []string{"l"},
			Usage:   "Set the log level (one of panic, fatal, error, warn, info, debug)",
		},
		&cli.BoolFlag{
			Name:    "cli",
			Aliases: []string{"c"},
			Usage:   "Use command line interface",
		},
		&cli.StringFlag{
			Name:   "restart",
			Usage:  "The number of times the application has already restarted",
			Hidden: true,
		},
		&cli.StringFlag{
			Name:   "launcher",
			Usage:  "The launcher to use to restart the application",
			Hidden: true,
		},
	}

	return app
}

// SetToRestart sets the app to restart the next time it is closed.
func (b *Base) SetToRestart() {
	b.restart = true
}

func (b *Base) run(appMainLoop func(*Base, *cli.Context) error) cli.ActionFunc { // nolint[funlen]
	return func(c *cli.Context) error {
		defer b.CrashHandler.HandlePanic()
		defer func() { _ = b.Lock.Close() }()

		if doCPUProfile := c.Bool("cpu-prof"); doCPUProfile {
			startCPUProfile()
			defer pprof.StopCPUProfile()
		}

		if doMemoryProfile := c.Bool("mem-prof"); doMemoryProfile {
			defer makeMemoryProfile()
		}

		logging.SetLevel(c.String("log-level"))

		logrus.
			WithField("appName", b.name).
			WithField("version", constants.Version).
			WithField("revision", constants.Revision).
			WithField("build", constants.BuildTime).
			WithField("runtime", runtime.GOOS).
			WithField("args", os.Args).
			Info("Run app")

		b.CrashHandler.AddRecoveryAction(func(interface{}) error {
			if c.Int("restart") > maxAllowedRestarts {
				logrus.
					WithField("restart", c.Int("restart")).
					Warn("Not restarting, already restarted too many times")

				return nil
			}

			return restartApp(c.String("launcher"), true)
		})

		if err := appMainLoop(b, c); err != nil {
			return err
		}

		if b.restart {
			return restartApp(c.String("launcher"), false)
		}

		if err := b.Versioner.RemoveOldVersions(); err != nil {
			return err
		}

		return nil
	}
}
