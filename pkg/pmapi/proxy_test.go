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
	"crypto/tls"
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

// getTrustedServer returns a server and sets its public key as one of the pinned ones.
func getTrustedServer() *httptest.Server {
	proxy := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))

	pin := certFingerprint(proxy.Certificate())
	TrustedAPIPins = append(TrustedAPIPins, pin)

	return proxy
}

// server.crt
const servercrt = `
-----BEGIN CERTIFICATE-----
MIIE5TCCA82gAwIBAgIJAKsmhcMFGfGcMA0GCSqGSIb3DQEBCwUAMIGsMQswCQYD
VQQGEwJVUzEUMBIGA1UECAwLUmFuZG9tU3RhdGUxEzARBgNVBAcMClJhbmRvbUNp
dHkxGzAZBgNVBAoMElJhbmRvbU9yZ2FuaXphdGlvbjEfMB0GA1UECwwWUmFuZG9t
T3JnYW5pemF0aW9uVW5pdDEgMB4GCSqGSIb3DQEJARYRaGVsbG9AZXhhbXBsZS5j
b20xEjAQBgNVBAMMCTEyNy4wLjAuMTAeFw0yMDA0MjQxMzI3MzdaFw0yMTA5MDYx
MzI3MzdaMIGsMQswCQYDVQQGEwJVUzEUMBIGA1UECAwLUmFuZG9tU3RhdGUxEzAR
BgNVBAcMClJhbmRvbUNpdHkxGzAZBgNVBAoMElJhbmRvbU9yZ2FuaXphdGlvbjEf
MB0GA1UECwwWUmFuZG9tT3JnYW5pemF0aW9uVW5pdDEgMB4GCSqGSIb3DQEJARYR
aGVsbG9AZXhhbXBsZS5jb20xEjAQBgNVBAMMCTEyNy4wLjAuMTCCASIwDQYJKoZI
hvcNAQEBBQADggEPADCCAQoCggEBANAnYyqhosWwNzGjBwSwmDUINOaPs4TSTgKt
r6CE01atxAWzWUCyYqnQ4fPe5q2tx5t/VrmnTNpzycammKJszGLlmj9DFxSiYVw2
pTTK3DBWFkfTwxq98mM7wMnCWy1T2L2pmuYjnd7Pa6pQa9OHYoJwRzlIl2Q3YVdM
GIBDbkW728A1dcelkIdFpv3r3ayTZv01vU8JMXd4PLHwXU0x0hHlH52+kx+9Ndru
rdqqV6LqVfNlSR1jFZkwLBBqvh3XrJRD9Q01EAX6m+ufZ0yq8mK9ifMRtwQet10c
kKMnx63MwvxDFmqrBj4HMtIRUpK+LBDs1ke7DvS0eLqaojWl28ECAwEAAaOCAQYw
ggECMIHLBgNVHSMEgcMwgcChgbKkga8wgawxCzAJBgNVBAYTAlVTMRQwEgYDVQQI
DAtSYW5kb21TdGF0ZTETMBEGA1UEBwwKUmFuZG9tQ2l0eTEbMBkGA1UECgwSUmFu
ZG9tT3JnYW5pemF0aW9uMR8wHQYDVQQLDBZSYW5kb21Pcmdhbml6YXRpb25Vbml0
MSAwHgYJKoZIhvcNAQkBFhFoZWxsb0BleGFtcGxlLmNvbTESMBAGA1UEAwwJMTI3
LjAuMC4xggkAvCxbs152YckwCQYDVR0TBAIwADALBgNVHQ8EBAMCBPAwGgYDVR0R
BBMwEYIJMTI3LjAuMC4xhwR/AAABMA0GCSqGSIb3DQEBCwUAA4IBAQAC7ZycZMZ5
L+cjIpwSj0cemLkVD+kcFUCkI7ket5gbX1PmavmnpuFl9Sru0eJ5wyJ+97MQElPA
CNFgXoX7DbJWkcd/LSksvZoJnpc1sTqFKMWFmOUxmUD62lCacuhqE27ZTThQ/53P
3doLa74rKzUqlPI8OL4R34FY2deL7t5l2KSnpf7CKNeF5bkinAsn6NBqyZs2KPmg
yT1/POdlRewzGSqBTMdktNQ4vKSfdFjcfVeo8PSHBgbGXZ5KoHZ6R6DNJehEh27l
z3OteROLGoii+w3OllLq6JATif2MDIbH0s/KjGjbXSSGbM/rZu5eBZm5/vksGAzc
u53wgIhCJGuX
-----END CERTIFICATE-----
`

const serverkey = `
-----BEGIN PRIVATE KEY-----
MIIEvAIBADANBgkqhkiG9w0BAQEFAASCBKYwggSiAgEAAoIBAQDQJ2MqoaLFsDcx
owcEsJg1CDTmj7OE0k4Cra+ghNNWrcQFs1lAsmKp0OHz3uatrcebf1a5p0zac8nG
ppiibMxi5Zo/QxcUomFcNqU0ytwwVhZH08MavfJjO8DJwlstU9i9qZrmI53ez2uq
UGvTh2KCcEc5SJdkN2FXTBiAQ25Fu9vANXXHpZCHRab9692sk2b9Nb1PCTF3eDyx
8F1NMdIR5R+dvpMfvTXa7q3aqlei6lXzZUkdYxWZMCwQar4d16yUQ/UNNRAF+pvr
n2dMqvJivYnzEbcEHrddHJCjJ8etzML8QxZqqwY+BzLSEVKSviwQ7NZHuw70tHi6
mqI1pdvBAgMBAAECggEAOqqPOYm63arPs462QK0hCPlaJ41i1FGNqRWYxU4KXoi1
EcI9qo1cX24+8MPnEhZDhuD56XNsprkxqmpz5Htzk4AQ3DmlfKxTcnD4WQu/yWPJ
/c6CU7wrX6qMqJC9r+XM1Y/C15A8Q3sEZkkqSsECk67fdBawjI9LQRZyZVwb7U0F
qtvbKM7VQA6hrgdSmXWJ+spp5yymVFF22Ssz31SSbCI93bnp3mukRCKWdRmA9pmT
VXa0HzJ5p70WC+Se9nA/1riWGKt4HCmjVeEtZuiwaUTlXDSeYpu2e4QrX1OnUXBu
Z7yfviTqA8o7KfiA6urumFbAMJcibxkWJoWacc5tTQKBgQD39ZdtNz8B6XJy7f5h
bo9Ag9OrkVX+HITQyWKpcCDba9SuIX3/F++2AK4oeJ3aHKMJWiP19hQvGS1xE67X
TKejOsQxORn6nAYQpFd3AOBOtKAC+VQITBqlfq2ukGmvcQ1O31hMOFbZagFA5cpU
LYb9VVDsZzhM7CccIn/EGEZjgwKBgQDW51rUA2S9naV/iEGhw1tuhoQ5OADD/n8f
pPIkbGxmACDaX/7jt+UwlDU0EsI+aBlJUDqGiEZ5z3UPmaSJUdfRCeJEdKIe1GLm
nqF3sF6Aq+S/79v/wKYn+MHcoiWog5n3McLzZ3+0rwrhMREjE2eWPwVHz/jJIFP3
Pp3+UZVsawKBgB4Az5PdjXgzwS968L7lW9wYl3I5Iciftsp0s8WA1dj3EUMItnA5
ez3wkyI+hgswT+H/0D4gyoxwZXk7Qnq2wcoUgEzcdfJHEszMtfCmYH3liT8S4EIo
w0inLWjj/IXIDi4vBEYkww2HsCMkKvlIkP7yZdpVGxDjuk/DNOaLcWj1AoGAXuyK
PiPRl7/Onmp9MwqrlEJunSeTjv8W/89H9ba+mr9rw4mreMJ9xdtxNLMkgZRRtwRt
FYeUObHdLyradp1kCr2m6D3sblm55cwj3k5VL9i9jdpQ/sMFoZpLZz1oDOs0Uu/0
ALeyvQikcZvOygOEOeVUW8gNSCmzbP6HoxI+QkkCgYBCI6oL4GPcPPqzd+2djbOD
z3rVUyHzYc1KUcBixK/uaRQKM886k4CL8/GvbHHI/yoZ7xWJGnBi59DtpqnGTZJ2
FDJwYIlQKhZmsyVcZu/4smsaejGnHn/liksVlgesSwCtOrsd2AC8fBXSyrTWJx8o
vwRMog6lPhlRhHh/FZ43Cg==
-----END PRIVATE KEY-----
`

// getUntrustedServer returns a server but it doesn't add its public key to the list of pinned ones.
func getUntrustedServer() *httptest.Server {
	server := httptest.NewUnstartedServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))

	cert, err := tls.X509KeyPair([]byte(servercrt), []byte(serverkey))
	if err != nil {
		panic(err)
	}
	server.TLS = &tls.Config{Certificates: []tls.Certificate{cert}}

	server.StartTLS()
	return server
}

// closeServer closes the given server. If it is a trusted server, its cert is removed from the trusted public keys.
func closeServer(server *httptest.Server) {
	pin := certFingerprint(server.Certificate())

	for i := range TrustedAPIPins {
		if TrustedAPIPins[i] == pin {
			TrustedAPIPins = append(TrustedAPIPins[:i], TrustedAPIPins[i:]...)
			break
		}
	}

	server.Close()
}

func TestProxyProvider_FindProxy(t *testing.T) {
	blockAPI()
	defer unblockAPI()

	proxy := getTrustedServer()
	defer closeServer(proxy)

	p := newProxyProvider([]string{"not used"}, "not used")
	p.dohLookup = func(q, p string) ([]string, error) { return []string{proxy.URL}, nil }

	url, err := p.findReachableServer()
	require.NoError(t, err)
	require.Equal(t, proxy.URL, url)
}

func TestProxyProvider_FindProxy_ChooseReachableProxy(t *testing.T) {
	blockAPI()
	defer unblockAPI()

	reachableProxy := getTrustedServer()
	defer closeServer(reachableProxy)

	// We actually close the unreachable proxy straight away rather than deferring the closure.
	unreachableProxy := getTrustedServer()
	closeServer(unreachableProxy)

	p := newProxyProvider([]string{"not used"}, "not used")
	p.dohLookup = func(q, p string) ([]string, error) { return []string{reachableProxy.URL, unreachableProxy.URL}, nil }

	url, err := p.findReachableServer()
	require.NoError(t, err)
	require.Equal(t, reachableProxy.URL, url)
}

func TestProxyProvider_FindProxy_ChooseTrustedProxy(t *testing.T) {
	blockAPI()
	defer unblockAPI()

	trustedProxy := getTrustedServer()
	defer closeServer(trustedProxy)

	untrustedProxy := getUntrustedServer()
	defer closeServer(untrustedProxy)

	p := newProxyProvider([]string{"not used"}, "not used")
	p.dohLookup = func(q, p string) ([]string, error) { return []string{untrustedProxy.URL, trustedProxy.URL}, nil }

	url, err := p.findReachableServer()
	require.NoError(t, err)
	require.Equal(t, trustedProxy.URL, url)
}

func TestProxyProvider_FindProxy_FailIfNoneReachable(t *testing.T) {
	blockAPI()
	defer unblockAPI()

	unreachableProxy1 := getTrustedServer()
	closeServer(unreachableProxy1)

	unreachableProxy2 := getTrustedServer()
	closeServer(unreachableProxy2)

	p := newProxyProvider([]string{"not used"}, "not used")
	p.dohLookup = func(q, p string) ([]string, error) {
		return []string{unreachableProxy1.URL, unreachableProxy2.URL}, nil
	}

	_, err := p.findReachableServer()
	require.Error(t, err)
}

func TestProxyProvider_FindProxy_FailIfNoneTrusted(t *testing.T) {
	blockAPI()
	defer unblockAPI()

	untrustedProxy1 := getUntrustedServer()
	defer closeServer(untrustedProxy1)

	untrustedProxy2 := getUntrustedServer()
	defer closeServer(untrustedProxy2)

	p := newProxyProvider([]string{"not used"}, "not used")
	p.dohLookup = func(q, p string) ([]string, error) {
		return []string{untrustedProxy1.URL, untrustedProxy2.URL}, nil
	}

	_, err := p.findReachableServer()
	require.Error(t, err)
}

func TestProxyProvider_FindProxy_LookupTimeout(t *testing.T) {
	blockAPI()
	defer unblockAPI()

	p := newProxyProvider([]string{"not used"}, "not used")
	p.lookupTimeout = time.Second
	p.dohLookup = func(q, p string) ([]string, error) { time.Sleep(2 * time.Second); return nil, nil }

	// The findReachableServer should fail because lookup takes 2 seconds but we only allow 1 second.
	_, err := p.findReachableServer()
	require.Error(t, err)
}

func TestProxyProvider_FindProxy_FindTimeout(t *testing.T) {
	blockAPI()
	defer unblockAPI()

	slowProxy := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(2 * time.Second)
	}))
	defer slowProxy.Close()

	p := newProxyProvider([]string{"not used"}, "not used")
	p.findTimeout = time.Second
	p.dohLookup = func(q, p string) ([]string, error) { return []string{slowProxy.URL}, nil }

	// The findReachableServer should fail because lookup takes 2 seconds but we only allow 1 second.
	_, err := p.findReachableServer()
	require.Error(t, err)
}

func TestProxyProvider_UseProxy(t *testing.T) {
	blockAPI()
	defer unblockAPI()

	cm := newTestClientManager(testClientConfig)

	trustedProxy := getTrustedServer()
	defer closeServer(trustedProxy)

	p := newProxyProvider([]string{"not used"}, "not used")
	cm.proxyProvider = p

	p.dohLookup = func(q, p string) ([]string, error) { return []string{trustedProxy.URL}, nil }
	url, err := cm.switchToReachableServer()
	require.NoError(t, err)
	require.Equal(t, trustedProxy.URL, url)
	require.Equal(t, trustedProxy.URL, cm.getHost())
}

func TestProxyProvider_UseProxy_MultipleTimes(t *testing.T) {
	blockAPI()
	defer unblockAPI()

	cm := newTestClientManager(testClientConfig)

	proxy1 := getTrustedServer()
	defer closeServer(proxy1)
	proxy2 := getTrustedServer()
	defer closeServer(proxy2)
	proxy3 := getTrustedServer()
	defer closeServer(proxy3)

	p := newProxyProvider([]string{"not used"}, "not used")
	cm.proxyProvider = p

	p.dohLookup = func(q, p string) ([]string, error) { return []string{proxy1.URL}, nil }
	url, err := cm.switchToReachableServer()
	require.NoError(t, err)
	require.Equal(t, proxy1.URL, url)
	require.Equal(t, proxy1.URL, cm.getHost())

	// Have to wait so as to not get rejected.
	time.Sleep(proxyLookupWait)

	p.dohLookup = func(q, p string) ([]string, error) { return []string{proxy2.URL}, nil }
	url, err = cm.switchToReachableServer()
	require.NoError(t, err)
	require.Equal(t, proxy2.URL, url)
	require.Equal(t, proxy2.URL, cm.getHost())

	// Have to wait so as to not get rejected.
	time.Sleep(proxyLookupWait)

	p.dohLookup = func(q, p string) ([]string, error) { return []string{proxy3.URL}, nil }
	url, err = cm.switchToReachableServer()
	require.NoError(t, err)
	require.Equal(t, proxy3.URL, url)
	require.Equal(t, proxy3.URL, cm.getHost())
}

func TestProxyProvider_UseProxy_RevertAfterTime(t *testing.T) {
	blockAPI()
	defer unblockAPI()

	cm := newTestClientManager(testClientConfig)

	trustedProxy := getTrustedServer()
	defer closeServer(trustedProxy)

	p := newProxyProvider([]string{"not used"}, "not used")
	cm.proxyProvider = p
	cm.proxyUseDuration = time.Second

	p.dohLookup = func(q, p string) ([]string, error) { return []string{trustedProxy.URL}, nil }
	url, err := cm.switchToReachableServer()
	require.NoError(t, err)
	require.Equal(t, trustedProxy.URL, url)
	require.Equal(t, trustedProxy.URL, cm.getHost())

	time.Sleep(2 * time.Second)
	require.Equal(t, rootURL, cm.getHost())
}

func TestProxyProvider_UseProxy_RevertIfProxyStopsWorkingAndOriginalAPIIsReachable(t *testing.T) {
	blockAPI()
	defer unblockAPI()

	cm := newTestClientManager(testClientConfig)

	trustedProxy := getTrustedServer()

	p := newProxyProvider([]string{"not used"}, "not used")
	cm.proxyProvider = p

	p.dohLookup = func(q, p string) ([]string, error) { return []string{trustedProxy.URL}, nil }
	url, err := cm.switchToReachableServer()
	require.NoError(t, err)
	require.Equal(t, trustedProxy.URL, url)
	require.Equal(t, trustedProxy.URL, cm.getHost())

	// Simulate that the proxy stops working and that the standard api is reachable again.
	closeServer(trustedProxy)
	unblockAPI()
	time.Sleep(proxyLookupWait)

	// We should now find the original API URL if it is working again.
	// The error should be ErrAPINotReachable because the connection dropped intermittently but
	// the original API is now reachable (see Alternative-Routing-v2 spec for details).
	url, err = cm.switchToReachableServer()
	require.EqualError(t, err, ErrAPINotReachable.Error())
	require.Equal(t, rootURL, url)
	require.Equal(t, rootURL, cm.getHost())
}

func TestProxyProvider_UseProxy_FindSecondAlternativeIfFirstFailsAndAPIIsStillBlocked(t *testing.T) {
	blockAPI()
	defer unblockAPI()

	cm := newTestClientManager(testClientConfig)

	// proxy1 is closed later in this test so we don't defer it here.
	proxy1 := getTrustedServer()

	proxy2 := getTrustedServer()
	defer closeServer(proxy2)

	p := newProxyProvider([]string{"not used"}, "not used")
	cm.proxyProvider = p

	// Find a proxy.
	p.dohLookup = func(q, p string) ([]string, error) { return []string{proxy1.URL, proxy2.URL}, nil }
	url, err := cm.switchToReachableServer()
	require.NoError(t, err)
	require.Equal(t, proxy1.URL, url)
	require.Equal(t, proxy1.URL, cm.getHost())

	// Have to wait so as to not get rejected.
	time.Sleep(proxyLookupWait)

	// The proxy stops working and the protonmail API is still blocked.
	proxy1.Close()

	// Should switch to the second proxy because both the first proxy and the protonmail API are blocked.
	url, err = cm.switchToReachableServer()
	require.NoError(t, err)
	require.Equal(t, proxy2.URL, url)
	require.Equal(t, proxy2.URL, cm.getHost())
}

func TestProxyProvider_DoHLookup_Quad9(t *testing.T) {
	p := newProxyProvider([]string{TestQuad9Provider, TestGoogleProvider}, TestDoHQuery)

	records, err := p.dohLookup(TestDoHQuery, TestQuad9Provider)
	require.NoError(t, err)
	require.NotEmpty(t, records)
}

func TestProxyProvider_DoHLookup_Google(t *testing.T) {
	p := newProxyProvider([]string{TestQuad9Provider, TestGoogleProvider}, TestDoHQuery)

	records, err := p.dohLookup(TestDoHQuery, TestGoogleProvider)
	require.NoError(t, err)
	require.NotEmpty(t, records)
}

func TestProxyProvider_DoHLookup_FindProxy(t *testing.T) {
	p := newProxyProvider([]string{TestQuad9Provider, TestGoogleProvider}, TestDoHQuery)

	url, err := p.findReachableServer()
	require.NoError(t, err)
	require.NotEmpty(t, url)
}

func TestProxyProvider_DoHLookup_FindProxyFirstProviderUnreachable(t *testing.T) {
	p := newProxyProvider([]string{"https://unreachable", TestGoogleProvider}, TestDoHQuery)

	url, err := p.findReachableServer()
	require.NoError(t, err)
	require.NotEmpty(t, url)
}

// testAPIURLBackup is used to hold the globalOriginalURL because we clear it for test purposes and need to restore it.
var testAPIURLBackup = rootURL

// blockAPI prevents tests from reaching the standard API, forcing them to find a proxy.
func blockAPI() {
	rootURL = ""
}

// unblockAPI allow tests to reach the standard API again.
func unblockAPI() {
	rootURL = testAPIURLBackup
}
