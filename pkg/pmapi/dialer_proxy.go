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
func (d *ProxyTLSDialer) DialTLS(network, address string) (conn net.Conn, err error) {
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
