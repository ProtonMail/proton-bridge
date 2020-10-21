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

package rfc5322

import (
	"encoding/xml"
	"io"
	"net/mail"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseSingleAddress(t *testing.T) {
	tests := []struct {
		input string
		addrs []*mail.Address
	}{
		{
			input: `user@example.com`,
			addrs: []*mail.Address{{
				Address: `user@example.com`,
			}},
		},
		{
			input: `John Doe <jdoe@machine.example>`,
			addrs: []*mail.Address{{
				Name:    `John Doe`,
				Address: `jdoe@machine.example`,
			}},
		},
		{
			input: `Mary Smith <mary@example.net>`,
			addrs: []*mail.Address{{
				Name:    `Mary Smith`,
				Address: `mary@example.net`,
			}},
		},
		{
			input: `"Joe Q. Public" <john.q.public@example.com>`,
			addrs: []*mail.Address{{
				Name:    `Joe Q. Public`,
				Address: `john.q.public@example.com`,
			}},
		},
		{
			input: `Mary Smith <mary@x.test>`,
			addrs: []*mail.Address{{
				Name:    `Mary Smith`,
				Address: `mary@x.test`,
			}},
		},
		{
			input: `jdoe@example.org`,
			addrs: []*mail.Address{{
				Address: `jdoe@example.org`,
			}},
		},
		{
			input: `Who? <one@y.test>`,
			addrs: []*mail.Address{{
				Name:    `Who?`,
				Address: `one@y.test`,
			}},
		},
		{
			input: `<boss@nil.test>`,
			addrs: []*mail.Address{{
				Address: `boss@nil.test`,
			}},
		},
		{
			input: `"Giant; \"Big\" Box" <sysservices@example.net>`,
			addrs: []*mail.Address{{
				Name:    `Giant; "Big" Box`,
				Address: `sysservices@example.net`,
			}},
		},
		{
			input: `Pete <pete@silly.example>`,
			addrs: []*mail.Address{{
				Name:    `Pete`,
				Address: `pete@silly.example`,
			}},
		},
		{
			input: `"Mary Smith: Personal Account" <smith@home.example>`,
			addrs: []*mail.Address{{
				Name:    `Mary Smith: Personal Account`,
				Address: `smith@home.example`,
			}},
		},
		{
			input: `Pete(A nice \) chap) <pete(his account)@silly.test(his host)>`,
			addrs: []*mail.Address{{
				Name:    `Pete`,
				Address: `pete@silly.test`,
			}},
		},
		{
			input: `Gogh Fir <gf@example.com>`,
			addrs: []*mail.Address{{
				Name:    `Gogh Fir`,
				Address: `gf@example.com`,
			}},
		},
		{
			input: `normal name  <username@server.com>`,
			addrs: []*mail.Address{{
				Name:    `normal name`,
				Address: `username@server.com`,
			}},
		},
		{
			input: `"comma, name"  <username@server.com>`,
			addrs: []*mail.Address{{
				Name:    `comma, name`,
				Address: `username@server.com`,
			}},
		},
		{
			input: `name  <username@server.com> (ignore comment)`,
			addrs: []*mail.Address{{
				Name:    `name`,
				Address: `username@server.com`,
			}},
		},
		{
			input: `"Mail Robot" <>`,
			addrs: []*mail.Address{{
				Name: `Mail Robot`,
			}},
		},
		{
			input: `Michal Hořejšek <hořejšek@mail.com>`,
			addrs: []*mail.Address{{
				Name:    `Michal Hořejšek`,
				Address: `hořejšek@mail.com`, // Not his real address.
			}},
		},
		{
			input: `First Last <user@domain.com >`,
			addrs: []*mail.Address{{
				Name:    `First Last`,
				Address: `user@domain.com`,
			}},
		},
		{
			input: `First Last <user@domain.com. >`,
			addrs: []*mail.Address{{
				Name:    `First Last`,
				Address: `user@domain.com.`,
			}},
		},
		{
			input: `First Last <user@domain.com.>`,
			addrs: []*mail.Address{{
				Name:    `First Last`,
				Address: `user@domain.com.`,
			}},
		},
		{
			input: `First Last <user@domain.com:25>`,
			addrs: []*mail.Address{{
				Name:    `First Last`,
				Address: `user@domain.com:25`,
			}},
		},
		{
			input: `First Last <user@[10.0.0.1]>`,
			addrs: []*mail.Address{{
				Name:    `First Last`,
				Address: `user@[10.0.0.1]`,
			}},
		},
		{
			input: `<postmaster@[10.10.10.10]>`,
			addrs: []*mail.Address{
				{
					Address: `postmaster@[10.10.10.10]`,
				},
			},
		},
		{
			input: `user@domain <user@domain.com>`,
			addrs: []*mail.Address{{
				// Name:    `user@domain`,
				Name:    `user @ domain`,
				Address: `user@domain.com`,
			}},
		},
		{
			input: `First Last < user@domain.com>`,
			addrs: []*mail.Address{{
				Name:    `First Last`,
				Address: `user@domain.com`,
			}},
		},
		{
			input: `First Middle @ Last <user@domain.com>`,
			addrs: []*mail.Address{{
				Name:    `First Middle @ Last`,
				Address: `user@domain.com`,
			}},
		},
		{
			input: `user@domain.com,`,
			addrs: []*mail.Address{
				{
					Address: `user@domain.com`,
				},
			},
		},
		{
			input: `First Middle "Last" <user@domain.com>`,
			addrs: []*mail.Address{
				{
					Name:    `First Middle Last`,
					Address: `user@domain.com`,
				},
			},
		},
		{
			input: `First Middle Last <user@domain.com>`,
			addrs: []*mail.Address{
				{
					Name:    `First Middle Last`,
					Address: `user@domain.com`,
				},
			},
		},
		{
			input: `First Middle"Last" <user@domain.com>`,
			addrs: []*mail.Address{
				{
					Name:    `First Middle Last`,
					Address: `user@domain.com`,
				},
			},
		},
		{
			input: `First Middle "Last"<user@domain.com>`,
			addrs: []*mail.Address{
				{
					Name:    `First Middle Last`,
					Address: `user@domain.com`,
				},
			},
		},
		{
			input: `First "Middle" "Last" <user@domain.com>`,
			addrs: []*mail.Address{
				{
					Name:    `First Middle Last`,
					Address: `user@domain.com`,
				},
			},
		},
		{
			input: `First "Middle""Last" <user@domain.com>`,
			addrs: []*mail.Address{
				{
					Name:    `First Middle Last`,
					Address: `user@domain.com`,
				},
			},
		},
	}
	for _, test := range tests {
		test := test

		t.Run(test.input, func(t *testing.T) {
			addrs, err := ParseAddressList(test.input)
			assert.NoError(t, err)
			assert.ElementsMatch(t, test.addrs, addrs)
		})
	}
}

func TestParseSingleAddressEncodedWord(t *testing.T) {
	tests := []struct {
		input string
		addrs []*mail.Address
	}{
		{
			input: `=?US-ASCII?Q?Keith_Moore?= <moore@cs.utk.edu>`,
			addrs: []*mail.Address{{
				Name:    `Keith Moore`,
				Address: `moore@cs.utk.edu`,
			}},
		},
		{
			input: `=?ISO-8859-1?Q?Keld_J=F8rn_Simonsen?= <keld@dkuug.dk>`,
			addrs: []*mail.Address{{
				Name:    `Keld Jørn Simonsen`,
				Address: `keld@dkuug.dk`,
			}},
		},
		{
			input: `=?ISO-8859-1?Q?Andr=E9?= Pirard <PIRARD@vm1.ulg.ac.be>`,
			addrs: []*mail.Address{{
				Name:    `André Pirard`,
				Address: `PIRARD@vm1.ulg.ac.be`,
			}},
		},
		{
			input: `=?ISO-8859-1?Q?Olle_J=E4rnefors?= <ojarnef@admin.kth.se>`,
			addrs: []*mail.Address{{
				Name:    `Olle Järnefors`,
				Address: `ojarnef@admin.kth.se`,
			}},
		},
		{
			input: `=?ISO-8859-1?Q?Patrik_F=E4ltstr=F6m?= <paf@nada.kth.se>`,
			addrs: []*mail.Address{{
				Name:    `Patrik Fältström`,
				Address: `paf@nada.kth.se`,
			}},
		},
		{
			input: `Nathaniel Borenstein <nsb@thumper.bellcore.com> (=?iso-8859-8?b?7eXs+SDv4SDp7Oj08A==?=)`,
			addrs: []*mail.Address{{
				Name:    `Nathaniel Borenstein`,
				Address: `nsb@thumper.bellcore.com`,
			}},
		},
		{
			input: `=?UTF-8?B?PEJlemUgam3DqW5hPg==?= <user@domain.com>`,
			addrs: []*mail.Address{
				{
					Name:    `<Beze jména>`,
					Address: `user@domain.com`,
				},
			},
		},
		{
			input: `First Middle =?utf-8?Q?Last?= <user@domain.com>`,
			addrs: []*mail.Address{
				{
					Name:    `First Middle Last`,
					Address: `user@domain.com`,
				},
			},
		},
		/*
			{
				input: `First Middle=?utf-8?Q?Last?= <user@domain.com>`,
				addrs: []*mail.Address{
					{
						Name:    `First Middle Last`,
						Address: `user@domain.com`,
					},
				},
			},
		*/
		{
			input: `First Middle =?utf-8?Q?Last?=<user@domain.com>`,
			addrs: []*mail.Address{
				{
					Name:    `First Middle Last`,
					Address: `user@domain.com`,
				},
			},
		},
		{
			input: `First =?utf-8?Q?Middle?= =?utf-8?Q?Last?= <user@domain.com>`,
			addrs: []*mail.Address{
				{
					Name:    `First Middle Last`,
					Address: `user@domain.com`,
				},
			},
		},
		{
			input: `First =?utf-8?Q?Middle?==?utf-8?Q?Last?= <user@domain.com>`,
			addrs: []*mail.Address{
				{
					Name:    `First Middle Last`,
					Address: `user@domain.com`,
				},
			},
		},
		{
			input: `First "Middle"=?utf-8?Q?Last?= <user@domain.com>`,
			addrs: []*mail.Address{
				{
					Name:    `First Middle Last`,
					Address: `user@domain.com`,
				},
			},
		},
		{
			input: `First "Middle" =?utf-8?Q?Last?= <user@domain.com>`,
			addrs: []*mail.Address{
				{
					Name:    `First Middle Last`,
					Address: `user@domain.com`,
				},
			},
		},
		{
			input: `First "Middle" =?utf-8?Q?Last?=<user@domain.com>`,
			addrs: []*mail.Address{
				{
					Name:    `First Middle Last`,
					Address: `user@domain.com`,
				},
			},
		},
		{
			input: `=?UTF-8?B?PEJlemUgam3DqW5hPg==?= <user@domain.com>`,
			addrs: []*mail.Address{
				{
					Name:    `<Beze jména>`,
					Address: `user@domain.com`,
				},
			},
		},
	}
	for _, test := range tests {
		test := test

		t.Run(test.input, func(t *testing.T) {
			addrs, err := ParseAddressList(test.input)
			assert.NoError(t, err)
			assert.ElementsMatch(t, test.addrs, addrs)
		})
	}
}

func TestParseAddressList(t *testing.T) {
	tests := []struct {
		input string
		addrs []*mail.Address
	}{
		{
			input: `Alice <alice@example.com>, Bob <bob@example.com>, Eve <eve@example.com>`,
			addrs: []*mail.Address{
				{
					Name:    `Alice`,
					Address: `alice@example.com`,
				},
				{
					Name:    `Bob`,
					Address: `bob@example.com`,
				},
				{
					Name:    `Eve`,
					Address: `eve@example.com`,
				},
			},
		},
		{
			input: `Ed Jones <c@a.test>,joe@where.test,John <jdoe@one.test>`,
			addrs: []*mail.Address{
				{
					Name:    `Ed Jones`,
					Address: `c@a.test`,
				},
				{
					Address: `joe@where.test`,
				},
				{
					Name:    `John`,
					Address: `jdoe@one.test`,
				},
			},
		},
		{
			input: `name (ignore comment)  <username@server.com>,  (Comment as name) username2@server.com`,
			addrs: []*mail.Address{
				{
					Name:    `name`,
					Address: `username@server.com`,
				},
				{
					Address: `username2@server.com`,
				},
			},
		},
		{
			input: `"normal name"  <username@server.com>, "comma, name" <address@server.com>`,
			addrs: []*mail.Address{
				{
					Name:    `normal name`,
					Address: `username@server.com`,
				},
				{
					Name:    `comma, name`,
					Address: `address@server.com`,
				},
			},
		},
		{
			input: `"comma, one"  <username@server.com>, "comma, two" <address@server.com>`,
			addrs: []*mail.Address{
				{
					Name:    `comma, one`,
					Address: `username@server.com`,
				},
				{
					Name:    `comma, two`,
					Address: `address@server.com`,
				},
			},
		},
		{
			input: `normal name  <username@server.com>, (comment)All.(around)address@(the)server.com`,
			addrs: []*mail.Address{
				{
					Name:    `normal name`,
					Address: `username@server.com`,
				},
				{
					Address: `All.address@server.com`,
				},
			},
		},
		{
			input: `normal name  <username@server.com>, All.("comma, in comment")address@(the)server.com`,
			addrs: []*mail.Address{
				{
					Name:    `normal name`,
					Address: `username@server.com`,
				},
				{
					Address: `All.address@server.com`,
				},
			},
		},
	}
	for _, test := range tests {
		test := test

		t.Run(test.input, func(t *testing.T) {
			addrs, err := ParseAddressList(test.input)
			assert.NoError(t, err)
			assert.ElementsMatch(t, test.addrs, addrs)
		})
	}
}

func TestParseGroup(t *testing.T) {
	tests := []struct {
		input string
		addrs []*mail.Address
	}{
		{
			input: `A Group:Ed Jones <c@a.test>,joe@where.test,John <jdoe@one.test>;`,
			addrs: []*mail.Address{
				{
					Name:    `Ed Jones`,
					Address: `c@a.test`,
				},
				{
					Address: `joe@where.test`,
				},
				{
					Name:    `John`,
					Address: `jdoe@one.test`,
				},
			},
		},
		{
			input: `Undisclosed recipients:;`,
			addrs: []*mail.Address{},
		},
		{
			input: `(Empty list)(start)Hidden recipients  :(nobody(that I know))  ;`,
			addrs: []*mail.Address{},
		},
	}
	for _, test := range tests {
		test := test

		t.Run(test.input, func(t *testing.T) {
			addrs, err := ParseAddressList(test.input)
			assert.NoError(t, err)
			assert.ElementsMatch(t, test.addrs, addrs)
		})
	}
}

// TestParseRejectedAddresses tests that weird addresses that are rejected by
// serverside are also rejected by us. If for some reason we end up being able
// to parse these malformed addresses, great! For now let's collect them here.
func TestParseRejectedAddresses(t *testing.T) {
	tests := []struct {
		input string
		addrs []*mail.Address
	}{
		{input: `"comma, name"  <username@server.com>, another, name <address@server.com>`},
		{input: `username`},
		{input: `undisclosed-recipients:`},
		{input: `=?ISO-8859-2?Q?First_Last?= <user@domain.com>, <user@domain.com,First/AAA/BBB/CCC,>`},
		{input: `user@domain...com`},
		{input: `=?windows-1250?Q?Spr=E1vce_syst=E9mu?=`},
		{input: `"'user@domain.com.'"`},
		{input: `<this is not an email address>`},
	}
	for _, test := range tests {
		test := test

		t.Run(test.input, func(t *testing.T) {
			_, err := ParseAddressList(test.input)
			assert.Error(t, err)
		})
	}
}

// TestIsEmailValidCategory runs over the "IsEmail" standard tests,
// ensuring it can at least recognize all emails in the "valid" category.
// In future, we should expand these tests to run over more categories.
func TestIsEmailValidCategory(t *testing.T) {
	f, err := os.Open("tests.xml")
	require.NoError(t, err)
	defer func() { require.NoError(t, err) }()

	for test := range readTestCases(f) {
		test := test

		if test.category != "ISEMAIL_VALID_CATEGORY" {
			continue
		}

		t.Run(test.id, func(t *testing.T) {
			_, err := ParseAddressList(test.address)
			assert.NoError(t, err)
		})
	}
}

type testCase struct {
	id        string
	address   string
	category  string
	diagnosis string
}

func readTestCases(r io.Reader) chan testCase {
	ch := make(chan testCase)

	var (
		test testCase
		data string
	)

	go func() {
		decoder := xml.NewDecoder(r)

		for token, err := decoder.Token(); err == nil; token, err = decoder.Token() {
			switch t := token.(type) {
			case xml.StartElement:
				if t.Name.Local == "test" {
					test = testCase{
						id: t.Attr[0].Value,
					}
				}

			case xml.EndElement:
				switch t.Name.Local {
				case "test":
					ch <- test

				case "address":
					test.address = data

				case "category":
					test.category = data

				case "diagnosis":
					test.diagnosis = data
				}

			case xml.CharData:
				data = string(t)
			}
		}

		close(ch)
	}()

	return ch
}
