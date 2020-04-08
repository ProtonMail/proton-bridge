// Copyright (c) 2020 Proton Technologies AG
//
// This file is part of ProtonMail Bridge.
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
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

const (
	TestDoHQuery       = "dMFYGSLTQOJXXI33ONVQWS3BOMNUA.protonpro.xyz"
	TestQuad9Provider  = "https://dns11.quad9.net/dns-query"
	TestGoogleProvider = "https://dns.google/dns-query"
)

func TestProxyManager_FindProxy(t *testing.T) {
	blockAPI()
	defer unblockAPI()

	proxy := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
	defer proxy.Close()

	p := newProxyManager([]string{"not used"}, "not used")
	p.dohLookup = func(q, p string) ([]string, error) { return []string{proxy.URL}, nil }

	url, err := p.findProxy()
	require.NoError(t, err)
	require.Equal(t, proxy.URL, url)
}

func TestProxyManager_FindProxy_ChooseReachableProxy(t *testing.T) {
	blockAPI()
	defer unblockAPI()

	badProxy := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
	goodProxy := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))

	// Close the bad proxy first so it isn't reachable; we should then choose the good proxy.
	badProxy.Close()
	defer goodProxy.Close()

	p := newProxyManager([]string{"not used"}, "not used")
	p.dohLookup = func(q, p string) ([]string, error) { return []string{badProxy.URL, goodProxy.URL}, nil }

	url, err := p.findProxy()
	require.NoError(t, err)
	require.Equal(t, goodProxy.URL, url)
}

func TestProxyManager_FindProxy_FailIfNoneReachable(t *testing.T) {
	blockAPI()
	defer unblockAPI()

	badProxy := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
	anotherBadProxy := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))

	// Close the proxies to simulate them not being reachable.
	badProxy.Close()
	anotherBadProxy.Close()

	p := newProxyManager([]string{"not used"}, "not used")
	p.dohLookup = func(q, p string) ([]string, error) { return []string{badProxy.URL, anotherBadProxy.URL}, nil }

	_, err := p.findProxy()
	require.Error(t, err)
}

func TestProxyManager_FindProxy_LookupTimeout(t *testing.T) {
	blockAPI()
	defer unblockAPI()

	proxy := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
	defer proxy.Close()

	p := newProxyManager([]string{"not used"}, "not used")
	p.lookupTimeout = time.Second
	p.dohLookup = func(q, p string) ([]string, error) { time.Sleep(2 * time.Second); return nil, nil }

	// The findProxy should fail because lookup takes 2 seconds but we only allow 1 second.
	_, err := p.findProxy()
	require.Error(t, err)
}

func TestProxyManager_FindProxy_FindTimeout(t *testing.T) {
	blockAPI()
	defer unblockAPI()

	slowProxy := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(2 * time.Second)
	}))
	defer slowProxy.Close()

	p := newProxyManager([]string{"not used"}, "not used")
	p.findTimeout = time.Second
	p.dohLookup = func(q, p string) ([]string, error) { return []string{slowProxy.URL}, nil }

	// The findProxy should fail because lookup takes 2 seconds but we only allow 1 second.
	_, err := p.findProxy()
	require.Error(t, err)
}

func TestProxyManager_UseProxy(t *testing.T) {
	blockAPI()
	defer unblockAPI()

	proxy := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
	defer proxy.Close()

	p := newProxyManager([]string{"not used"}, "not used")
	p.dohLookup = func(q, p string) ([]string, error) { return []string{proxy.URL}, nil }

	url, err := p.findProxy()
	require.NoError(t, err)

	p.useProxy(url)
	require.Equal(t, proxy.URL, GlobalGetRootURL())
}

func TestProxyManager_UseProxy_MultipleTimes(t *testing.T) {
	blockAPI()
	defer unblockAPI()

	proxy1 := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
	defer proxy1.Close()
	proxy2 := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
	defer proxy2.Close()
	proxy3 := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
	defer proxy3.Close()

	p := newProxyManager([]string{"not used"}, "not used")

	p.dohLookup = func(q, p string) ([]string, error) { return []string{proxy1.URL}, nil }
	url, err := p.findProxy()
	require.NoError(t, err)
	p.useProxy(url)
	require.Equal(t, proxy1.URL, GlobalGetRootURL())

	// Have to wait so as to not get rejected.
	time.Sleep(proxyLookupWait)

	p.dohLookup = func(q, p string) ([]string, error) { return []string{proxy2.URL}, nil }
	url, err = p.findProxy()
	require.NoError(t, err)
	p.useProxy(url)
	require.Equal(t, proxy2.URL, GlobalGetRootURL())

	// Have to wait so as to not get rejected.
	time.Sleep(proxyLookupWait)

	p.dohLookup = func(q, p string) ([]string, error) { return []string{proxy3.URL}, nil }
	url, err = p.findProxy()
	require.NoError(t, err)
	p.useProxy(url)
	require.Equal(t, proxy3.URL, GlobalGetRootURL())
}

func TestProxyManager_UseProxy_RevertAfterTime(t *testing.T) {
	blockAPI()
	defer unblockAPI()

	proxy := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
	defer proxy.Close()

	p := newProxyManager([]string{"not used"}, "not used")
	p.useDuration = time.Second
	p.dohLookup = func(q, p string) ([]string, error) { return []string{proxy.URL}, nil }

	url, err := p.findProxy()
	require.NoError(t, err)
	require.Equal(t, proxy.URL, url)

	p.useProxy(url)
	require.Equal(t, proxy.URL, GlobalGetRootURL())

	time.Sleep(2 * time.Second)
	require.Equal(t, globalOriginalURL, GlobalGetRootURL())
}

func TestProxyManager_UseProxy_RevertIfProxyStopsWorkingAndOriginalAPIIsReachable(t *testing.T) {
	// Don't block the API here because we want it to be working so the test can find it.
	defer unblockAPI()

	proxy := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
	defer proxy.Close()

	p := newProxyManager([]string{"not used"}, "not used")
	p.dohLookup = func(q, p string) ([]string, error) { return []string{proxy.URL}, nil }

	url, err := p.findProxy()
	require.NoError(t, err)
	require.Equal(t, proxy.URL, url)

	p.useProxy(url)
	require.Equal(t, proxy.URL, GlobalGetRootURL())

	// Simulate that the proxy stops working.
	proxy.Close()
	time.Sleep(proxyLookupWait)

	// We should now find the original API URL if it is working again.
	url, err = p.findProxy()
	require.NoError(t, err)
	require.Equal(t, globalOriginalURL, url)

	p.useProxy(url)
	require.Equal(t, globalOriginalURL, GlobalGetRootURL())
}

func TestProxyManager_UseProxy_FindSecondAlternativeIfFirstFailsAndAPIIsStillBlocked(t *testing.T) {
	blockAPI()
	defer unblockAPI()

	proxy1 := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
	defer proxy1.Close()
	proxy2 := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
	defer proxy2.Close()

	p := newProxyManager([]string{"not used"}, "not used")
	p.dohLookup = func(q, p string) ([]string, error) { return []string{proxy1.URL, proxy2.URL}, nil }

	// Find a proxy.
	url, err := p.findProxy()
	require.NoError(t, err)
	p.useProxy(url)
	require.Equal(t, proxy1.URL, GlobalGetRootURL())

	// Have to wait so as to not get rejected.
	time.Sleep(proxyLookupWait)

	// The proxy stops working and the protonmail API is still blocked.
	proxy1.Close()

	// Should switch to the second proxy because both the first proxy and the protonmail API are blocked.
	url, err = p.findProxy()
	require.NoError(t, err)
	p.useProxy(url)
	require.Equal(t, proxy2.URL, GlobalGetRootURL())
}

func TestProxyManager_DoHLookup_Quad9(t *testing.T) {
	p := newProxyManager([]string{TestQuad9Provider, TestGoogleProvider}, TestDoHQuery)

	records, err := p.dohLookup(TestDoHQuery, TestQuad9Provider)
	require.NoError(t, err)
	require.NotEmpty(t, records)
}

func TestProxyManager_DoHLookup_Google(t *testing.T) {
	p := newProxyManager([]string{TestQuad9Provider, TestGoogleProvider}, TestDoHQuery)

	records, err := p.dohLookup(TestDoHQuery, TestGoogleProvider)
	require.NoError(t, err)
	require.NotEmpty(t, records)
}

func TestProxyManager_DoHLookup_FindProxy(t *testing.T) {
	p := newProxyManager([]string{TestQuad9Provider, TestGoogleProvider}, TestDoHQuery)

	url, err := p.findProxy()
	require.NoError(t, err)
	require.NotEmpty(t, url)
}

func TestProxyManager_DoHLookup_FindProxyFirstProviderUnreachable(t *testing.T) {
	p := newProxyManager([]string{"https://unreachable", TestGoogleProvider}, TestDoHQuery)

	url, err := p.findProxy()
	require.NoError(t, err)
	require.NotEmpty(t, url)
}

// testAPIURLBackup is used to hold the globalOriginalURL because we clear it for test purposes and need to restore it.
var testAPIURLBackup = globalOriginalURL

// blockAPI prevents tests from reaching the standard API, forcing them to find a proxy.
func blockAPI() {
	globalSetRootURL("")
	globalOriginalURL = ""
}

// unblockAPI allow tests to reach the standard API again.
func unblockAPI() {
	globalOriginalURL = testAPIURLBackup
	globalSetRootURL(globalOriginalURL)
}
