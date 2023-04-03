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

package tests

import (
	"context"
	"fmt"
	"os"
	"runtime"

	"github.com/ProtonMail/gluon/async"
	"github.com/ProtonMail/go-proton-api"
	"github.com/ProtonMail/gopenpgp/v2/crypto"
	"github.com/bradenaw/juniper/stream"
)

// withProton executes the given function with a proton manager configured to use the test API.
func (t *testCtx) withProton(fn func(*proton.Manager) error) error {
	m := proton.New(
		proton.WithHostURL(t.api.GetHostURL()),
		proton.WithTransport(proton.InsecureTransport()),
		proton.WithAppVersion(t.api.GetAppVersion()),
		proton.WithDebug(os.Getenv("FEATURE_API_DEBUG") != ""),
	)
	defer m.Close()

	return fn(m)
}

// withClient executes the given function with a client that is logged in as the given (known) user.
func (t *testCtx) withClient(ctx context.Context, username string, fn func(context.Context, *proton.Client) error) error {
	return t.withClientPass(ctx, username, t.getUserByName(username).getUserPass(), fn)
}

// withClient executes the given function with a client that is logged in with the given username and password.
func (t *testCtx) withClientPass(ctx context.Context, username, password string, fn func(context.Context, *proton.Client) error) error {
	return t.withProton(func(m *proton.Manager) error {
		c, _, err := m.NewClientWithLogin(ctx, username, []byte(password))
		if err != nil {
			return fmt.Errorf("failed to create client: %w", err)
		}
		defer c.Close()

		if err := fn(ctx, c); err != nil {
			return fmt.Errorf("failed to execute with client: %w", err)
		}

		if err := c.AuthDelete(ctx); err != nil {
			return fmt.Errorf("failed to delete auth: %w", err)
		}

		return nil
	})
}

// runQuarkCmd runs the given quark command with the given arguments.
func (t *testCtx) runQuarkCmd(ctx context.Context, command string, args ...string) ([]byte, error) {
	var out []byte

	if err := t.withProton(func(m *proton.Manager) error {
		res, err := m.QuarkRes(ctx, command, args...)
		if err != nil {
			return err
		}

		out = res

		return nil
	}); err != nil {
		return nil, err
	}

	return out, nil
}

func (t *testCtx) withAddrKR(
	ctx context.Context,
	c *proton.Client,
	username, addrID string,
	fn func(context.Context, *crypto.KeyRing) error,
) error {
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

	keyPass, err := salt.SaltForKey([]byte(t.getUserByName(username).getUserPass()), user.Keys.Primary().ID)
	if err != nil {
		return err
	}

	_, addrKRs, err := proton.Unlock(user, addr, keyPass, async.NoopPanicHandler{})
	if err != nil {
		return err
	}

	return fn(ctx, addrKRs[addrID])
}

func (t *testCtx) createMessages(ctx context.Context, username, addrID string, req []proton.ImportReq) error {
	return t.withClient(ctx, username, func(ctx context.Context, c *proton.Client) error {
		return t.withAddrKR(ctx, c, username, addrID, func(ctx context.Context, addrKR *crypto.KeyRing) error {
			str, err := c.ImportMessages(
				ctx,
				addrKR,
				runtime.NumCPU(),
				runtime.NumCPU(),
				req...,
			)
			if err != nil {
				return fmt.Errorf("failed to prepare messages for import: %w", err)
			}

			if _, err := stream.Collect(ctx, str); err != nil {
				return fmt.Errorf("failed to import messages: %w", err)
			}

			return nil
		})
	})
}
