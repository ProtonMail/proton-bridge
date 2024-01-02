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
// along with Proton Mail Bridge. If not, see <https://www.gnu.org/licenses/>.

package dialer

import (
	"net"
	"net/http"
	"time"
)

const (
	// ClientTimeout is the timeout for the whole request (from dial to
	// receiving the response body). It should be large enough to download
	// even the largest attachments or the new binary of the Bridge, but
	// should be hit if the server hangs (default is infinite which is bad).
	clientTimeout = 30 * time.Minute
	dialTimeout   = 3 * time.Second
)

// DialTimeoutClient creates client with overridden dialTimeout.
func DialTimeoutClient() *http.Client {
	transport := &http.Transport{
		Dial: func(network, addr string) (net.Conn, error) {
			return net.DialTimeout(network, addr, dialTimeout)
		},
	}
	return &http.Client{
		Timeout:   clientTimeout,
		Transport: transport,
	}
}
