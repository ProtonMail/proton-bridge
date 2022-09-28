package pool

import (
	"context"
	"errors"
	"sync"

	"github.com/ProtonMail/gluon/queue"
)

// ErrJobCancelled indicates the job was cancelled.
var ErrJobCancelled = errors.New("Job cancelled by surrounding context")

// Pool is a worker pool that handles input of type In and returns results of type Out.
type Pool[In comparable, Out any] struct {
	queue *queue.QueuedChannel[*job[In, Out]]
	size  int
}

// doneFunc must be called to free up pool resources.
type doneFunc func()

// New returns a new pool.
func New[In comparable, Out any](size int, work func(context.Context, In) (Out, error)) *Pool[In, Out] {
	queue := queue.NewQueuedChannel[*job[In, Out]](0, 0)

	for i := 0; i < size; i++ {
		go func() {
			for job := range queue.GetChannel() {
				select {
				case <-job.ctx.Done():
					job.postFailure(ErrJobCancelled)

				default:
					res, err := work(job.ctx, job.req)
					if err != nil {
						job.postFailure(err)
					} else {
						job.postSuccess(res)
					}

					job.waitDone()
				}
			}
		}()
	}

	return &Pool[In, Out]{
		queue: queue,
		size:  size,
	}
}

// Process submits jobs to the pool. The callback provides access to the result, or an error if one occurred.
func (pool *Pool[In, Out]) Process(ctx context.Context, reqs []In, fn func(In, Out, error) error) error {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	var (
		wg      sync.WaitGroup
		errList []error
		lock    sync.Mutex
	)

	for _, req := range reqs {
		req := req

		wg.Add(1)

		go func() {
			defer wg.Done()

			job, done := pool.newJob(ctx, req)
			defer done()

			res, err := job.result()

			if err := fn(req, res, err); err != nil {
				lock.Lock()
				defer lock.Unlock()

				// Cancel ongoing jobs.
				cancel()

				// Collect the error.
				errList = append(errList, err)
			}
		}()
	}

	wg.Wait()

	// TODO: Join the errors somehow?
	if len(errList) > 0 {
		return errList[0]
	}

	return nil
}

// ProcessAll submits jobs to the pool. All results are returned once available.
func (pool *Pool[In, Out]) ProcessAll(ctx context.Context, reqs []In) (map[In]Out, error) {
	var (
		data = make(map[In]Out)
		lock = sync.Mutex{}
	)

	if err := pool.Process(ctx, reqs, func(req In, res Out, err error) error {
		if err != nil {
			return err
		}

		lock.Lock()
		defer lock.Unlock()

		data[req] = res

		return nil
	}); err != nil {
		return nil, err
	}

	return data, nil
}

// ProcessOne submits one job to the pool and returns the result.
func (pool *Pool[In, Out]) ProcessOne(ctx context.Context, req In) (Out, error) {
	job, done := pool.newJob(ctx, req)
	defer done()

	return job.result()
}

func (pool *Pool[In, Out]) Done() {
	pool.queue.Close()
}

// newJob submits a job to the pool. It returns a job handle and a DoneFunc.
// The job handle allows the job result to be obtained. The DoneFunc is used to mark the job as done,
// which frees up the worker in the pool for reuse.
func (pool *Pool[In, Out]) newJob(ctx context.Context, req In) (*job[In, Out], doneFunc) {
	job := newJob[In, Out](ctx, req)

	pool.queue.Enqueue(job)

	return job, func() { close(job.done) }
}
