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

	"github.com/stretchr/testify/assert"
)

const liveAPI = "api.protonmail.ch"

var testLiveConfig = &ClientConfig{
	AppVersion: "Bridge_1.2.4-test",
	ClientID:   "Bridge",
}

func createAndSetPinningDialer(cm *ClientManager) (*int, *PinningTLSDialer) {
	called := 0

	dialer := NewPinningTLSDialer(NewBasicTLSDialer())
	dialer.SetTLSIssueNotifier(func() { called++ })
	cm.SetRoundTripper(CreateTransportWithDialer(dialer))

	return &called, dialer
}

func TestTLSPinValid(t *testing.T) {
	cm := newTestClientManager(testLiveConfig)
	cm.host = liveAPI
	rootScheme = "https"
	called, _ := createAndSetPinningDialer(cm)
	client := cm.GetClient("pmapi" + t.Name())

	_, err := client.AuthInfo("this.address.is.disabled")
	Ok(t, err)

	Equals(t, 0, *called)
}

func TestTLSPinBackup(t *testing.T) {
	cm := newTestClientManager(testLiveConfig)
	cm.host = liveAPI
	called, p := createAndSetPinningDialer(cm)
	p.pinChecker.trustedPins[1] = p.pinChecker.trustedPins[0]
	p.pinChecker.trustedPins[0] = ""

	client := cm.GetClient("pmapi" + t.Name())

	_, err := client.AuthInfo("this.address.is.disabled")
	Ok(t, err)

	Equals(t, 0, *called)
}

func _TestTLSPinNoMatch(t *testing.T) { // nolint[unused]
	cm := newTestClientManager(testLiveConfig)
	cm.host = liveAPI

	called, p := createAndSetPinningDialer(cm)
	for i := 0; i < len(p.pinChecker.trustedPins); i++ {
		p.pinChecker.trustedPins[i] = "testing"
	}

	client := cm.GetClient("pmapi" + t.Name())

	_, err := client.AuthInfo("this.address.is.disabled")
	Ok(t, err)

	// check that it will be called only once per session
	client = cm.GetClient("pmapi" + t.Name())
	_, err = client.AuthInfo("this.address.is.disabled")
	Ok(t, err)

	Equals(t, 1, *called)
}

func _TestTLSPinInvalid(t *testing.T) { // nolint[unused]
	cm := newTestClientManager(testLiveConfig)

	ts := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		writeJSONResponsefromFile(t, w, "/auth/info/post_response.json", 0)
	}))
	defer ts.Close()

	called, _ := createAndSetPinningDialer(cm)

	client := cm.GetClient("pmapi" + t.Name())

	cm.host = liveAPI
	_, err := client.AuthInfo("this.address.is.disabled")
	Ok(t, err)

	cm.host = ts.URL
	_, err = client.AuthInfo("this.address.is.disabled")
	Assert(t, err != nil, "error is expected but have %v", err)

	Equals(t, 1, *called)
}

func TestTLSSignedCertWrongPublicKey(t *testing.T) { // nolint[unused]
	cm := newTestClientManager(testLiveConfig)
	_, dialer := createAndSetPinningDialer(cm)
	_, err := dialer.DialTLS("tcp", "rsa4096.badssl.com:443")
	assert.Error(t, err, "expected dial to fail because of wrong public key")
}

func TestTLSSignedCertTrustedPublicKey(t *testing.T) { // nolint[unused]
	cm := newTestClientManager(testLiveConfig)
	_, dialer := createAndSetPinningDialer(cm)
	dialer.pinChecker.trustedPins = append(dialer.pinChecker.trustedPins, `pin-sha256="W8/42Z0ffufwnHIOSndT+eVzBJSC0E8uTIC8O6mEliQ="`)
	_, err := dialer.DialTLS("tcp", "rsa4096.badssl.com:443")
	assert.NoError(t, err, "expected dial to succeed because public key is known and cert is signed by CA")
}

func TestTLSSelfSignedCertTrustedPublicKey(t *testing.T) { // nolint[unused]
	cm := newTestClientManager(testLiveConfig)
	_, dialer := createAndSetPinningDialer(cm)
	dialer.pinChecker.trustedPins = append(dialer.pinChecker.trustedPins, `pin-sha256="9SLklscvzMYj8f+52lp5ze/hY0CFHyLSPQzSpYYIBm8="`)
	_, err := dialer.DialTLS("tcp", "self-signed.badssl.com:443")
	assert.NoError(t, err, "expected dial to succeed because public key is known despite cert being self-signed")
}
