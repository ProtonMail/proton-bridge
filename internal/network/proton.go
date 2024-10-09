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

package network

import (
	"context"
	"errors"
	"math/rand"
	"time"

	"github.com/ProtonMail/go-proton-api"
)

type CoolDownProvider interface {
	GetNextWaitTime() time.Duration
	Reset()
}

func jitter(max int) time.Duration { //nolint:predeclared
	return time.Duration(rand.Intn(max)) * time.Second //nolint:gosec
}

type ExpCoolDown struct {
	count int
}

func (c *ExpCoolDown) GetNextWaitTime() time.Duration {
	waitTimes := []time.Duration{
		20 * time.Second,
		40 * time.Second,
		80 * time.Second,
		160 * time.Second,
		300 * time.Second,
		600 * time.Second,
	}

	last := len(waitTimes) - 1

	if c.count >= last {
		return waitTimes[last] + jitter(10)
	}

	c.count++

	return waitTimes[c.count-1] + jitter(10)
}

func (c *ExpCoolDown) Reset() {
	c.count = 0
}

type NoCoolDown struct{}

func (c *NoCoolDown) GetNextWaitTime() time.Duration { return time.Millisecond }
func (c *NoCoolDown) Reset()                         {}

func Is429Or5XXError(err error) bool {
	var apiErr *proton.APIError
	if errors.As(err, &apiErr) {
		return apiErr.Status == 429 || apiErr.Status >= 500
	}

	return false
}

type ProtonClientRetryWrapper[T any] struct {
	client              T
	coolDown            CoolDownProvider
	encountered429or5xx bool
}

func NewClientRetryWrapper[T any](client T, coolDown CoolDownProvider) *ProtonClientRetryWrapper[T] {
	return &ProtonClientRetryWrapper[T]{client: client, coolDown: coolDown}
}

func (p *ProtonClientRetryWrapper[T]) DidEncounter429or5xx() bool {
	return p.encountered429or5xx
}

func (p *ProtonClientRetryWrapper[T]) Retry(ctx context.Context, f func(context.Context, T) error) error {
	p.coolDown.Reset()
	p.encountered429or5xx = false
	for {
		err := f(ctx, p.client)
		if Is429Or5XXError(err) {
			p.encountered429or5xx = true
			coolDown := p.coolDown.GetNextWaitTime()
			select {
			case <-ctx.Done():
			case <-time.After(coolDown):
			}
			continue
		}

		return err
	}
}

func RetryWithClient[T any, R any](ctx context.Context, p *ProtonClientRetryWrapper[T], f func(context.Context, T) (R, error)) (R, error) {
	var result R
	err := p.Retry(ctx, func(ctx context.Context, t T) error {
		r, err := f(ctx, t)
		result = r
		return err
	})

	return result, err
}
