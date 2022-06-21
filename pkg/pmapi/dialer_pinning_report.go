// Copyright (c) 2022 Proton AG
//
// This file is part of Proton Mail Bridge.Bridge.
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
	"bytes"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"strconv"
	"time"

	"github.com/sirupsen/logrus"
)

// tlsReport is inspired by https://tools.ietf.org/html/rfc7469#section-3.
// When a TLS key mismatch is detected, a tlsReport is posted to TLSReportURI.
type tlsReport struct {
	//  DateTime of observed pin validation in time.RFC3339 format.
	DateTime string `json:"date-time"`

	// Hostname to which the UA made original request that failed pin validation.
	Hostname string `json:"hostname"`

	// Port to which the UA made original request that failed pin validation.
	Port int `json:"port"`

	// EffectiveExpirationDate for noted pins in time.RFC3339 format.
	EffectiveExpirationDate string `json:"effective-expiration-date"`

	// IncludeSubdomains indicates whether or not the UA has noted the
	// includeSubDomains directive for the Known Pinned Host.
	IncludeSubdomains bool `json:"include-subdomains"`

	// NotedHostname indicates the hostname that the UA noted when it noted
	// the Known Pinned Host. This field allows operators to understand why
	// Pin Validation was performed for, e.g., foo.example.com when the
	// noted Known Pinned Host was example.com with includeSubDomains set.
	NotedHostname string `json:"noted-hostname"`

	// ServedCertificateChain is the certificate chain, as served by
	// the Known Pinned Host during TLS session setup.  It is provided as an
	// array of strings; each string pem1, ... pemN is the Privacy-Enhanced
	// Mail (PEM) representation of each X.509 certificate as described in
	// [RFC7468].
	ServedCertificateChain []string `json:"served-certificate-chain"`

	// ValidatedCertificateChain is the certificate chain, as
	// constructed by the UA during certificate chain verification.  (This
	// may differ from the served-certificate-chain.)  It is provided as an
	// array of strings; each string pem1, ... pemN is the PEM
	// representation of each X.509 certificate as described in [RFC7468].
	// UAs that build certificate chains in more than one way during the
	// validation process SHOULD send the last chain built.  In this way,
	// they can avoid keeping too much state during the validation process.
	ValidatedCertificateChain []string `json:"validated-certificate-chain"`

	// The known-pins are the Pins that the UA has noted for the Known
	// Pinned Host.  They are provided as an array of strings with the
	// syntax: known-pin = token "=" quoted-string
	// e.g.:
	// ```
	// "known-pins": [
	//   'pin-sha256="d6qzRu9zOECb90Uez27xWltNsj0e1Md7GkYYkVoZWmM="',
	//   "pin-sha256=\"E9CZ9INDbd+2eRQozYqqbQ2yXLVKB9+xcprMF+44U1g=\""
	// ]
	// ```
	KnownPins []string `json:"known-pins"`

	// AppVersion is used to set `x-pm-appversion` json format from datatheorem/TrustKit.
	AppVersion string `json:"app-version"`
}

// newTLSReport constructs a new tlsReport configured with the given app version and known pinned public keys.
// Temporal things (current date/time) are not set yet -- they are set when sendReport is called.
func newTLSReport(host, port, server string, certChain, knownPins []string, appVersion string) (report tlsReport) {
	// If we can't parse the port for whatever reason, it doesn't really matter; we should report anyway.
	intPort, _ := strconv.Atoi(port)

	report = tlsReport{
		Hostname:               host,
		Port:                   intPort,
		NotedHostname:          server,
		ServedCertificateChain: certChain,
		KnownPins:              knownPins,
		AppVersion:             appVersion,
	}

	return
}

// sendReport posts the given TLS report to the standard TLS Report URI.
func (r tlsReport) sendReport(cfg Config, uri string) {
	now := time.Now()
	r.DateTime = now.Format(time.RFC3339)
	r.EffectiveExpirationDate = now.Add(365 * 24 * 60 * 60 * time.Second).Format(time.RFC3339)

	b, err := json.Marshal(r)
	if err != nil {
		logrus.WithError(err).Error("Failed to marshal TLS report")
		return
	}

	req, err := http.NewRequest("POST", uri, bytes.NewReader(b))
	if err != nil {
		logrus.WithError(err).Error("Failed to create http request")
		return
	}

	req.Header.Add("Content-Type", "application/json")
	req.Header.Set("User-Agent", cfg.getUserAgent())
	req.Header.Set("x-pm-appversion", r.AppVersion)

	logrus.WithField("request", req).Warn("Reporting TLS mismatch")
	res, err := (&http.Client{Transport: CreateTransportWithDialer(NewBasicTLSDialer(cfg))}).Do(req)
	if err != nil {
		logrus.WithError(err).Error("Failed to report TLS mismatch")
		return
	}

	logrus.WithField("response", res).Error("Reported TLS mismatch")

	if res.StatusCode != http.StatusOK {
		logrus.WithField("status", http.StatusOK).Error("StatusCode was not OK")
	}

	_, _ = ioutil.ReadAll(res.Body)
	_ = res.Body.Close()
}
