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
	"errors"
	"fmt"
	"math"
	"runtime"
	"testing"
	"time"

	r "github.com/stretchr/testify/require"
)

//nolint:gochecknoglobals
var (
	testInput               = []interface{}{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}
	wantOutput              = []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}
	testProcessSleep        = 100 // ms
	runParallelTimeOverhead = 150 // ms
	windowsCIExtra          = 500 // ms - estimated experimentally
)

func TestParallel(t *testing.T) {
	workersTests := []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}

	for _, workers := range workersTests {
		workers := workers
		t.Run(fmt.Sprintf("%d", workers), func(t *testing.T) {
			collected := make([]int, 0)
			collect := func(idx int, value interface{}) error {
				collected = append(collected, value.(int)) //nolint:forcetypeassert
				return nil
			}

			tstart := time.Now()
			err := RunParallel(workers, testInput, processSleep, collect)
			duration := time.Since(tstart)

			r.Nil(t, err)
			r.Equal(t, wantOutput, collected) // Check the order is always kept.

			wantMinDuration := int(math.Ceil(float64(len(testInput))/float64(workers))) * testProcessSleep
			wantMaxDuration := wantMinDuration + runParallelTimeOverhead
			if runtime.GOOS == "windows" {
				wantMaxDuration += windowsCIExtra
			}
			r.True(t, duration.Nanoseconds() > int64(wantMinDuration*1000000), "Duration too short: %v (expected: %v)", duration, wantMinDuration)
			r.True(t, duration.Nanoseconds() < int64(wantMaxDuration*1000000), "Duration too long: %v (expected: %v)", duration, wantMaxDuration)
		})
	}
}

func TestParallelEmptyInput(t *testing.T) {
	workersTests := []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}

	for _, workers := range workersTests {
		workers := workers
		t.Run(fmt.Sprintf("%d", workers), func(t *testing.T) {
			err := RunParallel(workers, []interface{}{}, processSleep, collectNil)
			r.Nil(t, err)
		})
	}
}

func TestParallelErrorInProcess(t *testing.T) {
	workersTests := []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}

	for _, workers := range workersTests {
		workers := workers
		t.Run(fmt.Sprintf("%d", workers), func(t *testing.T) {
			var lastCollected int
			process := func(value interface{}) (interface{}, error) {
				time.Sleep(10 * time.Millisecond)
				if value.(int) == 5 { //nolint:forcetypeassert
					return nil, errors.New("Error")
				}
				return value, nil
			}
			collect := func(idx int, value interface{}) error {
				lastCollected = value.(int) //nolint:forcetypeassert
				return nil
			}

			err := RunParallel(workers, testInput, process, collect)
			r.EqualError(t, err, "Error")

			time.Sleep(10 * time.Millisecond)
			r.True(t, lastCollected < 5, "Last collected cannot be higher that 5, got: %d", lastCollected)
		})
	}
}

func TestParallelErrorInCollect(t *testing.T) {
	workersTests := []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}

	for _, workers := range workersTests {
		workers := workers
		t.Run(fmt.Sprintf("%d", workers), func(t *testing.T) {
			collect := func(idx int, value interface{}) error {
				if value.(int) == 5 { //nolint:forcetypeassert
					return errors.New("Error")
				}
				return nil
			}

			err := RunParallel(workers, testInput, processSleep, collect)
			r.EqualError(t, err, "Error")
		})
	}
}

func processSleep(value interface{}) (interface{}, error) {
	time.Sleep(time.Duration(testProcessSleep) * time.Millisecond)
	return value.(int), nil //nolint:forcetypeassert
}

func collectNil(idx int, value interface{}) error {
	return nil
}
