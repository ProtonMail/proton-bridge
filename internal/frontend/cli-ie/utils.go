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

package cli

import (
	"strings"

	pmapi "github.com/ProtonMail/proton-bridge/pkg/pmapi"
	"github.com/fatih/color"
)

const (
	maxInputRepeat = 2
)

var (
	bold = color.New(color.Bold).SprintFunc() //nolint[gochecknoglobals]
)

func isNotEmpty(val string) bool {
	return val != ""
}

func (f *frontendCLI) yesNoQuestion(question string) bool {
	f.Print(question, "? yes/"+bold("no")+": ")
	yes := "yes"
	answer := strings.ToLower(f.ReadLine())
	for i := 0; i < len(answer); i++ {
		if i >= len(yes) || answer[i] != yes[i] {
			return false // Everything else is false.
		}
	}
	return len(answer) > 0 // Empty is false.
}

func (f *frontendCLI) readStringInAttempts(title string, readFunc func() string, isOK func(string) bool) (value string) {
	f.Printf("%s: ", title)
	value = readFunc()
	title = strings.ToLower(string(title[0])) + title[1:]
	for i := 0; !isOK(value); i++ {
		if i >= maxInputRepeat {
			f.Println("Too many attempts")
			return ""
		}
		f.Printf("Please fill %s: ", title)
		value = readFunc()
	}
	return
}

func (f *frontendCLI) printAndLogError(args ...interface{}) {
	log.Error(args...)
	f.Println(args...)
}

func (f *frontendCLI) processAPIError(err error) {
	log.Warn("API error: ", err)
	switch err {
	case pmapi.ErrAPINotReachable:
		f.notifyInternetOff()
	case pmapi.ErrUpgradeApplication:
		f.notifyNeedUpgrade()
	default:
		f.Println("Server error:", err.Error())
	}
}

func (f *frontendCLI) notifyInternetOff() {
	f.Println("Internet connection is not available.")
}

func (f *frontendCLI) notifyInternetOn() {
	f.Println("Internet connection is available again.")
}

func (f *frontendCLI) notifyLogout(address string) {
	f.Printf("Account %s is disconnected. Login to continue using this account with email client.", address)
}

func (f *frontendCLI) notifyNeedUpgrade() {
	f.Println("Please download and install the newest version of application from", f.updates.GetDownloadLink())
}

func (f *frontendCLI) notifyCredentialsError() {
	// Print in 80-column width.
	f.Println("ProtonMail Import/Export is not able to detect a supported password manager")
	f.Println("(pass, gnome-keyring). Please install and set up a supported password manager")
	f.Println("and restart the application.")
}

func (f *frontendCLI) notifyCertIssue() {
	// Print in 80-column width.
	f.Println(`Connection security error: Your network connection to Proton services may
be insecure.

Description:
ProtonMail Import/Export was not able to establish a secure connection to Proton
servers due to a TLS certificate error. This means your connection may
potentially be insecure and susceptible to monitoring by third parties.

Recommendation:
* If you trust your network operator, you can continue to use ProtonMail
  as usual.
* If you don't trust your network operator, reconnect to ProtonMail over a VPN
  (such as ProtonVPN) which encrypts your Internet connection, or use
  a different network to access ProtonMail.
`)
}
