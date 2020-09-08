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
	"testing"

	r "github.com/stretchr/testify/require"
)

func TestLeastUsedColor(t *testing.T) {
	var mailboxes []Mailbox
	// Unset mailboxes, should use first available color
	mailboxes = nil
	r.Equal(t, "#7272a7", LeastUsedColor(mailboxes))

	// No mailboxes at all, should use first available color
	mailboxes = []Mailbox{}
	r.Equal(t, "#7272a7", LeastUsedColor(mailboxes))

	// All colors have same frequency, should use first available color
	mailboxes = []Mailbox{
		{Name: "Mbox1", Color: "#7272a7"},
		{Name: "Mbox2", Color: "#cf5858"},
		{Name: "Mbox3", Color: "#c26cc7"},
		{Name: "Mbox4", Color: "#7569d1"},
		{Name: "Mbox5", Color: "#69a9d1"},
		{Name: "Mbox6", Color: "#5ec7b7"},
		{Name: "Mbox7", Color: "#72bb75"},
		{Name: "Mbox8", Color: "#c3d261"},
		{Name: "Mbox9", Color: "#e6c04c"},
		{Name: "Mbox10", Color: "#e6984c"},
		{Name: "Mbox11", Color: "#8989ac"},
		{Name: "Mbox12", Color: "#cf7e7e"},
		{Name: "Mbox13", Color: "#c793ca"},
		{Name: "Mbox14", Color: "#9b94d1"},
		{Name: "Mbox15", Color: "#a8c4d5"},
		{Name: "Mbox16", Color: "#97c9c1"},
		{Name: "Mbox17", Color: "#9db99f"},
		{Name: "Mbox18", Color: "#c6cd97"},
		{Name: "Mbox19", Color: "#e7d292"},
		{Name: "Mbox20", Color: "#dfb286"},
	}
	r.Equal(t, "#7272a7", LeastUsedColor(mailboxes))

	// First three colors already used, but others wasn't. Should use first non-used one.
	mailboxes = []Mailbox{
		{Name: "Mbox1", Color: "#7272a7"},
		{Name: "Mbox2", Color: "#cf5858"},
		{Name: "Mbox3", Color: "#c26cc7"},
	}
	r.Equal(t, "#7569d1", LeastUsedColor(mailboxes))
}

func TestFindMatchingMailboxes(t *testing.T) {
	mailboxes := []Mailbox{
		{Name: "Inbox", IsExclusive: true},
		{Name: "Sent", IsExclusive: true},
		{Name: "Archive", IsExclusive: true},
		{Name: "Foo", IsExclusive: false},
		{Name: "hello/world", IsExclusive: true},
		{Name: "Hello", IsExclusive: false},
		{Name: "WORLD", IsExclusive: true},
		{Name: "Trash", IsExclusive: true},
		{Name: "Drafts", IsExclusive: true},
	}

	tests := []struct {
		name      string
		wantNames []string
	}{
		{"inbox", []string{"Inbox"}},
		{"foo", []string{"Foo"}},
		{"hello", []string{"Hello"}},
		{"world", []string{"WORLD"}},
		{"hello/world", []string{"hello/world", "Hello"}},
		{"hello|world", []string{"WORLD", "Hello"}},
		{"nomailbox", []string{}},
		{"bin", []string{"Trash"}},
		{"root/bin", []string{"Trash"}},
		{"draft", []string{"Drafts"}},
		{"root/draft", []string{"Drafts"}},
	}
	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			mailbox := Mailbox{Name: tc.name}
			got := mailbox.findMatchingMailboxes(mailboxes)
			gotNames := []string{}
			for _, m := range got {
				gotNames = append(gotNames, m.Name)
			}
			r.Equal(t, tc.wantNames, gotNames)
		})
	}
}
