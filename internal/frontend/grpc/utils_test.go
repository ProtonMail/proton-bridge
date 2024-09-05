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

package grpc

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func Test_GetInitials(t *testing.T) {
	require.Equal(t, "?", getInitials(""))
	require.Equal(t, "T", getInitials(" test"))
	require.Equal(t, "T", getInitials("test "))
	require.Equal(t, "T", getInitials(" test "))
	require.Equal(t, "JD", getInitials(" John Doe "))
	require.Equal(t, "J", getInitials(" JohnDoe@proton.me "))
	require.Equal(t, "JD", getInitials("\t\r\n John Doe \t\r\n "))

	require.Equal(t, "T", getInitials("TestTestman"))
	require.Equal(t, "TT", getInitials("Test Testman"))
	require.Equal(t, "J", getInitials("JamesJoyce"))
	require.Equal(t, "J", getInitials("JamesJoyceJeremy"))
	require.Equal(t, "J", getInitials("james.joyce"))
	require.Equal(t, "JJ", getInitials("James Joyce"))
	require.Equal(t, "JM", getInitials("James Joyce Mahabharata"))
	require.Equal(t, "JL", getInitials("James Joyce Jeremy Lin"))
	require.Equal(t, "JM", getInitials("Jean Michel"))
	require.Equal(t, "GC", getInitials("George Michael Carrie"))
}
