// Copyright (c) 2024 Proton AG
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

// Package cli provides CLI interface of the Bridge.
package cli

import (
	"errors"
	"fmt"
	"os"
	"runtime"

	"github.com/ProtonMail/gluon/async"
	"github.com/ProtonMail/proton-bridge/v3/internal/bridge"
	"github.com/ProtonMail/proton-bridge/v3/internal/constants"
	"github.com/ProtonMail/proton-bridge/v3/internal/events"
	"github.com/ProtonMail/proton-bridge/v3/pkg/restarter"

	"github.com/abiosoft/ishell"
	"github.com/sirupsen/logrus"
)

var log = logrus.WithField("pkg", "frontend/cli") //nolint:gochecknoglobals

type frontendCLI struct {
	*ishell.Shell

	bridge    *bridge.Bridge
	restarter *restarter.Restarter

	badUserID string

	panicHandler async.PanicHandler
}

// New returns a new CLI frontend configured with the given options.
func New(
	bridge *bridge.Bridge,
	restarter *restarter.Restarter,
	eventCh <-chan events.Event,
	panicHandler async.PanicHandler,
	quitCh <-chan struct{},
) *frontendCLI { //nolint:revive
	fe := &frontendCLI{
		Shell:        ishell.New(),
		bridge:       bridge,
		restarter:    restarter,
		badUserID:    "",
		panicHandler: panicHandler,
	}

	// We want to exit at the first Ctrl+C. By default, ishell requires two.
	fe.Interrupt(func(_ *ishell.Context, _ int, _ string) {
		os.Exit(1)
	})

	// Clear commands.
	clearCmd := &ishell.Cmd{
		Name:    "clear",
		Help:    "remove stored accounts and preferences. (alias: cl)",
		Aliases: []string{"cl"},
	}
	clearCmd.AddCmd(&ishell.Cmd{
		Name:    "accounts",
		Help:    "remove all accounts from keychain. (aliases: a, k, keychain)",
		Aliases: []string{"a", "k", "keychain"},
		Func:    fe.deleteAccounts,
	})
	clearCmd.AddCmd(&ishell.Cmd{
		Name:    "everything",
		Help:    "remove everything",
		Aliases: []string{"a", "k", "keychain"},
		Func:    fe.deleteEverything,
	})
	fe.AddCmd(clearCmd)

	// Change commands.
	changeCmd := &ishell.Cmd{
		Name:    "change",
		Help:    "change server or account settings (aliases: ch, switch)",
		Aliases: []string{"ch", "switch"},
	}
	changeCmd.AddCmd(&ishell.Cmd{
		Name:      "mode",
		Help:      "switch between combined addresses and split addresses mode for account. Use index or account name as parameter. (alias: m)",
		Aliases:   []string{"m"},
		Func:      fe.changeMode,
		Completer: fe.completeUsernames,
	})
	changeCmd.AddCmd(&ishell.Cmd{
		Name: "change-location",
		Help: "change the location of the encrypted message cache",
		Func: fe.setGluonLocation,
	})
	changeCmd.AddCmd(&ishell.Cmd{
		Name: "imap-port",
		Help: "change port number of IMAP server.",
		Func: fe.changeIMAPPort,
	})
	changeCmd.AddCmd(&ishell.Cmd{
		Name: "smtp-port",
		Help: "change port number of SMTP server.",
		Func: fe.changeSMTPPort,
	})
	changeCmd.AddCmd(&ishell.Cmd{
		Name:    "imap-security",
		Help:    "change IMAP SSL settings servers.(alias: ssl-imap, starttls-imap)",
		Aliases: []string{"ssl-imap", "starttls-imap"},
		Func:    fe.changeIMAPSecurity,
	})
	changeCmd.AddCmd(&ishell.Cmd{
		Name:    "smtp-security",
		Help:    "change SMTP SSL settings servers.(alias: ssl-smtp, starttls-smtp)",
		Aliases: []string{"ssl-smtp", "starttls-smtp"},
		Func:    fe.changeSMTPSecurity,
	})
	fe.AddCmd(changeCmd)

	// DoH commands.
	dohCmd := &ishell.Cmd{
		Name: "proxy",
		Help: "allow or disallow bridge to securely connect to proton via a third party when it is being blocked",
	}
	dohCmd.AddCmd(&ishell.Cmd{
		Name: "allow",
		Help: "allow bridge to securely connect to proton via a third party when it is being blocked",
		Func: fe.allowProxy,
	})
	dohCmd.AddCmd(&ishell.Cmd{
		Name: "disallow",
		Help: "disallow bridge to securely connect to proton via a third party when it is being blocked",
		Func: fe.disallowProxy,
	})
	fe.AddCmd(dohCmd)

	//goland:noinspection GoBoolExpressions
	if runtime.GOOS == "darwin" {
		// Apple Mail commands.
		configureCmd := &ishell.Cmd{
			Name: "configure-apple-mail",
			Help: "Configures Apple Mail to use ProtonMail Bridge",
			Func: fe.configureAppleMail,
		}
		fe.AddCmd(configureCmd)
	}

	// TLS commands.
	certCmd := &ishell.Cmd{
		Name: "cert",
		Help: "Manage the TLS certificate used by Bridge",
	}

	//goland:noinspection GoBoolExpressions
	if runtime.GOOS == "darwin" {
		certCmd.AddCmd(&ishell.Cmd{
			Name: "status",
			Help: "Check if the TLS certificate used by Bridge is installed in the OS keychain",
			Func: fe.tlsCertStatus,
		})
		certCmd.AddCmd(&ishell.Cmd{
			Name: "install",
			Help: "Install TLS certificate used by Bridge in the OS keychain",
			Func: fe.installTLSCert,
		})
		certCmd.AddCmd(&ishell.Cmd{
			Name: "uninstall",
			Help: "Uninstall the TLS certificate used by Bridge from the OS keychain",
			Func: fe.uninstallTLSCert,
		})
	}
	certCmd.AddCmd(&ishell.Cmd{
		Name: "export",
		Help: "Export the TLS certificate used by Bridge",
		Func: fe.exportTLSCerts,
	})
	certCmd.AddCmd(&ishell.Cmd{
		Name: "import",
		Help: "Import a TLS certificate to be used by Bridge",
		Func: fe.importTLSCerts,
	})
	fe.AddCmd(certCmd)

	// All mail visibility commands.
	allMailCmd := &ishell.Cmd{
		Name: "all-mail-visibility",
		Help: "choose not to list the All Mail folder in your local client",
	}
	allMailCmd.AddCmd(&ishell.Cmd{
		Name: "hide",
		Help: "All Mail folder will not be listed in your local client",
		Func: fe.hideAllMail,
	})
	allMailCmd.AddCmd(&ishell.Cmd{
		Name: "show",
		Help: "All Mail folder will be listed in your local client",
		Func: fe.showAllMail,
	})
	fe.AddCmd(allMailCmd)

	// Updates commands.
	updatesCmd := &ishell.Cmd{
		Name: "updates",
		Help: "manage bridge updates",
	}
	updatesCmd.AddCmd(&ishell.Cmd{
		Name: "check",
		Help: "check for Bridge updates",
		Func: fe.checkUpdates,
	})
	autoUpdatesCmd := &ishell.Cmd{
		Name: "autoupdates",
		Help: "manage bridge updates",
	}
	updatesCmd.AddCmd(autoUpdatesCmd)
	autoUpdatesCmd.AddCmd(&ishell.Cmd{
		Name: "enable",
		Help: "automatically keep bridge up to date",
		Func: fe.enableAutoUpdates,
	})
	autoUpdatesCmd.AddCmd(&ishell.Cmd{
		Name: "disable",
		Help: "require bridge to be manually updated",
		Func: fe.disableAutoUpdates,
	})
	updatesChannelCmd := &ishell.Cmd{
		Name: "channel",
		Help: "switch updates channel",
	}
	updatesCmd.AddCmd(updatesChannelCmd)
	updatesChannelCmd.AddCmd(&ishell.Cmd{
		Name: "early",
		Help: "switch to the early-access updates channel",
		Func: fe.selectEarlyChannel,
	})
	updatesChannelCmd.AddCmd(&ishell.Cmd{
		Name: "stable",
		Help: "switch to the stable updates channel",
		Func: fe.selectStableChannel,
	})
	fe.AddCmd(updatesCmd)

	// Print info commands.
	fe.AddCmd(&ishell.Cmd{
		Name:    "log-dir",
		Help:    "print path to directory with logs. (aliases: log, logs)",
		Aliases: []string{"log", "logs"},
		Func:    fe.printLogDir,
	})
	fe.AddCmd(&ishell.Cmd{
		Name:    "manual",
		Help:    "print URL with instructions. (alias: man)",
		Aliases: []string{"man"},
		Func:    fe.printManual,
	})
	fe.AddCmd(&ishell.Cmd{
		Name: "credits",
		Help: "print used resources.",
		Func: fe.printCredits,
	})

	// Account commands.
	fe.AddCmd(&ishell.Cmd{
		Name:    "list",
		Help:    "print the list of accounts. (aliases: l, ls)",
		Func:    fe.noAccountWrapper(fe.listAccounts),
		Aliases: []string{"l", "ls"},
	})
	fe.AddCmd(&ishell.Cmd{
		Name:      "info",
		Help:      "print the configuration for account. Use index or account name as parameter. (alias: i)",
		Func:      fe.noAccountWrapper(fe.showAccountInfo),
		Completer: fe.completeUsernames,
		Aliases:   []string{"i"},
	})
	fe.AddCmd(&ishell.Cmd{
		Name:      "login",
		Help:      "login procedure to add or connect account. Optionally use index or account as parameter. (aliases: a, add, con, connect)",
		Func:      fe.loginAccount,
		Aliases:   []string{"add", "a", "con", "connect"},
		Completer: fe.completeUsernames,
	})
	fe.AddCmd(&ishell.Cmd{
		Name:      "logout",
		Help:      "disconnect the account. Use index or account name as parameter. (aliases: d, disconnect)",
		Func:      fe.noAccountWrapper(fe.logoutAccount),
		Aliases:   []string{"d", "disconnect"},
		Completer: fe.completeUsernames,
	})
	fe.AddCmd(&ishell.Cmd{
		Name:      "delete",
		Help:      "remove the account from keychain. Use index or account name as parameter. (aliases: del, rm, remove)",
		Func:      fe.noAccountWrapper(fe.deleteAccount),
		Aliases:   []string{"del", "rm", "remove"},
		Completer: fe.completeUsernames,
	})
	fe.AddCmd(&ishell.Cmd{
		Name:    "repair",
		Help:    "reload all accounts and cached data, re-download emails. Email clients remain connected. Logged out users will be repaired on next login. (aliases: rep)",
		Func:    fe.repair,
		Aliases: []string{"rep"},
	})

	badEventCmd := &ishell.Cmd{
		Name: "bad-event",
		Help: "manage actions when bad event error occurs",
	}
	badEventCmd.AddCmd(&ishell.Cmd{
		Name: "synchronize",
		Help: "synchronize your local database to resolve the bad event error",
		Func: fe.badEventSynchronize,
	})
	badEventCmd.AddCmd(&ishell.Cmd{
		Name: "logout",
		Help: "logout to deal with bad event error later",
		Func: fe.badEventLogout,
	})
	fe.AddCmd(badEventCmd)

	// Telemetry commands
	telemetryCmd := &ishell.Cmd{
		Name: "telemetry",
		Help: "choose whether usage diagnostics are collected or not",
	}
	telemetryCmd.AddCmd(&ishell.Cmd{
		Name: "enable",
		Help: "Usage diagnostics collection will be enabled",
		Func: fe.enableTelemetry,
	})
	telemetryCmd.AddCmd(&ishell.Cmd{
		Name: "disable",
		Help: "Usage diagnostics collection will be disabled",
		Func: fe.disableTelemetry,
	})
	fe.AddCmd(telemetryCmd)

	dbgCmd := &ishell.Cmd{
		Name: "debug",
		Help: "Debug diagnostics ",
	}

	dbgCmd.AddCmd(&ishell.Cmd{
		Name: "mailbox-state",
		Help: "Verify local mailbox state against proton server state",
		Func: fe.debugMailboxState,
	})

	fe.AddCmd(dbgCmd)

	go fe.watchEvents(eventCh)

	go func() {
		<-quitCh
		fe.Close()
	}()

	return fe
}

func (f *frontendCLI) watchEvents(eventCh <-chan events.Event) { // nolint:gocyclo
	defer async.HandlePanic(f.panicHandler)

	// GODT-1949: Better error events.
	for _, err := range f.bridge.GetErrors() {
		switch {
		case errors.Is(err, bridge.ErrVaultCorrupt):
			f.notifyCredentialsError()

		case errors.Is(err, bridge.ErrVaultInsecure):
			f.notifyCredentialsError()
		}
	}

	for event := range eventCh {
		switch event := event.(type) {
		case events.ConnStatusUp:
			f.notifyInternetOn()

		case events.ConnStatusDown:
			f.notifyInternetOff()

		case events.IMAPServerError:
			f.Println("IMAP server error:", event.Error)

		case events.SMTPServerError:
			f.Println("SMTP server error:", event.Error)

		case events.UserDeauth:
			user, err := f.bridge.GetUserInfo(event.UserID)
			if err != nil {
				return
			}

			f.notifyLogout(user.Username)

		case events.UserBadEvent:
			user, err := f.bridge.GetUserInfo(event.UserID)
			if err != nil {
				return
			}

			f.badUserID = event.UserID

			f.Printf("\nInternal Error\n\n")
			f.Printf("Bridge ran into an internal error and it is not able proceed with %s.\n", user.Username)
			f.Printf("Synchronize your local database now or logout to do it later.\n")
			f.Printf("Synchronization time depends on the size of your mailbox.\n")
			f.Printf("\n\n")
			f.Printf("The allowed actions are:\n")
			f.Printf("* bad-event synchronize\n")
			f.Printf("* bad-event logout\n\n")

		case events.IMAPLoginFailed:
			f.Printf("An IMAP login attempt failed for user %v\n", event.Username)

		case events.UserAddressEnabled:
			user, err := f.bridge.GetUserInfo(event.UserID)
			if err != nil {
				return
			}

			f.Printf("An address for %s was enabled. You may need to reconfigure your email client.\n", user.Username)

		case events.UserAddressDisabled:
			user, err := f.bridge.GetUserInfo(event.UserID)
			if err != nil {
				return
			}

			f.Printf("An address for %s was disabled. You may need to reconfigure your email client.\n", user.Username)

		case events.UserAddressDeleted:
			user, err := f.bridge.GetUserInfo(event.UserID)
			if err != nil {
				return
			}

			f.Printf("An address for %s was disabled. You may need to reconfigure your email client.\n", user.Username)

		case events.SyncStarted:
			user, err := f.bridge.GetUserInfo(event.UserID)
			if err != nil {
				return
			}

			f.Printf("A sync has begun for %s.\n", user.Username)

		case events.SyncFinished:
			user, err := f.bridge.GetUserInfo(event.UserID)
			if err != nil {
				return
			}

			f.Printf("A sync has finished for %s.\n", user.Username)

		case events.SyncProgress:
			user, err := f.bridge.GetUserInfo(event.UserID)
			if err != nil {
				return
			}

			f.Printf(
				"Sync (%v): %.1f%% (Elapsed: %0.1fs, ETA: %0.1fs)\n",
				user.Username,
				100*event.Progress,
				event.Elapsed.Seconds(),
				event.Remaining.Seconds(),
			)

		case events.UpdateAvailable:
			if !event.Compatible {
				f.Printf("A new version (%v) is available but it cannot be installed automatically.\n", event.Version.Version)
			} else if !event.Silent {
				f.Printf("A new version (%v) is available.\n", event.Version.Version)
			}

		case events.UpdateInstalled:
			f.Printf("A new version (%v) was installed.\n", event.Version.Version)

		case events.UpdateFailed:
			f.Printf("A new version (%v) failed to be installed (%v).\n", event.Version.Version, event.Error)

		case events.UpdateForced:
			f.notifyNeedUpgrade()

		case events.TLSIssue:
			f.notifyCertIssue()

		case events.Raise:
			f.Printf("Hello!")

		case events.UserNotification:
			user, err := f.bridge.GetUserInfo(event.UserID)
			if err != nil {
				return
			}

			fmt.Printf("\n--- NOTIFICATION ---\n\n")
			fmt.Printf("Sent to: %s\n", user.Username)
			fmt.Printf("Title: %s\n", event.Title)
			fmt.Printf("Subtitle: %s\n", event.Subtitle)
			fmt.Printf("Message: %s\n\n", event.Body)
		}
	}

	/*
		errorCh := f.eventListener.ProvideChannel(events.ErrorEvent)
		credentialsErrorCh := f.eventListener.ProvideChannel(events.CredentialsErrorEvent)
		for {
			select {
			case errorDetails := <-errorCh:
				f.Println("Bridge failed:", errorDetails)
			case <-credentialsErrorCh:
				f.notifyCredentialsError()
			case stat := <-internetConnChangedCh:
				if stat == events.InternetOff {
					f.notifyInternetOff()
				}
				if stat == events.InternetOn {
					f.notifyInternetOn()
				}
			}
		}
	*/
}

// Loop starts the frontend loop with an interactive shell.
func (f *frontendCLI) Loop() error {
	f.Printf(`
            Welcome to %s interactive shell
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
`, constants.FullAppName)
	f.Run()
	return nil
}
