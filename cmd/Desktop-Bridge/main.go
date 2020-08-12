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
	"io/ioutil"
	"os"
	"runtime/pprof"

	"github.com/ProtonMail/proton-bridge/internal/api"
	"github.com/ProtonMail/proton-bridge/internal/bridge"
	"github.com/ProtonMail/proton-bridge/internal/cmd"
	"github.com/ProtonMail/proton-bridge/internal/cookies"
	"github.com/ProtonMail/proton-bridge/internal/events"
	"github.com/ProtonMail/proton-bridge/internal/frontend"
	"github.com/ProtonMail/proton-bridge/internal/imap"
	"github.com/ProtonMail/proton-bridge/internal/preferences"
	"github.com/ProtonMail/proton-bridge/internal/smtp"
	"github.com/ProtonMail/proton-bridge/internal/updates"
	"github.com/ProtonMail/proton-bridge/internal/users/credentials"
	"github.com/ProtonMail/proton-bridge/pkg/config"
	"github.com/ProtonMail/proton-bridge/pkg/constants"
	"github.com/ProtonMail/proton-bridge/pkg/listener"
	"github.com/ProtonMail/proton-bridge/pkg/pmapi"
	"github.com/allan-simon/go-singleinstance"
	"github.com/sirupsen/logrus"
	"github.com/urfave/cli"
)

const (
	// cacheVersion is used for cache files such as lock, events, preferences, user_info, db files.
	// Different number will drop old files and create new ones.
	cacheVersion = "c11"

	appName = "bridge"
)

var (
	log = logrus.WithField("pkg", "main") //nolint[gochecknoglobals]
)

func main() {
	cmd.Main(
		"ProtonMail Bridge",
		"ProtonMail IMAP and SMTP Bridge",
		[]cli.Flag{
			cli.BoolFlag{
				Name:  "no-window",
				Usage: "Don't show window after start"},
			cli.BoolFlag{
				Name:  "noninteractive",
				Usage: "Start Bridge entirely noninteractively"},
		},
		run,
	)
}

// run initializes and starts everything in a precise order.
//
// IMPORTANT: ***Read the comments before CHANGING the order ***
func run(context *cli.Context) (contextError error) { // nolint[funlen]
	// We need to have config instance to setup a logs, panic handler, etc ...
	cfg := config.New(appName, constants.Version, constants.Revision, cacheVersion)

	// We want to know about any problem. Our PanicHandler calls sentry which is
	// not dependent on anything else. If that fails, it tries to create crash
	// report which will not be possible if no folder can be created. That's the
	// only problem we will not be notified about in any way.
	panicHandler := &cmd.PanicHandler{
		AppName: "ProtonMail Bridge",
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
	debugClient, debugServer := config.SetupLog(cfg, logLevel)

	// Doesn't make sense to continue when Bridge was invoked with wrong arguments.
	// We should tell that to the user before we do anything else.
	if context.Args().First() != "" {
		_ = cli.ShowAppHelp(context)
		return cli.NewExitError("Unknown argument", 4)
	}

	// It's safe to get version JSON file even when other instance is running.
	// (thus we put it before check of presence of other Bridge instance).
	updates := updates.NewBridge(cfg.GetUpdateDir())

	if dir := context.GlobalString("version-json"); dir != "" {
		cmd.GenerateVersionFiles(updates, dir)
		return nil
	}

	// Should be called after logs are configured but before preferences are created.
	migratePreferencesFromC10(cfg)

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
			cmd.DisableRestart()
			log.Error("Second instance: ", err)
		}
		return cli.NewExitError("Bridge is already running.", 3)
	}
	defer lock.Close() //nolint[errcheck]

	// In case user wants to do CPU or memory profiles...
	if doCPUProfile := context.GlobalBool("cpu-prof"); doCPUProfile {
		cmd.StartCPUProfile()
		defer pprof.StopCPUProfile()
	}

	if doMemoryProfile := context.GlobalBool("mem-prof"); doMemoryProfile {
		defer cmd.MakeMemoryProfile()
	}

	// Now we initialize all Bridge parts.
	log.Debug("Initializing bridge...")
	eventListener := listener.New()
	events.SetupEvents(eventListener)

	credentialsStore, credentialsError := credentials.NewStore(appName)
	if credentialsError != nil {
		log.Error("Could not get credentials store: ", credentialsError)
	}

	cm := pmapi.NewClientManager(cfg.GetAPIConfig())

	// Different build types have different roundtrippers (e.g. we want to enable
	// TLS fingerprint checks in production builds). GetRoundTripper has a different
	// implementation depending on whether build flag pmapi_prod is used or not.
	cm.SetRoundTripper(cfg.GetRoundTripper(cm, eventListener))

	// Cookies must be persisted across restarts.
	jar, err := cookies.NewCookieJar(pref)
	if err != nil {
		logrus.WithError(err).Warn("Could not create cookie jar")
	} else {
		cm.SetCookieJar(jar)
	}

	bridgeInstance := bridge.New(cfg, pref, panicHandler, eventListener, cm, credentialsStore)
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
	frontend := frontend.New(constants.Version, constants.BuildVersion, frontendMode, showWindowOnStart, panicHandler, cfg, pref, eventListener, updates, bridgeInstance, smtpBackend)

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

// migratePreferencesFromC10 will copy preferences from c10 folder to c11.
// It will happen only when c10/prefs.json exists and c11/prefs.json not.
// No configuration changed between c10 and c11 versions.
func migratePreferencesFromC10(cfg *config.Config) {
	pref10Path := config.New(appName, constants.Version, constants.Revision, "c10").GetPreferencesPath()
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

	err = ioutil.WriteFile(pref11Path, data, 0600)
	if err != nil {
		log.WithError(err).Error("Problem to migrate preferences")
		return
	}

	log.Info("Preferences migrated")
}
