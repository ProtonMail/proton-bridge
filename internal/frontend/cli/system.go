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
	"github.com/ProtonMail/proton-bridge/v3/internal/certs"
	"github.com/ProtonMail/proton-bridge/v3/pkg/ports"
	"github.com/abiosoft/ishell"
)

func (f *frontendCLI) printLogDir(_ *ishell.Context) {
	if path, err := f.bridge.GetLogsPath(); err != nil {
		f.Println("Failed to determine location of log files")
	} else {
		f.Println("Log files are stored in\n\n ", path)
	}
}

func (f *frontendCLI) printManual(_ *ishell.Context) {
	f.Println("More instructions about the Bridge can be found at\n\n  https://proton.me/mail/bridge")
}

func (f *frontendCLI) printCredits(_ *ishell.Context) {
	for _, pkg := range strings.Split(bridge.Credits, ";") {
		f.Println(pkg)
	}
}

func (f *frontendCLI) changeIMAPSecurity(_ *ishell.Context) {
	f.ShowPrompt(false)
	defer f.ShowPrompt(true)

	newSecurity := "SSL"
	if f.bridge.GetIMAPSSL() {
		newSecurity = "STARTTLS"
	}

	msg := fmt.Sprintf("Are you sure you want to change IMAP setting to %q", newSecurity)

	if f.yesNoQuestion(msg) {
		if err := f.bridge.SetIMAPSSL(context.Background(), !f.bridge.GetIMAPSSL()); err != nil {
			f.printAndLogError(err)
			return
		}
	}
}

func (f *frontendCLI) changeSMTPSecurity(_ *ishell.Context) {
	f.ShowPrompt(false)
	defer f.ShowPrompt(true)

	newSecurity := "SSL"
	if f.bridge.GetSMTPSSL() {
		newSecurity = "STARTTLS"
	}

	msg := fmt.Sprintf("Are you sure you want to change SMTP setting to %q", newSecurity)

	if f.yesNoQuestion(msg) {
		if err := f.bridge.SetSMTPSSL(context.Background(), !f.bridge.GetSMTPSSL()); err != nil {
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

	if err := f.bridge.SetIMAPPort(context.Background(), newIMAPPortInt); err != nil {
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

	if err := f.bridge.SetSMTPPort(context.Background(), newSMTPPortInt); err != nil {
		f.printAndLogError(err)
		return
	}
}

func (f *frontendCLI) allowProxy(_ *ishell.Context) {
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

func (f *frontendCLI) disallowProxy(_ *ishell.Context) {
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

func (f *frontendCLI) hideAllMail(_ *ishell.Context) {
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

func (f *frontendCLI) showAllMail(_ *ishell.Context) {
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

func (f *frontendCLI) enableTelemetry(_ *ishell.Context) {
	if !f.bridge.GetTelemetryDisabled() {
		f.Println("Usage diagnostics collection is enabled.")
		return
	}

	f.Println("Usage diagnostics collection is disabled right now.")

	if f.yesNoQuestion("Do you want to enable usage diagnostics collection") {
		if err := f.bridge.SetTelemetryDisabled(false); err != nil {
			f.printAndLogError(err)
			return
		}
	}
}

func (f *frontendCLI) disableTelemetry(_ *ishell.Context) {
	if f.bridge.GetTelemetryDisabled() {
		f.Println("Usage diagnostics collection is disabled.")
		return
	}

	f.Println("Usage diagnostics collection is enabled right now.")

	if f.yesNoQuestion("Do you want to disable usage diagnostics collection") {
		if err := f.bridge.SetTelemetryDisabled(true); err != nil {
			f.printAndLogError(err)
			return
		}
	}
}

func (f *frontendCLI) setGluonLocation(c *ishell.Context) {
	if gluonDir := f.bridge.GetGluonCacheDir(); gluonDir != "" {
		f.Println("The current message cache location is:", gluonDir)
	}

	if location := f.readStringInAttempts("Enter a new location for the message cache", c.ReadLine, f.isCacheLocationUsable); location != "" {
		if err := f.bridge.SetGluonDir(context.Background(), location); err != nil {
			f.printAndLogError(err)
			return
		}
	}
}

func (f *frontendCLI) tlsCertStatus(_ *ishell.Context) {
	cert, _ := f.bridge.GetBridgeTLSCert()
	installer := certs.NewInstaller()
	if installer.IsCertInstalled(cert) {
		f.Println("The Bridge TLS certificate is already installed in the OS keychain.")
	} else {
		f.Println("The Bridge TLS certificate is not installed in the OS keychain.")
	}
}

func (f *frontendCLI) installTLSCert(_ *ishell.Context) {
	cert, _ := f.bridge.GetBridgeTLSCert()
	installer := certs.NewInstaller()
	if installer.IsCertInstalled(cert) {
		f.printAndLogError(errors.New("the Bridge TLS certificate is already installed in the OS keychain"))
		return
	}

	f.Println("Please provide your credentials in the system popup dialog in order to continue.")
	if err := installer.InstallCert(cert); err != nil {
		f.printAndLogError(err)
		return
	}

	f.Println("The Bridge TLS certificate was successfully installed in the OS keychain.")
}

func (f *frontendCLI) uninstallTLSCert(_ *ishell.Context) {
	cert, _ := f.bridge.GetBridgeTLSCert()
	installer := certs.NewInstaller()
	if !installer.IsCertInstalled(cert) {
		f.printAndLogError(errors.New("the Bridge TLS certificate is not installed in the OS keychain"))
		return
	}

	f.Println("Please provide your credentials in the system popup dialog in order to continue.")
	if err := installer.UninstallCert(cert); err != nil {
		f.printAndLogError(err)
		return
	}

	f.Println("The Bridge TLS certificate was successfully uninstalled from the OS keychain.")
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

func (f *frontendCLI) importTLSCerts(c *ishell.Context) {
	certPath := f.readStringInAttempts("Enter the path to the cert.pem file", c.ReadLine, f.isFile)
	if certPath == "" {
		f.printAndLogError(errors.New("failed to get cert path"))
		return
	}

	keyPath := f.readStringInAttempts("Enter the path to the key.pem file", c.ReadLine, f.isFile)
	if keyPath == "" {
		f.printAndLogError(errors.New("failed to get key path"))
		return
	}

	if err := f.bridge.SetBridgeTLSCertPath(certPath, keyPath); err != nil {
		f.printAndLogError(err)
		return
	}

	f.Println("TLS certificate imported. Restart Bridge to use it.")
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

func (f *frontendCLI) isFile(location string) bool {
	stat, err := os.Stat(location)
	if err != nil {
		return false
	}

	return !stat.IsDir()
}

func (f *frontendCLI) repair(_ *ishell.Context) {
	if f.bridge.HasAPIConnection() {
		if f.yesNoQuestion("Are you sure you want to initialize a repair, this may take a while") {
			f.bridge.Repair()
		}
	} else {
		f.Println("Bridge cannot connect to the Proton servers. A connection is required to utilize this feature.")
	}
}
