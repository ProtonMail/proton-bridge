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

// Package cliie provides CLI interface of the Import-Export app.
package cliie

import (
	"github.com/ProtonMail/proton-bridge/internal/events"
	"github.com/ProtonMail/proton-bridge/internal/frontend/types"
	"github.com/ProtonMail/proton-bridge/internal/locations"
	"github.com/ProtonMail/proton-bridge/internal/updater"
	"github.com/ProtonMail/proton-bridge/pkg/listener"

	"github.com/abiosoft/ishell"
	"github.com/sirupsen/logrus"
)

var (
	log = logrus.WithField("pkg", "frontend/cli-ie") //nolint[gochecknoglobals]
)

type frontendCLI struct {
	*ishell.Shell

	locations     *locations.Locations
	eventListener listener.Listener
	updater       types.Updater
	ie            types.ImportExporter

	restarter types.Restarter
}

// New returns a new CLI frontend configured with the given options.
func New( //nolint[funlen]
	panicHandler types.PanicHandler,

	locations *locations.Locations,
	eventListener listener.Listener,
	updater types.Updater,
	ie types.ImportExporter,
	restarter types.Restarter,
) *frontendCLI { //nolint[golint]
	fe := &frontendCLI{
		Shell: ishell.New(),

		locations:     locations,
		eventListener: eventListener,
		updater:       updater,
		ie:            ie,

		restarter: restarter,
	}

	// Clear commands.
	clearCmd := &ishell.Cmd{Name: "clear",
		Help:    "remove stored accounts and preferences. (alias: cl)",
		Aliases: []string{"cl"},
	}
	clearCmd.AddCmd(&ishell.Cmd{Name: "accounts",
		Help:    "remove all accounts from keychain. (aliases: a, k, keychain)",
		Aliases: []string{"a", "k", "keychain"},
		Func:    fe.deleteAccounts,
	})
	fe.AddCmd(clearCmd)

	// Check commands.
	checkCmd := &ishell.Cmd{Name: "check", Help: "check internet connection or new version."}
	checkCmd.AddCmd(&ishell.Cmd{Name: "updates",
		Help:    "check for Import-Export updates. (aliases: u, v, version)",
		Aliases: []string{"u", "version", "v"},
		Func:    fe.checkUpdates,
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

	// Import-Export commands.
	importCmd := &ishell.Cmd{Name: "import",
		Help:    "import messages. (alias: imp)",
		Aliases: []string{"imp"},
	}
	importCmd.AddCmd(&ishell.Cmd{Name: "local",
		Help:    "import local messages. (aliases: loc)",
		Func:    fe.noAccountWrapper(fe.importLocalMessages),
		Aliases: []string{"loc"},
	})
	importCmd.AddCmd(&ishell.Cmd{Name: "remote",
		Help:    "import remote messages. (aliases: rem)",
		Func:    fe.noAccountWrapper(fe.importRemoteMessages),
		Aliases: []string{"rem"},
	})
	fe.AddCmd(importCmd)

	exportCmd := &ishell.Cmd{Name: "export",
		Help:    "export messages. (alias: exp)",
		Aliases: []string{"exp"},
	}
	exportCmd.AddCmd(&ishell.Cmd{Name: "eml",
		Help: "export messages to eml files.",
		Func: fe.noAccountWrapper(fe.exportMessagesToEML),
	})
	exportCmd.AddCmd(&ishell.Cmd{Name: "mbox",
		Help: "export messages to mbox files.",
		Func: fe.noAccountWrapper(fe.exportMessagesToMBOX),
	})
	fe.AddCmd(exportCmd)

	// System commands.
	fe.AddCmd(&ishell.Cmd{Name: "restart",
		Help: "restart the Import-Export app.",
		Func: fe.restart,
	})

	go func() {
		defer panicHandler.HandlePanic()
		fe.watchEvents()
	}()
	return fe
}

func (f *frontendCLI) watchEvents() {
	errorCh := f.eventListener.ProvideChannel(events.ErrorEvent)
	credentialsErrorCh := f.eventListener.ProvideChannel(events.CredentialsErrorEvent)
	internetOffCh := f.eventListener.ProvideChannel(events.InternetOffEvent)
	internetOnCh := f.eventListener.ProvideChannel(events.InternetOnEvent)
	addressChangedLogoutCh := f.eventListener.ProvideChannel(events.AddressChangedLogoutEvent)
	logoutCh := f.eventListener.ProvideChannel(events.LogoutEvent)
	certIssue := f.eventListener.ProvideChannel(events.TLSCertIssue)
	for {
		select {
		case errorDetails := <-errorCh:
			f.Println("Import-Export failed:", errorDetails)
		case <-credentialsErrorCh:
			f.notifyCredentialsError()
		case <-internetOffCh:
			f.notifyInternetOff()
		case <-internetOnCh:
			f.notifyInternetOn()
		case address := <-addressChangedLogoutCh:
			f.notifyLogout(address)
		case userID := <-logoutCh:
			user, err := f.ie.GetUser(userID)
			if err != nil {
				return
			}
			f.notifyLogout(user.Username())
		case <-certIssue:
			f.notifyCertIssue()
		}
	}
}

// Loop starts the frontend loop with an interactive shell.
func (f *frontendCLI) Loop() error {
	f.Print(`
Welcome to ProtonMail Import-Export app interactive shell

WARNING: The CLI is an experimental feature and does not yet cover all functionality.
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
