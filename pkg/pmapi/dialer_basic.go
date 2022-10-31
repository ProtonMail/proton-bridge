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

package pmapi

import (
	"crypto/tls"
	"net"
	"net/http"
	"time"
)

type TLSDialer interface {
	DialTLS(network, address string) (conn net.Conn, err error)
}

// CreateTransportWithDialer creates an http.Transport that uses the given dialer to make TLS connections.
func CreateTransportWithDialer(dialer TLSDialer) *http.Transport {
	return &http.Transport{
		DialTLS: dialer.DialTLS,

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
	cfg Config
}

// NewBasicTLSDialer returns a new BasicTLSDialer.
func NewBasicTLSDialer(cfg Config) *BasicTLSDialer {
	return &BasicTLSDialer{
		cfg: cfg,
	}
}

// DialTLS returns a connection to the given address using the given network.
func (d *BasicTLSDialer) DialTLS(network, address string) (conn net.Conn, err error) {
	dialer := &net.Dialer{Timeout: 30 * time.Second} // Alternative Routes spec says this should be a 30s timeout.

	var tlsConfig *tls.Config

	// If we are not dialing the standard API then we should skip cert verification checks.
	if address != d.cfg.HostURL {
		tlsConfig = &tls.Config{InsecureSkipVerify: true} //nolint:gosec
	}

	return tls.DialWithDialer(dialer, network, address, tlsConfig)
}
