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
// along with Proton Mail Bridge.  If not, see <https://www.gnu.org/licenses/>.

package main

import (
	"bytes"
	"crypto/tls"
	"flag"
	"fmt"
	"os"

	"github.com/emersion/go-sasl"
	"github.com/emersion/go-smtp"
)

var (
	serverURL    = flag.String("server", "127.0.0.1:1025", "SMTP server address:port")
	userName     = flag.String("user-name", "user", "SMTP user name")
	userPassword = flag.String("user-pwd", "password", "SMTP user password")
	toAddr       = flag.String("toAddr", "", "Address toAddr whom toAddr send the message")
)

func main() {
	flag.Parse()
	flag.Usage = func() {
		fmt.Printf("Usage %v [options] file0 ... fileN\n", os.Args[0])
		fmt.Printf("\nOptions:\n")
		flag.PrintDefaults()
	}

	if len(*toAddr) == 0 {
		panic(fmt.Errorf("to flag can't be empty"))
	}

	args := flag.Args()
	if len(args) == 0 {
		flag.Usage()
		return
	}

	smtpClient, err := smtp.Dial(*serverURL)
	if err != nil {
		panic(fmt.Errorf("failed to connect to server: %w", err))
	}
	defer func() { _ = smtpClient.Close() }()

	// Upgrade to TLS.
	if err := smtpClient.StartTLS(&tls.Config{InsecureSkipVerify: true}); err != nil {
		panic(fmt.Errorf("failed to starttls: %w", err))
	}

	// Authorize with SASL PLAIN.
	if err := smtpClient.Auth(sasl.NewPlainClient(
		*userName,
		*userName,
		*userPassword,
	)); err != nil {
		panic(fmt.Errorf("failed to login: %w", err))
	}

	for idx, v := range args {
		fileData, err := os.ReadFile(v)
		if err != nil {
			panic(fmt.Errorf("failed to read file:%v - %w", v, err))
		}

		// Send the message.
		if err := smtpClient.SendMail(
			*userName,
			[]string{*toAddr},
			bytes.NewReader(fileData),
		); err != nil {
			panic(fmt.Errorf("failed to send msg %v: %w", idx, err))
		}
	}
}
