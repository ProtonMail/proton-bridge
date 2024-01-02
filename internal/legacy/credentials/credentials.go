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

// Package credentials implements our struct stored in keychain.
// Store struct is kind of like a database client.
// Credentials struct is kind of like one record from the database.
package credentials

import (
	"encoding/base64"
	"errors"
	"fmt"
	"strings"

	"github.com/sirupsen/logrus"
)

const (
	sep = "\x00"

	itemLengthBridge       = 9
	itemLengthImportExport = 6 // Old format for Import-Export.
)

var (
	log = logrus.WithField("pkg", "credentials") //nolint:gochecknoglobals

	ErrWrongFormat = errors.New("malformed credentials")
)

type Credentials struct {
	UserID, // Do not marshal; used as a key.
	Name,
	Emails,
	APIToken string
	MailboxPassword []byte
	BridgePassword,
	Version string
	Timestamp int64
	IsHidden, // Deprecated.
	IsCombinedAddressMode bool
}

func (s *Credentials) Marshal() string {
	items := []string{
		s.Name,                    // 0
		s.Emails,                  // 1
		s.APIToken,                // 2
		string(s.MailboxPassword), // 3
		s.BridgePassword,          // 4
		s.Version,                 // 5
		"",                        // 6
		"",                        // 7
		"",                        // 8
	}

	items[6] = fmt.Sprint(s.Timestamp)

	if s.IsHidden {
		items[7] = "1"
	}

	if s.IsCombinedAddressMode {
		items[8] = "1"
	}

	str := strings.Join(items, sep)
	return base64.StdEncoding.EncodeToString([]byte(str))
}

func (s *Credentials) Unmarshal(secret string) error {
	b, err := base64.StdEncoding.DecodeString(secret)
	if err != nil {
		return err
	}
	items := strings.Split(string(b), sep)

	if len(items) != itemLengthBridge && len(items) != itemLengthImportExport {
		return ErrWrongFormat
	}

	s.Name = items[0]
	s.Emails = items[1]
	s.APIToken = items[2]
	s.MailboxPassword = []byte(items[3])

	switch len(items) {
	case itemLengthBridge:
		s.BridgePassword = items[4]
		s.Version = items[5]
		if _, err = fmt.Sscan(items[6], &s.Timestamp); err != nil {
			s.Timestamp = 0
		}
		if s.IsHidden = false; items[7] == "1" {
			s.IsHidden = true
		}
		if s.IsCombinedAddressMode = false; items[8] == "1" {
			s.IsCombinedAddressMode = true
		}

	case itemLengthImportExport:
		s.Version = items[4]
		if _, err = fmt.Sscan(items[5], &s.Timestamp); err != nil {
			s.Timestamp = 0
		}
	}
	return nil
}

func (s *Credentials) EmailList() []string {
	return strings.Split(s.Emails, ";")
}

func (s *Credentials) SplitAPIToken() (string, string, error) {
	split := strings.Split(s.APIToken, ":")

	if len(split) != 2 {
		return "", "", errors.New("malformed API token")
	}

	return split[0], split[1], nil
}
