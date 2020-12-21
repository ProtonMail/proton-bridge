// Copyright (c) 2021 Proton Technologies AG
//
// This file is part of ProtonMail Bridge.Bridge.
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

package pmapi

import (
	"net"
)

// ProxyTLSDialer wraps a TLSDialer to switch to a proxy if the initial dial fails.
type ProxyTLSDialer struct {
	dialer TLSDialer

	cm *ClientManager
}

// NewProxyTLSDialer constructs a dialer which provides a proxy-managing layer on top of an underlying dialer.
func NewProxyTLSDialer(dialer TLSDialer, cm *ClientManager) *ProxyTLSDialer {
	return &ProxyTLSDialer{
		dialer: dialer,
		cm:     cm,
	}
}

// DialTLS dials the given network/address. If it fails, it retries using a proxy.
func (d *ProxyTLSDialer) DialTLS(network, address string) (net.Conn, error) {
	conn, err := d.dialTLS(network, address)
	if err != nil {
		d.cm.config.NoConnectionHandler()
	} else {
		d.cm.config.ConnectionHandler()
	}
	return conn, err
}

func (d *ProxyTLSDialer) dialTLS(network, address string) (conn net.Conn, err error) {
	if conn, err = d.dialer.DialTLS(network, address); err == nil {
		return
	}

	if !d.cm.allowProxy {
		return
	}

	var proxy string

	if proxy, err = d.cm.switchToReachableServer(); err != nil {
		return
	}

	_, port, err := net.SplitHostPort(address)
	if err != nil {
		return
	}

	return d.dialer.DialTLS(network, net.JoinHostPort(proxy, port))
}
