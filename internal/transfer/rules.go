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
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/ProtonMail/proton-bridge/pkg/pmapi"
	"github.com/pkg/errors"
)

// transferRules maintains import rules, e.g. to which target mailbox should be
// source mailbox imported or what time spans.
type transferRules struct {
	filePath string

	// rules is map with key as hash of source mailbox to its rule.
	// Every source mailbox should have rule, at least disabled one.
	rules map[string]*Rule

	// globalMailbox is applied to every message in the import phase.
	// E.g., every message will be imported into this mailbox.
	globalMailbox *Mailbox

	// skipEncryptedMessages determines whether message which cannot
	// be decrypted should be exported or skipped.
	skipEncryptedMessages bool
}

// loadRules loads rules from `rulesPath` based on `ruleID`.
func loadRules(rulesPath, ruleID string) transferRules {
	fileName := fmt.Sprintf("rules_%s.json", ruleID)
	filePath := filepath.Join(rulesPath, fileName)

	var rules map[string]*Rule
	f, err := os.Open(filePath) //nolint[gosec]
	if err != nil {
		log.WithError(err).Debug("Problem to read rules")
	} else {
		defer f.Close() //nolint[errcheck]
		if err := json.NewDecoder(f).Decode(&rules); err != nil {
			log.WithError(err).Warn("Problem to umarshal rules")
		}
	}
	if rules == nil {
		rules = map[string]*Rule{}
	}

	return transferRules{
		filePath: filePath,
		rules:    rules,
	}
}

func (r *transferRules) setSkipEncryptedMessages(skip bool) {
	r.skipEncryptedMessages = skip
}

func (r *transferRules) setGlobalMailbox(mailbox *Mailbox) {
	r.globalMailbox = mailbox
}

func (r *transferRules) setGlobalTimeLimit(fromTime, toTime int64) {
	for _, rule := range r.rules {
		if !rule.HasTimeLimit() {
			rule.FromTime = fromTime
			rule.ToTime = toTime
		}
	}
}

func (r *transferRules) getRuleBySourceMailboxName(name string) (*Rule, error) {
	for _, rule := range r.rules {
		if rule.SourceMailbox.Name == name {
			return rule, nil
		}
	}
	return nil, fmt.Errorf("no rule for mailbox %s", name)
}

func (r *transferRules) iterateActiveRules() chan *Rule {
	ch := make(chan *Rule)
	go func() {
		for _, rule := range r.rules {
			if rule.Active {
				ch <- rule
			}
		}
		close(ch)
	}()
	return ch
}

// setDefaultRules iterates `sourceMailboxes` and sets missing rules with
// matching mailboxes from `targetMailboxes`. In case no matching mailbox
// is found, `defaultCallback` with a source mailbox as a parameter is used.
func (r *transferRules) setDefaultRules(sourceMailboxes []Mailbox, targetMailboxes []Mailbox, defaultCallback func(Mailbox) []Mailbox) {
	for _, sourceMailbox := range sourceMailboxes {
		h := sourceMailbox.Hash()
		if _, ok := r.rules[h]; ok {
			continue
		}

		targetMailboxes := sourceMailbox.findMatchingMailboxes(targetMailboxes)
		if len(targetMailboxes) == 0 {
			targetMailboxes = defaultCallback(sourceMailbox)
		}

		active := true
		if len(targetMailboxes) == 0 {
			active = false
		}

		// For both import to or export from ProtonMail, spam and draft
		// mailboxes are by default deactivated.
		for _, mailbox := range append([]Mailbox{sourceMailbox}, targetMailboxes...) {
			if mailbox.ID == pmapi.SpamLabel || mailbox.ID == pmapi.DraftLabel || mailbox.ID == pmapi.TrashLabel {
				active = false
				break
			}
		}

		r.rules[h] = &Rule{
			Active:          active,
			SourceMailbox:   sourceMailbox,
			TargetMailboxes: targetMailboxes,
		}
	}

	for _, rule := range r.rules {
		if !rule.Active {
			continue
		}
		found := false
		for _, sourceMailbox := range sourceMailboxes {
			if sourceMailbox.Name == rule.SourceMailbox.Name {
				found = true
			}
		}
		if !found {
			rule.Active = false
		}
	}

	r.save()
}

// setRule sets messages from `sourceMailbox` between `fromData` and `toDate`
// (if used) to be imported to all `targetMailboxes`.
func (r *transferRules) setRule(sourceMailbox Mailbox, targetMailboxes []Mailbox, fromTime, toTime int64) error {
	numberOfExclusiveMailboxes := 0
	for _, mailbox := range targetMailboxes {
		if mailbox.IsExclusive {
			numberOfExclusiveMailboxes++
		}
	}
	if numberOfExclusiveMailboxes > 1 {
		return errors.New("rule can have only one exclusive target mailbox")
	}

	h := sourceMailbox.Hash()
	r.rules[h] = &Rule{
		Active:          true,
		SourceMailbox:   sourceMailbox,
		TargetMailboxes: targetMailboxes,
		FromTime:        fromTime,
		ToTime:          toTime,
	}
	r.save()
	return nil
}

// unsetRule unsets messages from `sourceMailbox` to be exported.
func (r *transferRules) unsetRule(sourceMailbox Mailbox) {
	h := sourceMailbox.Hash()
	if rule, ok := r.rules[h]; ok {
		rule.Active = false
	} else {
		r.rules[h] = &Rule{
			Active:        false,
			SourceMailbox: sourceMailbox,
		}
	}
	r.save()
}

// getRule returns rule for `sourceMailbox` or nil if it does not exist.
func (r *transferRules) getRule(sourceMailbox Mailbox) *Rule {
	h := sourceMailbox.Hash()
	return r.rules[h]
}

// getRules returns all set rules.
func (r *transferRules) getRules() []*Rule {
	rules := []*Rule{}
	for _, rule := range r.rules {
		rules = append(rules, rule)
	}
	return rules
}

// reset wipes our all rules.
func (r *transferRules) reset() {
	r.rules = map[string]*Rule{}
	r.save()
}

// save saves rules to file.
func (r *transferRules) save() {
	f, err := os.Create(r.filePath)
	if err != nil {
		log.WithError(err).Warn("Problem to write rules")
		return
	}
	defer f.Close() //nolint[errcheck]

	if err := json.NewEncoder(f).Encode(r.rules); err != nil {
		log.WithError(err).Warn("Problem to marshal rules")
	}
}

// Rule is data holder of rule for one source mailbox used by `transferRules`.
type Rule struct {
	Active          bool      `json:"active"`
	SourceMailbox   Mailbox   `json:"source"`
	TargetMailboxes []Mailbox `json:"targets"`
	FromTime        int64     `json:"from"`
	ToTime          int64     `json:"to"`
}

// String returns textual representation for log purposes.
func (r *Rule) String() string {
	return fmt.Sprintf(
		"%s -> %s (%d - %d)",
		r.SourceMailbox.Name,
		strings.Join(r.TargetMailboxNames(), ", "),
		r.FromTime,
		r.ToTime,
	)
}

func (r *Rule) isTimeInRange(t int64) bool {
	if !r.HasTimeLimit() {
		return true
	}
	return r.FromTime <= t && t <= r.ToTime
}

// HasTimeLimit returns whether rule defines time limit.
func (r *Rule) HasTimeLimit() bool {
	return r.FromTime != 0 || r.ToTime != 0
}

// FromDate returns time struct based on `FromTime`.
func (r *Rule) FromDate() time.Time {
	return time.Unix(r.FromTime, 0)
}

// ToDate returns time struct based on `ToTime`.
func (r *Rule) ToDate() time.Time {
	return time.Unix(r.ToTime, 0)
}

// TargetMailboxNames returns array of target mailbox names.
func (r *Rule) TargetMailboxNames() (names []string) {
	for _, mailbox := range r.TargetMailboxes {
		names = append(names, mailbox.Name)
	}
	return
}
