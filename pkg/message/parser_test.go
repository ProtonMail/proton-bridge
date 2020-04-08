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

package message

import (
	"net/mail"
	"testing"
)

func TestRFC822AddressFormat(t *testing.T) { //nolint[funlen]
	tests := []struct {
		address  string
		expected []string
	}{
		{
			" normal name  <username@server.com>",
			[]string{
				"\"normal name\" <username@server.com>",
			},
		},
		{
			" \"comma, name\"  <username@server.com>",
			[]string{
				"\"comma, name\" <username@server.com>",
			},
		},
		{
			" name  <username@server.com> (ignore comment)",
			[]string{
				"\"name\" <username@server.com>",
			},
		},
		{
			" name (ignore comment)  <username@server.com>,  (Comment as name) username2@server.com",
			[]string{
				"\"name\" <username@server.com>",
				"<username2@server.com>",
			},
		},
		{
			" normal name  <username@server.com>, (comment)All.(around)address@(the)server.com",
			[]string{
				"\"normal name\" <username@server.com>",
				"<All.address@server.com>",
			},
		},
		{
			" normal name  <username@server.com>, All.(\"comma, in comment\")address@(the)server.com",
			[]string{
				"\"normal name\" <username@server.com>",
				"<All.address@server.com>",
			},
		},
		{
			" \"normal name\"  <username@server.com>, \"comma, name\" <address@server.com>",
			[]string{
				"\"normal name\" <username@server.com>",
				"\"comma, name\" <address@server.com>",
			},
		},
		{
			" \"comma, one\"  <username@server.com>, \"comma, two\" <address@server.com>",
			[]string{
				"\"comma, one\" <username@server.com>",
				"\"comma, two\" <address@server.com>",
			},
		},
		{
			" \"comma, name\"  <username@server.com>, another, name <address@server.com>",
			[]string{
				"\"comma, name\" <username@server.com>",
				"\"another, name\" <address@server.com>",
			},
		},
	}

	for _, data := range tests {
		uncommented := parseAddressComment(data.address)
		result, err := mail.ParseAddressList(uncommented)
		if err != nil {
			t.Errorf("Can not parse '%s' created from '%s': %v", uncommented, data.address, err)
		}
		if len(result) != len(data.expected) {
			t.Errorf("Wrong parsing of '%s' created from '%s': expected '%s' but have '%+v'", uncommented, data.address, data.expected, result)
		}
		for i, result := range result {
			if data.expected[i] != result.String() {
				t.Errorf("Wrong parsing\nof %q\ncreated from %q:\nexpected %q\nbut have %q", uncommented, data.address, data.expected, result.String())
			}
		}
	}
}
