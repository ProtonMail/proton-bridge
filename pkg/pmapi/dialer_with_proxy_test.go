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
)

const liveAPI = "https://api.protonmail.ch"

var testLiveConfig = &ClientConfig{
	AppVersion: "Bridge_1.2.4-test",
	ClientID:   "Bridge",
}

func newTestDialerWithPinning() (*int, *DialerWithPinning) {
	called := 0
	p := NewPMAPIPinning(testLiveConfig.AppVersion)
	p.ReportCertIssueLocal = func() { called++ }
	testLiveConfig.Transport = p.TransportWithPinning()
	return &called, p
}

func TestTLSPinValid(t *testing.T) {
	called, _ := newTestDialerWithPinning()

	RootURL = liveAPI
	client := NewClient(testLiveConfig, "pmapi"+t.Name())

	_, err := client.AuthInfo("this.address.is.disabled")
	Ok(t, err)

	Equals(t, 0, *called)
}

func TestTLSPinBackup(t *testing.T) {
	called, p := newTestDialerWithPinning()
	p.report.KnownPins[1] = p.report.KnownPins[0]
	p.report.KnownPins[0] = ""

	RootURL = liveAPI
	client := NewClient(testLiveConfig, "pmapi"+t.Name())

	_, err := client.AuthInfo("this.address.is.disabled")
	Ok(t, err)

	Equals(t, 0, *called)
}

func _TestTLSPinNoMatch(t *testing.T) { // nolint[unused]
	called, p := newTestDialerWithPinning()
	for i := 0; i < len(p.report.KnownPins); i++ {
		p.report.KnownPins[i] = "testing"
	}

	RootURL = liveAPI
	client := NewClient(testLiveConfig, "pmapi"+t.Name())

	_, err := client.AuthInfo("this.address.is.disabled")
	Ok(t, err)

	// check that it will be called only once per session
	client = NewClient(testLiveConfig, "pmapi"+t.Name())
	_, err = client.AuthInfo("this.address.is.disabled")
	Ok(t, err)

	Equals(t, 1, *called)
}

func _TestTLSPinInvalid(t *testing.T) { // nolint[unused]
	ts := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		writeJSONResponsefromFile(t, w, "/auth/info/post_response.json", 0)
	}))
	defer ts.Close()

	called, _ := newTestDialerWithPinning()

	client := NewClient(testLiveConfig, "pmapi"+t.Name())

	RootURL = liveAPI
	_, err := client.AuthInfo("this.address.is.disabled")
	Ok(t, err)

	RootURL = ts.URL
	_, err = client.AuthInfo("this.address.is.disabled")
	Assert(t, err != nil, "error is expected but have %v", err)

	Equals(t, 1, *called)
}

func _TestTLSSignedCertWrongPublicKey(t *testing.T) { // nolint[unused]
	_, dialer := newTestDialerWithPinning()
	_, err := dialer.dialAndCheckFingerprints("tcp", "rsa4096.badssl.com:443")
	Assert(t, err != nil, "expected dial to fail because of wrong public key: ", err.Error())
}

func _TestTLSSignedCertTrustedPublicKey(t *testing.T) { // nolint[unused]
	_, dialer := newTestDialerWithPinning()
	dialer.report.KnownPins = append(dialer.report.KnownPins, `pin-sha256="W8/42Z0ffufwnHIOSndT+eVzBJSC0E8uTIC8O6mEliQ="`)
	_, err := dialer.dialAndCheckFingerprints("tcp", "rsa4096.badssl.com:443")
	Assert(t, err == nil, "expected dial to succeed because public key is known and cert is signed by CA: ", err.Error())
}

func _TestTLSSelfSignedCertTrustedPublicKey(t *testing.T) { // nolint[unused]
	_, dialer := newTestDialerWithPinning()
	dialer.report.KnownPins = append(dialer.report.KnownPins, `pin-sha256="9SLklscvzMYj8f+52lp5ze/hY0CFHyLSPQzSpYYIBm8="`)
	_, err := dialer.dialAndCheckFingerprints("tcp", "self-signed.badssl.com:443")
	Assert(t, err == nil, "expected dial to succeed because public key is known despite cert being self-signed: ", err.Error())
}
