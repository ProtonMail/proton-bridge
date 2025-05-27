// Copyright (c) 2025 Proton AG
//
// This file is part of Proton Mail Bridge.
//
// Proton Mail Bridge is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// Proton Mail Bridge is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with Proton Mail Bridge.  If not, see <https://www.gnu.org/licenses/>.

package imapservice

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/ProtonMail/gluon/db"
	"github.com/ProtonMail/gluon/imap"
	"github.com/ProtonMail/gluon/reporter"
	"github.com/ProtonMail/go-proton-api"
	"github.com/ProtonMail/proton-bridge/v3/internal/unleash"
	"github.com/sirupsen/logrus"
)

type GluonLabelNameProvider interface {
	GetUserMailboxByName(ctx context.Context, addrID string, labelName []string) (imap.MailboxData, error)
}

type gluonIDProvider interface {
	GetGluonID(addrID string) (string, bool)
}

type sentryReporter interface {
	ReportMessageWithContext(string, reporter.Context) error
}

type apiClient interface {
	GetLabel(ctx context.Context, labelID string, labelTypes ...proton.LabelType) (proton.Label, error)
}

type mailboxFetcherFn func(ctx context.Context, label proton.Label) (imap.MailboxData, error)

type LabelConflictManager struct {
	gluonLabelNameProvider GluonLabelNameProvider
	gluonIDProvider        gluonIDProvider
	client                 apiClient
	reporter               sentryReporter
	getFeatureFlagValueFn  unleash.GetFlagValueFn
}

func NewLabelConflictManager(
	gluonLabelNameProvider GluonLabelNameProvider,
	gluonIDProvider gluonIDProvider,
	client apiClient,
	reporter sentryReporter,
	getFeatureFlagValueFn unleash.GetFlagValueFn) *LabelConflictManager {
	return &LabelConflictManager{
		gluonLabelNameProvider: gluonLabelNameProvider,
		gluonIDProvider:        gluonIDProvider,
		client:                 client,
		reporter:               reporter,
		getFeatureFlagValueFn:  getFeatureFlagValueFn,
	}
}

func (m *LabelConflictManager) generateMailboxFetcher(connectors []*Connector) mailboxFetcherFn {
	return func(ctx context.Context, label proton.Label) (imap.MailboxData, error) {
		for _, updateCh := range connectors {
			addrID, ok := m.gluonIDProvider.GetGluonID(updateCh.addrID)
			if !ok {
				continue
			}
			return m.gluonLabelNameProvider.GetUserMailboxByName(ctx, addrID, GetMailboxName(label))
		}
		return imap.MailboxData{}, errors.New("no gluon connectors found")
	}
}

type LabelConflictResolver interface {
	ResolveConflict(ctx context.Context, label proton.Label, visited map[string]bool) (func() []imap.Update, error)
}
type labelConflictResolverImpl struct {
	mailboxFetch mailboxFetcherFn
	client       apiClient
	reporter     sentryReporter
}

type nullLabelConflictResolverImpl struct {
}

func (r *nullLabelConflictResolverImpl) ResolveConflict(_ context.Context, _ proton.Label, _ map[string]bool) (func() []imap.Update, error) {
	return func() []imap.Update {
		return []imap.Update{}
	}, nil
}

func (m *LabelConflictManager) NewConflictResolver(connectors []*Connector) LabelConflictResolver {
	if m.getFeatureFlagValueFn(unleash.LabelConflictResolverDisabled) {
		return &nullLabelConflictResolverImpl{}
	}

	return &labelConflictResolverImpl{
		mailboxFetch: m.generateMailboxFetcher(connectors),
		client:       m.client,
		reporter:     m.reporter,
	}
}

func (r *labelConflictResolverImpl) ResolveConflict(ctx context.Context, label proton.Label, visited map[string]bool) (func() []imap.Update, error) {
	var updateFns []func() []imap.Update

	// There's a cycle, such as in a label swap operation, we'll need to temporarily rename the label.
	// The change will be overwritten by one of the previous recursive calls.
	if visited[label.ID] {
		fn := func() []imap.Update {
			return []imap.Update{newMailboxUpdatedOrCreated(imap.MailboxID(label.ID), getMailboxNameWithTempPrefix(label))}
		}
		updateFns = append(updateFns, fn)
		return combineIMAPUpdateFns(updateFns), nil
	}
	visited[label.ID] = true

	// Fetch the gluon mailbox data and verify whether there are conflicts with the name.
	mailboxData, err := r.mailboxFetch(ctx, label)
	if err != nil {
		// Name is free, create the mailbox.
		if db.IsErrNotFound(err) {
			fn := func() []imap.Update {
				return []imap.Update{newMailboxUpdatedOrCreated(imap.MailboxID(label.ID), GetMailboxName(label))}
			}
			updateFns = append(updateFns, fn)
			return combineIMAPUpdateFns(updateFns), nil
		}
		return combineIMAPUpdateFns(updateFns), err
	}

	// Verify whether the label name corresponds to the same label ID. If true terminate, we don't need to update.
	if mailboxData.RemoteID == label.ID {
		return combineIMAPUpdateFns(updateFns), nil
	}

	// If the label name belongs to some other label ID. Fetch it's state from the remote.
	conflictingLabel, err := r.client.GetLabel(ctx, mailboxData.RemoteID, proton.LabelTypeFolder, proton.LabelTypeLabel)
	if err != nil {
		// If it's not present on the remote we should delete it. And create the new label.
		if errors.Is(err, proton.ErrNoSuchLabel) {
			fn := func() []imap.Update {
				return []imap.Update{
					imap.NewMailboxDeleted(imap.MailboxID(mailboxData.RemoteID)),
					newMailboxUpdatedOrCreated(imap.MailboxID(label.ID), GetMailboxName(label)),
				}
			}
			updateFns = append(updateFns, fn)
			return combineIMAPUpdateFns(updateFns), nil
		}
		return combineIMAPUpdateFns(updateFns), err
	}

	// Check if the conflicting label name has changed. If not, then this is a BE inconsistency.
	if compareLabelNames(GetMailboxName(conflictingLabel), mailboxData.BridgeName) {
		if err := r.reporter.ReportMessageWithContext("Unexpected label conflict", reporter.Context{
			"labelID":            label.ID,
			"conflictingLabelID": conflictingLabel.ID,
		}); err != nil {
			logrus.WithError(err).Error("Failed to report update error")
		}

		err := fmt.Errorf("unexpected label conflict: the name of label ID %s is already used by label ID %s", label.ID, conflictingLabel.ID)
		return combineIMAPUpdateFns(updateFns), err
	}

	// The name of the conflicting label has changed on the remote. We need to verify that the new name does not conflict with anything else.
	// Thus, a recursive check can be performed.
	childUpdateFns, err := r.ResolveConflict(ctx, conflictingLabel, visited)
	if err != nil {
		return combineIMAPUpdateFns(updateFns), err
	}
	updateFns = append(updateFns, childUpdateFns)

	fn := func() []imap.Update {
		return []imap.Update{newMailboxUpdatedOrCreated(imap.MailboxID(label.ID), GetMailboxName(label))}
	}
	updateFns = append(updateFns, fn)

	return combineIMAPUpdateFns(updateFns), nil
}

func combineIMAPUpdateFns(updateFunctions []func() []imap.Update) func() []imap.Update {
	return func() []imap.Update {
		var updates []imap.Update
		for _, fn := range updateFunctions {
			updates = append(updates, fn()...)
		}
		return updates
	}
}

func compareLabelNames(labelName1, labelName2 []string) bool {
	name1 := strings.Join(labelName1, "")
	name2 := strings.Join(labelName2, "")
	return name1 == name2
}
