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

func newTestMBOXProvider(path string) *MBOXProvider {
	if path == "" {
		path = "testdata/mbox"
	}
	return NewMBOXProvider(path)
}

func TestMBOXProviderMailboxes(t *testing.T) {
	provider := newTestMBOXProvider("")

	tests := []struct {
		includeEmpty  bool
		wantMailboxes []Mailbox
	}{
		{true, []Mailbox{
			{Name: "Foo"},
			{Name: "Inbox"},
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

func TestMBOXProviderTransferTo(t *testing.T) {
	provider := newTestMBOXProvider("")

	rules, rulesClose := newTestRules(t)
	defer rulesClose()
	setupMBOXRules(rules)

	testTransferTo(t, rules, provider, []string{
		"Foo.mbox:1",
		"Inbox.mbox:1",
	})
}

func TestMBOXProviderTransferFrom(t *testing.T) {
	dir, err := ioutil.TempDir("", "eml")
	r.NoError(t, err)
	defer os.RemoveAll(dir) //nolint[errcheck]

	provider := newTestMBOXProvider(dir)

	rules, rulesClose := newTestRules(t)
	defer rulesClose()
	setupMBOXRules(rules)

	testTransferFrom(t, rules, provider, []Message{
		{ID: "Foo.mbox:1", Body: getTestMsgBody("msg"), Targets: []Mailbox{{Name: "Foo"}}},
	})

	checkMBOXFileStructure(t, dir, []string{
		"Foo.mbox",
	})
}

func TestMBOXProviderTransferFromTo(t *testing.T) {
	dir, err := ioutil.TempDir("", "eml")
	r.NoError(t, err)
	defer os.RemoveAll(dir) //nolint[errcheck]

	source := newTestMBOXProvider("")
	target := newTestMBOXProvider(dir)

	rules, rulesClose := newTestRules(t)
	defer rulesClose()
	setupEMLRules(rules)

	testTransferFromTo(t, rules, source, target, 5*time.Second)

	checkMBOXFileStructure(t, dir, []string{
		"Foo.mbox",
		"Inbox.mbox",
	})
}

func setupMBOXRules(rules transferRules) {
	_ = rules.setRule(Mailbox{Name: "Inbox"}, []Mailbox{{Name: "Inbox"}}, 0, 0)
	_ = rules.setRule(Mailbox{Name: "Foo"}, []Mailbox{{Name: "Foo"}}, 0, 0)
}

func checkMBOXFileStructure(t *testing.T, root string, expectedFiles []string) {
	files, err := getFilePathsWithSuffix(root, ".mbox")
	r.NoError(t, err)
	r.Equal(t, expectedFiles, files)
}
