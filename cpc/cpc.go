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

var ErrRequestHasNoReply = errors.New("request has no reply channel")
var ErrExpectedReply = errors.New("request does not have reply channel")

// Utilities to implement Chanel Procedure Calls. Similar in concept to RPC, but with between go-routines.

type RequestReply struct {
	Value any
	Error error
}

type Request struct {
	Value any
	Reply chan RequestReply
}

func NewRequest(value any) *Request {
	return &Request{
		Value: value,
		Reply: make(chan RequestReply),
	}
}

func NewRequestWithoutReply(value any) *Request {
	return &Request{
		Value: value,
		Reply: nil,
	}
}

func (r *Request) Close() {
	if r.Reply != nil {
		panic("request reply has not been sent")
	}
}

func (r *Request) SendReply(ctx context.Context, value any, err error) {
	if r.Reply == nil {
		panic("request has no reply")
	}

	defer func() {
		close(r.Reply)
		r.Reply = nil
	}()

	select {
	case <-ctx.Done():
	case r.Reply <- RequestReply{
		Value: value,
		Error: err,
	}:
	}
}

type CPC struct {
	request chan *Request
}

func NewCPC() *CPC {
	return &CPC{
		request: make(chan *Request),
	}
}

// Receive is meant to be called by the code that is supposed to handle the requests that arrive.
func (c *CPC) Receive(ctx context.Context, f func(context.Context, *Request)) {
	for request := range c.request {
		f(ctx, request)
		request.Close()
	}
}

func (c *CPC) Close() {
	close(c.request)
}

// SendNoReply sends a request which doesn't expect a reply.
func (c *CPC) SendNoReply(ctx context.Context, value any) error {
	return c.executeNoReplyImpl(ctx, NewRequestWithoutReply(value))
}

// SendWithReply sends a request which expects a reply.
func (c *CPC) SendWithReply(ctx context.Context, value any) (any, error) {
	return c.executeReplyImpl(ctx, NewRequest(value))
}

func (c *CPC) executeNoReplyImpl(ctx context.Context, request *Request) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	case c.request <- request:
	}

	return nil
}

func (c *CPC) executeReplyImpl(ctx context.Context, request *Request) (any, error) {
	if request.Reply == nil {
		return nil, ErrExpectedReply
	}

	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	case c.request <- request:
	}

	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	case reply := <-request.Reply:
		return reply.Value, reply.Error
	}
}
