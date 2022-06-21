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
// along with Proton Mail Bridge. If not, see <https://www.gnu.org/licenses/>.

package ports

import (
	"fmt"
	"net"
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
	if isOccupied(fmt.Sprintf("127.0.0.1:%d", port)) {
		return false
	}
	// Second, check also ports opened to public.
	return !isOccupied(fmt.Sprintf(":%d", port))
}

func isOccupied(port string) bool {
	// Try to create server at port.
	dummyserver, err := net.Listen("tcp", port)
	if err != nil {
		return true
	}
	_ = dummyserver.Close()
	return false
}

// FindFreePortFrom finds first empty port, starting with `startPort`.
func FindFreePortFrom(startPort int) int {
	loopedOnce := false
	freePort := startPort
	for !IsPortFree(freePort) {
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
