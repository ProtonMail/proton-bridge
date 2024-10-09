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
	"sync/atomic"

	"github.com/ProtonMail/gluon/async"
	"github.com/ProtonMail/gluon/logging"
	"github.com/ProtonMail/go-proton-api"
	"github.com/ProtonMail/proton-bridge/v3/internal/network"
	"github.com/bradenaw/juniper/parallel"
	"github.com/bradenaw/juniper/xslices"
	"github.com/sirupsen/logrus"
)

type DownloadRequest struct {
	childJob
	ids []string
}

type DownloadStageInput = StageInputConsumer[DownloadRequest]
type DownloadStageOutput = StageOutputProducer[BuildRequest]

// DownloadStage downloads the messages and attachments. It auto-throttles the download of the messages based on
// whether we run into 429|5xx codes.
type DownloadStage struct {
	input                DownloadStageInput
	output               DownloadStageOutput
	maxParallelDownloads int
	panicHandler         async.PanicHandler
	log                  *logrus.Entry
}

func NewDownloadStage(
	input DownloadStageInput,
	output DownloadStageOutput,
	maxParallelDownloads int,
	panicHandler async.PanicHandler,
) *DownloadStage {
	return &DownloadStage{
		input:                input,
		output:               output,
		maxParallelDownloads: maxParallelDownloads * 2,
		panicHandler:         panicHandler,
		log:                  logrus.WithField("sync-stage", "download"),
	}
}

func (d *DownloadStage) Run(group *async.Group) {
	group.Once(func(ctx context.Context) {
		logging.DoAnnotated(ctx, func(ctx context.Context) {
			d.run(ctx)
		}, logging.Labels{"sync-stage": "Download"})
	})
}

func (d *DownloadStage) run(ctx context.Context) {
	defer d.output.Close()

	newCoolDown := func() network.CoolDownProvider {
		return &network.ExpCoolDown{}
	}

	for {
		request, err := d.input.Consume(ctx)
		if err != nil {
			if !(errors.Is(err, ErrNoMoreInput) || errors.Is(err, context.Canceled)) {
				d.log.WithError(err).Error("Exiting state with error")
			}
			return
		}

		if request.checkCancelled() {
			continue
		}

		// Step 1: Download Messages.
		result, err := autoDownloadRate(
			request.getContext(),
			&DefaultDownloadRateModifier{},
			request.job.client,
			d.maxParallelDownloads,
			request.ids,
			newCoolDown,
			func(ctx context.Context, client APIClient, input string) (proton.FullMessage, error) {
				msg, err := downloadMessage(ctx, request.job.downloadCache, client, input)
				if err != nil {
					var apiErr *proton.APIError
					if errors.As(err, &apiErr) && apiErr.Status == 422 {
						return proton.FullMessage{}, nil
					}

					return proton.FullMessage{}, err
				}

				var attData [][]byte

				numAttachments := len(msg.Attachments)

				if numAttachments > 0 {
					attData = make([][]byte, numAttachments)
				}

				return proton.FullMessage{Message: msg, AttData: attData}, nil
			},
		)
		if err != nil {
			request.job.onError(err)
			continue
		}

		// Step 2: Prepare attachment ids for download.
		type attachmentMeta struct {
			msgIdx int
			attIdx int
		}

		// Filter out any messages that don't exist.
		result = xslices.Filter(result, func(t proton.FullMessage) bool {
			return t.ID != ""
		})

		attachmentIndices := make([]attachmentMeta, 0, len(result))
		attachmentIDs := make([]string, 0, len(result))

		for msgIdx, v := range result {
			numAttachments := len(v.Attachments)
			for attIdx := 0; attIdx < numAttachments; attIdx++ {
				attachmentIndices = append(attachmentIndices, attachmentMeta{
					msgIdx: msgIdx,
					attIdx: attIdx,
				})

				attachmentIDs = append(attachmentIDs, result[msgIdx].Attachments[attIdx].ID)
			}
		}

		// Step 3: Download attachments data to the message.
		attachments, err := autoDownloadRate(
			request.getContext(),
			&DefaultDownloadRateModifier{},
			request.job.client,
			d.maxParallelDownloads,
			attachmentIndices,
			newCoolDown,
			func(ctx context.Context, client APIClient, input attachmentMeta) ([]byte, error) {
				attachment := result[input.msgIdx].Attachments[input.attIdx]
				return downloadAttachment(ctx, request.job.downloadCache, client, attachment.ID, attachment.Size)
			},
		)
		if err != nil {
			request.job.onError(err)
			continue
		}

		// Step 4: attach attachment data to the message.
		for i, meta := range attachmentIndices {
			result[meta.msgIdx].AttData[meta.attIdx] = attachments[i]
		}

		request.cachedAttachmentIDs = attachmentIDs
		request.cachedMessageIDs = request.ids

		// Step 5: Publish result.
		request.onStageCompleted(ctx)

		if err := d.output.Produce(ctx, BuildRequest{
			batch:    result,
			childJob: request.childJob,
		}); err != nil {
			request.job.onError(fmt.Errorf("failed to produce output for next stage: %w", err))
		}
	}
}

func downloadMessage(ctx context.Context, cache *DownloadCache, client APIClient, id string) (proton.Message, error) {
	msg, ok := cache.GetMessage(id)
	if ok {
		return msg, nil
	}

	msg, err := client.GetMessage(ctx, id)
	if err != nil {
		return proton.Message{}, err
	}

	cache.StoreMessage(msg)

	return msg, nil
}

func downloadAttachment(ctx context.Context, cache *DownloadCache, client APIClient, id string, size int64) ([]byte, error) {
	data, ok := cache.GetAttachment(id)
	if ok {
		return data, nil
	}

	var buffer bytes.Buffer

	buffer.Grow(int(size))

	if err := client.GetAttachmentInto(ctx, id, &buffer); err != nil {
		return nil, err
	}

	data = buffer.Bytes()

	cache.StoreAttachment(id, data)

	return data, nil
}

type DownloadRateModifier interface {
	Apply(wasSuccess bool, current int, max int) int //nolint:predeclared
}

func autoDownloadRate[T any, R any](
	ctx context.Context,
	modifier DownloadRateModifier,
	client APIClient,
	maxParallelDownloads int,
	data []T,
	newCoolDown func() network.CoolDownProvider,
	f func(ctx context.Context, client APIClient, input T) (R, error),
) ([]R, error) {
	result := make([]R, 0, len(data))

	proton429or5xxCounter := int32(0)
	parallelTasks := maxParallelDownloads
	for _, chunk := range xslices.Chunk(data, maxParallelDownloads) {
		if err := ctx.Err(); err != nil {
			return nil, err
		}

		parallelTasks = modifier.Apply(atomic.LoadInt32(&proton429or5xxCounter) != 0, parallelTasks, maxParallelDownloads)

		atomic.StoreInt32(&proton429or5xxCounter, 0)

		chunkResult, err := parallel.MapContext(
			ctx,
			parallelTasks,
			chunk,
			func(ctx context.Context, in T) (R, error) {
				wrapper := network.NewClientRetryWrapper(client, newCoolDown())
				msg, err := network.RetryWithClient(ctx, wrapper, func(ctx context.Context, c APIClient) (R, error) {
					return f(ctx, c, in)
				})

				if wrapper.DidEncounter429or5xx() {
					atomic.AddInt32(&proton429or5xxCounter, 1)
				}

				return msg, err
			})

		if err != nil {
			return nil, err
		}

		result = append(result, chunkResult...)
	}

	return result, nil
}

type DefaultDownloadRateModifier struct{}

func (d DefaultDownloadRateModifier) Apply(wasSuccess bool, current int, max int) int { //nolint:predeclared
	if !wasSuccess {
		return 2
	}

	parallelTasks := current * 2
	if parallelTasks > max {
		parallelTasks = max
	}

	return parallelTasks
}
