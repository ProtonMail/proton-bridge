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

// +build darwin

package autoconfig

import (
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	mobileconfig "github.com/ProtonMail/go-apple-mobileconfig"
	"github.com/ProtonMail/proton-bridge/internal/bridge"
	"github.com/ProtonMail/proton-bridge/internal/frontend/types"
)

func init() { //nolint[gochecknoinit]
	available = append(available, &appleMail{})
}

type appleMail struct{}

func (c *appleMail) Name() string {
	return "Apple Mail"
}

func (c *appleMail) Configure(imapPort, smtpPort int, imapSSL, smtpSSL bool, user types.User, addressIndex int) error { //nolint[funlen]
	var addresses string
	var displayName string

	if user.IsCombinedAddressMode() {
		displayName = user.GetPrimaryAddress()
		addresses = strings.Join(user.GetAddresses(), ",")
	} else {
		for idx, address := range user.GetAddresses() {
			if idx == addressIndex {
				displayName = address
				break
			}
		}
		addresses = displayName
	}

	timestamp := strconv.FormatInt(time.Now().Unix(), 10)

	mc := &mobileconfig.Config{
		EmailAddress: addresses,
		DisplayName:  displayName,
		Identifier:   "protonmail " + displayName + timestamp,
		Imap: &mobileconfig.Imap{
			Hostname: bridge.Host,
			Port:     imapPort,
			Tls:      imapSSL,
			Username: displayName,
			Password: user.GetBridgePassword(),
		},
		Smtp: &mobileconfig.Smtp{
			Hostname: bridge.Host,
			Port:     smtpPort,
			Tls:      smtpSSL,
			Username: displayName,
		},
	}

	dir, err := ioutil.TempDir("", "protonmail-autoconfig")
	if err != nil {
		return err
	}

	// Make sure the temporary file is deleted.
	go (func() {
		<-time.After(10 * time.Minute)
		_ = os.RemoveAll(dir)
	})()

	// Make sure the file is only readable for the current user.
	f, err := os.OpenFile(filepath.Join(dir, "protonmail.mobileconfig"), os.O_RDWR|os.O_CREATE, 0600)
	if err != nil {
		return err
	}

	if err := mc.WriteTo(f); err != nil {
		_ = f.Close()
		return err
	}
	_ = f.Close()

	return exec.Command("open", f.Name()).Run() // nolint[gosec]
}
