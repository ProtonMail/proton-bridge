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

package cli

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/ProtonMail/proton-bridge/internal/config/settings"
	"github.com/ProtonMail/proton-bridge/pkg/ports"
	"github.com/abiosoft/ishell"
)

var (
	currentPort = "" //nolint[gochecknoglobals]
)

func (f *frontendCLI) restart(c *ishell.Context) {
	if f.yesNoQuestion("Are you sure you want to restart the Bridge") {
		f.Println("Restarting Bridge...")
		f.restarter.SetToRestart()
		f.Stop()
	}
}

func (f *frontendCLI) checkInternetConnection(c *ishell.Context) {
	if f.bridge.CheckConnection() == nil {
		f.Println("Internet connection is available.")
	} else {
		f.Println("Can not contact the server, please check your internet connection.")
	}
}

func (f *frontendCLI) printLogDir(c *ishell.Context) {
	if path, err := f.locations.ProvideLogsPath(); err != nil {
		f.Println("Failed to determine location of log files")
	} else {
		f.Println("Log files are stored in\n\n ", path)
	}
}

func (f *frontendCLI) printManual(c *ishell.Context) {
	f.Println("More instructions about the Bridge can be found at\n\n  https://protonmail.com/bridge")
}

func (f *frontendCLI) deleteCache(c *ishell.Context) {
	f.ShowPrompt(false)
	defer f.ShowPrompt(true)

	if !f.yesNoQuestion("Do you really want to remove all stored preferences") {
		return
	}
	if err := f.bridge.ClearData(); err != nil {
		f.printAndLogError("Cache clear failed: ", err.Error())
		return
	}
	f.Println("Cached cleared, restarting bridge")
	// Clearing data removes everything (db, preferences, ...)
	// so everything has to be stopped and started again.
	f.restarter.SetToRestart()
	f.Stop()
}

func (f *frontendCLI) changeSMTPSecurity(c *ishell.Context) {
	f.ShowPrompt(false)
	defer f.ShowPrompt(true)

	isSSL := f.settings.GetBool(settings.SMTPSSLKey)
	newSecurity := "SSL"
	if isSSL {
		newSecurity = "STARTTLS"
	}

	msg := fmt.Sprintf("Are you sure you want to change SMTP setting to %q and restart the Bridge", newSecurity)

	if f.yesNoQuestion(msg) {
		f.settings.SetBool(settings.SMTPSSLKey, !isSSL)
		f.Println("Restarting Bridge...")
		f.restarter.SetToRestart()
		f.Stop()
	}
}

func (f *frontendCLI) changePort(c *ishell.Context) {
	f.ShowPrompt(false)
	defer f.ShowPrompt(true)

	currentPort = f.settings.Get(settings.IMAPPortKey)
	newIMAPPort := f.readStringInAttempts("Set IMAP port (current "+currentPort+")", c.ReadLine, f.isPortFree)
	if newIMAPPort == "" {
		newIMAPPort = currentPort
	}
	imapPortChanged := newIMAPPort != currentPort

	currentPort = f.settings.Get(settings.SMTPPortKey)
	newSMTPPort := f.readStringInAttempts("Set SMTP port (current "+currentPort+")", c.ReadLine, f.isPortFree)
	if newSMTPPort == "" {
		newSMTPPort = currentPort
	}
	smtpPortChanged := newSMTPPort != currentPort

	if newIMAPPort == newSMTPPort {
		f.Println("SMTP and IMAP ports must be different!")
		return
	}

	if imapPortChanged || smtpPortChanged {
		f.Println("Saving values IMAP:", newIMAPPort, "SMTP:", newSMTPPort)
		f.settings.Set(settings.IMAPPortKey, newIMAPPort)
		f.settings.Set(settings.SMTPPortKey, newSMTPPort)
		f.Println("Restarting Bridge...")
		f.restarter.SetToRestart()
		f.Stop()
	} else {
		f.Println("Nothing changed")
	}
}

func (f *frontendCLI) allowProxy(c *ishell.Context) {
	if f.settings.GetBool(settings.AllowProxyKey) {
		f.Println("Bridge is already set to use alternative routing to connect to Proton if it is being blocked.")
		return
	}

	f.Println("Bridge is currently set to NOT use alternative routing to connect to Proton if it is being blocked.")

	if f.yesNoQuestion("Are you sure you want to allow bridge to do this") {
		f.settings.SetBool(settings.AllowProxyKey, true)
		f.bridge.AllowProxy()
	}
}

func (f *frontendCLI) disallowProxy(c *ishell.Context) {
	if !f.settings.GetBool(settings.AllowProxyKey) {
		f.Println("Bridge is already set to NOT use alternative routing to connect to Proton if it is being blocked.")
		return
	}

	f.Println("Bridge is currently set to use alternative routing to connect to Proton if it is being blocked.")

	if f.yesNoQuestion("Are you sure you want to stop bridge from doing this") {
		f.settings.SetBool(settings.AllowProxyKey, false)
		f.bridge.DisallowProxy()
	}
}

func (f *frontendCLI) isPortFree(port string) bool {
	port = strings.ReplaceAll(port, ":", "")
	if port == "" || port == currentPort {
		return true
	}
	number, err := strconv.Atoi(port)
	if err != nil || number < 0 || number > 65535 {
		f.Println("Input", port, "is not a valid port number.")
		return false
	}
	if !ports.IsPortFree(number) {
		f.Println("Port", number, "is occupied by another process.")
		return false
	}
	return true
}
