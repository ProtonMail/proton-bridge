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
	"net/http/httptest"
	"testing"
	"time"

	a "github.com/stretchr/testify/assert"
	r "github.com/stretchr/testify/require"
)

func TestTLSPinValid(t *testing.T) {
	called, _, cm := createClientWithPinningDialer(getRootURL())

	_, _ = cm.getAuthInfo(context.Background(), GetAuthInfoReq{Username: "username"})
	checkTLSIssueHandler(t, 0, called)
}

func TestTLSPinBackup(t *testing.T) {
	called, dialer, cm := createClientWithPinningDialer(getRootURL())
	copyTrustedPins(dialer.pinChecker)
	dialer.pinChecker.trustedPins[1] = dialer.pinChecker.trustedPins[0]
	dialer.pinChecker.trustedPins[0] = ""

	_, _ = cm.getAuthInfo(context.Background(), GetAuthInfoReq{Username: "username"})
	checkTLSIssueHandler(t, 0, called)
}

func TestTLSPinInvalid(t *testing.T) {
	ts := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		writeJSONResponsefromFile(t, w, "/auth/info/post_response.json", 0)
	}))
	defer ts.Close()

	called, _, cm := createClientWithPinningDialer(ts.URL)

	_, _ = cm.getAuthInfo(context.Background(), GetAuthInfoReq{Username: "username"})
	checkTLSIssueHandler(t, 1, called)
}

func TestTLSPinNoMatch(t *testing.T) {
	skipIfProxyIsSet(t)

	called, dialer, cm := createClientWithPinningDialer(getRootURL())

	copyTrustedPins(dialer.pinChecker)
	for i := 0; i < len(dialer.pinChecker.trustedPins); i++ {
		dialer.pinChecker.trustedPins[i] = "testing"
	}

	_, _ = cm.getAuthInfo(context.Background(), GetAuthInfoReq{Username: "username"})
	_, _ = cm.getAuthInfo(context.Background(), GetAuthInfoReq{Username: "username"})

	// Check that it will be reported only once per session, but notified every time.
	r.Equal(t, 1, len(dialer.reporter.sentReports))
	checkTLSIssueHandler(t, 2, called)
}

func TestTLSSignedCertWrongPublicKey(t *testing.T) {
	skipIfProxyIsSet(t)

	_, dialer, _ := createClientWithPinningDialer("")
	_, err := dialer.DialTLS("tcp", "rsa4096.badssl.com:443")
	r.Error(t, err, "expected dial to fail because of wrong public key")
}

func TestTLSSignedCertTrustedPublicKey(t *testing.T) {
	skipIfProxyIsSet(t)

	_, dialer, _ := createClientWithPinningDialer("")
	copyTrustedPins(dialer.pinChecker)
	dialer.pinChecker.trustedPins = append(dialer.pinChecker.trustedPins, `pin-sha256="2opdB7b5INED5jS7duIDR7dM8Er99i7trnwKuW3GMCY="`)
	_, err := dialer.DialTLS("tcp", "rsa4096.badssl.com:443")
	r.NoError(t, err, "expected dial to succeed because public key is known and cert is signed by CA")
}

func TestTLSSelfSignedCertTrustedPublicKey(t *testing.T) {
	skipIfProxyIsSet(t)

	_, dialer, _ := createClientWithPinningDialer("")
	copyTrustedPins(dialer.pinChecker)
	dialer.pinChecker.trustedPins = append(dialer.pinChecker.trustedPins, `pin-sha256="9SLklscvzMYj8f+52lp5ze/hY0CFHyLSPQzSpYYIBm8="`)
	_, err := dialer.DialTLS("tcp", "self-signed.badssl.com:443")
	r.NoError(t, err, "expected dial to succeed because public key is known despite cert being self-signed")
}

func createClientWithPinningDialer(hostURL string) (*int, *PinningTLSDialer, *manager) {
	called := 0

	cfg := Config{
		AppVersion:      "Bridge_1.2.4-test",
		HostURL:         hostURL,
		TLSIssueHandler: func() { called++ },
	}

	dialer := NewPinningTLSDialer(cfg, NewBasicTLSDialer(cfg))

	cm := newManager(cfg)
	cm.SetTransport(CreateTransportWithDialer(dialer))

	return &called, dialer, cm
}

func copyTrustedPins(pinChecker *pinChecker) {
	copiedPins := make([]string, len(pinChecker.trustedPins))
	copy(copiedPins, pinChecker.trustedPins)
	pinChecker.trustedPins = copiedPins
}

func checkTLSIssueHandler(t *testing.T, wantCalledAtLeast int, called *int) {
	// TLSIssueHandler is called in goroutine se we need to wait a bit to be sure it was called.
	a.Eventually(
		t,
		func() bool {
			if wantCalledAtLeast == 0 {
				return *called == 0
			}
			// Dialer can do more attempts resulting in more calls.
			return *called >= wantCalledAtLeast
		},
		time.Second,
		10*time.Millisecond,
	)
	// Repeated again so it generates nice message.
	if wantCalledAtLeast == 0 {
		r.Equal(t, 0, *called)
	} else {
		r.GreaterOrEqual(t, *called, wantCalledAtLeast)
	}
}
