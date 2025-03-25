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

package syncservice

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/ProtonMail/gluon/async"
	"github.com/ProtonMail/go-proton-api"
	"github.com/bradenaw/juniper/xmaps"
	"github.com/golang/mock/gomock"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/require"
)

func TestTask_NoStateAndSucceeds(t *testing.T) {
	const MessageTotal int64 = 50
	const MessageID string = "foo"
	const MessageDelta int64 = 10

	labels := getTestLabels()
	mockCtrl := gomock.NewController(t)

	tt := newTestHandler(mockCtrl, "u")

	tt.addMessageSyncCompletedExpectation(MessageID, MessageDelta)

	{
		call1 := tt.syncState.EXPECT().GetSyncStatus(gomock.Any()).DoAndReturn(func(_ context.Context) (Status, error) {
			return Status{
				HasLabels:           false,
				HasMessages:         false,
				HasMessageCount:     false,
				FailedMessages:      xmaps.SetFromSlice([]string{}),
				LastSyncedMessageID: "",
				NumSyncedMessages:   0,
				TotalMessageCount:   0,
			}, nil
		})
		call2 := tt.syncState.EXPECT().SetHasLabels(gomock.Any(), gomock.Eq(true)).After(call1).Times(1).Return(nil)
		call3 := tt.syncState.EXPECT().SetMessageCount(gomock.Any(), gomock.Eq(MessageTotal)).After(call2).Times(1).Return(nil)
		tt.syncReporter.EXPECT().InitializeProgressCounter(gomock.Any(), gomock.Any(), gomock.Eq(MessageTotal*NumSyncStages))
		call4 := tt.syncState.EXPECT().SetLastMessageID(gomock.Any(), gomock.Eq(MessageID), gomock.Eq(MessageDelta)).After(call3).Times(1).Return(nil)
		call5 := tt.syncState.EXPECT().SetHasMessages(gomock.Any(), gomock.Eq(true)).After(call4).Times(1).Return(nil)
		tt.syncState.EXPECT().GetSyncStatus(gomock.Any()).After(call5).Times(1).DoAndReturn(func(_ context.Context) (Status, error) {
			return Status{
				HasLabels:           true,
				HasMessages:         true,
				HasMessageCount:     true,
				FailedMessages:      xmaps.SetFromSlice([]string{}),
				LastSyncedMessageID: MessageID,
				NumSyncedMessages:   MessageDelta,
				TotalMessageCount:   MessageTotal,
			}, nil
		})
	}

	{
		tt.updateApplier.EXPECT().SyncLabels(gomock.Any(), gomock.Eq(labels)).Times(2).Return(nil)
	}

	{
		tt.client.EXPECT().GetGroupedMessageCount(gomock.Any()).Return([]proton.MessageGroupCount{
			{
				LabelID: proton.AllMailLabel,
				Total:   int(MessageTotal),
				Unread:  0,
			},
		}, nil)
	}

	tt.syncReporter.EXPECT().OnProgress(gomock.Any(), gomock.Eq(MessageDelta))

	// First run.
	err := tt.task.run(context.Background(), tt.syncReporter, labels, tt.updateApplier, tt.messageBuilder)
	require.NoError(t, err)

	// Second Run, it's completed sync labels only.
	err = tt.task.run(context.Background(), tt.syncReporter, labels, tt.updateApplier, tt.messageBuilder)
	require.NoError(t, err)
}

func TestTask_StateHasLabels(t *testing.T) {
	const MessageTotal int64 = 50
	const MessageID string = "foo"
	const MessageDelta int64 = 10

	labels := getTestLabels()

	mockCtrl := gomock.NewController(t)
	tt := newTestHandler(mockCtrl, "u")

	tt.addMessageSyncCompletedExpectation(MessageID, MessageDelta)

	{
		call2 := tt.syncState.EXPECT().GetSyncStatus(gomock.Any()).DoAndReturn(func(_ context.Context) (Status, error) {
			return Status{
				HasLabels:           true,
				HasMessages:         false,
				HasMessageCount:     false,
				FailedMessages:      xmaps.SetFromSlice([]string{}),
				LastSyncedMessageID: "",
				NumSyncedMessages:   0,
				TotalMessageCount:   0,
			}, nil
		})
		call3 := tt.syncState.EXPECT().SetMessageCount(gomock.Any(), gomock.Eq(MessageTotal)).After(call2).Times(1).Return(nil)
		call4 := tt.syncState.EXPECT().SetLastMessageID(gomock.Any(), gomock.Eq(MessageID), gomock.Eq(MessageDelta)).After(call3).Times(1).Return(nil)
		tt.syncState.EXPECT().SetHasMessages(gomock.Any(), gomock.Eq(true)).After(call4).Times(1).Return(nil)
		tt.syncReporter.EXPECT().InitializeProgressCounter(gomock.Any(), gomock.Any(), gomock.Eq(MessageTotal*NumSyncStages))
	}

	{
		tt.client.EXPECT().GetGroupedMessageCount(gomock.Any()).Return([]proton.MessageGroupCount{
			{
				LabelID: proton.AllMailLabel,
				Total:   int(MessageTotal),
				Unread:  0,
			},
		}, nil)
	}

	tt.syncReporter.EXPECT().OnProgress(gomock.Any(), gomock.Eq(MessageDelta))

	err := tt.task.run(context.Background(), tt.syncReporter, labels, tt.updateApplier, tt.messageBuilder)
	require.NoError(t, err)
}

func TestTask_StateHasLabelsAndMessageCount(t *testing.T) {
	const MessageTotal int64 = 50
	const MessageID string = "foo"
	const MessageDelta int64 = 10

	labels := getTestLabels()

	mockCtrl := gomock.NewController(t)

	tt := newTestHandler(mockCtrl, "u")

	tt.addMessageSyncCompletedExpectation(MessageID, MessageDelta)

	{
		call3 := tt.syncState.EXPECT().GetSyncStatus(gomock.Any()).DoAndReturn(func(_ context.Context) (Status, error) {
			return Status{
				HasLabels:           true,
				HasMessages:         false,
				HasMessageCount:     true,
				FailedMessages:      xmaps.SetFromSlice([]string{}),
				LastSyncedMessageID: "",
				NumSyncedMessages:   0,
				TotalMessageCount:   MessageTotal,
			}, nil
		})
		call4 := tt.syncState.EXPECT().SetLastMessageID(gomock.Any(), gomock.Eq(MessageID), gomock.Eq(MessageDelta)).After(call3).Times(1).Return(nil)
		tt.syncState.EXPECT().SetHasMessages(gomock.Any(), gomock.Eq(true)).After(call4).Times(1).Return(nil)
		tt.syncReporter.EXPECT().InitializeProgressCounter(gomock.Any(), gomock.Any(), gomock.Eq(MessageTotal*NumSyncStages))
	}

	tt.syncReporter.EXPECT().OnProgress(gomock.Any(), gomock.Eq(MessageDelta))

	err := tt.task.run(context.Background(), tt.syncReporter, labels, tt.updateApplier, tt.messageBuilder)
	require.NoError(t, err)
}

func TestTask_StateHasSyncedState(t *testing.T) {
	const MessageTotal int64 = 50
	const MessageID string = "foo"

	labels := getTestLabels()

	mockCtrl := gomock.NewController(t)

	tt := newTestHandler(mockCtrl, "u")

	tt.syncState.EXPECT().GetSyncStatus(gomock.Any()).DoAndReturn(func(_ context.Context) (Status, error) {
		return Status{
			HasLabels:           true,
			HasMessages:         true,
			HasMessageCount:     true,
			FailedMessages:      xmaps.SetFromSlice([]string{}),
			LastSyncedMessageID: MessageID,
			NumSyncedMessages:   MessageTotal,
			TotalMessageCount:   MessageTotal,
		}, nil
	})

	tt.updateApplier.EXPECT().SyncLabels(gomock.Any(), gomock.Eq(labels)).Return(nil)

	err := tt.task.run(context.Background(), tt.syncReporter, labels, tt.updateApplier, tt.messageBuilder)
	require.NoError(t, err)
}

func TestTask_RepeatsOnSyncFailure(t *testing.T) {
	const MessageTotal int64 = 50
	const MessageID string = "foo"
	const MessageDelta int64 = 10

	labels := getTestLabels()

	mockCtrl := gomock.NewController(t)

	tt := newTestHandler(mockCtrl, "u")

	tt.addMessageSyncCompletedExpectation(MessageID, MessageDelta)

	tt.syncReporter.EXPECT().InitializeProgressCounter(gomock.Any(), gomock.Any(), gomock.Eq(MessageTotal*NumSyncStages))

	{
		call0 := tt.syncState.EXPECT().GetSyncStatus(gomock.Any()).DoAndReturn(func(_ context.Context) (Status, error) {
			return Status{
				HasLabels:           false,
				HasMessages:         false,
				HasMessageCount:     false,
				FailedMessages:      xmaps.SetFromSlice([]string{}),
				LastSyncedMessageID: "",
				NumSyncedMessages:   0,
				TotalMessageCount:   0,
			}, nil
		})
		call1 := tt.syncState.EXPECT().GetSyncStatus(gomock.Any()).DoAndReturn(func(_ context.Context) (Status, error) {
			return Status{
				HasLabels:           false,
				HasMessages:         false,
				HasMessageCount:     false,
				FailedMessages:      xmaps.SetFromSlice([]string{}),
				LastSyncedMessageID: "",
				NumSyncedMessages:   0,
				TotalMessageCount:   0,
			}, nil
		}).After(call0)
		call2 := tt.syncState.EXPECT().SetHasLabels(gomock.Any(), gomock.Eq(true)).After(call1).Times(1).Return(nil)
		call3 := tt.syncState.EXPECT().SetMessageCount(gomock.Any(), gomock.Eq(MessageTotal)).After(call2).Times(1).Return(nil)
		call4 := tt.syncState.EXPECT().SetLastMessageID(gomock.Any(), gomock.Eq(MessageID), gomock.Eq(MessageDelta)).After(call3).Times(1).Return(nil)
		tt.syncState.EXPECT().SetHasMessages(gomock.Any(), gomock.Eq(true)).After(call4).Times(1).Return(nil)
	}

	{
		call0 := tt.updateApplier.EXPECT().SyncLabels(gomock.Any(), gomock.Eq(labels)).Times(1).Return(fmt.Errorf("failed"))
		tt.updateApplier.EXPECT().SyncLabels(gomock.Any(), gomock.Eq(labels)).Times(1).Return(nil).After(call0)
	}

	{
		tt.client.EXPECT().GetGroupedMessageCount(gomock.Any()).Return([]proton.MessageGroupCount{
			{
				LabelID: proton.AllMailLabel,
				Total:   int(MessageTotal),
				Unread:  0,
			},
		}, nil)
	}

	tt.syncReporter.EXPECT().OnStart(gomock.Any())
	tt.syncReporter.EXPECT().OnFinished(gomock.Any())
	tt.syncReporter.EXPECT().OnProgress(gomock.Any(), gomock.Eq(MessageDelta))

	tt.task.Execute(tt.syncReporter, labels, tt.updateApplier, tt.messageBuilder, time.Microsecond)
	require.NoError(t, <-tt.task.OnSyncFinishedCH())
}

func getTestLabels() map[string]proton.Label {
	return map[string]proton.Label{
		proton.AllMailLabel: {
			ID:   proton.AllMailLabel,
			Name: "All Mail",
			Type: proton.LabelTypeSystem,
		},
		proton.InboxLabel: {
			ID:   proton.InboxLabel,
			Name: "Inbox",
			Type: proton.LabelTypeSystem,
		},
		proton.DraftsLabel: {
			ID:   proton.DraftsLabel,
			Name: "Drafts",
			Type: proton.LabelTypeSystem,
		},
		proton.TrashLabel: {
			ID:   proton.DraftsLabel,
			Name: "Drafts",
			Type: proton.LabelTypeSystem,
		},
		"label1": {
			ID:   "label1",
			Name: "label1",
			Type: proton.LabelTypeLabel,
		},
		"folder1": {
			ID:   "folder1",
			Name: "folder1",
			Type: proton.LabelTypeFolder,
		},
		"folder2": {
			ID:       "folder2",
			Name:     "folder2",
			ParentID: "folder1",
			Type:     proton.LabelTypeFolder,
		},
	}
}

type thandler struct {
	task           *Handler
	regulator      *MockRegulator
	syncState      *MockStateProvider
	updateApplier  *MockUpdateApplier
	messageBuilder *MockMessageBuilder
	client         *MockAPIClient
	syncReporter   *MockReporter
}

func (t thandler) addMessageSyncCompletedExpectation(messageID string, delta int64) { //nolint:unparam
	t.regulator.EXPECT().Sync(gomock.Any(), gomock.Any()).Do(func(_ context.Context, job *Job) {
		job.begin()
		j := job.newChildJob(messageID, delta)
		j.onFinished(context.Background())
		job.end()
	})
}

func newTestHandler(mockCtrl *gomock.Controller, userID string) thandler { // nolint:unparam
	regulator := NewMockRegulator(mockCtrl)
	syncState := NewMockStateProvider(mockCtrl)
	updateApplier := NewMockUpdateApplier(mockCtrl)
	client := NewMockAPIClient(mockCtrl)
	messageBuilder := NewMockMessageBuilder(mockCtrl)
	syncReporter := NewMockReporter(mockCtrl)
	task := NewHandler(regulator, client, userID, syncState, logrus.WithField("test", "test"), &async.NoopPanicHandler{})

	return thandler{
		task:           task,
		regulator:      regulator,
		syncState:      syncState,
		updateApplier:  updateApplier,
		messageBuilder: messageBuilder,
		syncReporter:   syncReporter,
		client:         client,
	}
}
