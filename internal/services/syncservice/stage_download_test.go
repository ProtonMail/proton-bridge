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
	"bytes"
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/ProtonMail/gluon/async"
	"github.com/ProtonMail/go-proton-api"
	"github.com/ProtonMail/proton-bridge/v3/internal/network"
	"github.com/bradenaw/juniper/xslices"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/require"
)

func TestDownloadMessage_NotInCache(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	client := NewMockAPIClient(mockCtrl)
	cache := newDownloadCache()
	client.EXPECT().GetMessage(gomock.Any(), gomock.Any()).Return(proton.Message{}, nil)

	_, err := downloadMessage(context.Background(), cache, client, "msg")
	require.NoError(t, err)
}

func TestDownloadMessage_InCache(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	client := NewMockAPIClient(mockCtrl)
	cache := newDownloadCache()

	msg := proton.Message{
		MessageMetadata: proton.MessageMetadata{ID: "msg", Size: 1024},
	}

	cache.StoreMessage(msg)

	downloaded, err := downloadMessage(context.Background(), cache, client, "msg")
	require.NoError(t, err)
	require.Equal(t, msg, downloaded)
}

func TestDownloadAttachment_NotInCache(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	client := NewMockAPIClient(mockCtrl)
	cache := newDownloadCache()
	client.EXPECT().GetAttachmentInto(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)

	_, err := downloadAttachment(context.Background(), cache, client, "id", 1024)
	require.NoError(t, err)
}

func TestDownloadAttachment_InCache(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	client := NewMockAPIClient(mockCtrl)
	cache := newDownloadCache()
	attachment := []byte("hello world")
	cache.StoreAttachment("id", attachment)

	downloaded, err := downloadAttachment(context.Background(), cache, client, "id", 1024)
	require.NoError(t, err)
	require.Equal(t, attachment, downloaded)
}

func TestAutoDownloadScale_AllOkay(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	client := NewMockAPIClient(mockCtrl)
	data := buildDownloadScaleData(15)

	const MaxParallel = 5

	call1 := client.EXPECT().GetMessage(gomock.Any(), gomock.Any()).Times(5).DoAndReturn(autoDownloadScaleClientDoAndReturn)
	call2 := client.EXPECT().GetMessage(gomock.Any(), gomock.Any()).Times(5).After(call1).DoAndReturn(autoDownloadScaleClientDoAndReturn)
	client.EXPECT().GetMessage(gomock.Any(), gomock.Any()).Times(5).After(call2).DoAndReturn(autoDownloadScaleClientDoAndReturn)

	msgs, err := autoDownloadRate(
		context.Background(),
		&DefaultDownloadRateModifier{},
		client,
		MaxParallel,
		data,
		autoScaleCoolDown,
		func(ctx context.Context, client APIClient, input string) (proton.Message, error) {
			return client.GetMessage(ctx, input)
		},
	)

	require.NoError(t, err)
	require.Equal(t, xslices.Map(data, newDownloadScaleMessage), msgs)
}

func TestAutoDownloadScale_429or500x(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	client := NewMockAPIClient(mockCtrl)
	data := buildDownloadScaleData(32)

	rateModifier := NewMockDownloadRateModifier(mockCtrl)

	const MaxParallel = 8

	for _, d := range data {
		switch d {
		case "m7":
			call429 := client.EXPECT().GetMessage(gomock.Any(), gomock.Eq("m7")).DoAndReturn(func(_ context.Context, _ string) (proton.Message, error) {
				return proton.Message{}, &proton.APIError{Status: 429}
			})
			client.EXPECT().GetMessage(gomock.Any(), gomock.Eq("m7")).After(call429).DoAndReturn(autoDownloadScaleClientDoAndReturn)
		case "m23":
			call503 := client.EXPECT().GetMessage(gomock.Any(), gomock.Eq("m23")).DoAndReturn(func(_ context.Context, _ string) (proton.Message, error) {
				return proton.Message{}, &proton.APIError{Status: 503}
			})
			client.EXPECT().GetMessage(gomock.Any(), gomock.Eq("m23")).After(call503).DoAndReturn(autoDownloadScaleClientDoAndReturn)
		default:
			client.EXPECT().GetMessage(gomock.Any(), gomock.Eq(d)).DoAndReturn(autoDownloadScaleClientDoAndReturn)
		}
	}

	defaultRateModifier := DefaultDownloadRateModifier{}

	// First call catches failure in message m7, we throttle.
	call1 := rateModifier.EXPECT().Apply(gomock.Eq(false), gomock.Eq(8), gomock.Eq(8)).Return(defaultRateModifier.Apply(false, 8, 8))
	// Next batch succeeds. So we bump the parallel downloads by 2x.
	call2 := rateModifier.EXPECT().Apply(gomock.Eq(true), gomock.Eq(2), gomock.Eq(8)).Return(defaultRateModifier.Apply(true, 2, 8))
	// We now encounter a 503 with m23. Reset to 2 parallel downloads.
	call3 := rateModifier.EXPECT().Apply(gomock.Eq(false), gomock.Eq(4), gomock.Eq(8)).Return(defaultRateModifier.Apply(false, 4, 8))
	// The next batch succeeds once again.
	call4 := rateModifier.EXPECT().Apply(gomock.Eq(true), gomock.Eq(2), gomock.Eq(8)).Return(defaultRateModifier.Apply(true, 2, 8))

	gomock.InOrder(call1, call2, call3, call4)

	msgs, err := autoDownloadRate(
		context.Background(),
		rateModifier,
		client,
		MaxParallel,
		data,
		autoScaleCoolDown,
		func(ctx context.Context, client APIClient, input string) (proton.Message, error) {
			return client.GetMessage(ctx, input)
		},
	)

	require.NoError(t, err)
	require.Equal(t, xslices.Map(data, newDownloadScaleMessage), msgs)
}

func TestDownloadStage_Run(t *testing.T) {
	mockCtrl := gomock.NewController(t)

	input := NewChannelConsumerProducer[DownloadRequest]()
	output := NewChannelConsumerProducer[BuildRequest]()

	ctx, cancel := context.WithCancel(context.Background())

	tj := newTestJob(ctx, mockCtrl, "", map[string]proton.Label{})

	tj.syncReporter.EXPECT().OnProgress(gomock.Any(), gomock.Any())
	tj.state.EXPECT().SetLastMessageID(gomock.Any(), gomock.Eq("f"), gomock.Eq(int64(10))).Return(nil)

	tj.syncReporter.EXPECT().OnProgress(gomock.Any(), gomock.Eq(int64(10)))

	tj.job.begin()
	defer tj.job.end()
	childJob := tj.job.newChildJob("f", 10)

	stage := NewDownloadStage(input, output, 4, &async.NoopPanicHandler{})

	msgIDs, expected := buildDownloadStageData(&tj, 56, false)

	go func() {
		stage.run(ctx)
	}()

	require.NoError(t, input.Produce(ctx, DownloadRequest{
		childJob: childJob,
		ids:      msgIDs,
	}))

	out, err := output.Consume(ctx)
	require.NoError(t, err)
	require.Equal(t, expected, out.batch)
	out.onFinished(ctx)
	cancel()

	cachedMessages, cachedAttachments := tj.job.downloadCache.Count()
	require.Zero(t, cachedMessages)
	require.Zero(t, cachedAttachments)
}

func TestDownloadStage_RunWith422(t *testing.T) {
	mockCtrl := gomock.NewController(t)

	input := NewChannelConsumerProducer[DownloadRequest]()
	output := NewChannelConsumerProducer[BuildRequest]()

	ctx, cancel := context.WithCancel(context.Background())

	tj := newTestJob(ctx, mockCtrl, "", map[string]proton.Label{})

	tj.syncReporter.EXPECT().OnProgress(gomock.Any(), gomock.Any())
	tj.state.EXPECT().SetLastMessageID(gomock.Any(), gomock.Eq("f"), gomock.Eq(int64(10))).Return(nil)

	tj.syncReporter.EXPECT().OnProgress(gomock.Any(), gomock.Eq(int64(10)))

	tj.job.begin()
	defer tj.job.end()
	childJob := tj.job.newChildJob("f", 10)

	stage := NewDownloadStage(input, output, 4, &async.NoopPanicHandler{})

	msgIDs, expected := buildDownloadStageData(&tj, 56, true)

	go func() {
		stage.run(ctx)
	}()

	require.NoError(t, input.Produce(ctx, DownloadRequest{
		childJob: childJob,
		ids:      msgIDs,
	}))

	out, err := output.Consume(ctx)
	require.NoError(t, err)
	require.Equal(t, expected, out.batch)
	out.onFinished(ctx)
	cancel()

	cachedMessages, cachedAttachments := tj.job.downloadCache.Count()
	require.Zero(t, cachedMessages)
	require.Zero(t, cachedAttachments)
}

func TestDownloadStage_CancelledJobIsDiscarded(t *testing.T) {
	mockCtrl := gomock.NewController(t)

	input := NewChannelConsumerProducer[DownloadRequest]()
	output := NewChannelConsumerProducer[BuildRequest]()

	ctx, cancel := context.WithCancel(context.Background())

	jobCtx, jobCancel := context.WithCancel(context.Background())

	tj := newTestJob(jobCtx, mockCtrl, "", map[string]proton.Label{})

	tj.job.begin()
	defer tj.job.end()
	childJob := tj.job.newChildJob("f", 10)

	stage := NewDownloadStage(input, output, 4, &async.NoopPanicHandler{})

	go func() {
		stage.run(ctx)
	}()

	jobCancel()

	require.NoError(t, input.Produce(ctx, DownloadRequest{
		childJob: childJob,
		ids:      nil,
	}))

	go func() { cancel() }()

	_, err := output.Consume(context.Background())
	require.ErrorIs(t, err, ErrNoMoreInput)
}

func TestDownloadStage_JobAbortsOnMessageDownloadError(t *testing.T) {
	mockCtrl := gomock.NewController(t)

	input := NewChannelConsumerProducer[DownloadRequest]()
	output := NewChannelConsumerProducer[BuildRequest]()

	ctx, cancel := context.WithCancel(context.Background())

	jobCtx, jobCancel := context.WithCancel(context.Background())
	defer jobCancel()

	expectedErr := errors.New("fail")

	tj := newTestJob(jobCtx, mockCtrl, "", map[string]proton.Label{})
	tj.client.EXPECT().GetMessage(gomock.Any(), gomock.Any()).Return(proton.Message{}, expectedErr)

	tj.job.begin()
	childJob := tj.job.newChildJob("f", 10)
	tj.job.end()

	stage := NewDownloadStage(input, output, 4, &async.NoopPanicHandler{})

	go func() {
		stage.run(ctx)
	}()

	require.NoError(t, input.Produce(ctx, DownloadRequest{
		childJob: childJob,
		ids:      []string{"foo"},
	}))

	err := tj.job.waitAndClose(ctx)
	require.Equal(t, expectedErr, err)

	cancel()

	_, err = output.Consume(context.Background())
	require.ErrorIs(t, err, ErrNoMoreInput)
}

func TestDownloadStage_JobAbortsOnAttachmentDownloadError(t *testing.T) {
	mockCtrl := gomock.NewController(t)

	input := NewChannelConsumerProducer[DownloadRequest]()
	output := NewChannelConsumerProducer[BuildRequest]()

	ctx, cancel := context.WithCancel(context.Background())

	jobCtx, jobCancel := context.WithCancel(context.Background())
	defer jobCancel()

	expectedErr := errors.New("fail")

	tj := newTestJob(jobCtx, mockCtrl, "", map[string]proton.Label{})
	tj.client.EXPECT().GetMessage(gomock.Any(), gomock.Any()).Return(proton.Message{
		MessageMetadata: proton.MessageMetadata{
			ID: "msg",
		},
		Header:   "",
		Body:     "",
		MIMEType: "",
		Attachments: []proton.Attachment{{
			ID: "attach",
		}},
	}, nil)
	tj.client.EXPECT().GetAttachmentInto(gomock.Any(), gomock.Eq("attach"), gomock.Any()).Return(expectedErr)

	tj.job.begin()
	childJob := tj.job.newChildJob("f", 10)
	tj.job.end()

	stage := NewDownloadStage(input, output, 4, &async.NoopPanicHandler{})

	go func() {
		stage.run(ctx)
	}()

	require.NoError(t, input.Produce(ctx, DownloadRequest{
		childJob: childJob,
		ids:      []string{"foo"},
	}))

	err := tj.job.waitAndClose(ctx)
	require.Equal(t, expectedErr, err)

	cancel()

	_, err = output.Consume(context.Background())
	require.ErrorIs(t, err, ErrNoMoreInput)
}

func buildDownloadStageData(tj *tjob, numMessages int, with422 bool) ([]string, []proton.FullMessage) {
	result := make([]proton.FullMessage, numMessages)
	msgIDs := make([]string, numMessages)

	for i := 0; i < numMessages; i++ {
		msgID := fmt.Sprintf("msg-%v", i)
		msgIDs[i] = msgID
		result[i] = proton.FullMessage{
			Message: proton.Message{
				MessageMetadata: proton.MessageMetadata{
					ID:   msgID,
					Size: len([]byte(msgID)),
				},
				Header:      "",
				Body:        msgID,
				MIMEType:    "",
				Attachments: nil,
			},
			AttData: nil,
		}

		buildDownloadStageAttachments(&result[i], i)
	}

	for i, m := range result {
		if with422 && i%2 == 0 {
			tj.client.EXPECT().GetMessage(gomock.Any(), gomock.Eq(m.ID)).Return(proton.Message{}, &proton.APIError{Status: 422})
			continue
		}

		tj.client.EXPECT().GetMessage(gomock.Any(), gomock.Eq(m.ID)).Return(m.Message, nil)

		for idx, a := range m.Attachments {
			attData := m.AttData[idx]
			tj.client.EXPECT().GetAttachmentInto(gomock.Any(), gomock.Eq(a.ID), gomock.Any()).DoAndReturn(
				func(_ context.Context, _ string, b *bytes.Buffer) error {
					_, err := b.Write(attData)
					return err
				},
			)
		}
	}

	if with422 {
		result422 := make([]proton.FullMessage, 0, numMessages/2)

		for i := 0; i < numMessages; i++ {
			if i%2 == 0 {
				continue
			}
			result422 = append(result422, result[i])
		}

		return msgIDs, result422
	}

	return msgIDs, result
}

func buildDownloadStageAttachments(msg *proton.FullMessage, index int) {
	mod := index % 4
	if mod == 0 {
		return
	}

	genDownloadStageAttachmentInfo(msg, index, mod)
}

func genDownloadStageAttachmentInfo(msg *proton.FullMessage, msgIdx int, count int) {
	msg.Attachments = make([]proton.Attachment, count)
	msg.AttData = make([][]byte, count)

	for i := 0; i < count; i++ {
		data := fmt.Sprintf("msg-%v-att-%v", msgIdx, i)
		msg.Attachments[i] = proton.Attachment{
			ID:   data,
			Size: int64(len([]byte(data))),
		}
		msg.AttData[i] = []byte(data)
		msg.Size += len([]byte(data))
	}
}

func autoScaleCoolDown() network.CoolDownProvider {
	return &network.NoCoolDown{}
}

func buildDownloadScaleData(count int) []string {
	r := make([]string, count)
	for i := 0; i < count; i++ {
		r[i] = fmt.Sprintf("m%v", i)
	}

	return r
}

func newDownloadScaleMessage(id string) proton.Message {
	return proton.Message{
		MessageMetadata: proton.MessageMetadata{ID: id},
	}
}

func autoDownloadScaleClientDoAndReturn(_ context.Context, id string) (proton.Message, error) {
	return newDownloadScaleMessage(id), nil
}
