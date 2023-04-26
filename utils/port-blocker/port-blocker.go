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

// port-blocker is a command-line that ensure a port or range of ports is occupied by creating listeners.
package main

import (
	"fmt"
	"net"
	"os"
	"strconv"
)

func main() {
	argCount := len(os.Args)
	if (len(os.Args) < 2) || (argCount > 3) {
		exitWithUsage("Invalid number of arguments.")
	}

	startPort := parsePort(os.Args[1])
	endPort := startPort
	if argCount == 3 {
		endPort = parsePort(os.Args[2])
	}

	runBlocker(startPort, endPort)
}

func parsePort(portString string) int {
	result, err := strconv.Atoi(portString)
	if err != nil {
		exitWithUsage(fmt.Sprintf("Invalid port '%v'.", portString))
	}

	if (result < 1024) || (result > 65535) { // ports below 1024 are reserved.
		exitWithUsage("Ports must be in the range [1024-65535].")
	}

	return result
}

func exitWithUsage(message string) {
	fmt.Printf("Usage: port-blocker <startPort> [<endPort>]\n")
	if len(message) > 0 {
		fmt.Println(message)
	}
	os.Exit(1)
}

func runBlocker(startPort, endPort int) {
	if endPort < startPort {
		exitWithUsage("startPort must be less than or equal to endPort.")
	}

	for port := startPort; port <= endPort; port++ {
		listener, err := net.Listen("tcp", "127.0.0.1:"+strconv.Itoa(port))
		if err != nil {
			fmt.Printf("Port %v is already blocked. Skipping.\n", port)
		} else {
			//goland:noinspection GoDeferInLoop
			defer func() {
				_ = listener.Close()
			}()
		}
	}

	fmt.Println("Blocking requested ports. Press enter to exit.")
	_, _ = fmt.Scanln()
}
