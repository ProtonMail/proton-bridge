// Copyright (c) 2021 Proton Technologies AG
//
// This file is part of ProtonMail Bridge.
//
// ProtonMail Bridge is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// ProtonMail Bridge is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with ProtonMail Bridge.  If not, see <https://www.gnu.org/licenses/>.

package pmapi

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestMailboxPassword(t *testing.T) {
	// wantHash was generated with passprase and salt defined below. It
	// should not change when changing implementation of the function.
	wantHash := []byte("B5nwpsJQSTJ16ldr64Vdq6oeCCn32Fi")

	// Valid salt is 128bit long (16bytes)
	// $echo aaaabbbbccccdddd | base64
	salt := "YWFhYWJiYmJjY2NjZGRkZAo="

	passphrase := []byte("random")

	r := require.New(t)
	_, err := HashMailboxPassword(passphrase, "badsalt")
	r.Error(err)

	haveHash, err := HashMailboxPassword(passphrase, salt)
	r.NoError(err)
	r.Equal(wantHash, haveHash)
}
