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

package user

import (
	"context"
	"testing"

	"github.com/ProtonMail/go-proton-api"
	"github.com/ProtonMail/go-proton-api/server"
	"github.com/ProtonMail/gopenpgp/v2/crypto"
	"github.com/ProtonMail/proton-bridge/v3/internal/usertypes"
	"github.com/stretchr/testify/require"
)

func BenchmarkAddrKeyRing(b *testing.B) {
	b.StopTimer()

	withAPI(b, context.Background(), func(ctx context.Context, s *server.Server, m *proton.Manager) {
		withAccount(b, s, "username", "password", []string{"email@pm.me"}, func(_ string, _ []string) {
			withUser(b, ctx, s, m, "username", "password", func(user *User) {
				b.StartTimer()

				apiUser, err := user.identityService.GetAPIUser(ctx)
				require.NoError(b, err)

				apiAddrs, err := user.identityService.GetAddresses(ctx)
				require.NoError(b, err)

				for i := 0; i < b.N; i++ {
					require.NoError(b, usertypes.WithAddrKRs(apiUser, apiAddrs, user.vault.KeyPass(), func(_ *crypto.KeyRing, _ map[string]*crypto.KeyRing) error {
						return nil
					}))
				}
			})
		})
	})
}
