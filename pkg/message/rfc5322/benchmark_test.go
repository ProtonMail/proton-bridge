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
	"net/mail"
	"testing"
)

func BenchmarkStandardSet(b *testing.B) {
	for i := 0; i < b.N; i++ {
		parseSet(benchStandardSet)
	}
}

func BenchmarkStandardSetGolang(b *testing.B) {
	for i := 0; i < b.N; i++ {
		parseSetGolang(benchStandardSet)
	}
}

func BenchmarkEncodedSet(b *testing.B) {
	for i := 0; i < b.N; i++ {
		parseSet(benchEncodedSet)
	}
}

func BenchmarkEncodedSetGolang(b *testing.B) {
	for i := 0; i < b.N; i++ {
		parseSetGolang(benchEncodedSet)
	}
}

func BenchmarkAddressListSet(b *testing.B) {
	for i := 0; i < b.N; i++ {
		parseSet(benchAddressListSet)
	}
}

func BenchmarkAddressListSetGolang(b *testing.B) {
	for i := 0; i < b.N; i++ {
		parseSetGolang(benchAddressListSet)
	}
}

func BenchmarkGroupSet(b *testing.B) {
	for i := 0; i < b.N; i++ {
		parseSet(benchGroupSet)
	}
}

func BenchmarkGroupSetGolang(b *testing.B) {
	for i := 0; i < b.N; i++ {
		parseSetGolang(benchGroupSet)
	}
}

func parseSet(set []string) {
	for _, addr := range set {
		_, _ = ParseAddressList(addr)
	}
}

func parseSetGolang(set []string) {
	for _, addr := range set {
		_, _ = mail.ParseAddressList(addr)
	}
}

var benchStandardSet = []string{
	`user@example.com`,
	`John Doe <jdoe@machine.example>`,
	`Mary Smith <mary@example.net>`,
	`"Joe Q. Public" <john.q.public@example.com>`,
	`Mary Smith <mary@x.test>`,
	`jdoe@example.org`,
	`Who? <one@y.test>`,
	`<boss@nil.test>`,
	`"Giant; \"Big\" Box" <sysservices@example.net>`,
	`Pete <pete@silly.example>`,
	`"Mary Smith: Personal Account" <smith@home.example>`,
	`Pete(A nice \) chap) <pete(his account)@silly.test(his host)>`,
	`Gogh Fir <gf@example.com>`,
	`normal name  <username@server.com>`,
	`"comma, name"  <username@server.com>`,
	`name  <username@server.com> (ignore comment)`,
	`"Mail Robot" <>`,
	`Michal Hořejšek <hořejšek@mail.com>`,
	`First Last <user@domain.com >`,
	`First Last <user@domain.com. >`,
	`First Last <user@domain.com.>`,
	`First Last <user@domain.com:25>`,
	`First Last <user@[10.0.0.1]>`,
	`<postmaster@[10.10.10.10]>`,
	`user@domain <user@domain.com>`,
	`First Last < user@domain.com>`,
	`First Middle @ Last <user@domain.com>`,
	`user@domain.com,`,
	`First Middle "Last" <user@domain.com>`,
	`First Middle Last <user@domain.com>`,
	`First Middle"Last" <user@domain.com>`,
	`First Middle "Last"<user@domain.com>`,
	`First "Middle" "Last" <user@domain.com>`,
	`First "Middle""Last" <user@domain.com>`,
}

var benchEncodedSet = []string{
	`=?US-ASCII?Q?Keith_Moore?= <moore@cs.utk.edu>`,
	`=?ISO-8859-1?Q?Keld_J=F8rn_Simonsen?= <keld@dkuug.dk>`,
	`=?ISO-8859-1?Q?Andr=E9?= Pirard <PIRARD@vm1.ulg.ac.be>`,
	`=?ISO-8859-1?Q?Olle_J=E4rnefors?= <ojarnef@admin.kth.se>`,
	`=?ISO-8859-1?Q?Patrik_F=E4ltstr=F6m?= <paf@nada.kth.se>`,
	`Nathaniel Borenstein <nsb@thumper.bellcore.com> (=?iso-8859-8?b?7eXs+SDv4SDp7Oj08A==?=)`,
	`=?UTF-8?B?PEJlemUgam3DqW5hPg==?= <user@domain.com>`,
	`First Middle =?utf-8?Q?Last?= <user@domain.com>`,
	`First Middle =?utf-8?Q?Last?=<user@domain.com>`,
	`First =?utf-8?Q?Middle?= =?utf-8?Q?Last?= <user@domain.com>`,
	`First =?utf-8?Q?Middle?==?utf-8?Q?Last?= <user@domain.com>`,
	`First "Middle"=?utf-8?Q?Last?= <user@domain.com>`,
	`First "Middle" =?utf-8?Q?Last?= <user@domain.com>`,
	`First "Middle" =?utf-8?Q?Last?=<user@domain.com>`,
	`=?UTF-8?B?PEJlemUgam3DqW5hPg==?= <user@domain.com>`,
	`=?utf-8?B?6YCZ5piv5ryi5a2X55qE5LiA5YCL5L6L5a2Q?= <proton.testqa@gmail.com>`,
	`=?utf-8?B?8J+MmfCfjbc=?= <user@domain.com>`,
	`=?utf-8?B?8J+MmfCfjbc=?= <user@domain.com>`,
}

var benchAddressListSet = []string{
	`Alice <alice@example.com>, Bob <bob@example.com>, Eve <eve@example.com>`,
	`Ed Jones <c@a.test>,joe@where.test,John <jdoe@one.test>`,
	`name (ignore comment)  <username@server.com>,  (Comment as name) username2@server.com`,
	`"normal name"  <username@server.com>, "comma, name" <address@server.com>`,
	`"comma, one"  <username@server.com>, "comma, two" <address@server.com>`,
	`normal name  <username@server.com>, (comment)All.(around)address@(the)server.com`,
	`normal name  <username@server.com>, All.("comma, in comment")address@(the)server.com`,
}

var benchGroupSet = []string{
	`A Group:Ed Jones <c@a.test>,joe@where.test,John <jdoe@one.test>;`,
	`Undisclosed recipients:;`,
	`(Empty list)(start)Hidden recipients  :(nobody(that I know))  ;`,
}
