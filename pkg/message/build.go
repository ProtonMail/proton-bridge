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
	"io/ioutil"
	"sync"

	"github.com/ProtonMail/gopenpgp/v2/crypto"
	"github.com/ProtonMail/proton-bridge/pkg/pmapi"
	"github.com/ProtonMail/proton-bridge/pkg/pool"
	"github.com/pkg/errors"
)

var (
	ErrDecryptionFailed = errors.New("message could not be decrypted")
	ErrNoSuchKeyRing    = errors.New("the keyring to decrypt this message could not be found")
)

const (
	BackgroundPriority = 1 << iota
	ForegroundPriority
)

type Builder struct {
	pool *pool.Pool
	jobs map[string]*Job
	lock sync.Mutex
}

type Fetcher interface {
	GetMessage(context.Context, string) (*pmapi.Message, error)
	GetAttachment(context.Context, string) (io.ReadCloser, error)
	KeyRingForAddressID(string) (*crypto.KeyRing, error)
}

// NewBuilder creates a new builder which manages the given number of fetch/attach/build workers.
//  - fetchWorkers: the number of workers which fetch messages from API
//  - attachWorkers: the number of workers which fetch attachments from API.
//
// The returned builder is ready to handle jobs -- see (*Builder).NewJob for more information.
//
// Call (*Builder).Done to shut down the builder and stop all workers.
func NewBuilder(fetchWorkers, attachWorkers int) *Builder {
	attacherPool := pool.New(attachWorkers, newAttacherWorkFunc())

	fetcherPool := pool.New(fetchWorkers, newFetcherWorkFunc(attacherPool))

	return &Builder{
		pool: fetcherPool,
		jobs: make(map[string]*Job),
	}
}

func (builder *Builder) NewJob(ctx context.Context, fetcher Fetcher, messageID string, prio int) (*Job, pool.DoneFunc) {
	return builder.NewJobWithOptions(ctx, fetcher, messageID, JobOptions{}, prio)
}

func (builder *Builder) NewJobWithOptions(ctx context.Context, fetcher Fetcher, messageID string, opts JobOptions, prio int) (*Job, pool.DoneFunc) {
	builder.lock.Lock()
	defer builder.lock.Unlock()

	if job, ok := builder.jobs[messageID]; ok {
		if job.GetPriority() < prio {
			job.SetPriority(prio)
		}

		return job, job.done
	}

	job, done := builder.pool.NewJob(
		&fetchReq{
			fetcher:   fetcher,
			messageID: messageID,
			options:   opts,
		},
		prio,
	)

	buildJob := &Job{
		Job:  job,
		done: done,
	}

	builder.jobs[messageID] = buildJob

	return buildJob, func() {
		builder.lock.Lock()
		defer builder.lock.Unlock()

		// Remove the job from the builder.
		delete(builder.jobs, messageID)

		// And mark it as done.
		done()
	}
}

func (builder *Builder) Done() {
	// NOTE(GODT-1158): Stop worker pool.
}

type fetchReq struct {
	fetcher   Fetcher
	messageID string
	options   JobOptions
}

type attachReq struct {
	fetcher Fetcher
	message *pmapi.Message
}

type Job struct {
	*pool.Job

	done pool.DoneFunc
}

func (job *Job) GetResult() ([]byte, error) {
	res, err := job.Job.GetResult()
	if err != nil {
		return nil, err
	}

	return res.([]byte), nil
}

func newAttacherWorkFunc() pool.WorkFunc {
	return func(payload interface{}, prio int) (interface{}, error) {
		req, ok := payload.(*attachReq)
		if !ok {
			panic("bad payload type")
		}

		res := make(map[string][]byte)

		for _, att := range req.message.Attachments {
			rc, err := req.fetcher.GetAttachment(context.Background(), att.ID)
			if err != nil {
				return nil, err
			}

			b, err := ioutil.ReadAll(rc)
			if err != nil {
				return nil, err
			}

			if err := rc.Close(); err != nil {
				return nil, err
			}

			res[att.ID] = b
		}

		return res, nil
	}
}

func newFetcherWorkFunc(attacherPool *pool.Pool) pool.WorkFunc {
	return func(payload interface{}, prio int) (interface{}, error) {
		req, ok := payload.(*fetchReq)
		if !ok {
			panic("bad payload type")
		}

		msg, err := req.fetcher.GetMessage(context.Background(), req.messageID)
		if err != nil {
			return nil, err
		}

		attJob, attDone := attacherPool.NewJob(&attachReq{
			fetcher: req.fetcher,
			message: msg,
		}, prio)
		defer attDone()

		val, err := attJob.GetResult()
		if err != nil {
			return nil, err
		}

		attData, ok := val.(map[string][]byte)
		if !ok {
			panic("bad response type")
		}

		kr, err := req.fetcher.KeyRingForAddressID(msg.AddressID)
		if err != nil {
			return nil, ErrNoSuchKeyRing
		}

		return buildRFC822(kr, msg, attData, req.options)
	}
}
