// Copyright (c) 2022 Proton AG
//
// This file is part of Proton Mail Bridge.Bridge.
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

package context

import (
	"os"

	"github.com/ProtonMail/proton-bridge/v2/test/accounts"
)

func newTestAccounts() *accounts.TestAccounts {
	envFile := os.Getenv("TEST_ACCOUNTS")
	data, err := accounts.Load(envFile)
	if err != nil {
		panic(err)
	}
	return data
}

func (ctx *TestContext) GetTestAccount(bddUserID string) *accounts.TestAccount {
	return ctx.testAccounts.GetTestAccount(bddUserID)
}

func (ctx *TestContext) GetTestAccountWithAddress(bddUserID, bddAddressID string) *accounts.TestAccount {
	return ctx.testAccounts.GetTestAccountWithAddress(bddUserID, bddAddressID)
}

func (ctx *TestContext) EnsureAddressID(bddUserID, addressOrAddressTestID string) string {
	account := ctx.GetTestAccount(bddUserID)
	if account == nil {
		return addressOrAddressTestID
	}
	return account.EnsureAddressID(addressOrAddressTestID)
}

func (ctx *TestContext) EnsureAddress(bddUserID, addressOrAddressTestID string) string {
	account := ctx.GetTestAccount(bddUserID)
	if account == nil {
		return addressOrAddressTestID
	}
	return account.EnsureAddress(addressOrAddressTestID)
}

func (ctx *TestContext) ReorderAddresses(userID, addressA, addressB string) {
	addresses := ctx.testAccounts.Addresses[userID]

	addresses[addressA], addresses[addressB] = addresses[addressB], addresses[addressA]
}
