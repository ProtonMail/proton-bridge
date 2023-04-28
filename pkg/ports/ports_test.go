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
// along with Proton Mail Bridge. If not, see <https://www.gnu.org/licenses/>.

package ports

import (
	"net"
	"strconv"
	"testing"

	"github.com/stretchr/testify/require"
)

const testPort = 18080

func TestFreePort(t *testing.T) {
	require.True(t, IsPortFree(testPort), "port should be empty")
}

func TestOccupiedPort(t *testing.T) {
	dummyServer, err := net.Listen("tcp", ":"+strconv.Itoa(testPort))
	require.NoError(t, err)

	require.True(t, !IsPortFree(testPort), "port should be occupied")

	_ = dummyServer.Close()
}

func TestFindFreePortFromDirectly(t *testing.T) {
	foundPort := FindFreePortFrom(testPort)
	require.Equal(t, testPort, foundPort)
}

func TestFindFreePortFromNextOne(t *testing.T) {
	dummyServer, err := net.Listen("tcp", ":"+strconv.Itoa(testPort))
	require.NoError(t, err)

	foundPort := FindFreePortFrom(testPort)
	require.Equal(t, testPort+1, foundPort)

	_ = dummyServer.Close()
}

func TestFindFreePortExcluding(t *testing.T) {
	dummyServer, err := net.Listen("tcp", ":"+strconv.Itoa(testPort))
	require.NoError(t, err)

	foundPort := FindFreePortFrom(testPort, testPort+1, testPort+2)
	require.Equal(t, testPort+3, foundPort)

	_ = dummyServer.Close()
}
