// Copyright (c) 2026 Proton AG
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

package imapsmtpserver

import (
	"crypto/tls"
	"net"
	"strconv"
)

func newListener(bindAddress string, port int, useTLS bool, tlsConfig *tls.Config) (net.Listener, error) {
	address := net.JoinHostPort(bindAddress, strconv.Itoa(port))

	if useTLS {
		tlsListener, err := tls.Listen("tcp", address, tlsConfig)
		if err != nil {
			return nil, err
		}

		return tlsListener, nil
	}

	netListener, err := net.Listen("tcp", address)
	if err != nil {
		return nil, err
	}

	return netListener, nil
}

func getPort(addr net.Addr) int {
	switch addr := addr.(type) {
	case *net.TCPAddr:
		return addr.Port

	case *net.UDPAddr:
		return addr.Port

	default:
		return 0
	}
}
