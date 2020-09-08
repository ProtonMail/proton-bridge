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
	"testing"

	r "github.com/stretchr/testify/require"
)

func newTestLocalProvider(path string) *LocalProvider {
	if path == "" {
		path = "testdata/emlmbox"
	}
	return NewLocalProvider(path)
}

func TestLocalProviderMailboxes(t *testing.T) {
	provider := newTestLocalProvider("")

	tests := []struct {
		includeEmpty  bool
		wantMailboxes []Mailbox
	}{
		{true, []Mailbox{
			{Name: "Foo"},
			{Name: "emlmbox"},
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

func TestLocalProviderTransferTo(t *testing.T) {
	provider := newTestLocalProvider("")

	rules, rulesClose := newTestRules(t)
	defer rulesClose()
	setupEMLMBOXRules(rules)

	testTransferTo(t, rules, provider, []string{
		"Foo/msg.eml",
		"Inbox.mbox:1",
	})
}

func setupEMLMBOXRules(rules transferRules) {
	_ = rules.setRule(Mailbox{Name: "Inbox"}, []Mailbox{{Name: "Inbox"}}, 0, 0)
	_ = rules.setRule(Mailbox{Name: "Foo"}, []Mailbox{{Name: "Foo"}}, 0, 0)
}
