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

package transfer

import (
	"fmt"
	"io/ioutil"
	"os"
	"testing"
	"time"

	r "github.com/stretchr/testify/require"
)

func newTestEMLProvider(path string) *EMLProvider {
	if path == "" {
		path = "testdata/eml"
	}
	return NewEMLProvider(path)
}

func TestEMLProviderMailboxes(t *testing.T) {
	provider := newTestEMLProvider("")

	tests := []struct {
		includeEmpty  bool
		wantMailboxes []Mailbox
	}{
		{true, []Mailbox{
			{Name: "Foo"},
			{Name: "Inbox"},
			{Name: "eml"},
		}},
		{false, []Mailbox{
			{Name: "Foo"},
			{Name: "Inbox"},
		}},
	}
	for _, tc := range tests {
		tc := tc
		t.Run(fmt.Sprintf("%v", tc.includeEmpty), func(t *testing.T) {
			mailboxes, err := provider.Mailboxes(tc.includeEmpty, false)
			r.NoError(t, err)
			r.Equal(t, tc.wantMailboxes, mailboxes)
		})
	}
}

func TestEMLProviderTransferTo(t *testing.T) {
	provider := newTestEMLProvider("")

	rules, rulesClose := newTestRules(t)
	defer rulesClose()
	setupEMLRules(rules)

	testTransferTo(t, rules, provider, []string{
		"Foo/msg.eml",
		"Inbox/msg.eml",
	})
}

func TestEMLProviderTransferFrom(t *testing.T) {
	dir, err := ioutil.TempDir("", "eml")
	r.NoError(t, err)
	defer os.RemoveAll(dir) //nolint[errcheck]

	provider := newTestEMLProvider(dir)

	rules, rulesClose := newTestRules(t)
	defer rulesClose()
	setupEMLRules(rules)

	testTransferFrom(t, rules, provider, []Message{
		{ID: "Foo/msg.eml", Body: getTestMsgBody("msg"), Targets: []Mailbox{{Name: "Foo"}}},
	})

	checkEMLFileStructure(t, dir, []string{
		"Foo/msg.eml",
	})
}

func TestEMLProviderTransferFromTo(t *testing.T) {
	dir, err := ioutil.TempDir("", "eml")
	r.NoError(t, err)
	defer os.RemoveAll(dir) //nolint[errcheck]

	source := newTestEMLProvider("")
	target := newTestEMLProvider(dir)

	rules, rulesClose := newTestRules(t)
	defer rulesClose()
	setupEMLRules(rules)

	testTransferFromTo(t, rules, source, target, 5*time.Second)

	checkEMLFileStructure(t, dir, []string{
		"Foo/msg.eml",
		"Inbox/msg.eml",
	})
}

func setupEMLRules(rules transferRules) {
	_ = rules.setRule(Mailbox{Name: "Inbox"}, []Mailbox{{Name: "Inbox"}}, 0, 0)
	_ = rules.setRule(Mailbox{Name: "Foo"}, []Mailbox{{Name: "Foo"}}, 0, 0)
}

func checkEMLFileStructure(t *testing.T, root string, expectedFiles []string) {
	files, err := getFilePathsWithSuffix(root, ".eml")
	r.NoError(t, err)
	r.Equal(t, expectedFiles, files)
}
