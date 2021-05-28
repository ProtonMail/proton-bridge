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
	"net/url"
	"sync"
	"time"

	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

// ProxyTLSDialer wraps a TLSDialer to switch to a proxy if the initial dial fails.
type ProxyTLSDialer struct {
	dialer TLSDialer

	locker           sync.RWMutex
	directAddress    string
	proxyAddress     string
	allowProxy       bool
	proxyProvider    *proxyProvider
	proxyUseDuration time.Duration
}

// NewProxyTLSDialer constructs a dialer which provides a proxy-managing layer on top of an underlying dialer.
func NewProxyTLSDialer(cfg Config, dialer TLSDialer) *ProxyTLSDialer {
	return &ProxyTLSDialer{
		dialer:           dialer,
		locker:           sync.RWMutex{},
		directAddress:    formatAsAddress(cfg.HostURL),
		proxyAddress:     formatAsAddress(cfg.HostURL),
		proxyProvider:    newProxyProvider(cfg, dohProviders, proxyQuery),
		proxyUseDuration: proxyUseDuration,
	}
}

// formatAsAddress returns URL as `host:port` for easy comparison in DialTLS.
func formatAsAddress(rawURL string) string {
	url, err := url.Parse(rawURL)
	if err != nil {
		// This means wrong configuration.
		// Developer should get feedback right away.
		panic(err)
	}

	port := "443"
	if url.Scheme == "http" {
		port = "80"
	}
	return net.JoinHostPort(url.Host, port)
}

// DialTLS dials the given network/address. If it fails, it retries using a proxy.
func (d *ProxyTLSDialer) DialTLS(network, address string) (net.Conn, error) {
	if address == d.directAddress {
		address = d.proxyAddress
	}

	conn, err := d.dialer.DialTLS(network, address)
	if err == nil || !d.allowProxy {
		return conn, err
	}

	err = d.switchToReachableServer()
	if err != nil {
		return nil, err
	}

	return d.dialer.DialTLS(network, d.proxyAddress)
}

// switchToReachableServer switches to using a reachable server (either proxy or standard API).
func (d *ProxyTLSDialer) switchToReachableServer() error {
	d.locker.Lock()
	defer d.locker.Unlock()

	logrus.Info("Attempting to switch to a proxy")

	proxy, err := d.proxyProvider.findReachableServer()
	if err != nil {
		return errors.Wrap(err, "failed to find a usable proxy")
	}

	proxyAddress := formatAsAddress(proxy)

	// If the chosen proxy is the standard API, we want to use it but still show the troubleshooting screen.
	if proxyAddress == d.directAddress {
		logrus.Info("The standard API is reachable again; connection drop was only intermittent")
		d.proxyAddress = proxyAddress
		return ErrNoConnection
	}

	logrus.WithField("proxy", proxyAddress).Info("Switching to a proxy")

	// If the host is currently the rootURL, it's the first time we are enabling a proxy.
	// This means we want to disable it again in 24 hours.
	if d.proxyAddress == d.directAddress {
		go func() {
			<-time.After(d.proxyUseDuration)

			d.locker.Lock()
			defer d.locker.Unlock()

			d.proxyAddress = d.directAddress
		}()
	}

	d.proxyAddress = proxyAddress
	return nil
}

// AllowProxy allows the dialer to switch to a proxy if need be.
func (d *ProxyTLSDialer) AllowProxy() {
	d.locker.Lock()
	defer d.locker.Unlock()

	d.allowProxy = true
}

// DisallowProxy prevents the dialer from switching to a proxy if need be.
func (d *ProxyTLSDialer) DisallowProxy() {
	d.locker.Lock()
	defer d.locker.Unlock()

	d.allowProxy = false
	d.proxyAddress = d.directAddress
}
