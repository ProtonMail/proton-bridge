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

// Package cli provides CLI interface of the Bridge.
package cli

import (
	"github.com/ProtonMail/proton-bridge/internal/config/settings"
	"github.com/ProtonMail/proton-bridge/internal/events"
	"github.com/ProtonMail/proton-bridge/internal/frontend/types"
	"github.com/ProtonMail/proton-bridge/internal/locations"
	"github.com/ProtonMail/proton-bridge/internal/updater"
	"github.com/ProtonMail/proton-bridge/pkg/listener"

	"github.com/abiosoft/ishell"
	"github.com/sirupsen/logrus"
)

var (
	log = logrus.WithField("pkg", "frontend/cli") //nolint[gochecknoglobals]
)

type frontendCLI struct {
	*ishell.Shell

	locations     *locations.Locations
	settings      *settings.Settings
	eventListener listener.Listener
	updater       types.Updater
	bridge        types.Bridger

	restarter types.Restarter
}

// New returns a new CLI frontend configured with the given options.
func New( //nolint[funlen]
	panicHandler types.PanicHandler,

	locations *locations.Locations,
	settings *settings.Settings,
	eventListener listener.Listener,
	updater types.Updater,
	bridge types.Bridger,
	restarter types.Restarter,
) *frontendCLI { //nolint[golint]
	fe := &frontendCLI{
		Shell: ishell.New(),

		locations:     locations,
		settings:      settings,
		eventListener: eventListener,
		updater:       updater,
		bridge:        bridge,

		restarter: restarter,
	}

	// Clear commands.
	clearCmd := &ishell.Cmd{Name: "clear",
		Help:    "remove stored accounts and preferences. (alias: cl)",
		Aliases: []string{"cl"},
	}
	clearCmd.AddCmd(&ishell.Cmd{Name: "cache",
		Help:    "remove stored preferences for accounts (aliases: c, prefs, preferences)",
		Aliases: []string{"c", "prefs", "preferences"},
		Func:    fe.deleteCache,
	})
	clearCmd.AddCmd(&ishell.Cmd{Name: "accounts",
		Help:    "remove all accounts from keychain. (aliases: a, k, keychain)",
		Aliases: []string{"a", "k", "keychain"},
		Func:    fe.deleteAccounts,
	})
	fe.AddCmd(clearCmd)

	// Change commands.
	changeCmd := &ishell.Cmd{Name: "change",
		Help:    "change server or account settings (aliases: ch, switch)",
		Aliases: []string{"ch", "switch"},
	}
	changeCmd.AddCmd(&ishell.Cmd{Name: "mode",
		Help:      "switch between combined addresses and split addresses mode for account. Use index or account name as parameter. (alias: m)",
		Aliases:   []string{"m"},
		Func:      fe.changeMode,
		Completer: fe.completeUsernames,
	})
	changeCmd.AddCmd(&ishell.Cmd{Name: "port",
		Help:    "change port numbers of IMAP and SMTP servers. (alias: p)",
		Aliases: []string{"p"},
		Func:    fe.changePort,
	})
	changeCmd.AddCmd(&ishell.Cmd{Name: "proxy",
		Help: "allow or disallow bridge to securely connect to proton via a third party when it is being blocked",
		Func: fe.toggleAllowProxy,
	})
	changeCmd.AddCmd(&ishell.Cmd{Name: "smtp-security",
		Help:    "change port numbers of IMAP and SMTP servers.(alias: ssl, starttls)",
		Aliases: []string{"ssl", "starttls"},
		Func:    fe.changeSMTPSecurity,
	})
	fe.AddCmd(changeCmd)

	// Check commands.
	checkCmd := &ishell.Cmd{Name: "check", Help: "check internet connection or new version."}
	checkCmd.AddCmd(&ishell.Cmd{Name: "updates",
		Help:    "check for Bridge updates. (aliases: u, v, version)",
		Aliases: []string{"u", "version", "v"},
		Func:    fe.checkUpdates,
	})
	checkCmd.AddCmd(&ishell.Cmd{Name: "internet",
		Help:    "check internet connection. (aliases: i, conn, connection)",
		Aliases: []string{"i", "con", "connection"},
		Func:    fe.checkInternetConnection,
	})
	fe.AddCmd(checkCmd)

	// Print info commands.
	fe.AddCmd(&ishell.Cmd{Name: "log-dir",
		Help:    "print path to directory with logs. (aliases: log, logs)",
		Aliases: []string{"log", "logs"},
		Func:    fe.printLogDir,
	})
	fe.AddCmd(&ishell.Cmd{Name: "manual",
		Help:    "print URL with instructions. (alias: man)",
		Aliases: []string{"man"},
		Func:    fe.printManual,
	})

	fe.AddCmd(&ishell.Cmd{Name: "credits",
		Help: "print used resources.",
		Func: fe.printCredits,
	})

	// Account commands.
	fe.AddCmd(&ishell.Cmd{Name: "list",
		Help:    "print the list of accounts. (aliases: l, ls)",
		Func:    fe.noAccountWrapper(fe.listAccounts),
		Aliases: []string{"l", "ls"},
	})
	fe.AddCmd(&ishell.Cmd{Name: "info",
		Help:      "print the configuration for account. Use index or account name as parameter. (alias: i)",
		Func:      fe.noAccountWrapper(fe.showAccountInfo),
		Completer: fe.completeUsernames,
		Aliases:   []string{"i"},
	})
	fe.AddCmd(&ishell.Cmd{Name: "login",
		Help:      "login procedure to add or connect account. Optionally use index or account as parameter. (aliases: a, add, con, connect)",
		Func:      fe.loginAccount,
		Aliases:   []string{"add", "a", "con", "connect"},
		Completer: fe.completeUsernames,
	})
	fe.AddCmd(&ishell.Cmd{Name: "logout",
		Help:      "disconnect the account. Use index or account name as parameter. (aliases: d, disconnect)",
		Func:      fe.noAccountWrapper(fe.logoutAccount),
		Aliases:   []string{"d", "disconnect"},
		Completer: fe.completeUsernames,
	})
	fe.AddCmd(&ishell.Cmd{Name: "delete",
		Help:      "remove the account from keychain. Use index or account name as parameter. (aliases: del, rm, remove)",
		Func:      fe.noAccountWrapper(fe.deleteAccount),
		Aliases:   []string{"del", "rm", "remove"},
		Completer: fe.completeUsernames,
	})

	// System commands.
	fe.AddCmd(&ishell.Cmd{Name: "restart",
		Help: "restart the bridge.",
		Func: fe.restart,
	})

	go func() {
		defer panicHandler.HandlePanic()
		fe.watchEvents()
	}()
	return fe
}

func (f *frontendCLI) watchEvents() {
	errorCh := f.getEventChannel(events.ErrorEvent)
	credentialsErrorCh := f.getEventChannel(events.CredentialsErrorEvent)
	internetOffCh := f.getEventChannel(events.InternetOffEvent)
	internetOnCh := f.getEventChannel(events.InternetOnEvent)
	addressChangedCh := f.getEventChannel(events.AddressChangedEvent)
	addressChangedLogoutCh := f.getEventChannel(events.AddressChangedLogoutEvent)
	logoutCh := f.getEventChannel(events.LogoutEvent)
	certIssue := f.getEventChannel(events.TLSCertIssue)
	for {
		select {
		case errorDetails := <-errorCh:
			f.Println("Bridge failed:", errorDetails)
		case <-credentialsErrorCh:
			f.notifyCredentialsError()
		case <-internetOffCh:
			f.notifyInternetOff()
		case <-internetOnCh:
			f.notifyInternetOn()
		case address := <-addressChangedCh:
			f.Printf("Address changed for %s. You may need to reconfigure your email client.", address)
		case address := <-addressChangedLogoutCh:
			f.notifyLogout(address)
		case userID := <-logoutCh:
			user, err := f.bridge.GetUser(userID)
			if err != nil {
				return
			}
			f.notifyLogout(user.Username())
		case <-certIssue:
			f.notifyCertIssue()
		}
	}
}

func (f *frontendCLI) getEventChannel(event string) <-chan string {
	ch := make(chan string)
	f.eventListener.Add(event, ch)
	f.eventListener.RetryEmit(event)
	return ch
}

// Loop starts the frontend loop with an interactive shell.
func (f *frontendCLI) Loop() error {
	f.Print(`
            Welcome to ProtonMail Bridge interactive shell
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
`)
	f.Run()
	return nil
}

func (f *frontendCLI) NotifyManualUpdate(update updater.VersionInfo, canInstall bool) {
	// NOTE: Save the update somewhere so that it can be installed when user chooses "install now".
}

func (f *frontendCLI) WaitUntilFrontendIsReady()              {}
func (f *frontendCLI) SetVersion(version updater.VersionInfo) {}
func (f *frontendCLI) NotifySilentUpdateInstalled()           {}
func (f *frontendCLI) NotifySilentUpdateError(err error)      {}
