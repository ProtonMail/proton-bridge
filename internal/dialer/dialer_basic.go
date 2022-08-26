// Copyright (c) 2022 Proton AG
//
// This file is part of Proton Mail Bridge.Bridge.
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
	"context"
	"crypto/tls"
	"net"
	"net/http"
	"time"
)

type TLSDialer interface {
	DialTLSContext(ctx context.Context, network, address string) (conn net.Conn, err error)
}

// CreateTransportWithDialer creates an http.Transport that uses the given dialer to make TLS connections.
func CreateTransportWithDialer(dialer TLSDialer) *http.Transport {
	return &http.Transport{
		DialTLSContext: dialer.DialTLSContext,

		Proxy:               http.ProxyFromEnvironment,
		MaxIdleConns:        100,
		MaxIdleConnsPerHost: 100,
		IdleConnTimeout:     5 * time.Minute,

		ExpectContinueTimeout: 500 * time.Millisecond,

		// GODT-126: this was initially 10s but logs from users showed a significant number
		// were hitting this timeout, possibly due to flaky wifi taking >10s to reconnect.
		// Bumping to 30s for now to avoid this problem.
		ResponseHeaderTimeout: 30 * time.Second,

		// If we allow up to 30 seconds for response headers, it is reasonable to allow up
		// to 30 seconds for the TLS handshake to take place.
		TLSHandshakeTimeout: 30 * time.Second,
	}
}

// BasicTLSDialer implements TLSDialer.
type BasicTLSDialer struct {
	hostURL string
}

// NewBasicTLSDialer returns a new BasicTLSDialer.
func NewBasicTLSDialer(hostURL string) *BasicTLSDialer {
	return &BasicTLSDialer{
		hostURL: hostURL,
	}
}

// DialTLS returns a connection to the given address using the given network.
func (d *BasicTLSDialer) DialTLSContext(ctx context.Context, network, address string) (conn net.Conn, err error) {
	return (&tls.Dialer{
		NetDialer: &net.Dialer{
			Timeout: 30 * time.Second,
		},
		Config: &tls.Config{
			InsecureSkipVerify: address != d.hostURL,
		},
	}).DialContext(ctx, network, address)
}
