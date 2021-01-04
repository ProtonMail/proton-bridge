// Copyright (c) 2021 Proton Technologies AG
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

package transfer

import (
	"fmt"
	"strings"
	"testing"

	r "github.com/stretchr/testify/require"
)

func TestGetGmailLabelsFromMboxReader(t *testing.T) {
	mboxFile := `From - Mon May  4 16:40:31 2020
Subject: Test 1
X-Gmail-Labels: Foo,Bar

hello

From - Mon May  4 16:40:31 2020
Subject: Test 2
X-Gmail-Labels: Foo , Baz

hello

From - Mon May  4 16:40:31 2020
Subject: Test 3
X-Gmail-Labels: ,

hello

From - Mon May  4 16:40:31 2020
Subject: Test 4
X-Gmail-Labels:

hello

From - Mon May  4 16:40:31 2020
Subject: Test 5

hello

`
	mboxReader := strings.NewReader(mboxFile)
	labels, err := getGmailLabelsFromMboxReader(mboxReader)
	r.NoError(t, err)
	r.Equal(t, toSet("Foo", "Bar", "Baz"), labels)
}

func TestGetGmailLabelsFromMessage(t *testing.T) {
	tests := []struct {
		body       string
		wantLabels stringSet
	}{
		{`Subject: One
X-Gmail-Labels: Foo,Bar

Hello
`, toSet("Foo", "Bar")},
		{`Subject: Two
X-Gmail-Labels: Foo , Bar ,

Hello
`, toSet("Foo", "Bar")},
		{`Subject: Three
X-Gmail-Labels: ,

Hello
`, toSet()},
		{`Subject: Four
X-Gmail-Labels:

Hello
`, toSet()},
		{`Subject: Five

Hello
`, toSet()},
	}
	for _, tc := range tests {
		tc := tc
		t.Run(fmt.Sprintf("%v", tc.body), func(t *testing.T) {
			labels, err := getGmailLabelsFromMessage([]byte(tc.body))
			r.NoError(t, err)
			r.Equal(t, tc.wantLabels, labels)
		})
	}
}

func TestGetGmailLabelsFromValue(t *testing.T) {
	tests := []struct {
		value      string
		wantLabels stringSet
	}{
		{"Foo,Bar", toSet("Foo", "Bar")},
		{" Foo , Bar ", toSet("Foo", "Bar")},
		{" Foo , Bar , ", toSet("Foo", "Bar")},
		{" Foo Bar ", toSet("Foo Bar")},
		{" , ", toSet()},
		{" ", toSet()},
		{"", toSet()},
		{"=?UTF-8?Q?Archived,Category_personal,test_=F0=9F=98=80=F0=9F=99=83?=", toSet("Archived", "Category personal", "test 😀🙃")},
		{"IMAP_NotJunk,Foo,Opened,bar,Unread", toSet("Foo", "bar")},
	}
	for _, tc := range tests {
		tc := tc
		t.Run(fmt.Sprintf("%v", tc.value), func(t *testing.T) {
			labels := getGmailLabelsFromValue(tc.value)
			r.Equal(t, tc.wantLabels, labels)
		})
	}
}

func toSet(items ...string) stringSet {
	set := map[string]bool{}
	for _, item := range items {
		set[item] = true
	}
	return set
}
