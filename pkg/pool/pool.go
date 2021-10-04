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

package pool

import (
	"sync"

	"github.com/ProtonMail/proton-bridge/pkg/pchan"
)

type WorkFunc func(interface{}, int) (interface{}, error)

type DoneFunc func()

type Pool struct {
	jobCh *pchan.PChan
}

func New(size int, work WorkFunc) *Pool {
	jobCh := pchan.New()

	for i := 0; i < size; i++ {
		go func() {
			for {
				val, prio, ok := jobCh.Pop()
				if !ok {
					return
				}

				job, ok := val.(*Job)
				if !ok {
					panic("bad result type")
				}

				res, err := work(job.req, prio)
				if err != nil {
					job.postFailure(err)
				} else {
					job.postSuccess(res)
				}

				job.waitDone()
			}
		}()
	}

	return &Pool{jobCh: jobCh}
}

func (pool *Pool) NewJob(req interface{}, prio int) (*Job, DoneFunc) {
	job := newJob(req)

	job.setItem(pool.jobCh.Push(job, prio))

	return job, job.markDone
}

type Job struct {
	req interface{}
	res interface{}
	err error

	item *pchan.Item

	ready, done chan struct{}
	once        sync.Once
}

func newJob(req interface{}) *Job {
	return &Job{
		req:   req,
		ready: make(chan struct{}),
		done:  make(chan struct{}),
	}
}

func (job *Job) GetResult() (interface{}, error) {
	<-job.ready

	return job.res, job.err
}

func (job *Job) GetPriority() int {
	return job.item.GetPriority()
}

func (job *Job) SetPriority(prio int) {
	job.item.SetPriority(prio)
}

func (job *Job) postSuccess(res interface{}) {
	defer close(job.ready)

	job.res = res
}

func (job *Job) postFailure(err error) {
	defer close(job.ready)

	job.err = err
}

func (job *Job) setItem(item *pchan.Item) {
	job.item = item
}

func (job *Job) markDone() {
	job.once.Do(func() { close(job.done) })
}

func (job *Job) waitDone() {
	<-job.done
}
