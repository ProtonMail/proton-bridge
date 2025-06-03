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
	"github.com/ProtonMail/proton-bridge/v3/pkg/algo"
	"github.com/sirupsen/logrus"
)

type labelDiscrepancyType int

const (
	discrepancyInternal labelDiscrepancyType = iota
	discrepancySystem
	discrepancyUser
)

func (t labelDiscrepancyType) String() string {
	switch t {
	case discrepancyInternal:
		return "internal"
	case discrepancySystem:
		return "system"
	case discrepancyUser:
		return "user"
	default:
		return "unknown"
	}
}

type labelDiscrepancy struct {
	labelName            string
	labelPath            string
	labelID              string
	conflictingLabelName string
	conflictingLabelID   string
	Type                 labelDiscrepancyType
}

func joinStrings(input []string) string {
	return strings.Join(input, "/")
}

func newLabelDiscrepancy(label proton.Label, mbox imap.MailboxData, dType labelDiscrepancyType) labelDiscrepancy {
	discrepancy := labelDiscrepancy{
		labelName:          label.Name,
		labelID:            label.ID,
		conflictingLabelID: mbox.RemoteID,
		Type:               dType,
	}

	if dType == discrepancyUser {
		discrepancy.labelName = algo.HashBase64SHA256(label.Name)
		discrepancy.labelPath = algo.HashBase64SHA256(joinStrings(label.Path))
		discrepancy.conflictingLabelName = algo.HashBase64SHA256(joinStrings(mbox.BridgeName))
	} else {
		discrepancy.labelName = label.Name
		discrepancy.labelPath = joinStrings(label.Path)
		discrepancy.conflictingLabelName = joinStrings(mbox.BridgeName)
	}

	return discrepancy
}

func discrepanciesToContext(discrepancies []labelDiscrepancy) reporter.Context {
	ctx := make(reporter.Context)

	for i, d := range discrepancies {
		prefix := fmt.Sprintf("discrepancy_%d_", i)

		ctx[prefix+"type"] = d.Type.String()
		ctx[prefix+"label_name"] = d.labelName
		ctx[prefix+"label_path"] = d.labelPath
		ctx[prefix+"label_id"] = d.labelID
		ctx[prefix+"conflicting_label_name"] = d.conflictingLabelName
		ctx[prefix+"conflicting_label_id"] = d.conflictingLabelID
	}

	ctx["discrepancy_count"] = len(discrepancies)
	return ctx
}

type ConnectorGetter interface {
	getConnectors() []*Connector
}

type LabelConflictChecker struct {
	gluonLabelNameProvider GluonLabelNameProvider
	gluonIDProvider        gluonIDProvider
	connectorGetter        ConnectorGetter
	reporter               reporter.Reporter
	logger                 *logrus.Entry
}

func NewConflictChecker(connectorGetter ConnectorGetter, reporter reporter.Reporter, provider gluonIDProvider, nameProvider GluonLabelNameProvider) *LabelConflictChecker {
	return &LabelConflictChecker{
		gluonLabelNameProvider: nameProvider,
		gluonIDProvider:        provider,
		connectorGetter:        connectorGetter,
		reporter:               reporter,
		logger: logrus.WithFields(logrus.Fields{
			"pkg": "imapservice/labelConflictChecker",
		}),
	}
}

func (c *LabelConflictChecker) getFn() mailboxFetcherFn {
	connectors := c.connectorGetter.getConnectors()

	return func(ctx context.Context, label proton.Label) (imap.MailboxData, error) {
		for _, updateCh := range connectors {
			addrID, ok := c.gluonIDProvider.GetGluonID(updateCh.addrID)
			if !ok {
				continue
			}
			return c.gluonLabelNameProvider.GetUserMailboxByName(ctx, addrID, GetMailboxName(label))
		}
		return imap.MailboxData{}, errors.New("no gluon connectors found")
	}
}

func (c *LabelConflictChecker) CheckAndReportConflicts(ctx context.Context, labels map[string]proton.Label) error {
	labelDiscrepancies, err := c.checkConflicts(ctx, labels, c.getFn())
	if err != nil {
		return err
	}

	if len(labelDiscrepancies) == 0 {
		return nil
	}

	reporterCtx := discrepanciesToContext(labelDiscrepancies)
	if err := c.reporter.ReportMessageWithContext("Found label conflicts on Bridge start", reporterCtx); err != nil {
		c.logger.WithError(err).Error("Failed to report label conflicts to Sentry")
	}

	return nil
}

func (c *LabelConflictChecker) checkConflicts(ctx context.Context, labels map[string]proton.Label, mboxFetch mailboxFetcherFn) ([]labelDiscrepancy, error) {
	discrepancies := []labelDiscrepancy{}

	// Verify bridge internal mailboxes.
	for _, prefix := range []string{folderPrefix, labelPrefix} {
		label := proton.Label{
			Path: []string{prefix},
			ID:   prefix,
			Name: prefix,
		}

		mbox, err := mboxFetch(ctx, label)
		if err != nil {
			if db.IsErrNotFound(err) {
				continue
			}
			return nil, err
		}

		if mbox.RemoteID != label.ID {
			discrepancies = append(discrepancies, newLabelDiscrepancy(label, mbox, discrepancyInternal))
		}
	}

	// Verify system and user mailboxes.
	for _, label := range labels {
		if !WantLabel(label) {
			continue
		}

		mbox, err := mboxFetch(ctx, label)
		if err != nil {
			if db.IsErrNotFound(err) {
				continue
			}
			return nil, err
		}

		if mbox.RemoteID != label.ID {
			var dType labelDiscrepancyType
			switch label.Type {
			case proton.LabelTypeSystem:
				dType = discrepancySystem
			case proton.LabelTypeFolder, proton.LabelTypeLabel:
				dType = discrepancyUser
			case proton.LabelTypeContactGroup:
				fallthrough
			default:
				dType = discrepancySystem
			}
			discrepancies = append(discrepancies, newLabelDiscrepancy(label, mbox, dType))
		}
	}

	return discrepancies, nil
}
