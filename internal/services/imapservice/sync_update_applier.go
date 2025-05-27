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
	"fmt"

	"github.com/ProtonMail/gluon/imap"
	"github.com/ProtonMail/go-proton-api"
	"github.com/ProtonMail/proton-bridge/v3/internal/services/syncservice"
	"github.com/ProtonMail/proton-bridge/v3/internal/usertypes"
	"github.com/bradenaw/juniper/xslices"
	"github.com/sirupsen/logrus"
	"golang.org/x/exp/maps"
)

type SyncUpdateApplier struct {
	requestCh            chan updateRequest
	replyCh              chan updateReply
	labelConflictManager *LabelConflictManager
}

type updateReply struct {
	updates []imap.Update
	err     error
}

type updateRequest = func(ctx context.Context, mode usertypes.AddressMode, connectors map[string]*Connector) ([]imap.Update, error)

func NewSyncUpdateApplier(labelConflictManager *LabelConflictManager) *SyncUpdateApplier {
	return &SyncUpdateApplier{
		requestCh:            make(chan updateRequest),
		replyCh:              make(chan updateReply),
		labelConflictManager: labelConflictManager,
	}
}

func (s *SyncUpdateApplier) Close() {
	close(s.requestCh)
	close(s.replyCh)
}

func (s *SyncUpdateApplier) ApplySyncUpdates(ctx context.Context, updates []syncservice.BuildResult) error {
	request := func(ctx context.Context, mode usertypes.AddressMode, connectors map[string]*Connector) ([]imap.Update, error) {
		if mode == usertypes.AddressModeCombined {
			if len(connectors) != 1 {
				return nil, fmt.Errorf("unexpected connecto list state")
			}

			c := maps.Values(connectors)[0]

			update := imap.NewMessagesCreated(true, xslices.Map(updates, func(b syncservice.BuildResult) *imap.MessageCreated {
				return b.Update
			})...)

			c.publishUpdate(ctx, update)

			return []imap.Update{update}, nil
		}

		updateMap := make(map[string]*imap.MessagesCreated, len(connectors))
		result := make([]imap.Update, 0, len(connectors))

		for _, up := range updates {
			update, ok := updateMap[up.AddressID]
			if !ok {
				update = imap.NewMessagesCreated(true)
				updateMap[up.AddressID] = update
				result = append(result, update)
			}

			update.Messages = append(update.Messages, up.Update)
		}

		for addrID, update := range updateMap {
			c, ok := connectors[addrID]
			if !ok {
				logrus.Warnf("Could not find connector for address %v", addrID)
				continue
			}

			c.publishUpdate(ctx, update)
		}

		return result, nil
	}

	result, err := s.sendRequest(ctx, request)
	if err != nil {
		return err
	}

	if err := waitOnIMAPUpdates(ctx, result); err != nil {
		return fmt.Errorf("could not apply updates: %w", err)
	}

	return nil
}

func (s *SyncUpdateApplier) SyncLabels(ctx context.Context, labels map[string]proton.Label) error {
	request := func(ctx context.Context, _ usertypes.AddressMode, connectors map[string]*Connector) ([]imap.Update, error) {
		return syncLabels(ctx, labels, maps.Values(connectors), s.labelConflictManager)
	}

	updates, err := s.sendRequest(ctx, request)
	if err != nil {
		return err
	}
	if err := waitOnIMAPUpdates(ctx, updates); err != nil {
		return fmt.Errorf("could not sync labels: %w", err)
	}

	return nil
}

// nolint:exhaustive
func syncLabels(ctx context.Context, labels map[string]proton.Label, connectors []*Connector, labelConflictManager *LabelConflictManager) ([]imap.Update, error) {
	var updates []imap.Update

	labelConflictResolver := labelConflictManager.NewConflictResolver(connectors)

	// Create placeholder Folders/Labels mailboxes with the \Noselect attribute.
	for _, prefix := range []string{folderPrefix, labelPrefix} {
		for _, updateCh := range connectors {
			update := newPlaceHolderMailboxCreatedUpdate(prefix)
			updateCh.publishUpdate(ctx, update)
			updates = append(updates, update)
		}
	}

	// Sync the user's labels.
	for labelID, label := range labels {
		if !WantLabel(label) {
			continue
		}

		switch label.Type {
		case proton.LabelTypeSystem:
			for _, updateCh := range connectors {
				update := newSystemMailboxCreatedUpdate(imap.MailboxID(label.ID), label.Name)
				updateCh.publishUpdate(ctx, update)
				updates = append(updates, update)
			}

		case proton.LabelTypeFolder, proton.LabelTypeLabel:
			conflictUpdatesGenerator, err := labelConflictResolver.ResolveConflict(ctx, label, make(map[string]bool))
			if err != nil {
				return updates, err
			}

			for _, updateCh := range connectors {
				conflictUpdates := conflictUpdatesGenerator()
				updateCh.publishUpdate(ctx, conflictUpdates...)
				updates = append(updates, conflictUpdates...)

				update := newMailboxCreatedUpdate(imap.MailboxID(labelID), GetMailboxName(label))
				updateCh.publishUpdate(ctx, update)
				updates = append(updates, update)
			}

		default:
			logrus.Errorf("Unknown label type: %d", label.Type)
			continue
		}
	}

	return updates, nil
}

func (s *SyncUpdateApplier) sendRequest(ctx context.Context, request updateRequest) ([]imap.Update, error) {
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	case s.requestCh <- request:
	}

	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	case reply, ok := <-s.replyCh:
		if !ok {
			return nil, fmt.Errorf("no reply")
		}

		if reply.err != nil {
			return nil, reply.err
		}

		return reply.updates, nil
	}
}

func (s *SyncUpdateApplier) reply(ctx context.Context, updates []imap.Update, err error) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	case s.replyCh <- updateReply{
		updates: updates,
		err:     err,
	}:
		return nil
	}
}
