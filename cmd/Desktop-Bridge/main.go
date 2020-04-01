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

/*
                             ___....___
   ^^                __..-:'':__:..:__:'':-..__
                 _.-:__:.-:'':  :  :  :'':-.:__:-._
               .':.-:  :  :  :  :  :  :  :  :  :._:'.
            _ :.':  :  :  :  :  :  :  :  :  :  :  :'.: _
           [ ]:  :  :  :  :  :  :  :  :  :  :  :  :  :[ ]
           [ ]:  :  :  :  :  :  :  :  :  :  :  :  :  :[ ]
  :::::::::[ ]:__:__:__:__:__:__:__:__:__:__:__:__:__:[ ]:::::::::::
  !!!!!!!!![ ]!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!![ ]!!!!!!!!!!!
  ^^^^^^^^^[ ]^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^[ ]^^^^^^^^^^^
           [ ]                                        [ ]
           [ ]                                        [ ]
     jgs   [ ]                                        [ ]
   ~~^_~^~/   \~^-~^~ _~^-~_^~-^~_^~~-^~_~^~-~_~-^~_^/   \~^ ~~_ ^
*/

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"runtime/pprof"
	"strconv"
	"strings"

	"github.com/ProtonMail/proton-bridge/internal/api"
	"github.com/ProtonMail/proton-bridge/internal/bridge"
	"github.com/ProtonMail/proton-bridge/internal/bridge/credentials"
	"github.com/ProtonMail/proton-bridge/internal/events"
	"github.com/ProtonMail/proton-bridge/internal/frontend"
	"github.com/ProtonMail/proton-bridge/internal/imap"
	"github.com/ProtonMail/proton-bridge/internal/pmapifactory"
	"github.com/ProtonMail/proton-bridge/internal/preferences"
	"github.com/ProtonMail/proton-bridge/internal/smtp"
	"github.com/ProtonMail/proton-bridge/pkg/args"
	"github.com/ProtonMail/proton-bridge/pkg/config"
	"github.com/ProtonMail/proton-bridge/pkg/listener"
	"github.com/ProtonMail/proton-bridge/pkg/pmapi"
	"github.com/ProtonMail/proton-bridge/pkg/updates"
	"github.com/allan-simon/go-singleinstance"
	"github.com/getsentry/raven-go"
	"github.com/sirupsen/logrus"
	"github.com/urfave/cli"
)

// cacheVersion is used for cache files such as lock, events, preferences, user_info, db files.
// Different number will drop old files and create new ones.
const cacheVersion = "c11"

// Following variables are set via ldflags during build.
var (
	// Version of the build.
	Version = "" //nolint[gochecknoglobals]
	// Revision is current hash of the build.
	Revision = "" //nolint[gochecknoglobals]
	// BuildTime stamp of the build.
	BuildTime = "" //nolint[gochecknoglobals]
	// AppShortName to make setup
	AppShortName = "bridge" //nolint[gochecknoglobals]
	// DSNSentry client keys to be able to report crashes to Sentry
	DSNSentry = "" //nolint[gochecknoglobals]
)

var (
	longVersion  = Version + " (" + Revision + ")" //nolint[gochecknoglobals]
	buildVersion = longVersion + " " + BuildTime   //nolint[gochecknoglobals]

	log = config.GetLogEntry("main") //nolint[gochecknoglobals]

	// How many crashes in a row.
	numberOfCrashes = 0 //nolint[gochecknoglobals]
	// After how many crashes bridge gives up starting.
	maxAllowedCrashes = 10 //nolint[gochecknoglobals]
)

func main() {
	if err := raven.SetDSN(DSNSentry); err != nil {
		log.WithError(err).Errorln("Can not setup sentry DSN")
	}
	raven.SetRelease(Revision)
	bridge.UpdateCurrentUserAgent(Version, runtime.GOOS, "", "")

	args.FilterProcessSerialNumberFromArgs()
	filterRestartNumberFromArgs()

	app := cli.NewApp()
	app.Name = "Protonmail Bridge"
	app.Version = buildVersion
	app.Flags = []cli.Flag{
		cli.StringFlag{
			Name:  "log-level, l",
			Usage: "Set the log level (one of panic, fatal, error, warn, info, debug, debug-client, debug-server)"},
		cli.BoolFlag{
			Name:  "no-window",
			Usage: "Don't show window after start"},
		cli.BoolFlag{
			Name:  "cli, c",
			Usage: "Use command line interface"},
		cli.BoolFlag{
			Name:  "noninteractive",
			Usage: "Start Bridge entirely noninteractively"},
		cli.StringFlag{
			Name:  "version-json, g",
			Usage: "Generate json version file"},
		cli.BoolFlag{
			Name:  "mem-prof, m",
			Usage: "Generate memory profile"},
		cli.BoolFlag{
			Name:  "cpu-prof, p",
			Usage: "Generate CPU profile"},
	}
	app.Usage = "ProtonMail IMAP and SMTP Bridge"
	app.Action = run

	// Always log the basic info about current bridge.
	logrus.SetLevel(logrus.InfoLevel)
	log.WithField("version", Version).
		WithField("revision", Revision).
		WithField("runtime", runtime.GOOS).
		WithField("build", BuildTime).
		WithField("args", os.Args).
		WithField("appLong", app.Name).
		WithField("appShort", AppShortName).
		Info("Run app")
	if err := app.Run(os.Args); err != nil {
		log.Error("Program exited with error: ", err)
	}
}

type panicHandler struct {
	cfg *config.Config
	err *error // Pointer to error of cli action.
}

func (ph *panicHandler) HandlePanic() {
	r := recover()
	if r == nil {
		return
	}

	config.HandlePanic(ph.cfg, fmt.Sprintf("Recover: %v", r))
	frontend.HandlePanic()

	*ph.err = cli.NewExitError("Panic and restart", 666)
	numberOfCrashes++
	log.Error("Restarting after panic")
	restartApp()
	os.Exit(666)
}

// run initializes and starts everything in a precise order.
//
// IMPORTANT: ***Read the comments before CHANGING the order ***
func run(context *cli.Context) (contextError error) { // nolint[funlen]
	// We need to have config instance to setup a logs, panic handler, etc ...
	cfg := config.New(AppShortName, Version, Revision, cacheVersion)

	// We want to know about any problem. Our PanicHandler calls sentry which is
	// not dependent on anything else. If that fails, it tries to create crash
	// report which will not be possible if no folder can be created. That's the
	// only problem we will not be notified about in any way.
	panicHandler := &panicHandler{cfg, &contextError}
	defer panicHandler.HandlePanic()

	// First we need config and create necessary folder; it's dependency for everything.
	if err := cfg.CreateDirs(); err != nil {
		log.Fatal("Cannot create necessary folders: ", err)
	}

	// Setup of logs should be as soon as possible to ensure we record every wanted report in the log.
	logLevel := context.GlobalString("log-level")
	debugClient, debugServer := config.SetupLog(cfg, logLevel)

	// Should be called after logs are configured but before preferences are created.
	migratePreferencesFromC10(cfg)

	if err := cfg.ClearOldData(); err != nil {
		log.Error("Cannot clear old data: ", err)
	}

	// Doesn't make sense to continue when Bridge was invoked with wrong arguments.
	// We should tell that to the user before we do anything else.
	if context.Args().First() != "" {
		_ = cli.ShowAppHelp(context)
		return cli.NewExitError("Unknown argument", 4)
	}

	// It's safe to get version JSON file even when other instance is running.
	// (thus we put it before check of presence of other Bridge instance).
	updates := updates.New(AppShortName, Version, Revision, BuildTime, bridge.ReleaseNotes, bridge.ReleaseFixedBugs, cfg.GetUpdateDir())
	if dir := context.GlobalString("version-json"); dir != "" {
		generateVersionFiles(updates, dir)
		return nil
	}

	// ClearOldData before starting new bridge to do a proper setup.
	//
	// IMPORTANT: If you the change position of this you will need to wait
	// until force-update to be applied on all currently used bridge
	// versions
	if err := cfg.ClearOldData(); err != nil {
		log.Error("Cannot clear old data: ", err)
	}

	// GetTLSConfig is needed for IMAP, SMTL and local bridge API (to check second instance).
	//
	// This should be called after ClearOldData, in order to re-create the
	// certificates if clean data will remove them (accidentally or on purpose).
	tls, err := config.GetTLSConfig(cfg)
	if err != nil {
		log.WithError(err).Fatal("Cannot get TLS certificate")
	}

	pref := preferences.New(cfg)

	// Now we can try to proceed with starting the bridge. First we need to ensure
	// this is the only instance. If not, we will end and focus the existing one.
	lock, err := singleinstance.CreateLockFile(cfg.GetLockPath())
	if err != nil {
		log.Warn("Bridge is already running")
		if err := api.CheckOtherInstanceAndFocus(pref.GetInt(preferences.APIPortKey), tls); err != nil {
			numberOfCrashes = maxAllowedCrashes
			log.Error("Second instance: ", err)
		}
		return cli.NewExitError("Bridge is already running.", 3)
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
		defer makeMemoryProfile()
	}

	// Now we initialize all Bridge parts.
	log.Debug("Initializing bridge...")
	eventListener := listener.New()
	events.SetupEvents(eventListener)

	credentialsStore, credentialsError := credentials.NewStore()
	if credentialsError != nil {
		log.Error("Could not get credentials store: ", credentialsError)
	}

	clientman := pmapi.NewClientManager(pmapifactory.GetClientConfig(cfg, eventListener))
	bridgeInstance := bridge.New(cfg, pref, panicHandler, eventListener, Version, clientman, credentialsStore)
	imapBackend := imap.NewIMAPBackend(panicHandler, eventListener, cfg, bridgeInstance)
	smtpBackend := smtp.NewSMTPBackend(panicHandler, eventListener, pref, bridgeInstance)

	go func() {
		defer panicHandler.HandlePanic()
		apiServer := api.NewAPIServer(pref, tls, cfg.GetTLSCertPath(), cfg.GetTLSKeyPath(), eventListener)
		apiServer.ListenAndServe()
	}()

	go func() {
		defer panicHandler.HandlePanic()
		imapPort := pref.GetInt(preferences.IMAPPortKey)
		imapServer := imap.NewIMAPServer(debugClient, debugServer, imapPort, tls, imapBackend, eventListener)
		imapServer.ListenAndServe()
	}()

	go func() {
		defer panicHandler.HandlePanic()
		smtpPort := pref.GetInt(preferences.SMTPPortKey)
		useSSL := pref.GetBool(preferences.SMTPSSLKey)
		smtpServer := smtp.NewSMTPServer(debugClient || debugServer, smtpPort, useSSL, tls, smtpBackend, eventListener)
		smtpServer.ListenAndServe()
	}()

	// Decide about frontend mode before initializing rest of bridge.
	var frontendMode string

	switch {
	case context.GlobalBool("cli"):
		frontendMode = "cli"
	case context.GlobalBool("noninteractive"):
		frontendMode = "noninteractive"
	default:
		frontendMode = "qt"
	}

	log.WithField("mode", frontendMode).Debug("Determined frontend mode to use")

	// If we are starting bridge in noninteractive mode, simply block instead of starting a frontend.
	if frontendMode == "noninteractive" {
		<-(make(chan struct{}))
		return nil
	}

	showWindowOnStart := !context.GlobalBool("no-window")
	frontend := frontend.New(Version, buildVersion, frontendMode, showWindowOnStart, panicHandler, cfg, pref, eventListener, updates, bridgeInstance, smtpBackend)

	// Last part is to start everything.
	log.Debug("Starting frontend...")
	if err := frontend.Loop(credentialsError); err != nil {
		log.Error("Frontend failed with error: ", err)
		return cli.NewExitError("Frontend error", 2)
	}

	if frontend.IsAppRestarting() {
		restartApp()
	}

	return nil
}

// migratePreferencesFromC10 will copy preferences from c10 folder to c11.
// It will happen only when c10/prefs.json exists and c11/prefs.json not.
// No configuration changed between c10 and c11 versions.
func migratePreferencesFromC10(cfg *config.Config) {
	pref10Path := config.New(AppShortName, Version, Revision, "c10").GetPreferencesPath()
	if _, err := os.Stat(pref10Path); os.IsNotExist(err) {
		log.WithField("path", pref10Path).Trace("Old preferences does not exist, migration skipped")
		return
	}

	pref11Path := cfg.GetPreferencesPath()
	if _, err := os.Stat(pref11Path); err == nil {
		log.WithField("path", pref11Path).Trace("New preferences already exists, migration skipped")
		return
	}

	data, err := ioutil.ReadFile(pref10Path) //nolint[gosec]
	if err != nil {
		log.WithError(err).Error("Problem to load old preferences")
		return
	}

	err = ioutil.WriteFile(pref11Path, data, 0644)
	if err != nil {
		log.WithError(err).Error("Problem to migrate preferences")
		return
	}

	log.Info("Preferences migrated")
}

// generateVersionFiles writes a JSON file with details about current build.
// Those files are used for upgrading the app.
func generateVersionFiles(updates *updates.Updates, dir string) {
	log.Info("Generating version files")
	for _, goos := range []string{"windows", "darwin", "linux"} {
		log.Debug("Generating JSON for ", goos)
		if err := updates.CreateJSONAndSign(dir, goos); err != nil {
			log.Error(err)
		}
	}
}

func makeMemoryProfile() {
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

// restartApp starts a new instance in background.
func restartApp() {
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
