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

func TestFindMatchingMailboxes(t *testing.T) {
	mailboxes := []Mailbox{
		{Name: "Inbox", IsExclusive: true},
		{Name: "Sent", IsExclusive: true},
		{Name: "Archive", IsExclusive: true},
		{Name: "Foo", IsExclusive: false},
		{Name: "hello/world", IsExclusive: true},
		{Name: "Hello", IsExclusive: false},
		{Name: "WORLD", IsExclusive: true},
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
