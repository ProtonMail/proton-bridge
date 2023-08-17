// Copyright (c) 2023 Proton AG
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

	"github.com/ProtonMail/gluon/async"
	"github.com/golang/mock/gomock"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/require"
	"go.uber.org/goleak"
)

func setupGoLeak() goleak.Option {
	logrus.Trace("prepare for go leak")
	return goleak.IgnoreCurrent()
}

func TestJob_WaitsOnChildren(t *testing.T) {
	options := setupGoLeak()
	defer goleak.VerifyNone(t, options)

	mockCtrl := gomock.NewController(t)

	tj := newTestJob(context.Background(), mockCtrl, "u", getTestLabels())

	tj.state.EXPECT().SetLastMessageID(gomock.Any(), gomock.Eq("1"), gomock.Eq(int64(0))).Return(nil)
	tj.state.EXPECT().SetLastMessageID(gomock.Any(), gomock.Eq("2"), gomock.Eq(int64(1))).Return(nil)
	tj.syncReporter.EXPECT().OnProgress(gomock.Any(), gomock.Any()).Times(2)

	go func() {
		tj.job.begin()
		job1 := tj.job.newChildJob("1", 0)
		job2 := tj.job.newChildJob("2", 1)

		job1.onFinished(context.Background())
		job2.onFinished(context.Background())
		tj.job.end()
	}()

	require.NoError(t, tj.job.wait(context.Background()))
	tj.job.Close()
}

func TestJob_WaitsOnAllChildrenOnError(t *testing.T) {
	options := setupGoLeak()
	defer goleak.VerifyNone(t, options)

	mockCtrl := gomock.NewController(t)

	tj := newTestJob(context.Background(), mockCtrl, "u", getTestLabels())

	tj.state.EXPECT().SetLastMessageID(gomock.Any(), gomock.Eq("1"), gomock.Eq(int64(0))).Return(nil)
	tj.syncReporter.EXPECT().OnProgress(gomock.Any(), gomock.Any())

	jobErr := errors.New("failed")

	go func() {
		job1 := tj.job.newChildJob("1", 0)
		job2 := tj.job.newChildJob("2", 1)

		job1.onFinished(context.Background())
		job2.onError(jobErr)
	}()

	err := tj.job.wait(context.Background())
	require.Error(t, err)
	require.ErrorIs(t, err, jobErr)
	tj.job.Close()
}

func TestJob_MultipleChildrenReportError(t *testing.T) {
	options := setupGoLeak()
	defer goleak.VerifyNone(t, options)

	mockCtrl := gomock.NewController(t)

	tj := newTestJob(context.Background(), mockCtrl, "u", getTestLabels())

	jobErr := errors.New("failed")

	startCh := make(chan struct{})

	for i := 0; i < 10; i++ {
		go func() {
			job := tj.job.newChildJob("1", 0)
			<-startCh

			job.onError(jobErr)
		}()
	}

	close(startCh)
	err := tj.job.wait(context.Background())
	require.Error(t, err)
	require.ErrorIs(t, err, jobErr)
	tj.job.Close()
}

func TestJob_ChildFailureCancelsAllOtherChildJobs(t *testing.T) {
	options := setupGoLeak()
	defer goleak.VerifyNone(t, options)

	mockCtrl := gomock.NewController(t)

	tj := newTestJob(context.Background(), mockCtrl, "u", getTestLabels())

	jobErr := errors.New("failed")

	failJob := tj.job.newChildJob("0", 1)

	for i := 0; i < 10; i++ {
		go func() {
			job := tj.job.newChildJob("1", 0)
			<-job.getContext().Done()
			require.ErrorIs(t, job.getContext().Err(), context.Canceled)
			require.True(t, job.checkCancelled())
		}()
	}
	go func() {
		failJob.onError(jobErr)
	}()

	err := tj.job.wait(context.Background())
	require.Error(t, err)
	require.ErrorIs(t, err, jobErr)
	tj.job.Close()
}

func TestJob_CtxCancelCancelsAllChildren(t *testing.T) {
	options := setupGoLeak()
	defer goleak.VerifyNone(t, options)

	mockCtrl := gomock.NewController(t)

	ctx, cancel := context.WithCancel(context.Background())
	tj := newTestJob(ctx, mockCtrl, "u", getTestLabels())

	for i := 0; i < 10; i++ {
		go func() {
			job := tj.job.newChildJob("1", 0)
			<-job.getContext().Done()
			require.ErrorIs(t, job.getContext().Err(), context.Canceled)
			require.True(t, job.checkCancelled())
		}()
	}

	go func() {
		cancel()
	}()

	err := tj.job.wait(ctx)
	require.Error(t, err)
	require.ErrorIs(t, err, context.Canceled)
	tj.job.Close()
}

func TestJob_WithoutChildJobsCanBeTerminated(t *testing.T) {
	options := setupGoLeak()
	defer goleak.VerifyNone(t, options)

	mockCtrl := gomock.NewController(t)

	ctx := context.Background()

	tj := newTestJob(ctx, mockCtrl, "u", getTestLabels())
	go func() {
		tj.job.begin()
		tj.job.end()
	}()
	err := tj.job.wait(ctx)
	require.NoError(t, err)
	tj.job.Close()
}

type tjob struct {
	job            *Job
	client         *MockAPIClient
	messageBuilder *MockMessageBuilder
	updateApplier  *MockUpdateApplier
	syncReporter   *MockReporter
	state          *MockStateProvider
}

func newTestJob(
	ctx context.Context,
	mockCtrl *gomock.Controller,
	userID string,
	labels LabelMap,
) tjob {
	client := NewMockAPIClient(mockCtrl)
	messageBuilder := NewMockMessageBuilder(mockCtrl)
	updateApplier := NewMockUpdateApplier(mockCtrl)
	syncReporter := NewMockReporter(mockCtrl)
	state := NewMockStateProvider(mockCtrl)

	job := NewJob(
		ctx,
		client,
		userID,
		labels,
		messageBuilder,
		updateApplier,
		syncReporter,
		state,
		&async.NoopPanicHandler{},
		newDownloadCache(),
		logrus.WithField("s", "test"),
	)

	return tjob{
		job:            job,
		client:         client,
		messageBuilder: messageBuilder,
		updateApplier:  updateApplier,
		syncReporter:   syncReporter,
		state:          state,
	}
}
