// Copyright (c) 2023 Proton AG
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

package cli

import (
	"context"
	"strings"

	"github.com/ProtonMail/go-proton-api"
	"github.com/ProtonMail/proton-bridge/v3/internal/bridge"
	"github.com/ProtonMail/proton-bridge/v3/internal/constants"
	"github.com/ProtonMail/proton-bridge/v3/internal/vault"
	"github.com/abiosoft/ishell"
)

func (f *frontendCLI) listAccounts(_ *ishell.Context) {
	spacing := "%-2d: %-20s (%-15s, %-15s)\n"
	f.Printf(bold(strings.ReplaceAll(spacing, "d", "s")), "#", "account", "status", "address mode")
	for idx, userID := range f.bridge.GetUserIDs() {
		user, err := f.bridge.GetUserInfo(userID)
		if err != nil {
			panic(err)
		}

		var state string
		switch user.State {
		case bridge.SignedOut:
			state = "signed out"
		case bridge.Locked:
			state = "locked"
		case bridge.Connected:
			state = "connected"
		default:
			panic("Unknown user state")
		}

		f.Printf(spacing, idx, user.Username, state, user.AddressMode)
	}
	f.Println()
}

func (f *frontendCLI) showAccountInfo(c *ishell.Context) {
	user := f.askUserByIndexOrName(c)
	if user.UserID == "" {
		return
	}

	switch user.State {
	case bridge.SignedOut:
		f.Printf("Please login to %s to get email client configuration.\n", bold(user.Username))
		return
	case bridge.Locked:
		f.Printf("User %s is currently locked. Please wait and try again.\n", bold(user.Username))
		return
	case bridge.Connected:
	default:
	}

	if user.AddressMode == vault.CombinedMode {
		f.showAccountAddressInfo(user, user.Addresses[0])
	} else {
		for _, address := range user.Addresses {
			f.showAccountAddressInfo(user, address)
		}
	}
}

func (f *frontendCLI) showAccountAddressInfo(user bridge.UserInfo, address string) {
	const (
		StartTLS = "STARTTLS"
		SSL      = "SSL"
	)

	imapSecurity := StartTLS
	if f.bridge.GetIMAPSSL() {
		imapSecurity = SSL
	}

	smtpSecurity := StartTLS
	if f.bridge.GetSMTPSSL() {
		smtpSecurity = SSL
	}

	f.Println(bold("Configuration for " + address))
	f.Printf("IMAP Settings\nAddress:   %s\nIMAP port: %d\nUsername:  %s\nPassword:  %s\nSecurity:  %s\n",
		constants.Host,
		f.bridge.GetIMAPPort(),
		address,
		user.BridgePass,
		imapSecurity,
	)
	f.Println("")
	f.Printf("SMTP Settings\nAddress:   %s\nSMTP port: %d\nUsername:  %s\nPassword:  %s\nSecurity:  %s\n",
		constants.Host,
		f.bridge.GetSMTPPort(),
		address,
		user.BridgePass,
		smtpSecurity,
	)
	f.Println("")
}

func (f *frontendCLI) loginAccount(c *ishell.Context) {
	f.ShowPrompt(false)
	defer f.ShowPrompt(true)

	loginName := ""
	if len(c.Args) > 0 {
		user := f.getUserByIndexOrName(c.Args[0])
		if user.UserID != "" {
			loginName = user.Addresses[0]
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

	client, auth, err := f.bridge.LoginAuth(context.Background(), loginName, []byte(password))
	if err != nil {
		f.printAndLogError("Cannot login: ", err)
		return
	}

	if auth.TwoFA.Enabled&proton.HasTOTP != 0 {
		code := f.readStringInAttempts("Two factor code", c.ReadLine, isNotEmpty)
		if code == "" {
			f.printAndLogError("Cannot login: need two factor code")
			return
		}

		if err := client.Auth2FA(context.Background(), proton.Auth2FAReq{TwoFactorCode: code}); err != nil {
			f.printAndLogError("Cannot login: ", err)
			return
		}
	}

	var keyPass []byte

	if auth.PasswordMode == proton.TwoPasswordMode {
		keyPass = []byte(f.readStringInAttempts("Mailbox password", c.ReadPassword, isNotEmpty))
		if len(keyPass) == 0 {
			f.printAndLogError("Cannot login: need mailbox password")
			return
		}
	} else {
		keyPass = []byte(password)
	}

	userID, err := f.bridge.LoginUser(context.Background(), client, auth, keyPass)
	if err != nil {
		f.processAPIError(err)
		return
	}

	user, err := f.bridge.GetUserInfo(userID)
	if err != nil {
		panic(err)
	}

	f.Printf("Account %s was added successfully.\n", bold(user.Username))
}

func (f *frontendCLI) logoutAccount(c *ishell.Context) {
	f.ShowPrompt(false)
	defer f.ShowPrompt(true)

	user := f.askUserByIndexOrName(c)
	if user.UserID == "" {
		return
	}

	if f.yesNoQuestion("Are you sure you want to logout account " + bold(user.Username)) {
		if err := f.bridge.LogoutUser(context.Background(), user.UserID); err != nil {
			f.printAndLogError("Logging out failed: ", err)
		}
	}
}

func (f *frontendCLI) deleteAccount(c *ishell.Context) {
	f.ShowPrompt(false)
	defer f.ShowPrompt(true)

	user := f.askUserByIndexOrName(c)
	if user.UserID == "" {
		return
	}

	if f.yesNoQuestion("Are you sure you want to " + bold("remove account "+user.Username)) {
		if err := f.bridge.DeleteUser(context.Background(), user.UserID); err != nil {
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

	for _, userID := range f.bridge.GetUserIDs() {
		user, err := f.bridge.GetUserInfo(userID)
		if err != nil {
			f.printAndLogError("Cannot get user info: ", err)
			return
		}

		if err := f.bridge.DeleteUser(context.Background(), user.UserID); err != nil {
			f.printAndLogError("Cannot delete account ", user.Username, ": ", err)
			return
		}
	}

	c.Println("Keychain cleared")
}

func (f *frontendCLI) deleteEverything(c *ishell.Context) {
	f.ShowPrompt(false)
	defer f.ShowPrompt(true)

	if !f.yesNoQuestion("Do you really want remove everything") {
		return
	}

	f.bridge.FactoryReset(context.Background())

	c.Println("Everything cleared")

	f.restarter.Set(true, false)

	f.Stop()
}

func (f *frontendCLI) changeMode(c *ishell.Context) {
	user := f.askUserByIndexOrName(c)
	if user.UserID == "" {
		return
	}

	var targetMode vault.AddressMode

	if user.AddressMode == vault.CombinedMode {
		targetMode = vault.SplitMode
	} else {
		targetMode = vault.CombinedMode
	}

	if !f.yesNoQuestion("Are you sure you want to change the mode for account " + bold(user.Username) + " to " + bold(targetMode.String())) {
		return
	}

	if err := f.bridge.SetAddressMode(context.Background(), user.UserID, targetMode); err != nil {
		f.printAndLogError("Cannot switch address mode:", err)
	}

	f.Printf("Address mode for account %s changed to %s\n", user.Username, targetMode)
}

func (f *frontendCLI) configureAppleMail(c *ishell.Context) {
	user := f.askUserByIndexOrName(c)
	if user.UserID == "" {
		return
	}

	if !f.yesNoQuestion("Are you sure you want to configure Apple Mail for " + bold(user.Username) + " with address " + bold(user.Addresses[0])) {
		return
	}

	if err := f.bridge.ConfigureAppleMail(context.Background(), user.UserID, user.Addresses[0]); err != nil {
		f.printAndLogError(err)
		return
	}

	f.Printf("Apple Mail configured for %v with address %v\n", user.Username, user.Addresses[0])
}

func (f *frontendCLI) badEventSynchronize(_ *ishell.Context) {
	f.badEventFeedback(true)
}

func (f *frontendCLI) badEventLogout(_ *ishell.Context) {
	f.badEventFeedback(false)
}

func (f *frontendCLI) badEventFeedback(doResync bool) {
	if f.badUserID == "" {
		f.Printf("Error: There was no unresolved bad event.")
		return
	}

	action := "synchronize"
	if !doResync {
		action = "logout"
	}

	if err := f.bridge.SendBadEventUserFeedback(context.Background(), f.badUserID, doResync); err != nil {
		f.Printf("Error: failed to send %s feedback: %w", action, err)
		return
	}

	f.badUserID = ""
}
