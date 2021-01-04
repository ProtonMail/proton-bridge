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

package imap

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestDoNotCache(t *testing.T) {
	var dnc doNotCacheError
	require.NoError(t, dnc.errorOrNil())
	_, ok := dnc.errorOrNil().(*doNotCacheError)
	require.True(t, !ok, "should not be type doNotCacheError")

	dnc.add(errors.New("first"))
	require.True(t, dnc.errorOrNil() != nil, "should be error")
	_, ok = dnc.errorOrNil().(*doNotCacheError)
	require.True(t, ok, "should be type doNotCacheError")

	dnc.add(errors.New("second"))
	dnc.add(errors.New("third"))
	t.Log(dnc.errorOrNil())
}
