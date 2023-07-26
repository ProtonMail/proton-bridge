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

package user

import (
	"context"
	"fmt"
	"io"
	"strings"
	"testing"
	"time"

	"github.com/ProtonMail/gluon/async"
	"github.com/ProtonMail/go-proton-api"
	"github.com/ProtonMail/proton-bridge/v3/internal/user/mocks"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/require"
)

func TestSyncDownloader_Stage1_429(t *testing.T) {
	// Check 429 is correctly caught and download state recorded correctly
	// Message 1: All ok
	// Message 2: Message failed
	// Message 3: One attachment failed.
	mockCtrl := gomock.NewController(t)
	messageDownloader := mocks.NewMockMessageDownloader(mockCtrl)
	panicHandler := &async.NoopPanicHandler{}
	ctx := context.Background()

	requests := downloadRequest{
		ids:          []string{"Msg1", "Msg2", "Msg3"},
		expectedSize: 0,
		err:          nil,
	}

	messageDownloader.EXPECT().GetMessage(gomock.Any(), gomock.Eq("Msg1")).Times(1).Return(proton.Message{
		MessageMetadata: proton.MessageMetadata{
			ID:             "MsgID1",
			NumAttachments: 1,
		},
		Attachments: []proton.Attachment{
			{
				ID: "Attachment1_1",
			},
		},
	}, nil)

	messageDownloader.EXPECT().GetMessage(gomock.Any(), gomock.Eq("Msg2")).Times(1).Return(proton.Message{}, &proton.APIError{Status: 429})
	messageDownloader.EXPECT().GetMessage(gomock.Any(), gomock.Eq("Msg3")).Times(1).Return(proton.Message{
		MessageMetadata: proton.MessageMetadata{
			ID:             "MsgID3",
			NumAttachments: 2,
		},
		Attachments: []proton.Attachment{
			{
				ID: "Attachment3_1",
			},
			{
				ID: "Attachment3_2",
			},
		},
	}, nil)

	const attachmentData = "attachment data"

	messageDownloader.EXPECT().GetAttachmentInto(gomock.Any(), gomock.Eq("Attachment1_1"), gomock.Any()).Times(1).DoAndReturn(func(_ context.Context, _ string, r io.ReaderFrom) error {
		_, err := r.ReadFrom(strings.NewReader(attachmentData))
		return err
	})

	messageDownloader.EXPECT().GetAttachmentInto(gomock.Any(), gomock.Eq("Attachment3_1"), gomock.Any()).Times(1).Return(&proton.APIError{Status: 429})
	messageDownloader.EXPECT().GetAttachmentInto(gomock.Any(), gomock.Eq("Attachment3_2"), gomock.Any()).Times(1).DoAndReturn(func(_ context.Context, _ string, r io.ReaderFrom) error {
		_, err := r.ReadFrom(strings.NewReader(attachmentData))
		return err
	})

	cache := newSyncDownloadCache()
	attachmentDownloader := newAttachmentDownloader(ctx, panicHandler, messageDownloader, cache, 1)
	defer attachmentDownloader.close()

	result, err := downloadMessageStage1(ctx, panicHandler, requests, messageDownloader, attachmentDownloader, cache, 1)
	require.NoError(t, err)
	require.Equal(t, 3, len(result))
	// Check message 1
	require.Equal(t, result[0].State, downloadStateFinished)
	require.Equal(t, result[0].Message.ID, "MsgID1")
	require.NotEmpty(t, result[0].Message.AttData)
	require.NotEqual(t, attachmentData, result[0].Message.AttData[0])
	require.NotNil(t, result[0].Message.AttData[0])
	require.Nil(t, result[0].err)

	// Check message 2
	require.Equal(t, result[1].State, downloadStateZero)
	require.Empty(t, result[1].Message.ID)
	require.NotNil(t, result[1].err)

	require.Equal(t, result[2].State, downloadStateHasMessage)
	require.Equal(t, result[2].Message.ID, "MsgID3")
	require.Equal(t, 2, len(result[2].Message.AttData))
	require.NotNil(t, result[2].err)
	require.Nil(t, result[2].Message.AttData[0])
	require.NotEqual(t, attachmentData, result[2].Message.AttData[1])
	require.NotNil(t, result[2].err)

	_, ok := cache.GetMessage("MsgID1")
	require.True(t, ok)
	_, ok = cache.GetMessage("MsgID3")
	require.True(t, ok)
	att, ok := cache.GetAttachment("Attachment1_1")
	require.True(t, ok)
	require.Equal(t, attachmentData, string(att))
}

func TestSyncDownloader_Stage2_Everything200(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	messageDownloader := mocks.NewMockMessageDownloader(mockCtrl)
	ctx := context.Background()
	cache := newSyncDownloadCache()

	downloadResult := []downloadResult{
		{
			ID:    "Msg1",
			State: downloadStateFinished,
		},
		{
			ID:    "Msg2",
			State: downloadStateFinished,
		},
	}

	result, err := downloadMessagesStage2(ctx, downloadResult, messageDownloader, cache, time.Millisecond)
	require.NoError(t, err)
	require.Equal(t, 2, len(result))
}

func TestSyncDownloader_Stage2_Not429(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	messageDownloader := mocks.NewMockMessageDownloader(mockCtrl)
	ctx := context.Background()
	cache := newSyncDownloadCache()

	msgErr := fmt.Errorf("something not 429")
	downloadResult := []downloadResult{
		{
			ID:    "Msg1",
			State: downloadStateFinished,
		},
		{
			ID:    "Msg2",
			State: downloadStateHasMessage,
			err:   msgErr,
		},
		{
			ID:    "Msg3",
			State: downloadStateFinished,
		},
	}

	_, err := downloadMessagesStage2(ctx, downloadResult, messageDownloader, cache, time.Millisecond)
	require.Error(t, err)
	require.Equal(t, msgErr, err)
}

func TestSyncDownloader_Stage2_API500(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	messageDownloader := mocks.NewMockMessageDownloader(mockCtrl)
	ctx := context.Background()
	cache := newSyncDownloadCache()

	msgErr := &proton.APIError{Status: 500}
	downloadResult := []downloadResult{
		{
			ID:    "Msg2",
			State: downloadStateHasMessage,
			err:   msgErr,
		},
		{
			ID:    "Msg3",
			State: downloadStateFinished,
		},
	}

	_, err := downloadMessagesStage2(ctx, downloadResult, messageDownloader, cache, time.Millisecond)
	require.Error(t, err)
	require.Equal(t, msgErr, err)
}

func TestSyncDownloader_Stage2_Some429(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	messageDownloader := mocks.NewMockMessageDownloader(mockCtrl)
	ctx := context.Background()
	cache := newSyncDownloadCache()

	const attachmentData1 = "attachment data 1"
	const attachmentData2 = "attachment data 2"
	const attachmentData3 = "attachment data 3"
	const attachmentData4 = "attachment data 4"

	err429 := &proton.APIError{Status: 429}
	downloadResult := []downloadResult{
		{
			// Full message , but missing 1 of 2 attachments
			ID: "Msg1",
			Message: proton.FullMessage{
				Message: proton.Message{
					MessageMetadata: proton.MessageMetadata{
						ID:             "Msg1",
						NumAttachments: 2,
					},
					Attachments: []proton.Attachment{
						{
							ID: "A3",
						},
						{
							ID: "A4",
						},
					},
				},
				AttData: [][]byte{
					nil,
					[]byte(attachmentData4),
				},
			},
			State: downloadStateHasMessage,
			err:   err429,
		},
		{
			// Full message, but missing all attachments
			ID: "Msg2",
			Message: proton.FullMessage{
				Message: proton.Message{
					MessageMetadata: proton.MessageMetadata{
						ID:             "Msg2",
						NumAttachments: 2,
					},
					Attachments: []proton.Attachment{
						{
							ID: "A1",
						},
						{
							ID: "A2",
						},
					},
				},
				AttData: nil,
			},
			State: downloadStateHasMessage,
			err:   err429,
		},
		{
			// Missing everything
			ID:    "Msg3",
			State: downloadStateZero,
			Message: proton.FullMessage{
				Message: proton.Message{MessageMetadata: proton.MessageMetadata{ID: "Msg3"}},
			},
			err: err429,
		},
	}

	{
		// Simulate 2 failures for message 3 body.
		firstCall := messageDownloader.EXPECT().GetMessage(gomock.Any(), gomock.Eq("Msg3")).Times(2).Return(proton.Message{}, err429)
		messageDownloader.EXPECT().GetMessage(gomock.Any(), gomock.Eq("Msg3")).After(firstCall).Times(1).Return(proton.Message{
			MessageMetadata: proton.MessageMetadata{
				ID: "Msg3",
			},
		}, nil)
	}

	{
		// Simulate failures for message 2 attachments.
		firstCall := messageDownloader.EXPECT().GetAttachmentInto(gomock.Any(), gomock.Eq("A1"), gomock.Any()).Times(2).Return(err429)
		messageDownloader.EXPECT().GetAttachmentInto(gomock.Any(), gomock.Eq("A1"), gomock.Any()).After(firstCall).Times(1).DoAndReturn(func(_ context.Context, _ string, r io.ReaderFrom) error {
			_, err := r.ReadFrom(strings.NewReader(attachmentData1))
			return err
		})
		messageDownloader.EXPECT().GetAttachmentInto(gomock.Any(), gomock.Eq("A2"), gomock.Any()).Times(1).DoAndReturn(func(_ context.Context, _ string, r io.ReaderFrom) error {
			_, err := r.ReadFrom(strings.NewReader(attachmentData2))
			return err
		})
	}

	{
		messageDownloader.EXPECT().GetAttachmentInto(gomock.Any(), gomock.Eq("A3"), gomock.Any()).Times(1).DoAndReturn(func(_ context.Context, _ string, r io.ReaderFrom) error {
			_, err := r.ReadFrom(strings.NewReader(attachmentData3))
			return err
		})
	}

	messages, err := downloadMessagesStage2(ctx, downloadResult, messageDownloader, cache, time.Millisecond)
	require.NoError(t, err)
	require.Equal(t, 3, len(messages))

	require.Equal(t, messages[0].Message.ID, "Msg1")
	require.Equal(t, messages[1].Message.ID, "Msg2")
	require.Equal(t, messages[2].Message.ID, "Msg3")

	// check attachments
	require.Equal(t, attachmentData3, string(messages[0].AttData[0]))
	require.Equal(t, attachmentData4, string(messages[0].AttData[1]))
	require.Equal(t, attachmentData1, string(messages[1].AttData[0]))
	require.Equal(t, attachmentData2, string(messages[1].AttData[1]))
	require.Empty(t, messages[2].AttData)

	_, ok := cache.GetMessage("Msg3")
	require.True(t, ok)

	att3, ok := cache.GetAttachment("A3")
	require.True(t, ok)
	require.Equal(t, attachmentData3, string(att3))

	att1, ok := cache.GetAttachment("A1")
	require.True(t, ok)
	require.Equal(t, attachmentData1, string(att1))

	att2, ok := cache.GetAttachment("A2")
	require.True(t, ok)
	require.Equal(t, attachmentData2, string(att2))
}

func TestSyncDownloader_Stage2_ErrorOnNon429MessageDownload(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	messageDownloader := mocks.NewMockMessageDownloader(mockCtrl)
	ctx := context.Background()
	cache := newSyncDownloadCache()

	err429 := &proton.APIError{Status: 429}
	err500 := &proton.APIError{Status: 500}
	downloadResult := []downloadResult{
		{
			// Missing everything
			ID:    "Msg3",
			State: downloadStateZero,
			Message: proton.FullMessage{
				Message: proton.Message{MessageMetadata: proton.MessageMetadata{ID: "Msg3"}},
			},
			err: err429,
		},
		{
			// Full message , but missing 1 of 2 attachments
			ID: "Msg1",
			Message: proton.FullMessage{
				Message: proton.Message{
					MessageMetadata: proton.MessageMetadata{
						ID:             "Msg1",
						NumAttachments: 2,
					},
					Attachments: []proton.Attachment{
						{
							ID: "A3",
						},
						{
							ID: "A4",
						},
					},
				},
			},
			State: downloadStateHasMessage,
			err:   err429,
		},
	}

	{
		// Simulate 2 failures for message 3 body,
		messageDownloader.EXPECT().GetMessage(gomock.Any(), gomock.Eq("Msg3")).Times(1).Return(proton.Message{}, err500)
	}

	messages, err := downloadMessagesStage2(ctx, downloadResult, messageDownloader, cache, time.Millisecond)
	require.Error(t, err)
	require.Empty(t, 0, messages)
}

func TestSyncDownloader_Stage2_ErrorOnNon429AttachmentDownload(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	messageDownloader := mocks.NewMockMessageDownloader(mockCtrl)
	ctx := context.Background()
	cache := newSyncDownloadCache()

	err429 := &proton.APIError{Status: 429}
	err500 := &proton.APIError{Status: 500}
	downloadResult := []downloadResult{
		{
			// Full message , but missing 1 of 2 attachments
			ID: "Msg1",
			Message: proton.FullMessage{
				Message: proton.Message{
					MessageMetadata: proton.MessageMetadata{
						ID:             "Msg1",
						NumAttachments: 2,
					},
					Attachments: []proton.Attachment{
						{
							ID: "A3",
						},
						{
							ID: "A4",
						},
					},
				},
			},
			State: downloadStateHasMessage,
			err:   err429,
		},
	}

	// 429 for first attachment
	messageDownloader.EXPECT().GetAttachmentInto(gomock.Any(), gomock.Eq("A3"), gomock.Any()).Times(1).Return(err429)
	// 500 for second attachment
	messageDownloader.EXPECT().GetAttachmentInto(gomock.Any(), gomock.Eq("A4"), gomock.Any()).Times(1).Return(err500)

	messages, err := downloadMessagesStage2(ctx, downloadResult, messageDownloader, cache, time.Millisecond)
	require.Error(t, err)
	require.Empty(t, 0, messages)
}

func TestSyncDownloader_Stage1_DoNotDownloadIfAlreadyInCache(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	messageDownloader := mocks.NewMockMessageDownloader(mockCtrl)
	panicHandler := &async.NoopPanicHandler{}
	ctx := context.Background()

	requests := downloadRequest{
		ids:          []string{"Msg1", "Msg3"},
		expectedSize: 0,
		err:          nil,
	}

	cache := newSyncDownloadCache()
	attachmentDownloader := newAttachmentDownloader(ctx, panicHandler, messageDownloader, cache, 1)
	defer attachmentDownloader.close()

	const attachmentData = "attachment data"

	cache.StoreMessage(proton.Message{MessageMetadata: proton.MessageMetadata{ID: "Msg1", NumAttachments: 1}, Attachments: []proton.Attachment{{ID: "A1"}}})
	cache.StoreMessage(proton.Message{MessageMetadata: proton.MessageMetadata{ID: "Msg3", NumAttachments: 2}, Attachments: []proton.Attachment{{ID: "A2"}}})

	cache.StoreAttachment("A1", []byte(attachmentData))
	cache.StoreAttachment("A2", []byte(attachmentData))

	result, err := downloadMessageStage1(ctx, panicHandler, requests, messageDownloader, attachmentDownloader, cache, 1)
	require.NoError(t, err)
	require.Equal(t, 2, len(result))

	require.Equal(t, result[0].State, downloadStateFinished)
	require.Equal(t, result[0].Message.ID, "Msg1")
	require.NotEmpty(t, result[0].Message.AttData)
	require.NotEqual(t, attachmentData, result[0].Message.AttData[0])
	require.NotNil(t, result[0].Message.AttData[0])
	require.Nil(t, result[0].err)

	require.Equal(t, result[1].State, downloadStateFinished)
	require.Equal(t, result[1].Message.ID, "Msg3")
	require.NotEmpty(t, result[1].Message.AttData)
	require.NotEqual(t, attachmentData, result[1].Message.AttData[0])
	require.NotNil(t, result[1].Message.AttData[0])
	require.Nil(t, result[1].err)
}
