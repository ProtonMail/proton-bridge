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

package mobileconfig

import (
	"io"
	"text/template"

	"github.com/google/uuid"
)

// Config represents an Apple mobileconfig file.
type Config struct {
	EmailAddress       string
	DisplayName        string
	Identifier         string
	Organization       string
	AccountName        string
	AccountDescription string

	IMAP *IMAP
	SMTP *SMTP

	Description string
	ContentUUID string
	UUID        string
}

type IMAP struct {
	Hostname string
	Port     int
	TLS      bool

	Username string
	Password string
}

type SMTP struct {
	Hostname string
	Port     int
	TLS      bool

	// Leave Username blank to do not use SMTP authentication.
	Username string
	// Leave Password blank to use IMAP credentials.
	Password string
}

func (c *Config) WriteOut(w io.Writer) error {
	if c.ContentUUID == "" {
		uuid := uuid.New()
		c.ContentUUID = uuid.String()
	}

	if c.UUID == "" {
		uuid := uuid.New()
		c.UUID = uuid.String()
	}

	return template.Must(template.New("mobileconfig").Parse(mailTemplate)).Execute(w, c)
}
