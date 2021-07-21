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

package cliie

import (
	"context"
	"strings"

	"github.com/abiosoft/ishell"
)

func (f *frontendCLI) listAccounts(c *ishell.Context) {
	spacing := "%-2d: %-20s (%-15s, %-15s)\n"
	f.Printf(bold(strings.ReplaceAll(spacing, "d", "s")), "#", "account", "status", "address mode")
	for idx, user := range f.ie.GetUsers() {
		connected := "disconnected"
		if user.IsConnected() {
			connected = "connected"
		}
		mode := "split"
		if user.IsCombinedAddressMode() {
			mode = "combined"
		}
		f.Printf(spacing, idx, user.Username(), connected, mode)
	}
	f.Println()
}

func (f *frontendCLI) loginAccount(c *ishell.Context) { // nolint[funlen]
	f.ShowPrompt(false)
	defer f.ShowPrompt(true)

	loginName := ""
	if len(c.Args) > 0 {
		user := f.getUserByIndexOrName(c.Args[0])
		if user != nil {
			loginName = user.GetPrimaryAddress()
		}
	}

	if loginName == "" {
		loginName = f.readStringInAttempts("Username", c.ReadLine, isNotEmpty)
		if loginName == "" {
			return
		}
	} else {
		f.Println("Username:", loginName)
	}

	password := f.readStringInAttempts("Password", c.ReadPassword, isNotEmpty)
	if password == "" {
		return
	}

	f.Println("Authenticating ... ")
	client, auth, err := f.ie.Login(loginName, password)
	if err != nil {
		f.processAPIError(err)
		return
	}

	if auth.HasTwoFactor() {
		twoFactor := f.readStringInAttempts("Two factor code", c.ReadLine, isNotEmpty)
		if twoFactor == "" {
			return
		}

		err = client.Auth2FA(context.Background(), twoFactor)
		if err != nil {
			f.processAPIError(err)
			return
		}
	}

	mailboxPassword := password
	if auth.HasMailboxPassword() {
		mailboxPassword = f.readStringInAttempts("Mailbox password", c.ReadPassword, isNotEmpty)
	}
	if mailboxPassword == "" {
		return
	}

	f.Println("Adding account ...")
	user, err := f.ie.FinishLogin(client, auth, mailboxPassword)
	if err != nil {
		log.WithField("username", loginName).WithError(err).Error("Login was unsuccessful")
		f.Println("Adding account was unsuccessful:", err)
		return
	}

	f.Printf("Account %s was added successfully.\n", bold(user.Username()))
}

func (f *frontendCLI) logoutAccount(c *ishell.Context) {
	f.ShowPrompt(false)
	defer f.ShowPrompt(true)

	user := f.askUserByIndexOrName(c)
	if user == nil {
		return
	}
	if f.yesNoQuestion("Are you sure you want to logout account " + bold(user.Username())) {
		if err := user.Logout(); err != nil {
			f.printAndLogError("Logging out failed: ", err)
		}
	}
}

func (f *frontendCLI) deleteAccount(c *ishell.Context) {
	f.ShowPrompt(false)
	defer f.ShowPrompt(true)

	user := f.askUserByIndexOrName(c)
	if user == nil {
		return
	}
	if f.yesNoQuestion("Are you sure you want to " + bold("remove account "+user.Username())) {
		clearCache := f.yesNoQuestion("Do you want to remove cache for this account")
		if err := f.ie.DeleteUser(user.ID(), clearCache); err != nil {
			f.printAndLogError("Cannot delete account: ", err)
			return
		}
	}
}

func (f *frontendCLI) deleteAccounts(c *ishell.Context) {
	f.ShowPrompt(false)
	defer f.ShowPrompt(true)

	if !f.yesNoQuestion("Do you really want remove all accounts") {
		return
	}
	for _, user := range f.ie.GetUsers() {
		if err := f.ie.DeleteUser(user.ID(), false); err != nil {
			f.printAndLogError("Cannot delete account ", user.Username(), ": ", err)
		}
	}
	c.Println("Keychain cleared")
}
