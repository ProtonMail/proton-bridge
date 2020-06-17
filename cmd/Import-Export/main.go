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
	"runtime/pprof"

	"github.com/ProtonMail/proton-bridge/internal/cmd"
	"github.com/ProtonMail/proton-bridge/internal/events"
	"github.com/ProtonMail/proton-bridge/internal/frontend"
	"github.com/ProtonMail/proton-bridge/internal/importexport"
	"github.com/ProtonMail/proton-bridge/internal/users/credentials"
	"github.com/ProtonMail/proton-bridge/pkg/config"
	"github.com/ProtonMail/proton-bridge/pkg/constants"
	"github.com/ProtonMail/proton-bridge/pkg/listener"
	"github.com/ProtonMail/proton-bridge/pkg/pmapi"
	"github.com/ProtonMail/proton-bridge/pkg/updates"
	"github.com/allan-simon/go-singleinstance"
	"github.com/sirupsen/logrus"
	"github.com/urfave/cli"
)

const (
	appName     = "importExport"
	appNameDash = "import-export"
)

var (
	log = logrus.WithField("pkg", "main") //nolint[gochecknoglobals]
)

func main() {
	cmd.Main(
		"ProtonMail Import/Export",
		"ProtonMail Import/Export tool",
		nil,
		run,
	)
}

// run initializes and starts everything in a precise order.
//
// IMPORTANT: ***Read the comments before CHANGING the order ***
func run(context *cli.Context) (contextError error) { // nolint[funlen]
	// We need to have config instance to setup a logs, panic handler, etc ...
	cfg := config.New(appName, constants.Version, constants.Revision, "")

	// We want to know about any problem. Our PanicHandler calls sentry which is
	// not dependent on anything else. If that fails, it tries to create crash
	// report which will not be possible if no folder can be created. That's the
	// only problem we will not be notified about in any way.
	panicHandler := &cmd.PanicHandler{
		AppName: "ProtonMail Import/Export",
		Config:  cfg,
		Err:     &contextError,
	}
	defer panicHandler.HandlePanic()

	// First we need config and create necessary folder; it's dependency for everything.
	if err := cfg.CreateDirs(); err != nil {
		log.Fatal("Cannot create necessary folders: ", err)
	}

	// Setup of logs should be as soon as possible to ensure we record every wanted report in the log.
	logLevel := context.GlobalString("log-level")
	_, _ = config.SetupLog(cfg, logLevel)

	// Doesn't make sense to continue when Import/Export was invoked with wrong arguments.
	// We should tell that to the user before we do anything else.
	if context.Args().First() != "" {
		_ = cli.ShowAppHelp(context)
		return cli.NewExitError("Unknown argument", 4)
	}

	// It's safe to get version JSON file even when other instance is running.
	// (thus we put it before check of presence of other Import/Export instance).
	updates := updates.NewImportExport(cfg.GetUpdateDir())

	if dir := context.GlobalString("version-json"); dir != "" {
		cmd.GenerateVersionFiles(updates, dir)
		return nil
	}

	// Now we can try to proceed with starting the import/export. First we need to ensure
	// this is the only instance. If not, we will end and focus the existing one.
	lock, err := singleinstance.CreateLockFile(cfg.GetLockPath())
	if err != nil {
		log.Warn("Import/Export is already running")
		return cli.NewExitError("Import/Export is already running.", 3)
	}
	defer lock.Close() //nolint[errcheck]

	// In case user wants to do CPU or memory profiles...
	if doCPUProfile := context.GlobalBool("cpu-prof"); doCPUProfile {
		f, err := os.Create("cpu.pprof")
		if err != nil {
			log.Fatal("Could not create CPU profile: ", err)
		}
		if err := pprof.StartCPUProfile(f); err != nil {
			log.Fatal("Could not start CPU profile: ", err)
		}
		defer pprof.StopCPUProfile()
	}

	if doMemoryProfile := context.GlobalBool("mem-prof"); doMemoryProfile {
		defer cmd.MakeMemoryProfile()
	}

	// Now we initialize all Import/Export parts.
	log.Debug("Initializing import/export...")
	eventListener := listener.New()
	events.SetupEvents(eventListener)

	credentialsStore, credentialsError := credentials.NewStore(appNameDash)
	if credentialsError != nil {
		log.Error("Could not get credentials store: ", credentialsError)
	}

	cm := pmapi.NewClientManager(cfg.GetAPIConfig())

	// Different build types have different roundtrippers (e.g. we want to enable
	// TLS fingerprint checks in production builds). GetRoundTripper has a different
	// implementation depending on whether build flag pmapi_prod is used or not.
	cm.SetRoundTripper(cfg.GetRoundTripper(cm, eventListener))

	importexportInstance := importexport.New(cfg, panicHandler, eventListener, cm, credentialsStore)

	// Decide about frontend mode before initializing rest of import/export.
	var frontendMode string
	switch {
	case context.GlobalBool("cli"):
		frontendMode = "cli"
	default:
		frontendMode = "qt"
	}
	log.WithField("mode", frontendMode).Debug("Determined frontend mode to use")

	frontend := frontend.NewImportExport(constants.Version, constants.BuildVersion, frontendMode, panicHandler, cfg, eventListener, updates, importexportInstance)

	// Last part is to start everything.
	log.Debug("Starting frontend...")
	if err := frontend.Loop(credentialsError); err != nil {
		log.Error("Frontend failed with error: ", err)
		return cli.NewExitError("Frontend error", 2)
	}

	if frontend.IsAppRestarting() {
		cmd.RestartApp()
	}

	return nil
}
