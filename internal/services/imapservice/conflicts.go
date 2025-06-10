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
	"github.com/ProtonMail/proton-bridge/v3/pkg/algo"
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
	ReportWarningWithContext(string, reporter.Context) error
}

type apiClient interface {
	GetLabel(ctx context.Context, labelID string, labelTypes ...proton.LabelType) (proton.Label, error)
}

type mailboxFetcherFn func(ctx context.Context, label proton.Label) (imap.MailboxData, error)

type mailboxMessageCountFetcherFn func(ctx context.Context, internalMailboxID imap.InternalMailboxID) (int, error)

type LabelConflictManager struct {
	gluonLabelNameProvider GluonLabelNameProvider
	gluonIDProvider        gluonIDProvider
	client                 apiClient
	reporter               sentryReporter
	featureFlagProvider    unleash.FeatureFlagValueProvider
}

func NewLabelConflictManager(
	gluonLabelNameProvider GluonLabelNameProvider,
	gluonIDProvider gluonIDProvider,
	client apiClient,
	reporter sentryReporter,
	featureFlagProvider unleash.FeatureFlagValueProvider) *LabelConflictManager {
	return &LabelConflictManager{
		gluonLabelNameProvider: gluonLabelNameProvider,
		gluonIDProvider:        gluonIDProvider,
		client:                 client,
		reporter:               reporter,
		featureFlagProvider:    featureFlagProvider,
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

func (m *LabelConflictManager) generateMailboxMessageCountFetcher(connectors []*Connector) mailboxMessageCountFetcherFn {
	return func(ctx context.Context, id imap.InternalMailboxID) (int, error) {
		var countSum int
		var errs []error
		for _, conn := range connectors {
			count, err := conn.GetMailboxMessageCount(ctx, id)
			countSum += count
			errs = append(errs, err)
		}

		return countSum, errors.Join(errs...)
	}
}

type LabelConflictResolver interface {
	ResolveConflict(ctx context.Context, label proton.Label, visited map[string]bool) (func() []imap.Update, error)
}
type labelConflictResolverImpl struct {
	mailboxFetch mailboxFetcherFn
	client       apiClient
	reporter     sentryReporter
	log          *logrus.Entry
}

type nullLabelConflictResolverImpl struct {
}

func (r *nullLabelConflictResolverImpl) ResolveConflict(_ context.Context, _ proton.Label, _ map[string]bool) (func() []imap.Update, error) {
	return func() []imap.Update {
		return []imap.Update{}
	}, nil
}

func (m *LabelConflictManager) NewConflictResolver(connectors []*Connector) LabelConflictResolver {
	if m.featureFlagProvider.GetFlagValue(unleash.LabelConflictResolverDisabled) {
		return &nullLabelConflictResolverImpl{}
	}

	return &labelConflictResolverImpl{
		mailboxFetch: m.generateMailboxFetcher(connectors),
		client:       m.client,
		reporter:     m.reporter,
		log: logrus.WithFields(logrus.Fields{
			"pkg":                "imapservice/labelConflictResolver",
			"numberOfConnectors": len(connectors),
		}),
	}
}

func (r *labelConflictResolverImpl) ResolveConflict(ctx context.Context, label proton.Label, visited map[string]bool) (func() []imap.Update, error) {
	logger := r.log.WithFields(logrus.Fields{
		"labelID":   label.ID,
		"labelPath": hashLabelPaths(GetMailboxName(label)),
	})

	// For system type labels we shouldn't care.
	var updateFns []func() []imap.Update

	// There's a cycle, such as in a label swap operation, we'll need to temporarily rename the label.
	// The change will be overwritten by one of the previous recursive calls.
	if visited[label.ID] {
		logrus.Info("Cycle detected, applying temporary rename")
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
			logger.Info("Label not found in DB, creating mailbox.")
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
		logger.Info("Mailbox name matches label ID, no conflict.")
		return combineIMAPUpdateFns(updateFns), nil
	}

	// This means we've found a conflict. So let's log it.
	logger = logger.WithFields(logrus.Fields{
		"conflictingLabelID":   mailboxData.RemoteID,
		"conflictingLabelPath": hashLabelPaths(mailboxData.BridgeName),
	})
	logger.Info("Label conflict found")

	// If the label name belongs to some other label ID. Fetch it's state from the remote.
	conflictingLabel, err := r.client.GetLabel(ctx, mailboxData.RemoteID, proton.LabelTypeFolder, proton.LabelTypeLabel, proton.LabelTypeSystem)
	if err != nil {
		// If it's not present on the remote we should delete it. And create the new label.
		if errors.Is(err, proton.ErrNoSuchLabel) {
			logger.Info("Conflicting label does not exist on remote. Deleting.")
			fn := func() []imap.Update {
				return []imap.Update{
					imap.NewMailboxDeleted(imap.MailboxID(mailboxData.RemoteID)), // Should this be with remote ID
					newMailboxUpdatedOrCreated(imap.MailboxID(label.ID), GetMailboxName(label)),
				}
			}
			updateFns = append(updateFns, fn)
			return combineIMAPUpdateFns(updateFns), nil
		}
		logger.WithError(err).Error("Failed to fetch conflicting label from remote.")
		return combineIMAPUpdateFns(updateFns), err
	}

	// Check if the conflicting label name has changed. If not, then this is a BE inconsistency.
	if compareLabelNames(GetMailboxName(conflictingLabel), mailboxData.BridgeName) {
		if err := r.reporter.ReportMessageWithContext("Unexpected label conflict", reporter.Context{
			"labelID":            label.ID,
			"conflictingLabelID": conflictingLabel.ID,
		}); err != nil {
			logger.WithError(err).Error("Failed to report update error")
		}

		err := fmt.Errorf("unexpected label conflict: the name of label ID %s is already used by label ID %s", label.ID, conflictingLabel.ID)
		return combineIMAPUpdateFns(updateFns), err
	}

	// The name of the conflicting label has changed on the remote. We need to verify that the new name does not conflict with anything else.
	// Thus, a recursive check can be performed.
	logger.WithField("conflictingLabelNewPath", hashLabelPaths(conflictingLabel.Path)).
		Info("Conflicting label name has changed. Recursively resolving conflict.")
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

func hashLabelPaths(path []string) string {
	return algo.HashBase64SHA256(strings.Join(path, ""))
}

type InternalLabelConflictResolver interface {
	ResolveConflict(ctx context.Context, apiLabels map[string]proton.Label) (func() []imap.Update, error)
}

type internalLabelConflictResolverImpl struct {
	mailboxFetch                 mailboxFetcherFn
	mailboxMessageCountFetch     mailboxMessageCountFetcherFn
	userLabelConflictResolver    LabelConflictResolver
	allowNonEmptyMailboxDeletion bool
	client                       apiClient
	reporter                     sentryReporter
	log                          *logrus.Entry
}

type nullInternalLabelConflictResolver struct{}

func (r *nullInternalLabelConflictResolver) ResolveConflict(_ context.Context, _ map[string]proton.Label) (func() []imap.Update, error) {
	return func() []imap.Update { return []imap.Update{} }, nil
}

func (m *LabelConflictManager) NewInternalLabelConflictResolver(connectors []*Connector) InternalLabelConflictResolver {
	if m.featureFlagProvider.GetFlagValue(unleash.InternalLabelConflictResolverDisabled) {
		return &nullInternalLabelConflictResolver{}
	}

	return &internalLabelConflictResolverImpl{
		mailboxFetch:                 m.generateMailboxFetcher(connectors),
		mailboxMessageCountFetch:     m.generateMailboxMessageCountFetcher(connectors),
		userLabelConflictResolver:    m.NewConflictResolver(connectors),
		allowNonEmptyMailboxDeletion: m.featureFlagProvider.GetFlagValue(unleash.ItnternalLabelConflictNonEmptyMailboxDeletion),
		client:                       m.client,
		reporter:                     m.reporter,
		log: logrus.WithFields(logrus.Fields{
			"pkg":                "imapservice/internalLabelConflictResolver",
			"numberOfConnectors": len(connectors),
		}),
	}
}

func (r *internalLabelConflictResolverImpl) ResolveConflict(ctx context.Context, apiLabels map[string]proton.Label) (func() []imap.Update, error) {
	updateFns := []func() []imap.Update{}

	for _, prefix := range []string{folderPrefix, labelPrefix} {
		internalLabel := proton.Label{
			Path: []string{prefix},
			ID:   prefix,
			Name: prefix,
		}

		mbox, err := r.mailboxFetch(ctx, internalLabel)
		if err != nil {
			if db.IsErrNotFound(err) {
				continue
			}
			return nil, err
		}

		// If the ID's match then we don't have a discrepancy.
		if mbox.RemoteID == internalLabel.ID {
			continue
		}

		logFields := logrus.Fields{
			"internalLabelID":      internalLabel.ID,
			"internalLabelName":    internalLabel.Name,
			"conflictingLabelID":   mbox.RemoteID,
			"conflictingLabelName": strings.Join(mbox.BridgeName, "/"),
		}
		reporterContext := reporter.Context(logFields)
		logger := r.log.WithFields(logFields)
		logger.Info("Encountered conflict, resolving.")

		// There is a discrepancy, let's see if it comes from API.
		apiLabel, ok := apiLabels[mbox.RemoteID]
		if !ok {
			// Label does not come from API, we should delete it.
			// Due diligence, check if there are any messages associated with the mailbox.
			msgCount, _ := r.mailboxMessageCountFetch(ctx, mbox.InternalID)
			if msgCount != 0 {
				logger.WithField("conflictingLabelMessageCount", msgCount).Info("Non-API conflicting label has associated messages")

				reporterContext["conflictingLabelMessageCount"] = msgCount
				if rerr := r.reporter.ReportWarningWithContext("Internal mailbox name conflict. Conflicting non-API label has messages.",
					reporterContext); rerr != nil {
					logger.WithError(rerr).Error("Failed to send report to sentry")
				}

				if !r.allowNonEmptyMailboxDeletion {
					return combineIMAPUpdateFns(updateFns), fmt.Errorf("internal mailbox conflicting non-api label has associated messages")
				}
			}

			fn := func() []imap.Update {
				return []imap.Update{imap.NewMailboxDeletedSilent(imap.MailboxID(mbox.RemoteID))}
			}
			updateFns = append(updateFns, fn)
			continue
		}

		reporterContext["conflictingLabelType"] = apiLabel.Type

		// Label is indeed from API let's see if it's name has changed.
		if compareLabelNames(GetMailboxName(apiLabel), internalLabel.Path) {
			logger.Error("Conflict, same-name mailbox is returned by API")

			if err := r.reporter.ReportMessageWithContext("Internal mailbox name conflict. Same-name mailbox is returned by API", reporterContext); err != nil {
				logger.WithError(err).Error("Could not send report to sentry")
			}

			return combineIMAPUpdateFns(updateFns), fmt.Errorf("API label %s conflicts with internal label %s",
				GetMailboxName(apiLabel),
				strings.Join(mbox.BridgeName, "/"),
			)
		}

		// If it's name has changed then we ought to rename it while still taking care of potential conflicts.
		labelRenameUpdates, err := r.userLabelConflictResolver.ResolveConflict(ctx, apiLabel, make(map[string]bool))
		if err != nil {
			reporterContext["err"] = err.Error()
			if rerr := r.reporter.ReportMessageWithContext("Failed to resolve internal mailbox conflict", reporterContext); rerr != nil {
				logger.WithError(rerr).Error("Could not send report to sentry")
			}
			return combineIMAPUpdateFns(updateFns),
				fmt.Errorf("failed to resolve user label conflict for '%s': %w", apiLabel.Name, err)
		}
		updateFns = append(updateFns, labelRenameUpdates)
	}
	return combineIMAPUpdateFns(updateFns), nil
}
