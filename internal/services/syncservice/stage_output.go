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

package syncservice

import (
	"context"
	"errors"
)

type StageOutputProducer[T any] interface {
	Produce(ctx context.Context, value T) error
	Close()
}

var ErrNoMoreInput = errors.New("no more input")

type StageInputConsumer[T any] interface {
	Consume(ctx context.Context) (T, error)
}

type ChannelConsumerProducer[T any] struct {
	ch chan T
}

func NewChannelConsumerProducer[T any]() *ChannelConsumerProducer[T] {
	return &ChannelConsumerProducer[T]{ch: make(chan T)}
}

func (c ChannelConsumerProducer[T]) Produce(ctx context.Context, value T) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	case c.ch <- value:
		return nil
	}
}

func (c ChannelConsumerProducer[T]) Close() {
	close(c.ch)
}

func (c ChannelConsumerProducer[T]) Consume(ctx context.Context) (T, error) {
	select {
	case <-ctx.Done():
		var t T
		return t, ctx.Err()
	case t, ok := <-c.ch:
		if !ok {
			return t, ErrNoMoreInput
		}

		return t, nil
	}
}
