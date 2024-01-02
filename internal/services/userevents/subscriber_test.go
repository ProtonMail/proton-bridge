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

package userevents

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestChanneledSubscriber_CtxTimeoutDoesNotBlockFutureEvents(t *testing.T) {
	wg := sync.WaitGroup{}
	wg.Add(1)

	subscriber := newChanneledSubscriber[int]("test")
	defer subscriber.close()

	go func() {
		defer wg.Done()

		// Send one event, that succeeds.
		require.NoError(t, subscriber.handle(context.Background(), 30))

		// Add an impossible deadline that fails immediately to simulate on event taking too long to send.
		ctx, cancel := context.WithDeadline(context.Background(), time.Now().Add(time.Microsecond))
		defer cancel()

		err := subscriber.handle(ctx, 20)
		require.Error(t, err)
		require.True(t, errors.Is(err, context.DeadlineExceeded))
	}()

	// Receive first event. Notify success.
	event, ok := <-subscriber.OnEventCh()
	require.True(t, ok)
	event.Consume(func(event int) error {
		require.Equal(t, 30, event)
		return nil
	})
	wg.Wait()

	// Simulate reception of another event
	wg.Add(1)
	go func() {
		defer wg.Done()
		require.NoError(t, subscriber.handle(context.Background(), 40))
	}()

	event, ok = <-subscriber.OnEventCh()
	require.True(t, ok)
	event.Consume(func(event int) error {
		require.Equal(t, 40, event)
		return nil
	})

	wg.Wait()
}

func TestChanneledSubscriber_ErrorReported(t *testing.T) {
	wg := sync.WaitGroup{}
	wg.Add(1)

	subscriber := newChanneledSubscriber[int]("test")
	defer subscriber.close()
	reportedErr := fmt.Errorf("request failed")

	go func() {
		defer wg.Done()

		// Send one event, that succeeds.
		err := subscriber.handle(context.Background(), 30)
		require.Error(t, err)
		require.Equal(t, reportedErr, err)
	}()

	// Receive first event. Notify success.
	event, ok := <-subscriber.OnEventCh()
	require.True(t, ok)
	event.Consume(func(event int) error {
		require.Equal(t, 30, event)
		return reportedErr
	})

	wg.Wait()
}
