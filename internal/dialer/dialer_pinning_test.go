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

package dialer

import (
	"context"
	"testing"
	"time"

	"github.com/ProtonMail/proton-bridge/v2/internal/useragent"
	a "github.com/stretchr/testify/assert"
	r "github.com/stretchr/testify/require"
	"gitlab.protontech.ch/go/liteapi"
	"gitlab.protontech.ch/go/liteapi/server"
)

func getRootURL() string {
	return "https://api.protonmail.ch"
}

func TestTLSPinValid(t *testing.T) {
	called, _, _, _, cm := createClientWithPinningDialer(getRootURL())

	_, _, _ = cm.NewClientWithLogin(context.Background(), "username", "password")

	checkTLSIssueHandler(t, 0, called)
}

func TestTLSPinBackup(t *testing.T) {
	called, _, _, checker, cm := createClientWithPinningDialer(getRootURL())
	copyTrustedPins(checker)
	checker.trustedPins[1] = checker.trustedPins[0]
	checker.trustedPins[0] = ""

	_, _, _ = cm.NewClientWithLogin(context.Background(), "username", "password")

	checkTLSIssueHandler(t, 0, called)
}

func TestTLSPinInvalid(t *testing.T) {
	s := server.NewTLS()
	defer s.Close()

	called, _, _, _, cm := createClientWithPinningDialer(s.GetHostURL())

	_, _, _ = cm.NewClientWithLogin(context.Background(), "username", "password")

	checkTLSIssueHandler(t, 1, called)
}

func TestTLSPinNoMatch(t *testing.T) {
	skipIfProxyIsSet(t)

	called, _, reporter, checker, cm := createClientWithPinningDialer(getRootURL())

	copyTrustedPins(checker)
	for i := 0; i < len(checker.trustedPins); i++ {
		checker.trustedPins[i] = "testing"
	}

	_, _, _ = cm.NewClientWithLogin(context.Background(), "username", "password")
	_, _, _ = cm.NewClientWithLogin(context.Background(), "username", "password")

	// Check that it will be reported only once per session, but notified every time.
	r.Equal(t, 1, len(reporter.sentReports))
	checkTLSIssueHandler(t, 2, called)
}

func TestTLSSignedCertWrongPublicKey(t *testing.T) {
	skipIfProxyIsSet(t)

	_, dialer, _, _, _ := createClientWithPinningDialer("")
	_, err := dialer.DialTLSContext(context.Background(), "tcp", "rsa4096.badssl.com:443")
	r.Error(t, err, "expected dial to fail because of wrong public key")
}

func TestTLSSignedCertTrustedPublicKey(t *testing.T) {
	skipIfProxyIsSet(t)

	_, dialer, _, checker, _ := createClientWithPinningDialer("")
	copyTrustedPins(checker)
	checker.trustedPins = append(checker.trustedPins, `pin-sha256="LwnIKjNLV3z243ap8y0yXNPghsqE76J08Eq3COvUt2E="`)
	_, err := dialer.DialTLSContext(context.Background(), "tcp", "rsa4096.badssl.com:443")
	r.NoError(t, err, "expected dial to succeed because public key is known and cert is signed by CA")
}

func TestTLSSelfSignedCertTrustedPublicKey(t *testing.T) {
	skipIfProxyIsSet(t)

	_, dialer, _, checker, _ := createClientWithPinningDialer("")
	copyTrustedPins(checker)
	checker.trustedPins = append(checker.trustedPins, `pin-sha256="9SLklscvzMYj8f+52lp5ze/hY0CFHyLSPQzSpYYIBm8="`)
	_, err := dialer.DialTLSContext(context.Background(), "tcp", "self-signed.badssl.com:443")
	r.NoError(t, err, "expected dial to succeed because public key is known despite cert being self-signed")
}

func createClientWithPinningDialer(hostURL string) (*int, *PinningTLSDialer, *TLSReporter, *TLSPinChecker, *liteapi.Manager) {
	called := 0

	reporter := NewTLSReporter(hostURL, "appVersion", useragent.New(), TrustedAPIPins)
	checker := NewTLSPinChecker(TrustedAPIPins)
	dialer := NewPinningTLSDialer(NewBasicTLSDialer(hostURL), reporter, checker)

	go func() {
		for range dialer.GetTLSIssueCh() {
			called++
		}
	}()

	return &called, dialer, reporter, checker, liteapi.New(
		liteapi.WithHostURL(hostURL),
		liteapi.WithTransport(CreateTransportWithDialer(dialer)),
	)
}

func copyTrustedPins(pinChecker *TLSPinChecker) {
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
