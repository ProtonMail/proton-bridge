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

package smtp

import (
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestAccountTimeout(t *testing.T) {
	account := smtpAccountState{errTimeout: 5 * time.Second}
	err := errors.New("fail")

	for i := 0; i <= maxFailedCommands; i++ {
		requestTime := time.Now()
		assert.Nil(t, account.canMakeRequest(requestTime))
		account.handleSMTPErr(requestTime, err)
	}
	{
		requestTime := time.Now()
		assert.ErrorIs(t, account.canMakeRequest(requestTime), ErrTooManyErrors)
	}

	assert.Eventually(t, func() bool {
		requestTime := time.Now()
		return account.canMakeRequest(requestTime) == nil
	}, 10*time.Second, time.Second)
}
