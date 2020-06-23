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

// Package transfer provides tools to export messages from one provider and
// import them to another provider. Provider can be EML, MBOX, IMAP or PMAPI.
package transfer

import (
	"crypto/sha256"
	"fmt"

	"github.com/sirupsen/logrus"
)

var log = logrus.WithField("pkg", "transfer") //nolint[gochecknoglobals]

// Transfer is facade on top of import rules, progress manager and source
// and target providers. This is the main object which should be used.
type Transfer struct {
	panicHandler    PanicHandler
	id              string
	dir             string
	rules           transferRules
	source          SourceProvider
	target          TargetProvider
	sourceMboxCache []Mailbox
	targetMboxCache []Mailbox
}

// New creates Transfer for specific source and target. Usage:
//
//   source := transfer.NewEMLProvider(...)
//   target := transfer.NewPMAPIProvider(...)
//   transfer.New(source, target, ...)
func New(panicHandler PanicHandler, transferDir string, source SourceProvider, target TargetProvider) (*Transfer, error) {
	transferID := fmt.Sprintf("%x", sha256.Sum256([]byte(source.ID()+"-"+target.ID())))
	rules := loadRules(transferDir, transferID)
	transfer := &Transfer{
		panicHandler: panicHandler,
		id:           transferID,
		dir:          transferDir,
		rules:        rules,
		source:       source,
		target:       target,
	}
	if err := transfer.setDefaultRules(); err != nil {
		return nil, err
	}
	return transfer, nil
}

// SetDefaultRules sets missing rules for source mailboxes with matching
// target mailboxes. In case no matching mailbox is found, `defaultCallback`
// with a source mailbox as a parameter is used.
func (t *Transfer) setDefaultRules() error {
	sourceMailboxes, err := t.SourceMailboxes()
	if err != nil {
		return err
	}

	targetMailboxes, err := t.TargetMailboxes()
	if err != nil {
		return err
	}

	defaultCallback := func(sourceMailbox Mailbox) []Mailbox {
		return t.target.DefaultMailboxes(sourceMailbox)
	}

	t.rules.setDefaultRules(sourceMailboxes, targetMailboxes, defaultCallback)
	return nil
}

// SetSkipEncryptedMessages sets whether message which cannot be decrypted
// should be exported or skipped.
func (t *Transfer) SetSkipEncryptedMessages(skip bool) {
	t.rules.setSkipEncryptedMessages(skip)
}

// SetGlobalMailbox sets mailbox that is applied to every message in
// the import phase.
func (t *Transfer) SetGlobalMailbox(mailbox *Mailbox) {
	t.rules.setGlobalMailbox(mailbox)
}

// SetGlobalTimeLimit sets time limit that is applied to rules without any
// specified time limit.
func (t *Transfer) SetGlobalTimeLimit(fromTime, toTime int64) {
	t.rules.setGlobalTimeLimit(fromTime, toTime)
}

// SetRule sets sourceMailbox for transfer.
func (t *Transfer) SetRule(sourceMailbox Mailbox, targetMailboxes []Mailbox, fromTime, toTime int64) error {
	return t.rules.setRule(sourceMailbox, targetMailboxes, fromTime, toTime)
}

// UnsetRule unsets sourceMailbox from transfer.
func (t *Transfer) UnsetRule(sourceMailbox Mailbox) {
	t.rules.unsetRule(sourceMailbox)
}

// ResetRules unsets all rules.
func (t *Transfer) ResetRules() {
	t.rules.reset()
}

// GetRule returns rule for given mailbox.
func (t *Transfer) GetRule(sourceMailbox Mailbox) *Rule {
	return t.rules.getRule(sourceMailbox)
}

// GetRules returns all set transfer rules.
func (t *Transfer) GetRules() []*Rule {
	return t.rules.getRules()
}

// SourceMailboxes returns mailboxes available at source side.
func (t *Transfer) SourceMailboxes() (m []Mailbox, err error) {
	if t.sourceMboxCache == nil {
		t.sourceMboxCache, err = t.source.Mailboxes(false, true)
	}
	return t.sourceMboxCache, err
}

// TargetMailboxes returns mailboxes available at target side.
func (t *Transfer) TargetMailboxes() (m []Mailbox, err error) {
	if t.targetMboxCache == nil {
		t.targetMboxCache, err = t.target.Mailboxes(true, false)
	}
	return t.targetMboxCache, err
}

// CreateTargetMailbox creates mailbox in target provider.
func (t *Transfer) CreateTargetMailbox(mailbox Mailbox) (Mailbox, error) {
	t.targetMboxCache = nil

	return t.target.CreateMailbox(mailbox)
}

// ChangeTarget changes the target. It is safe to change target for export,
// must not be changed for import. Do not set after you started transfer.
func (t *Transfer) ChangeTarget(target TargetProvider) {
	t.targetMboxCache = nil

	t.target = target
}

// Start starts the transfer from source to target.
func (t *Transfer) Start() *Progress {
	log.Debug("Transfer started")
	t.rules.save()
	t.rules.propagateGlobalTime()

	log := log.WithField("id", t.id)
	reportFile := newFileReport(t.dir, t.id)
	progress := newProgress(log, reportFile)

	ch := make(chan Message)

	go func() {
		defer t.panicHandler.HandlePanic()

		progress.start()
		t.source.TransferTo(t.rules, &progress, ch)
		close(ch)
	}()

	go func() {
		defer t.panicHandler.HandlePanic()

		t.target.TransferFrom(t.rules, &progress, ch)
		progress.finish()
	}()

	return &progress
}
