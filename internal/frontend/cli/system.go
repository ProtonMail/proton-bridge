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
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/ProtonMail/proton-bridge/v3/internal/bridge"
	"github.com/ProtonMail/proton-bridge/v3/pkg/ports"
	"github.com/abiosoft/ishell"
)

func (f *frontendCLI) printLogDir(c *ishell.Context) {
	if path, err := f.bridge.GetLogsPath(); err != nil {
		f.Println("Failed to determine location of log files")
	} else {
		f.Println("Log files are stored in\n\n ", path)
	}
}

func (f *frontendCLI) printManual(c *ishell.Context) {
	f.Println("More instructions about the Bridge can be found at\n\n  https://protonmail.com/bridge")
}

func (f *frontendCLI) printCredits(c *ishell.Context) {
	for _, pkg := range strings.Split(bridge.Credits, ";") {
		f.Println(pkg)
	}
}

func (f *frontendCLI) changeIMAPSecurity(c *ishell.Context) {
	f.ShowPrompt(false)
	defer f.ShowPrompt(true)

	newSecurity := "SSL"
	if f.bridge.GetIMAPSSL() {
		newSecurity = "STARTTLS"
	}

	msg := fmt.Sprintf("Are you sure you want to change IMAP setting to %q", newSecurity)

	if f.yesNoQuestion(msg) {
		if err := f.bridge.SetIMAPSSL(!f.bridge.GetIMAPSSL()); err != nil {
			f.printAndLogError(err)
			return
		}
	}
}

func (f *frontendCLI) changeSMTPSecurity(c *ishell.Context) {
	f.ShowPrompt(false)
	defer f.ShowPrompt(true)

	newSecurity := "SSL"
	if f.bridge.GetSMTPSSL() {
		newSecurity = "STARTTLS"
	}

	msg := fmt.Sprintf("Are you sure you want to change SMTP setting to %q", newSecurity)

	if f.yesNoQuestion(msg) {
		if err := f.bridge.SetSMTPSSL(!f.bridge.GetSMTPSSL()); err != nil {
			f.printAndLogError(err)
			return
		}
	}
}

func (f *frontendCLI) changeIMAPPort(c *ishell.Context) {
	f.ShowPrompt(false)
	defer f.ShowPrompt(true)

	newIMAPPort := f.readStringInAttempts(fmt.Sprintf("Set IMAP port (current %v)", f.bridge.GetIMAPPort()), c.ReadLine, f.isPortFree)
	if newIMAPPort == "" {
		f.printAndLogError(errors.New("failed to get new port"))
		return
	}

	newIMAPPortInt, err := strconv.Atoi(newIMAPPort)
	if err != nil {
		f.printAndLogError(err)
		return
	}

	if err := f.bridge.SetIMAPPort(newIMAPPortInt); err != nil {
		f.printAndLogError(err)
		return
	}
}

func (f *frontendCLI) changeSMTPPort(c *ishell.Context) {
	f.ShowPrompt(false)
	defer f.ShowPrompt(true)

	newSMTPPort := f.readStringInAttempts(fmt.Sprintf("Set SMTP port (current %v)", f.bridge.GetSMTPPort()), c.ReadLine, f.isPortFree)
	if newSMTPPort == "" {
		f.printAndLogError(errors.New("failed to get new port"))
		return
	}

	newSMTPPortInt, err := strconv.Atoi(newSMTPPort)
	if err != nil {
		f.printAndLogError(err)
		return
	}

	if err := f.bridge.SetSMTPPort(newSMTPPortInt); err != nil {
		f.printAndLogError(err)
		return
	}
}

func (f *frontendCLI) allowProxy(c *ishell.Context) {
	if f.bridge.GetProxyAllowed() {
		f.Println("Bridge is already set to use alternative routing to connect to Proton if it is being blocked.")
		return
	}

	f.Println("Bridge is currently set to NOT use alternative routing to connect to Proton if it is being blocked.")

	if f.yesNoQuestion("Are you sure you want to allow bridge to do this") {
		if err := f.bridge.SetProxyAllowed(true); err != nil {
			f.printAndLogError(err)
			return
		}
	}
}

func (f *frontendCLI) disallowProxy(c *ishell.Context) {
	if !f.bridge.GetProxyAllowed() {
		f.Println("Bridge is already set to NOT use alternative routing to connect to Proton if it is being blocked.")
		return
	}

	f.Println("Bridge is currently set to use alternative routing to connect to Proton if it is being blocked.")

	if f.yesNoQuestion("Are you sure you want to stop bridge from doing this") {
		if err := f.bridge.SetProxyAllowed(false); err != nil {
			f.printAndLogError(err)
			return
		}
	}
}

func (f *frontendCLI) hideAllMail(c *ishell.Context) {
	if !f.bridge.GetShowAllMail() {
		f.Println("All Mail folder is not listed in your local client.")
		return
	}

	f.Println("All Mail folder is listed in your client right now.")

	if f.yesNoQuestion("Do you want to hide All Mail folder") {
		if err := f.bridge.SetShowAllMail(false); err != nil {
			f.printAndLogError(err)
			return
		}
	}
}

func (f *frontendCLI) showAllMail(c *ishell.Context) {
	if f.bridge.GetShowAllMail() {
		f.Println("All Mail folder is listed in your local client.")
		return
	}

	f.Println("All Mail folder is not listed in your client right now.")

	if f.yesNoQuestion("Do you want to show All Mail folder") {
		if err := f.bridge.SetShowAllMail(true); err != nil {
			f.printAndLogError(err)
			return
		}
	}
}

func (f *frontendCLI) setGluonLocation(c *ishell.Context) {
	if gluonDir := f.bridge.GetGluonDir(); gluonDir != "" {
		f.Println("The current message cache location is:", gluonDir)
	}

	if location := f.readStringInAttempts("Enter a new location for the message cache", c.ReadLine, f.isCacheLocationUsable); location != "" {
		if err := f.bridge.SetGluonDir(context.Background(), location); err != nil {
			f.printAndLogError(err)
			return
		}
	}
}

func (f *frontendCLI) exportTLSCerts(c *ishell.Context) {
	if location := f.readStringInAttempts("Enter a path to which to export the TLS certificate used for IMAP and SMTP", c.ReadLine, f.isCacheLocationUsable); location != "" {
		cert, key := f.bridge.GetBridgeTLSCert()

		if err := os.WriteFile(filepath.Join(location, "cert.pem"), cert, 0o600); err != nil {
			f.printAndLogError(err)
			return
		}

		if err := os.WriteFile(filepath.Join(location, "key.pem"), key, 0o600); err != nil {
			f.printAndLogError(err)
			return
		}

		f.Println("TLS certificate exported to", location)
	}
}

func (f *frontendCLI) isPortFree(port string) bool {
	port = strings.ReplaceAll(port, ":", "")
	if port == "" {
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

// NOTE(GODT-1158): Check free space in location.
func (f *frontendCLI) isCacheLocationUsable(location string) bool {
	stat, err := os.Stat(location)
	if err != nil {
		return false
	}

	return stat.IsDir()
}
