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
	"fmt"
	"io"
	"sync"
	"testing"

	"github.com/ProtonMail/gluon/async"
	"github.com/ProtonMail/go-proton-api"
	"github.com/ProtonMail/proton-bridge/v3/internal/network"
	"github.com/bradenaw/juniper/xslices"
	"github.com/golang/mock/gomock"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/require"
)

const TestMetadataPageSize = 5
const TestMaxDownloadMem = Kilobyte
const TestMaxMessages = 10

func TestMetadataStage_RunFinishesWith429(t *testing.T) {
	mockCtrl := gomock.NewController(t)

	tj := newTestJob(context.Background(), mockCtrl, "u", getTestLabels())
	tj.state.EXPECT().GetSyncStatus(gomock.Any()).Return(Status{
		LastSyncedMessageID: "",
	}, nil)

	input := NewChannelConsumerProducer[*Job]()
	output := NewChannelConsumerProducer[DownloadRequest]()

	ctx, cancel := context.WithCancel(context.Background())
	metadata := NewMetadataStage(input, output, TestMaxDownloadMem, &async.NoopPanicHandler{})

	numMessages := 50
	messageSize := 100

	msgs := setupMetadataSuccessRunWith429(&tj, numMessages, messageSize)

	go func() {
		metadata.run(ctx, TestMetadataPageSize, TestMaxMessages, &network.NoCoolDown{})
	}()

	require.NoError(t, input.Produce(ctx, tj.job))

	for _, chunk := range xslices.Chunk(msgs, TestMaxMessages) {
		tj.syncReporter.EXPECT().OnProgress(gomock.Any(), gomock.Eq(int64(len(chunk))))
		req, err := output.Consume(ctx)
		require.NoError(t, err)
		require.Equal(t, req.ids, xslices.Map(chunk, func(m proton.MessageMetadata) string {
			return m.ID
		}))
	}
	cancel()
}

func TestMetadataStage_JobCorrectlyFinishesAfterCancel(t *testing.T) {
	mockCtrl := gomock.NewController(t)

	jobCtx, jobCancel := context.WithCancel(context.Background())
	tj := newTestFixedMetadataJob(jobCtx, mockCtrl, "u", getTestLabels())
	tj.state.EXPECT().GetSyncStatus(gomock.Any()).Return(Status{
		LastSyncedMessageID: "",
	}, nil)
	tj.state.EXPECT().SetLastMessageID(gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes()
	tj.syncReporter.EXPECT().OnProgress(gomock.Any(), gomock.Any()).AnyTimes()

	input := NewChannelConsumerProducer[*Job]()
	output := NewChannelConsumerProducer[DownloadRequest]()

	ctx, cancel := context.WithCancel(context.Background())
	metadata := NewMetadataStage(input, output, TestMaxDownloadMem, &async.NoopPanicHandler{})

	go func() {
		metadata.run(ctx, TestMetadataPageSize, TestMaxMessages, &network.NoCoolDown{})
	}()

	{
		err := input.Produce(ctx, tj.job)
		require.NoError(t, err)
	}

	// read one output then cancel
	request, err := output.Consume(ctx)
	require.NoError(t, err)
	request.onFinished(ctx)
	// cancel job context
	jobCancel()

	wg := sync.WaitGroup{}
	wg.Add(1)
	// The next stages should check whether the job has been cancelled or not. Here we need to do it manually.
	go func() {
		wg.Done()
		for {
			req, err := output.Consume(ctx)
			if err != nil {
				return
			}

			req.checkCancelled()
		}
	}()
	wg.Wait()
	err = tj.job.waitAndClose(ctx)
	require.Error(t, err)
	require.ErrorIs(t, err, context.Canceled)
	cancel()
}

func TestMetadataStage_RunInterleaved(t *testing.T) {
	mockCtrl := gomock.NewController(t)

	tj1 := newTestJob(context.Background(), mockCtrl, "u", getTestLabels())
	tj1.state.EXPECT().GetSyncStatus(gomock.Any()).Return(Status{}, nil)
	tj1.state.EXPECT().SetLastMessageID(gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes()
	tj1.syncReporter.EXPECT().OnProgress(gomock.Any(), gomock.Any()).AnyTimes()

	tj2 := newTestJob(context.Background(), mockCtrl, "u", getTestLabels())
	tj2.state.EXPECT().GetSyncStatus(gomock.Any()).Return(Status{}, nil)
	tj2.state.EXPECT().SetLastMessageID(gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes()
	tj2.syncReporter.EXPECT().OnProgress(gomock.Any(), gomock.Any()).AnyTimes()

	input := NewChannelConsumerProducer[*Job]()
	output := NewChannelConsumerProducer[DownloadRequest]()

	ctx, cancel := context.WithCancel(context.Background())
	metadata := NewMetadataStage(input, output, TestMaxDownloadMem, &async.NoopPanicHandler{})

	numMessages := 50
	messageSize := 100

	setupMetadataSuccessRunWith429(&tj1, numMessages, messageSize)
	setupMetadataSuccessRunWith429(&tj2, numMessages, messageSize)

	go func() {
		metadata.run(ctx, TestMetadataPageSize, TestMaxMessages, &network.NoCoolDown{})
	}()

	go func() {
		require.NoError(t, input.Produce(ctx, tj1.job))
		require.NoError(t, input.Produce(ctx, tj2.job))
	}()

	go func() {
		for {
			req, err := output.Consume(ctx)
			if err != nil {
				require.ErrorIs(t, err, context.Canceled)
				return
			}

			req.onFinished(ctx)
		}
	}()

	require.NoError(t, tj1.job.waitAndClose(ctx))
	require.NoError(t, tj2.job.waitAndClose(ctx))
	cancel()
}

func TestMetadataIterator_ExitNoMoreMetadata(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	ctx := context.Background()
	tj := newTestJob(ctx, mockCtrl, "u", getTestLabels())

	tj.state.EXPECT().GetSyncStatus(gomock.Any()).Return(Status{
		LastSyncedMessageID: "foo",
	}, nil)

	tj.client.EXPECT().GetMessageMetadataPage(gomock.Any(), gomock.Eq(0), gomock.Eq(TestMetadataPageSize), gomock.Eq(proton.MessageFilter{Desc: true, EndID: "foo"})).Return(nil, nil)

	iter, err := newMetadataIterator(ctx, tj.job, TestMetadataPageSize, &network.NoCoolDown{})
	require.NoError(t, err)

	j, hasMore, err := iter.Next(TestMaxDownloadMem, TestMetadataPageSize, TestMaxMessages)
	require.NoError(t, err)
	require.False(t, hasMore)
	require.Empty(t, j.ids)
}

func TestMetadataIterator_ExitLastCallAlwaysReturnLastMessageID(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	ctx := context.Background()
	tj := newTestJob(ctx, mockCtrl, "u", getTestLabels())

	tj.state.EXPECT().GetSyncStatus(gomock.Any()).Return(Status{
		LastSyncedMessageID: "foo",
	}, nil)

	tj.client.EXPECT().GetMessageMetadataPage(
		gomock.Any(),
		gomock.Eq(0),
		gomock.Eq(TestMetadataPageSize),
		gomock.Eq(proton.MessageFilter{Desc: true, EndID: "foo"}),
	).Return([]proton.MessageMetadata{{
		ID:   "foo",
		Size: 100,
	}}, nil)

	iter, err := newMetadataIterator(ctx, tj.job, TestMetadataPageSize, &network.NoCoolDown{})
	require.NoError(t, err)

	j, hasMore, err := iter.Next(TestMaxDownloadMem, TestMetadataPageSize, TestMaxMessages)
	require.NoError(t, err)
	require.False(t, hasMore)
	require.Empty(t, j.ids)
}

func TestMetadataIterator_ExitWithRemainingReturnsNoMore(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	ctx := context.Background()
	tj := newTestJob(ctx, mockCtrl, "u", getTestLabels())

	tj.state.EXPECT().GetSyncStatus(gomock.Any()).Return(Status{}, nil)

	const MetadataPageSize = 2

	tj.client.EXPECT().GetMessageMetadataPage(
		gomock.Any(),
		gomock.Eq(0),
		gomock.Eq(MetadataPageSize),
		gomock.Eq(proton.MessageFilter{
			Desc: true,
		}),
	).Return([]proton.MessageMetadata{
		{
			ID:   "foo",
			Size: 100,
		},
		{
			ID:   "bar",
			Size: 100,
		},
	}, nil)

	tj.client.EXPECT().GetMessageMetadataPage(
		gomock.Any(),
		gomock.Eq(0),
		gomock.Eq(MetadataPageSize),
		gomock.Eq(proton.MessageFilter{
			Desc:  true,
			EndID: "bar",
		}),
	).Return([]proton.MessageMetadata{
		{
			ID:   "bar",
			Size: 100,
		},
	}, nil)

	iter, err := newMetadataIterator(ctx, tj.job, MetadataPageSize, &network.NoCoolDown{})
	require.NoError(t, err)

	j, hasMore, err := iter.Next(TestMaxDownloadMem, MetadataPageSize, 3)
	require.NoError(t, err)
	require.False(t, hasMore)
	require.Equal(t, []string{"foo", "bar"}, j.ids)
}

func TestMetadataIterator_RespectsSizeLimit(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	ctx := context.Background()
	tj := newTestJob(ctx, mockCtrl, "u", getTestLabels())

	tj.state.EXPECT().GetSyncStatus(gomock.Any()).Return(Status{}, nil)

	// First call.
	tj.client.EXPECT().GetMessageMetadataPage(
		gomock.Any(),
		gomock.Eq(0),
		gomock.Eq(TestMetadataPageSize),
		gomock.Eq(proton.MessageFilter{Desc: true}),
	).Return([]proton.MessageMetadata{
		{
			ID:   testMsgID(0),
			Size: 256,
		},
		{
			ID:   testMsgID(1),
			Size: 512,
		},
		{
			ID:   testMsgID(2),
			Size: 128,
		},
		{
			ID:   testMsgID(3),
			Size: 256,
		},
	}, nil)

	// Second Call
	tj.client.EXPECT().GetMessageMetadataPage(
		gomock.Any(),
		gomock.Eq(0),
		gomock.Eq(TestMetadataPageSize),
		gomock.Eq(proton.MessageFilter{Desc: true, EndID: testMsgID(3)}),
	).Return([]proton.MessageMetadata{
		{
			ID:   testMsgID(3),
			Size: 256,
		},
	}, nil)

	iter, err := newMetadataIterator(ctx, tj.job, TestMetadataPageSize, &network.NoCoolDown{})
	require.NoError(t, err)

	j, hasMore, err := iter.Next(TestMaxDownloadMem, TestMetadataPageSize, TestMaxMessages)
	require.NoError(t, err)
	require.True(t, hasMore)
	require.Equal(t, []string{testMsgID(0), testMsgID(1), testMsgID(2)}, j.ids)

	j, hasMore, err = iter.Next(TestMaxDownloadMem, TestMetadataPageSize, TestMaxMessages)
	require.NoError(t, err)
	require.False(t, hasMore)
	require.Equal(t, []string{testMsgID(3)}, j.ids)
}

func testMsgID(i int) string {
	return fmt.Sprintf("msg-id-%v", i)
}

func setupMetadataSuccessRunWith429(tj *tjob, msgCount int, msgSize int) []proton.MessageMetadata {
	msgs := make([]proton.MessageMetadata, msgCount)

	for i := 0; i < msgCount; i++ {
		msgs[i].ID = testMsgID(i)
		msgs[i].Size = msgSize
	}

	// setup api call
	for i := 0; i < msgCount; i += TestMetadataPageSize - 1 {
		filter := proton.MessageFilter{
			Desc: true,
		}

		if i != 0 {
			filter.EndID = msgs[i].ID
		}

		if i+TestMetadataPageSize > msgCount {
			call := tj.client.EXPECT().GetMessageMetadataPage(gomock.Any(), gomock.Eq(0), gomock.Eq(TestMetadataPageSize), gomock.Eq(filter)).Return(
				nil, &proton.APIError{Status: 503},
			)
			tj.client.EXPECT().GetMessageMetadataPage(gomock.Any(), gomock.Eq(0), gomock.Eq(TestMetadataPageSize), gomock.Eq(filter)).Return(
				msgs[i:], nil,
			).After(call)
		} else {
			call := tj.client.EXPECT().GetMessageMetadataPage(gomock.Any(), gomock.Eq(0), gomock.Eq(TestMetadataPageSize), gomock.Eq(filter)).Return(
				nil, &proton.APIError{Status: 429},
			)

			tj.client.EXPECT().GetMessageMetadataPage(gomock.Any(), gomock.Eq(0), gomock.Eq(TestMetadataPageSize), gomock.Eq(filter)).Return(
				msgs[i:i+TestMetadataPageSize], nil,
			).After(call)
		}
	}

	// Last call with last metadata id
	tj.client.EXPECT().GetMessageMetadataPage(gomock.Any(), gomock.Eq(0), gomock.Eq(TestMetadataPageSize), gomock.Eq(proton.MessageFilter{Desc: true, EndID: msgs[msgCount-1].ID})).Return(
		msgs[msgCount-1:], nil,
	)

	return msgs
}

func newTestFixedMetadataJob(
	ctx context.Context,
	mockCtrl *gomock.Controller,
	userID string,
	labels LabelMap,
) tjob {
	messageBuilder := NewMockMessageBuilder(mockCtrl)
	updateApplier := NewMockUpdateApplier(mockCtrl)
	syncReporter := NewMockReporter(mockCtrl)
	state := NewMockStateProvider(mockCtrl)
	client := newFixedMetadataClient(50)

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
		client:         nil,
		messageBuilder: messageBuilder,
		updateApplier:  updateApplier,
		syncReporter:   syncReporter,
		state:          state,
	}
}

type fixedMetadataClient struct {
	msg    []proton.MessageMetadata
	offset int
}

func newFixedMetadataClient(msgCount int) APIClient {
	msgs := make([]proton.MessageMetadata, msgCount)

	for i := 0; i < msgCount; i++ {
		msgs[i].ID = testMsgID(i)
		msgs[i].Size = 100
	}

	return &fixedMetadataClient{msg: msgs}
}

func (c *fixedMetadataClient) GetGroupedMessageCount(_ context.Context) ([]proton.MessageGroupCount, error) {
	panic("should not be called")
}

func (c *fixedMetadataClient) GetLabels(_ context.Context, _ ...proton.LabelType) ([]proton.Label, error) {
	panic("should not be called")
}

func (c *fixedMetadataClient) GetMessage(_ context.Context, _ string) (proton.Message, error) {
	panic("should not be called")
}

func (c *fixedMetadataClient) GetMessageMetadataPage(_ context.Context, _, pageSize int, _ proton.MessageFilter) ([]proton.MessageMetadata, error) {
	result := c.msg[c.offset : c.offset+pageSize]
	c.offset += pageSize

	return result, nil
}

func (c *fixedMetadataClient) GetMessageIDs(_ context.Context, _ string) ([]string, error) {
	panic("should not be called")
}

func (c *fixedMetadataClient) GetFullMessage(_ context.Context, _ string, _ proton.Scheduler, _ proton.AttachmentAllocator) (proton.FullMessage, error) {
	panic("should not be called")
}

func (c *fixedMetadataClient) GetAttachmentInto(_ context.Context, _ string, _ io.ReaderFrom) error {
	panic("should not be called")
}

func (c *fixedMetadataClient) GetAttachment(_ context.Context, _ string) ([]byte, error) {
	panic("should not be called")
}
