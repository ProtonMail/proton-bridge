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
	"fmt"
	"net"

	"github.com/ProtonMail/proton-bridge/v3/internal/constants"
	"golang.org/x/exp/slices"
)

const (
	maxPortNumber = 65535
)

// IsPortFree checks if the port is free to use.
func IsPortFree(port int) bool {
	if !(0 < port && port < maxPortNumber) {
		return false
	}
	// First, check localhost only.
	if isOccupied(fmt.Sprintf("%v:%d", constants.Host, port)) {
		return false
	}
	// Second, check also ports opened to public.
	return !isOccupied(fmt.Sprintf(":%d", port))
}

func isOccupied(port string) bool {
	// Try to create server at port.
	dummyServer, err := net.Listen("tcp", port)
	if err != nil {
		return true
	}
	_ = dummyServer.Close()
	return false
}

// FindFreePortFrom finds first empty port, starting with `startPort`, and excluding ports listed in exclude.
func FindFreePortFrom(startPort int, exclude ...int) int {
	loopedOnce := false
	freePort := startPort
	for slices.Contains(exclude, freePort) || !IsPortFree(freePort) {
		freePort++
		if freePort >= maxPortNumber {
			freePort = 1
			if loopedOnce {
				freePort = startPort
				break
			}
			loopedOnce = true
		}
	}
	return freePort
}
