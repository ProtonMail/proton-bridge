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

package imapservice

import (
	"bytes"
	"context"
	"fmt"

	"github.com/ProtonMail/gluon/async"
	"github.com/ProtonMail/gluon/logging"
	"github.com/ProtonMail/go-proton-api"
)

type attachmentResult struct {
	attachment []byte
	err        error
}

type attachmentJob struct {
	id     string
	size   int64
	result chan attachmentResult
}

type attachmentDownloader struct {
	workerCh chan attachmentJob
	cancel   context.CancelFunc
}

func attachmentWorker(ctx context.Context, client APIClient, work <-chan attachmentJob) {
	for {
		select {
		case <-ctx.Done():
			return
		case job, ok := <-work:
			if !ok {
				return
			}
			var b bytes.Buffer
			b.Grow(int(job.size))
			err := client.GetAttachmentInto(ctx, job.id, &b)
			select {
			case <-ctx.Done():
				close(job.result)
				return
			case job.result <- attachmentResult{attachment: b.Bytes(), err: err}:
				close(job.result)
			}
		}
	}
}

func (s *syncJob) newAttachmentDownloader(ctx context.Context, client APIClient, workerCount int) *attachmentDownloader {
	workerCh := make(chan attachmentJob, (workerCount+2)*workerCount)
	ctx, cancel := context.WithCancel(ctx)
	for i := 0; i < workerCount; i++ {
		workerCh = make(chan attachmentJob)
		async.GoAnnotated(ctx, s.panicHandler, func(ctx context.Context) { attachmentWorker(ctx, client, workerCh) }, logging.Labels{
			"sync": fmt.Sprintf("att-downloader %v", i),
		})
	}

	return &attachmentDownloader{
		workerCh: workerCh,
		cancel:   cancel,
	}
}

func (a *attachmentDownloader) getAttachments(ctx context.Context, attachments []proton.Attachment) ([][]byte, error) {
	resultChs := make([]chan attachmentResult, len(attachments))
	for i, id := range attachments {
		resultChs[i] = make(chan attachmentResult, 1)
		select {
		case a.workerCh <- attachmentJob{id: id.ID, result: resultChs[i], size: id.Size}:
		case <-ctx.Done():
			return nil, ctx.Err()
		}
	}

	result := make([][]byte, len(attachments))
	var err error
	for i := 0; i < len(attachments); i++ {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case r := <-resultChs[i]:
			if r.err != nil {
				err = fmt.Errorf("failed to get attachment %v: %w", attachments[i], r.err)
			}
			result[i] = r.attachment
		}
	}

	return result, err
}

func (a *attachmentDownloader) close() {
	a.cancel()
}
