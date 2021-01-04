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
	"io/ioutil"
	"os"
	"testing"

	"github.com/ProtonMail/proton-bridge/pkg/pmapi"
	r "github.com/stretchr/testify/require"
)

func newTestRules(t *testing.T) (transferRules, func()) {
	path, err := ioutil.TempDir("", "rules")
	r.NoError(t, err)

	ruleID := "rule"
	rules := loadRules(path, ruleID)
	return rules, func() {
		_ = os.RemoveAll(path)
	}
}

func TestLoadRules(t *testing.T) {
	path, err := ioutil.TempDir("", "rules")
	r.NoError(t, err)
	defer os.RemoveAll(path) //nolint[errcheck]

	ruleID := "rule"
	rules := loadRules(path, ruleID)

	mailboxA := Mailbox{ID: "1", Name: "One", Color: "orange", IsExclusive: true}
	mailboxB := Mailbox{ID: "2", Name: "Two", Color: "", IsExclusive: true}
	mailboxC := Mailbox{ID: "3", Name: "Three", Color: "", IsExclusive: false}

	r.NoError(t, rules.setRule(mailboxA, []Mailbox{mailboxB, mailboxC}, 0, 0))
	r.NoError(t, rules.setRule(mailboxB, []Mailbox{mailboxB}, 10, 20))
	r.NoError(t, rules.setRule(mailboxC, []Mailbox{}, 0, 30))

	rules2 := loadRules(path, ruleID)
	r.Equal(t, map[string]*Rule{
		mailboxA.Hash(): {Active: true, SourceMailbox: mailboxA, TargetMailboxes: []Mailbox{mailboxB, mailboxC}, FromTime: 0, ToTime: 0},
		mailboxB.Hash(): {Active: true, SourceMailbox: mailboxB, TargetMailboxes: []Mailbox{mailboxB}, FromTime: 10, ToTime: 20},
		mailboxC.Hash(): {Active: true, SourceMailbox: mailboxC, TargetMailboxes: []Mailbox{}, FromTime: 0, ToTime: 30},
	}, rules2.rules)

	rules2.unsetRule(mailboxA)
	rules2.unsetRule(mailboxC)

	rules3 := loadRules(path, ruleID)
	r.Equal(t, map[string]*Rule{
		mailboxA.Hash(): {Active: false, SourceMailbox: mailboxA, TargetMailboxes: []Mailbox{mailboxB, mailboxC}, FromTime: 0, ToTime: 0},
		mailboxB.Hash(): {Active: true, SourceMailbox: mailboxB, TargetMailboxes: []Mailbox{mailboxB}, FromTime: 10, ToTime: 20},
		mailboxC.Hash(): {Active: false, SourceMailbox: mailboxC, TargetMailboxes: []Mailbox{}, FromTime: 0, ToTime: 30},
	}, rules3.rules)
}

func TestSetGlobalTimeLimit(t *testing.T) {
	path, err := ioutil.TempDir("", "rules")
	r.NoError(t, err)
	defer os.RemoveAll(path) //nolint[errcheck]

	rules := loadRules(path, "rule")

	mailboxA := Mailbox{Name: "One"}
	mailboxB := Mailbox{Name: "Two"}

	r.NoError(t, rules.setRule(mailboxA, []Mailbox{}, 10, 20))
	r.NoError(t, rules.setRule(mailboxB, []Mailbox{}, 0, 0))

	rules.setGlobalTimeLimit(30, 40)
	rules.propagateGlobalTime()

	r.Equal(t, map[string]*Rule{
		mailboxA.Hash(): {Active: true, SourceMailbox: mailboxA, TargetMailboxes: []Mailbox{}, FromTime: 10, ToTime: 20},
		mailboxB.Hash(): {Active: true, SourceMailbox: mailboxB, TargetMailboxes: []Mailbox{}, FromTime: 30, ToTime: 40},
	}, rules.rules)
}

func TestSetDefaultRules(t *testing.T) {
	path, err := ioutil.TempDir("", "rules")
	r.NoError(t, err)
	defer os.RemoveAll(path) //nolint[errcheck]

	rules := loadRules(path, "rule")

	mailbox1 := Mailbox{Name: "One"}                       // Set manually, default will not override it.
	mailbox2 := Mailbox{Name: "Two"}                       // Matched by `targetMailboxes`.
	mailbox3 := Mailbox{Name: "Three"}                     // Matched by `defaultCallback`, not included in `targetMailboxes`.
	mailbox4 := Mailbox{Name: "Four"}                      // Matched by nothing, will not be active.
	mailbox5 := Mailbox{Name: "Spam", ID: pmapi.SpamLabel} // Spam is inactive by default (ID found in source).
	mailbox6a := Mailbox{Name: "Draft"}                    // Draft is inactive by default (ID found in target, mailbox6b).
	mailbox6b := Mailbox{Name: "Draft", ID: pmapi.DraftLabel}

	sourceMailboxes := []Mailbox{mailbox1, mailbox2, mailbox3, mailbox4, mailbox5, mailbox6a}
	targetMailboxes := []Mailbox{mailbox1, mailbox2, mailbox6b}

	r.NoError(t, rules.setRule(mailbox1, []Mailbox{mailbox3}, 0, 0))

	defaultCallback := func(mailbox Mailbox) []Mailbox {
		if mailbox.Name == "Three" {
			return []Mailbox{mailbox3}
		}
		return []Mailbox{}
	}

	rules.setDefaultRules(sourceMailboxes, targetMailboxes, defaultCallback)

	r.Equal(t, map[string]*Rule{
		mailbox1.Hash():  {Active: true, SourceMailbox: mailbox1, TargetMailboxes: []Mailbox{mailbox3}},
		mailbox2.Hash():  {Active: true, SourceMailbox: mailbox2, TargetMailboxes: []Mailbox{mailbox2}},
		mailbox3.Hash():  {Active: true, SourceMailbox: mailbox3, TargetMailboxes: []Mailbox{mailbox3}},
		mailbox4.Hash():  {Active: false, SourceMailbox: mailbox4, TargetMailboxes: []Mailbox{}},
		mailbox5.Hash():  {Active: false, SourceMailbox: mailbox5, TargetMailboxes: []Mailbox{}},
		mailbox6a.Hash(): {Active: false, SourceMailbox: mailbox6a, TargetMailboxes: []Mailbox{mailbox6b}},
	}, rules.rules)
}

func TestSetDefaultRulesDeactivateMissing(t *testing.T) {
	path, err := ioutil.TempDir("", "rules")
	r.NoError(t, err)
	defer os.RemoveAll(path) //nolint[errcheck]

	rules := loadRules(path, "rule")

	mailboxA := Mailbox{ID: "1", Name: "One", Color: "", IsExclusive: true}
	mailboxB := Mailbox{ID: "2", Name: "Two", Color: "", IsExclusive: true}

	r.NoError(t, rules.setRule(mailboxA, []Mailbox{mailboxB}, 0, 0))
	r.NoError(t, rules.setRule(mailboxB, []Mailbox{mailboxB}, 0, 0))

	sourceMailboxes := []Mailbox{mailboxA}
	targetMailboxes := []Mailbox{mailboxA, mailboxB}
	defaultCallback := func(mailbox Mailbox) (mailboxes []Mailbox) {
		return
	}
	rules.setDefaultRules(sourceMailboxes, targetMailboxes, defaultCallback)

	r.Equal(t, map[string]*Rule{
		mailboxA.Hash(): {Active: true, SourceMailbox: mailboxA, TargetMailboxes: []Mailbox{mailboxB}, FromTime: 0, ToTime: 0},
	}, rules.rules)
}

func TestIsTimeInRange(t *testing.T) {
	tests := []struct {
		rule Rule
		time int64
		want bool
	}{
		{generateTimeRule(0, 0), 0, true},
		{generateTimeRule(0, 0), 10, true},
		{generateTimeRule(0, 15), 10, true},
		{generateTimeRule(5, 15), 10, true},
		{generateTimeRule(0, 5), 10, false},
		{generateTimeRule(5, 7), 10, false},
		{generateTimeRule(15, 30), 10, false},
		{generateTimeRule(15, 0), 10, false},
	}
	for _, tc := range tests {
		tc := tc
		t.Run(fmt.Sprintf("%v / %d", tc.rule, tc.time), func(t *testing.T) {
			got := tc.rule.isTimeInRange(tc.time)
			r.Equal(t, tc.want, got)
		})
	}
}

func TestHasTimeLimit(t *testing.T) {
	tests := []struct {
		rule Rule
		want bool
	}{
		{generateTimeRule(0, 0), false},
		{generateTimeRule(0, 1), true},
		{generateTimeRule(1, 2), true},
		{generateTimeRule(1, 0), true},
	}
	for _, tc := range tests {
		tc := tc
		t.Run(fmt.Sprintf("%v", tc.rule), func(t *testing.T) {
			r.Equal(t, tc.want, tc.rule.HasTimeLimit())
		})
	}
}

func generateTimeRule(from, to int64) Rule {
	return Rule{
		SourceMailbox:   Mailbox{},
		TargetMailboxes: []Mailbox{},
		FromTime:        from,
		ToTime:          to,
	}
}

func TestOrderRules(t *testing.T) {
	wantMailboxOrder := []Mailbox{
		{Name: "Inbox", IsExclusive: true},
		{Name: "Drafts", IsExclusive: true},
		{Name: "Sent", IsExclusive: true},
		{Name: "Starred", IsExclusive: true},
		{Name: "Archive", IsExclusive: true},
		{Name: "Spam", IsExclusive: true},
		{Name: "All Mail", IsExclusive: true},
		{Name: "Folder A", IsExclusive: true},
		{Name: "Folder B", IsExclusive: true},
		{Name: "Folder C", IsExclusive: true},
		{Name: "Label A", IsExclusive: false},
		{Name: "Label B", IsExclusive: false},
		{Name: "Label C", IsExclusive: false},
	}
	wantMailboxNames := []string{}

	rules := map[string]*Rule{}
	for _, mailbox := range wantMailboxOrder {
		wantMailboxNames = append(wantMailboxNames, mailbox.Name)
		rules[mailbox.Hash()] = &Rule{
			SourceMailbox: mailbox,
		}
	}
	transferRules := transferRules{
		rules: rules,
	}

	gotMailboxNames := []string{}
	for _, rule := range transferRules.getSortedRules() {
		gotMailboxNames = append(gotMailboxNames, rule.SourceMailbox.Name)
	}

	r.Equal(t, wantMailboxNames, gotMailboxNames)
}
