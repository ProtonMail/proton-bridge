// Copyright (c) 2026 Proton AG
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

package user

import (
	"context"
	"testing"

	"github.com/ProtonMail/proton-bridge/v3/internal/events/mocks"
	"github.com/ProtonMail/proton-bridge/v3/internal/sentry"
	"github.com/ProtonMail/proton-bridge/v3/internal/services/imapservice"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

func TestCheckIrrecoverableEventID_EventIDIsEmptyButNoSyncStarted(t *testing.T) {
	tmpDir := t.TempDir()
	userID := "foo"
	mockCtrl := gomock.NewController(t)
	publisher := mocks.NewMockEventPublisher(mockCtrl)

	require.NoError(t, checkIrrecoverableEventID(context.Background(), "", userID, tmpDir, publisher, sentry.NullSentryReporter{}))
}

func TestCheckIrrecoverableEventID_EventIDIsNotEmptyButNoSyncStarted(t *testing.T) {
	tmpDir := t.TempDir()
	userID := "foo"
	mockCtrl := gomock.NewController(t)
	publisher := mocks.NewMockEventPublisher(mockCtrl)

	require.NoError(t, checkIrrecoverableEventID(context.Background(), "ffoofo", userID, tmpDir, publisher, sentry.NullSentryReporter{}))
}

func TestCheckIrrecoverableEventID_EventIDIsEmptyButSyncStarted(t *testing.T) {
	tmpDir := t.TempDir()
	userID := "foo"
	mockCtrl := gomock.NewController(t)
	publisher := mocks.NewMockEventPublisher(mockCtrl)

	publisher.EXPECT().PublishEvent(gomock.Any(), gomock.Eq(newEmptyEventIDBadEvent(userID)))

	require.NoError(t, genSyncState(context.Background(), userID, tmpDir, false))
	require.NoError(t, checkIrrecoverableEventID(context.Background(), "", userID, tmpDir, publisher, sentry.NullSentryReporter{}))
}

func TestCheckIrrecoverableEventID_EventIDIsEmptyButSyncFinished(t *testing.T) {
	tmpDir := t.TempDir()
	userID := "foo"
	mockCtrl := gomock.NewController(t)
	publisher := mocks.NewMockEventPublisher(mockCtrl)

	publisher.EXPECT().PublishEvent(gomock.Any(), gomock.Eq(newEmptyEventIDBadEvent(userID)))

	require.NoError(t, genSyncState(context.Background(), userID, tmpDir, true))
	require.NoError(t, checkIrrecoverableEventID(context.Background(), "", userID, tmpDir, publisher, sentry.NullSentryReporter{}))
}

func TestCheckIrrecoverableEventID_EventIDIsNotEmptyButSyncFinished(t *testing.T) {
	tmpDir := t.TempDir()
	userID := "foo"
	mockCtrl := gomock.NewController(t)
	publisher := mocks.NewMockEventPublisher(mockCtrl)

	require.NoError(t, genSyncState(context.Background(), userID, tmpDir, true))
	require.NoError(t, checkIrrecoverableEventID(context.Background(), "some event", userID, tmpDir, publisher, sentry.NullSentryReporter{}))
}

func genSyncState(ctx context.Context, userID, dir string, finished bool) error {
	s, err := imapservice.NewSyncState(imapservice.GetSyncConfigPath(dir, userID), sentry.NullSentryReporter{})
	if err != nil {
		return err
	}

	if finished {
		if err := s.SetHasLabels(ctx, true); err != nil {
			return err
		}
		if err := s.SetHasMessages(ctx, true); err != nil {
			return err
		}
		if err := s.SetMessageCount(ctx, 10); err != nil {
			return err
		}
	} else {
		if err := s.SetHasLabels(ctx, true); err != nil {
			return err
		}
	}
	return nil
}
