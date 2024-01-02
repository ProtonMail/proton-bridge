// Copyright (c) 2024 Proton AG
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
	"errors"
	"testing"

	"github.com/ProtonMail/gluon/imap"
	"github.com/ProtonMail/go-proton-api"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/require"
)

func TestApplyStage_CancelledJobIsDiscarded(t *testing.T) {
	mockCtrl := gomock.NewController(t)

	input := NewChannelConsumerProducer[ApplyRequest]()

	stage := NewApplyStage(input)

	ctx, cancel := context.WithCancel(context.Background())

	jobCtx, jobCancel := context.WithCancel(context.Background())

	tj := newTestJob(jobCtx, mockCtrl, "", map[string]proton.Label{})

	tj.job.begin()
	childJob := tj.job.newChildJob("f", 10)
	tj.job.end()

	go func() {
		stage.run(ctx)
	}()

	jobCancel()
	require.NoError(t, input.Produce(ctx, ApplyRequest{
		childJob: childJob,
		messages: nil,
	}))

	err := tj.job.waitAndClose(ctx)
	require.ErrorIs(t, err, context.Canceled)
	cancel()
}

func TestApplyStage_JobWithNoMessagesIsFinalized(t *testing.T) {
	mockCtrl := gomock.NewController(t)

	input := NewChannelConsumerProducer[ApplyRequest]()

	stage := NewApplyStage(input)

	ctx, cancel := context.WithCancel(context.Background())

	jobCtx, jobCancel := context.WithCancel(context.Background())
	defer jobCancel()

	tj := newTestJob(jobCtx, mockCtrl, "", map[string]proton.Label{})
	tj.syncReporter.EXPECT().OnProgress(gomock.Any(), gomock.Any())
	tj.state.EXPECT().SetLastMessageID(gomock.Any(), gomock.Eq("f"), gomock.Eq(int64(10)))

	tj.job.begin()
	childJob := tj.job.newChildJob("f", 10)
	tj.job.end()

	go func() {
		stage.run(ctx)
	}()

	require.NoError(t, input.Produce(ctx, ApplyRequest{
		childJob: childJob,
		messages: nil,
	}))

	err := tj.job.waitAndClose(ctx)
	cancel()
	require.NoError(t, err)
}

func TestApplyStage_ErrorOnApplyIsReportedAndJobFails(t *testing.T) {
	mockCtrl := gomock.NewController(t)

	input := NewChannelConsumerProducer[ApplyRequest]()

	stage := NewApplyStage(input)

	ctx, cancel := context.WithCancel(context.Background())

	jobCtx, jobCancel := context.WithCancel(context.Background())
	defer jobCancel()

	buildResults := []BuildResult{
		{
			AddressID: "Foo",
			MessageID: "Bar",
			Update:    &imap.MessageCreated{},
		},
	}

	tj := newTestJob(jobCtx, mockCtrl, "", map[string]proton.Label{})

	applyErr := errors.New("apply failed")
	tj.updateApplier.EXPECT().ApplySyncUpdates(gomock.Any(), gomock.Eq(buildResults)).Return(applyErr)

	tj.job.begin()
	childJob := tj.job.newChildJob("f", 10)
	tj.job.end()

	go func() {
		stage.run(ctx)
	}()

	require.NoError(t, input.Produce(ctx, ApplyRequest{
		childJob: childJob,
		messages: buildResults,
	}))

	err := tj.job.waitAndClose(ctx)
	cancel()
	require.ErrorIs(t, err, applyErr)
}
