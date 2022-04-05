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

package pmapi

import (
	"context"
	"net/http"
	"testing"
	"time"

	r "github.com/stretchr/testify/require"
	"golang.org/x/net/http/httpproxy"
)

func TestProxyProvider_FindProxy(t *testing.T) {
	proxy := getTrustedServer()
	defer closeServer(proxy)

	p := newProxyProvider(Config{HostURL: ""}, []string{"not used"}, "not used")
	p.dohLookup = func(ctx context.Context, q, p string) ([]string, error) { return []string{proxy.URL}, nil }

	url, err := p.findReachableServer()
	r.NoError(t, err)
	r.Equal(t, proxy.URL, url)
}

func TestProxyProvider_FindProxy_ChooseReachableProxy(t *testing.T) {
	reachableProxy := getTrustedServer()
	defer closeServer(reachableProxy)

	// We actually close the unreachable proxy straight away rather than deferring the closure.
	unreachableProxy := getTrustedServer()
	closeServer(unreachableProxy)

	p := newProxyProvider(Config{HostURL: ""}, []string{"not used"}, "not used")
	p.dohLookup = func(ctx context.Context, q, p string) ([]string, error) {
		return []string{reachableProxy.URL, unreachableProxy.URL}, nil
	}

	url, err := p.findReachableServer()
	r.NoError(t, err)
	r.Equal(t, reachableProxy.URL, url)
}

func TestProxyProvider_FindProxy_ChooseTrustedProxy(t *testing.T) {
	trustedProxy := getTrustedServer()
	defer closeServer(trustedProxy)

	untrustedProxy := getUntrustedServer()
	defer closeServer(untrustedProxy)

	p := newProxyProvider(Config{HostURL: ""}, []string{"not used"}, "not used")
	p.dohLookup = func(ctx context.Context, q, p string) ([]string, error) {
		return []string{untrustedProxy.URL, trustedProxy.URL}, nil
	}

	url, err := p.findReachableServer()
	r.NoError(t, err)
	r.Equal(t, trustedProxy.URL, url)
}

func TestProxyProvider_FindProxy_FailIfNoneReachable(t *testing.T) {
	unreachableProxy1 := getTrustedServer()
	closeServer(unreachableProxy1)

	unreachableProxy2 := getTrustedServer()
	closeServer(unreachableProxy2)

	p := newProxyProvider(Config{HostURL: ""}, []string{"not used"}, "not used")
	p.dohLookup = func(ctx context.Context, q, p string) ([]string, error) {
		return []string{unreachableProxy1.URL, unreachableProxy2.URL}, nil
	}

	_, err := p.findReachableServer()
	r.Error(t, err)
}

func TestProxyProvider_FindProxy_FailIfNoneTrusted(t *testing.T) {
	untrustedProxy1 := getUntrustedServer()
	defer closeServer(untrustedProxy1)

	untrustedProxy2 := getUntrustedServer()
	defer closeServer(untrustedProxy2)

	p := newProxyProvider(Config{HostURL: ""}, []string{"not used"}, "not used")
	p.dohLookup = func(ctx context.Context, q, p string) ([]string, error) {
		return []string{untrustedProxy1.URL, untrustedProxy2.URL}, nil
	}

	_, err := p.findReachableServer()
	r.Error(t, err)
}

func TestProxyProvider_FindProxy_RefreshCacheTimeout(t *testing.T) {
	p := newProxyProvider(Config{HostURL: ""}, []string{"not used"}, "not used")
	p.cacheRefreshTimeout = 1 * time.Second
	p.dohLookup = func(ctx context.Context, q, p string) ([]string, error) { time.Sleep(2 * time.Second); return nil, nil }

	// We should fail to refresh the proxy cache because the doh provider
	// takes 2 seconds to respond but we timeout after just 1 second.
	_, err := p.findReachableServer()

	r.Error(t, err)
}

func TestProxyProvider_FindProxy_CanReachTimeout(t *testing.T) {
	slowProxy := getTrustedServerWithHandler(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {
		time.Sleep(2 * time.Second)
	}))
	defer closeServer(slowProxy)

	p := newProxyProvider(Config{HostURL: ""}, []string{"not used"}, "not used")
	p.canReachTimeout = 1 * time.Second
	p.dohLookup = func(ctx context.Context, q, p string) ([]string, error) { return []string{slowProxy.URL}, nil }

	// We should fail to reach the returned proxy because it takes 2 seconds
	// to reach it and we only allow 1.
	_, err := p.findReachableServer()

	r.Error(t, err)
}

func TestProxyProvider_DoHLookup_Quad9(t *testing.T) {
	p := newProxyProvider(Config{}, []string{Quad9Provider, GoogleProvider}, proxyQuery)

	records, err := p.dohLookup(context.Background(), proxyQuery, Quad9Provider)
	r.NoError(t, err)
	r.NotEmpty(t, records)
}

// DISABLEDTestProxyProvider_DoHLookup_Quad9Port cannot run on CI due to custom
// port filter. Basic functionality should be covered by other tests. Keeping
// code here to be able to run it locally if needed.
func DISABLEDTestProxyProviderDoHLookupQuad9Port(t *testing.T) {
	p := newProxyProvider(Config{}, []string{Quad9PortProvider, GoogleProvider}, proxyQuery)

	records, err := p.dohLookup(context.Background(), proxyQuery, Quad9PortProvider)
	r.NoError(t, err)
	r.NotEmpty(t, records)
}

func TestProxyProvider_DoHLookup_Google(t *testing.T) {
	p := newProxyProvider(Config{}, []string{Quad9Provider, GoogleProvider}, proxyQuery)

	records, err := p.dohLookup(context.Background(), proxyQuery, GoogleProvider)
	r.NoError(t, err)
	r.NotEmpty(t, records)
}

func TestProxyProvider_DoHLookup_FindProxy(t *testing.T) {
	skipIfProxyIsSet(t)

	p := newProxyProvider(Config{}, []string{Quad9Provider, GoogleProvider}, proxyQuery)

	url, err := p.findReachableServer()
	r.NoError(t, err)
	r.NotEmpty(t, url)
}

func TestProxyProvider_DoHLookup_FindProxyFirstProviderUnreachable(t *testing.T) {
	skipIfProxyIsSet(t)

	p := newProxyProvider(Config{}, []string{"https://unreachable", Quad9Provider, GoogleProvider}, proxyQuery)

	url, err := p.findReachableServer()
	r.NoError(t, err)
	r.NotEmpty(t, url)
}

// skipIfProxyIsSet skips the tests if HTTPS proxy is set.
// Should be used for tests depending on proper certificate checks which
// is not possible under our CI setup.
func skipIfProxyIsSet(t *testing.T) {
	if httpproxy.FromEnvironment().HTTPSProxy != "" {
		t.SkipNow()
	}
}
