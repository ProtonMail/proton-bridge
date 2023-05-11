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

package cpc

import (
	"context"
	"errors"
)

var ErrInvalidReplyType = errors.New("reply type does not match")

// Utilities to implement Chanel Procedure Calls. Similar in concept to RPC, but with between go-routines.

// Request contains the data for a request as well as the means to reply to a request.
type Request struct {
	value any
	reply chan reply
}

// Value returns the request value.
func (r *Request) Value() any {
	return r.value
}

// Reply should be used to send a reply to a given request.
func (r *Request) Reply(ctx context.Context, value any, err error) {
	defer close(r.reply)

	select {
	case <-ctx.Done():
	case r.reply <- reply{
		value: value,
		error: err,
	}:
	}
}

// CPC Channel Procedure Call. A play on RPC, but with channels. Use this type to send requests and wait for replies
// from a goroutine.
type CPC struct {
	request chan *Request
}

func NewCPC() *CPC {
	return &CPC{
		request: make(chan *Request),
	}
}

// Receive invokes the function on all the request that arrive.
func (c *CPC) Receive(ctx context.Context, f func(context.Context, *Request)) {
	for request := range c.request {
		f(ctx, request)
	}
}

// ReceiveCh returns the channel on which all requests are sent.
func (c *CPC) ReceiveCh() <-chan *Request {
	return c.request
}

// Close closes the CPC channel and no further requests should be made.
func (c *CPC) Close() {
	close(c.request)
}

// Send sends a request which expects a reply.
func (c *CPC) Send(ctx context.Context, value any) (any, error) {
	return c.execute(ctx, newRequest(value))
}

// SendTyped is similar to CPC.Send, but ensure that reply is of the given Type T.
func SendTyped[T any](ctx context.Context, c *CPC, value any) (T, error) {
	val, err := c.execute(ctx, newRequest(value))
	if err != nil {
		var t T
		return t, err
	}

	switch vt := val.(type) {
	case T:
		return vt, nil
	default:
		var t T
		return t, ErrInvalidReplyType
	}
}

type reply struct {
	value any
	error error
}

func (c *CPC) execute(ctx context.Context, request *Request) (any, error) {
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	case c.request <- request:
	}

	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	case r := <-request.reply:
		return r.value, r.error
	}
}

func newRequest(value any) *Request {
	return &Request{
		value: value,
		reply: make(chan reply),
	}
}
