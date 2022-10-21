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
// along with Proton Mail Bridge.  If not, see <https://www.gnu.org/licenses/>.

package tests

import (
	"context"
	"fmt"
	"runtime"

	"github.com/bradenaw/juniper/stream"
	"gitlab.protontech.ch/go/liteapi"
)

func (t *testCtx) withClient(ctx context.Context, username string, fn func(context.Context, *liteapi.Client) error) error {
	c, _, err := liteapi.New(
		liteapi.WithHostURL(t.api.GetHostURL()),
		liteapi.WithTransport(liteapi.InsecureTransport()),
	).NewClientWithLogin(ctx, username, []byte(t.getUserPass(t.getUserID(username))))
	if err != nil {
		return err
	}

	defer c.Close()

	if err := fn(ctx, c); err != nil {
		return fmt.Errorf("failed to execute with client: %w", err)
	}

	if err := c.AuthDelete(ctx); err != nil {
		return fmt.Errorf("failed to delete auth: %w", err)
	}

	return nil
}

func (t *testCtx) createMessages(ctx context.Context, username, addrID string, req []liteapi.ImportReq) error {
	return t.withClient(ctx, username, func(ctx context.Context, c *liteapi.Client) error {
		user, err := c.GetUser(ctx)
		if err != nil {
			return err
		}

		addr, err := c.GetAddresses(ctx)
		if err != nil {
			return err
		}

		salt, err := c.GetSalts(ctx)
		if err != nil {
			return err
		}

		keyPass, err := salt.SaltForKey([]byte(t.getUserPass(t.getUserID(username))), user.Keys.Primary().ID)
		if err != nil {
			return err
		}

		_, addrKRs, err := liteapi.Unlock(user, addr, keyPass)
		if err != nil {
			return err
		}

		if _, err := stream.Collect(ctx, c.ImportMessages(
			ctx,
			addrKRs[addrID],
			runtime.NumCPU(),
			runtime.NumCPU(),
			req...,
		)); err != nil {
			return err
		}

		return nil
	})
}
