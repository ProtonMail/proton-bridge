// Copyright (c) 2022 Proton AG
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
// along with Proton Mail Bridge. If not, see <https://www.gnu.org/licenses/>.

package parallel

import (
	"sync"
	"time"
)

// parallelJob is to be used for passing items between input, worker and
// collector. `idx` is there to know the original order.
type parallelJob struct {
	idx   int
	value interface{}
}

// RunParallel starts `workers` number of workers and feeds them with `input` data.
// Each worker calls `process`. Processed data is collected in the same order as
// the input and is passed in order to the `collect` callback. If an error
// occurs, the execution is stopped and the error returned.
// runParallel blocks until everything is done.
func RunParallel( //nolint:funlen
	workers int,
	input []interface{},
	process func(interface{}) (interface{}, error),
	collect func(int, interface{}) error,
) (resultError error) {
	wgProcess := &sync.WaitGroup{}
	wgCollect := &sync.WaitGroup{}

	// Optimise by not executing the code at all if there is no input
	// or run less workers than requested if there are few inputs.
	inputLen := len(input)
	if inputLen == 0 {
		return nil
	}
	if inputLen < workers {
		workers = inputLen
	}

	inputChan := make(chan *parallelJob)
	outputChan := make(chan *parallelJob)

	orderedCollectLock := &sync.Mutex{}
	orderedCollect := make(map[int]interface{})

	// Feed input channel used by workers with input data with index for ordering.
	go func() {
		defer close(inputChan)
		for idx, item := range input {
			if resultError != nil {
				break
			}
			inputChan <- &parallelJob{idx, item}
		}
	}()

	// Start workers and process all the inputs.
	wgProcess.Add(workers)
	for i := 0; i < workers; i++ {
		go func() {
			defer wgProcess.Done()
			for item := range inputChan {
				if output, err := process(item.value); err != nil {
					resultError = err
					break
				} else {
					outputChan <- &parallelJob{item.idx, output}
				}
			}
		}()
	}

	// Collect data into map with the original position in the array.
	wgCollect.Add(1)
	go func() {
		defer wgCollect.Done()
		for output := range outputChan {
			orderedCollectLock.Lock()
			orderedCollect[output.idx] = output.value
			orderedCollectLock.Unlock()
		}
	}()

	// Collect data in the same order as in the input array.
	wgCollect.Add(1)
	go func() {
		defer wgCollect.Done()
		idx := 0
		for {
			if idx >= inputLen || resultError != nil {
				break
			}
			orderedCollectLock.Lock()
			value, ok := orderedCollect[idx]
			if ok {
				if err := collect(idx, value); err != nil {
					resultError = err
				}
				delete(orderedCollect, idx)
				idx++
			}
			orderedCollectLock.Unlock()
			if !ok {
				time.Sleep(10 * time.Millisecond)
			}
		}
	}()

	// When input channel is closed, all workers will finish. We need to wait
	// for all of them and close the output channel only once.
	wgProcess.Wait()
	close(outputChan)

	// When workers are done, the last job is to finish collecting data. First
	// collector is finished when output channel is closed and the second one
	// when all items are passed to `collect` in the order or after an error.
	wgCollect.Wait()

	return resultError
}
