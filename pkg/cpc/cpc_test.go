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
	"sync"
	"testing"

	"github.com/stretchr/testify/require"
)

type sendIntRequest struct{}

type quitRequest struct{}

func TestCPC_Receive(t *testing.T) {
	const replyValue = 20

	cpc := NewCPC()

	wg := sync.WaitGroup{}

	go func() {
		defer wg.Done()

		wg.Add(1)

		cpc.Receive(context.Background(), func(ctx context.Context, request *Request) {
			switch request.Value().(type) {
			case sendIntRequest:
				request.Reply(ctx, replyValue, nil)
			case quitRequest:
				request.Reply(ctx, nil, nil)
			default:
				panic("unknown request")
			}
		})
	}()

	r, err := cpc.Send(context.Background(), sendIntRequest{})
	require.NoError(t, err)
	require.Equal(t, r, replyValue)

	_, err = cpc.Send(context.Background(), quitRequest{})
	require.NoError(t, err)

	cpc.Close()
	wg.Wait()
}
