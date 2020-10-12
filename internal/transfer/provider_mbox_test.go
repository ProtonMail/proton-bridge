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
	tests := []struct {
		provider      *MBOXProvider
		includeEmpty  bool
		wantMailboxes []Mailbox
	}{
		{newTestMBOXProvider(""), true, []Mailbox{
			{Name: "All Mail"},
			{Name: "Foo"},
			{Name: "Bar"},
			{Name: "Inbox"},
		}},
		{newTestMBOXProvider(""), false, []Mailbox{
			{Name: "All Mail"},
			{Name: "Foo"},
			{Name: "Bar"},
			{Name: "Inbox"},
		}},
	}
	for _, tc := range tests {
		tc := tc
		t.Run(fmt.Sprintf("%v", tc.includeEmpty), func(t *testing.T) {
			mailboxes, err := tc.provider.Mailboxes(tc.includeEmpty, false)
			r.NoError(t, err)
			r.ElementsMatch(t, tc.wantMailboxes, mailboxes)
		})
	}
}

func TestMBOXProviderTransferTo(t *testing.T) {
	provider := newTestMBOXProvider("")

	rules, rulesClose := newTestRules(t)
	defer rulesClose()
	setupMBOXRules(rules)

	msgs := testTransferTo(t, rules, provider, []string{
		"All Mail.mbox:1",
		"All Mail.mbox:2",
		"Foo.mbox:1",
		"Inbox.mbox:1",
	})
	got := map[string][]string{}
	for _, msg := range msgs {
		got[msg.ID] = msg.targetNames()
	}
	r.Equal(t, map[string][]string{
		"All Mail.mbox:1": {"Archive", "Foo"}, // Bar is not in rules.
		"All Mail.mbox:2": {"Archive", "Foo"},
		"Foo.mbox:1":      {"Foo"},
		"Inbox.mbox:1":    {"Inbox"},
	}, got)
}

func TestMBOXProviderTransferFrom(t *testing.T) {
	dir, err := ioutil.TempDir("", "mbox")
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
	dir, err := ioutil.TempDir("", "mbox")
	r.NoError(t, err)
	defer os.RemoveAll(dir) //nolint[errcheck]

	source := newTestMBOXProvider("")
	target := newTestMBOXProvider(dir)

	rules, rulesClose := newTestRules(t)
	defer rulesClose()
	setupMBOXRules(rules)

	testTransferFromTo(t, rules, source, target, 5*time.Second)

	checkMBOXFileStructure(t, dir, []string{
		"Archive.mbox",
		"Foo.mbox",
		"Inbox.mbox",
	})
}

func TestMBOXProviderGetMessageTargetsReturnsOnlyOneFolder(t *testing.T) {
	provider := newTestMBOXProvider("")

	folderA := Mailbox{Name: "Folder A", IsExclusive: true}
	folderB := Mailbox{Name: "Folder B", IsExclusive: true}
	labelA := Mailbox{Name: "Label A", IsExclusive: false}
	labelB := Mailbox{Name: "Label B", IsExclusive: false}
	labelC := Mailbox{Name: "Label C", IsExclusive: false}

	rule1 := &Rule{TargetMailboxes: []Mailbox{folderA, labelA, labelB}}
	rule2 := &Rule{TargetMailboxes: []Mailbox{folderB, labelC}}
	rule3 := &Rule{TargetMailboxes: []Mailbox{folderB}}

	tests := []struct {
		rules         []*Rule
		wantMailboxes []Mailbox
	}{
		{[]*Rule{}, []Mailbox{}},
		{[]*Rule{rule1}, []Mailbox{folderA, labelA, labelB}},
		{[]*Rule{rule1, rule2}, []Mailbox{folderA, labelA, labelB, labelC}},
		{[]*Rule{rule1, rule3}, []Mailbox{folderA, labelA, labelB}},
		{[]*Rule{rule3, rule1}, []Mailbox{folderB, labelA, labelB}},
	}
	for _, tc := range tests {
		tc := tc
		t.Run(fmt.Sprintf("%v", tc.rules), func(t *testing.T) {
			mailboxes := provider.getMessageTargets(tc.rules, "", []byte(""))
			r.Equal(t, tc.wantMailboxes, mailboxes)
		})
	}
}

func setupMBOXRules(rules transferRules) {
	_ = rules.setRule(Mailbox{Name: "All Mail"}, []Mailbox{{Name: "Archive"}}, 0, 0)
	_ = rules.setRule(Mailbox{Name: "Inbox"}, []Mailbox{{Name: "Inbox"}}, 0, 0)
	_ = rules.setRule(Mailbox{Name: "Foo"}, []Mailbox{{Name: "Foo"}}, 0, 0)
}

func checkMBOXFileStructure(t *testing.T, root string, expectedFiles []string) {
	files, err := getFilePathsWithSuffix(root, ".mbox")
	r.NoError(t, err)
	r.Equal(t, expectedFiles, files)
}
