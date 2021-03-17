// Copyright (c) 2021 Proton Technologies AG
//
// This file is part of ProtonMail Bridge.
//
// ProtonMail Bridge is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// ProtonMail Bridge is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with ProtonMail Bridge.  If not, see <https://www.gnu.org/licenses/>.

package message

import (
	"context"
	"io"
	"sync"

	"github.com/ProtonMail/gopenpgp/v2/crypto"
	"github.com/ProtonMail/proton-bridge/pkg/pmapi"
	"github.com/pkg/errors"
)

var (
	ErrDecryptionFailed = errors.New("message could not be decrypted")
	ErrNoSuchKeyRing    = errors.New("the keyring to decrypt this message could not be found")
)

type Builder struct {
	reqs   chan fetchReq
	done   chan struct{}
	jobs   map[string]*BuildJob
	locker sync.Mutex
}

type Fetcher interface {
	GetMessage(string) (*pmapi.Message, error)
	GetAttachment(string) (io.ReadCloser, error)
	KeyRingForAddressID(string) (*crypto.KeyRing, error)
}

func NewBuilder(fetchWorkers, attachWorkers, buildWorkers int) *Builder {
	b := newBuilder()

	fetchReqCh, fetchResCh := startFetchWorkers(fetchWorkers, attachWorkers)
	buildReqCh, buildResCh := startBuildWorkers(buildWorkers)

	go func() {
		defer close(fetchReqCh)

		for {
			select {
			case req := <-b.reqs:
				fetchReqCh <- req

			case <-b.done:
				return
			}
		}
	}()

	go func() {
		defer close(buildReqCh)

		for res := range fetchResCh {
			if res.err != nil {
				b.jobFailure(res.messageID, res.err)
			} else {
				buildReqCh <- res
			}
		}
	}()

	go func() {
		for res := range buildResCh {
			if res.err != nil {
				b.jobFailure(res.messageID, res.err)
			} else {
				b.jobSuccess(res.messageID, res.literal)
			}
		}
	}()

	return b
}

func newBuilder() *Builder {
	return &Builder{
		reqs: make(chan fetchReq),
		done: make(chan struct{}),
		jobs: make(map[string]*BuildJob),
	}
}

func (b *Builder) NewJob(ctx context.Context, api Fetcher, messageID string) *BuildJob {
	return b.NewJobWithOptions(ctx, api, messageID, JobOptions{})
}

func (b *Builder) NewJobWithOptions(ctx context.Context, api Fetcher, messageID string, opts JobOptions) *BuildJob {
	b.locker.Lock()
	defer b.locker.Unlock()

	if job, ok := b.jobs[messageID]; ok {
		return job
	}

	b.jobs[messageID] = newBuildJob(messageID)

	go func() { b.reqs <- fetchReq{ctx: ctx, api: api, messageID: messageID, opts: opts} }()

	return b.jobs[messageID]
}

func (b *Builder) Done() {
	b.locker.Lock()
	defer b.locker.Unlock()

	close(b.done)
}

func (b *Builder) jobSuccess(messageID string, literal []byte) {
	b.locker.Lock()
	defer b.locker.Unlock()

	b.jobs[messageID].postSuccess(literal)

	delete(b.jobs, messageID)
}

func (b *Builder) jobFailure(messageID string, err error) {
	b.locker.Lock()
	defer b.locker.Unlock()

	b.jobs[messageID].postFailure(err)

	delete(b.jobs, messageID)
}
